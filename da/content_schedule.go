package da

import (
	"context"
	"fmt"
	"strings"

	"github.com/KL-Engineering/common-log/log"

	"github.com/KL-Engineering/kidsloop-cms-service/utils"

	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type contentCondition struct {
	Typ entity.ContentPropertyType
	IDs []string
}

func (cd *ContentMySQLDA) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest, condOrgContent dbo.Conditions, programGroups []*entity.ProgramGroup) (total int, lps []*entity.LessonPlanForSchedule, err error) {
	lps = []*entity.LessonPlanForSchedule{}
	if len(cond.ProgramIDs) == 0 {
		return
	}

	sqlContents := `select * from cms_contents cc `
	var sqlContentsWheres []string
	var argContents []interface{}
	var whereIDSql string
	var whereIDWhere string
	var whereIDArgs []interface{}
	AddContentWhereCond := func(typ entity.ContentPropertyType, IDs []string) {
		if len(IDs) == 0 {
			return
		}
		if whereIDSql == "" {
			whereIDSql = `select ccp.content_id from cms_content_properties ccp `
			whereIDWhere = `
where ccp.property_type=? and ccp.property_id in (?)`
			whereIDArgs = append(whereIDArgs, typ, IDs)
		} else {
			whereIDSql += fmt.Sprintf(`
inner join cms_content_properties ccp_%v on ccp_%v.content_id=ccp.content_id and ccp_%v.property_type=? and ccp_%v.property_id in (?)`, typ, typ, typ, typ)
			argContents = append(argContents, typ, IDs)
		}
	}
	AddContentWhereCond(entity.ContentPropertyTypeProgram, cond.ProgramIDs)
	AddContentWhereCond(entity.ContentPropertyTypeSubject, cond.SubjectIDs)
	AddContentWhereCond(entity.ContentPropertyTypeCategory, cond.CategoryIDs)
	AddContentWhereCond(entity.ContentPropertyTypeSubCategory, cond.SubCategoryIDs)
	AddContentWhereCond(entity.ContentPropertyTypeAge, cond.AgeIDs)
	AddContentWhereCond(entity.ContentPropertyTypeGrade, cond.GradeIDs)
	if whereIDSql != "" {
		whereIDSql += whereIDWhere
		argContents = append(argContents, whereIDArgs...)
		sqlContentsWheres = append(sqlContentsWheres, fmt.Sprintf(`EXISTS (
	%s 
	and ccp.content_id = cc.id
)`, whereIDSql))
	}

	if cond.LessonPlanName != "" {
		sqlContentsWheres = append(sqlContentsWheres, "cc.content_name like ?")
		argContents = append(argContents, "%"+cond.LessonPlanName+"%")
	}
	if len(sqlContentsWheres) > 0 {
		sqlContents += " where " + strings.Join(sqlContentsWheres, " and ")
	}
	sbContents := NewSqlBuilder(ctx, sqlContents, argContents...)

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

func (cd *ContentMySQLDA) PageRawSQL(ctx context.Context, values interface{}, orderBy, sql string, pager dbo.Pager, args ...interface{}) (count int, err error) {
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
