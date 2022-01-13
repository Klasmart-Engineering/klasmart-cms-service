package da

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cd *DBContentDA) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest, condOrgContent dbo.Conditions, programGroups []*entity.ProgramGroup) (total int, lps []*entity.LessonPlanForSchedule, err error) {
	sqlContents := strings.Builder{}
	sqlContents.WriteString(`select
	distinct cc.*
from cms_contents cc
`)
	var argContents []interface{}
	innerJoinCPP := func(typ entity.ContentPropertyType, IDs []string) {
		if len(IDs) == 0 {
			return
		}
		sqlContents.WriteString(fmt.Sprintf("inner join cms_content_properties ccp_%v on ccp_%v.content_id =cc.id  and  ccp_%v.property_type =? and ccp_%v.property_id in (?) ", typ, typ, typ, typ))
		argContents = append(argContents, IDs)
	}
	innerJoinCPP(entity.ContentPropertyTypeProgram, cond.ProgramIDs)
	innerJoinCPP(entity.ContentPropertyTypeSubject, cond.SubjectIDs)
	innerJoinCPP(entity.ContentPropertyTypeCategory, cond.CategoryIDs)
	innerJoinCPP(entity.ContentPropertyTypeSubCategory, cond.SubCategoryIDs)
	innerJoinCPP(entity.ContentPropertyTypeAge, cond.AgeIDs)
	innerJoinCPP(entity.ContentPropertyTypeGrade, cond.GradeIDs)

	if cond.LessonPlanName != "" {
		sqlContents.WriteString("where cc.content_name like ?")
		argContents = append(argContents, "%"+cond.LessonPlanName+"%")
	}
	sbContents := NewSqlBuilder(ctx, sqlContents.String(), argContents...)

	var sqlArr []string
	var sbOrgContent *sqlBuilder
	if utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameOrganizationContent.String()) {
		sql := strings.Builder{}
		sql.WriteString(fmt.Sprintf(`select 
cc1.id, 
cc1.content_name as name,
'%s' as group_name,
cc1.create_at
from ({{.sbContents}}) cc1
`, entity.LessonPlanGroupNameOrganizationContent))

		sqlArr = append(sqlArr, "{{.sbOrgContent}}")
		wheres, args1 := condOrgContent.GetConditions()
		if len(wheres) == 0 {
			sbOrgContent = NewSqlBuilder(ctx, sql.String()).Replace(ctx, "sbContents", sbContents)
		} else {
			sql.WriteString("{{.sbOrgContentWhere}}")
			sbOrgContent = NewSqlBuilder(ctx, sql.String()).Replace(ctx, "sbContents", sbContents)
			sbOrgContentWhere := NewSqlBuilder(ctx, "where "+strings.Join(wheres, " and "), args1...)
			sbOrgContent = sbOrgContent.Replace(ctx, "sbOrgContentWhere", sbOrgContentWhere)
		}
	}

	var sbBadaContent *sqlBuilder
	needBadaContent := utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameBadanamuContent.String())
	needMoreFeatured := utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameMoreFeaturedContent.String())
	if needBadaContent || needMoreFeatured {
		sqlArr = append(sqlArr, "{{.sbBadaContent}}")
		paths, err1 := GetFolderDA().GetSharedContentParentPath(ctx, dbo.MustGetDB(ctx), []string{op.OrgID, constant.ShareToAll})
		if err1 != nil {
			err = err1
			return
		}
		sbBadaContent = NewSqlBuilder(ctx, `
{{.sbBadaContentSelect}}
from  ({{.sbContents}})  cc2 
left join cms_content_properties ccp on ccp.content_id =cc2.id
{{.sbBadaContentWhere}}
`).Replace(ctx, "sbContents", sbContents)
		var programIDs []string
		for _, pg := range programGroups {
			programIDs = append(programIDs, pg.ProgramID)
		}
		sbBadaContentSelect := NewSqlBuilder(ctx, `
select 
	cc2.id,
	cc2.content_name as name,
	if(ccp.property_id in (?),?,?) as group_name,
	cc2.create_at
`, programIDs, entity.LessonPlanGroupNameBadanamuContent, entity.LessonPlanGroupNameMoreFeaturedContent)
		sbBadaContent = sbBadaContent.Replace(ctx, "sbBadaContentSelect", sbBadaContentSelect)

		var condition string
		if len(paths) <= 0 {
			condition = "1=0"
		} else {
			condArr := make([]string, len(paths))
			for i, v := range paths {
				condArr[i] = "cc2.dir_path like " + "'" + v + "%" + "'"
			}
			condition = fmt.Sprintf("(%s)", strings.Join(condArr, " or "))
		}
		var args []interface{}
		args = append(args, entity.ContentTypePlan)
		args = append(args, entity.ContentStatusPublished)
		args = append(args, entity.ContentPropertyTypeProgram)
		var whereGroupName string
		if needBadaContent && !needMoreFeatured {
			whereGroupName = "and ccp.property_id  in (?)"
			args = append(args, programIDs)
		}
		if !needBadaContent && needMoreFeatured {
			whereGroupName = "and ccp.property_id not in (?)"
			args = append(args, programIDs)
		}
		sbBadaContentWhere := NewSqlBuilder(ctx, fmt.Sprintf(`
where 
	cc2.content_type=? 
	and cc2.publish_status in (?)
	and ccp.property_type =? 
	and %s 
	%s
`, condition, whereGroupName), args...)
		sbBadaContent = sbBadaContent.Replace(ctx, "sbBadaContentWhere", sbBadaContentWhere)
	}

	subSql := strings.Join(sqlArr, `
union all
`)
	sql := fmt.Sprintf(`select 
	t.id,
	t.name,
	t.group_name,
	t.create_at
from (
	%s
)t
`, subSql)
	sb := NewSqlBuilder(ctx, sql).
		Replace(ctx, "sbContents", sbContents).
		Replace(ctx, "sbOrgContent", sbOrgContent).
		Replace(ctx, "sbBadaContent", sbBadaContent)
	sql, args, err := sb.Build(ctx)
	if err != nil {
		return
	}
	lps = []*entity.LessonPlanForSchedule{}
	total, err = cd.PageRawSQL(ctx, &lps, cond.OrderBy, sql, dbo.Pager{
		Page:     int(cond.Pager.PageIndex),
		PageSize: int(cond.Pager.PageSize),
	}, args...)
	if err != nil {
		return
	}

	return
}

func (cd *DBContentDA) PageRawSQL(ctx context.Context, values interface{}, orderBy, sql string, pager dbo.Pager, args ...interface{}) (count int, err error) {
	log.Info(ctx, "start PageRawSQL",
		log.Any("sql", sql),
		log.Any("orderBy", orderBy),
		log.Any("pager", pager),
		log.Any("args", args),
	)
	sqlCount := fmt.Sprintf(`select count(*) as count from (%s)t`, sql)
	countRes := &struct {
		Count int `json:"count" gorm:"column:count" `
	}{}
	err = cd.s.QueryRawSQL(ctx, countRes, sqlCount, args...)
	if err != nil {
		return
	}
	count = countRes.Count

	offset, limit := pager.Offset()
	sqlQuery := fmt.Sprintf(`
select * from (%s)t 
order by t.%s limit %v offset %v 
`, sql, orderBy, limit, offset)
	err = cd.s.QueryRawSQL(ctx, values, sqlQuery, args...)
	if err != nil {
		return
	}
	return
}
