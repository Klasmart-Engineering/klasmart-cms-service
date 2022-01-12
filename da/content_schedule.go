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
	var sqlArr []string
	var args []interface{}
	if utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameOrganizationContent.String()) {
		wheres, args1 := condOrgContent.GetConditions()
		sql := fmt.Sprintf(`select 
id, 
content_name as name,
'%s' as group_name,
create_at
from cms_contents`, entity.LessonPlanGroupNameOrganizationContent)
		if len(wheres) > 0 {
			sql = fmt.Sprintf(`
%s where %s
`, sql, strings.Join(wheres, " and "))
		}
		sqlArr = append(sqlArr, sql)
		args = append(args, args1...)
	}

	needBadaContent := utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameBadanamuContent.String())
	needMoreFeatured := utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameMoreFeaturedContent.String())
	if needBadaContent || needMoreFeatured {
		paths, err1 := GetFolderDA().GetSharedContentParentPath(ctx, dbo.MustGetDB(ctx), []string{op.OrgID, constant.ShareToAll})
		if err1 != nil {
			err = err1
			return
		}

		var condition string
		if len(paths) <= 0 {
			condition = "1=0"
		} else {
			condArr := make([]string, len(paths))
			for i, v := range paths {
				condArr[i] = "cc.dir_path like " + "'" + v + "%" + "'"
			}
			condition = fmt.Sprintf("(%s)", strings.Join(condArr, " or "))
		}

		var whereGroupName string
		if needBadaContent && !needMoreFeatured {
			whereGroupName = "and ccp.property_id  in (?)"
		}
		if !needBadaContent && needMoreFeatured {
			whereGroupName = "and ccp.property_id not in (?)"
		}

		sql := fmt.Sprintf(`
select 
	cc.id,
	cc.content_name as name,
	if(ccp.property_id in (?),?,?) as group_name,
	cc.create_at
from cms_contents cc 
left join cms_content_properties ccp on ccp.content_id =cc.id
where 
	cc.content_type=? 
	and cc.publish_status in (?)
	and ccp.property_type =? 
	and %s 
	%s
order by cc.create_at 
`, condition, whereGroupName)
		var pgIDs []string
		for _, pg := range programGroups {
			pgIDs = append(pgIDs, pg.ProgramID)
		}
		sqlArr = append(sqlArr, sql)
		args = append(args, entity.ContentTypePlan,
			entity.ContentStatusPublished,
			entity.ContentPropertyTypeProgram,
			pgIDs,
			entity.LessonPlanGroupNameBadanamuContent,
			entity.LessonPlanGroupNameMoreFeaturedContent)
	}

	subSql := strings.Join(sqlArr, `

union all

`)
	sql := fmt.Sprintf(`select 
	t.id,
	t.name,
	t.group_name
from (
	%s
)t
`, subSql)
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
	db := dbo.MustGetDB(ctx).Raw(sqlCount, args...)
	if db.Error != nil {
		err = db.Error
		log.Error(ctx, "PageRawSQL:QueryCount failed",
			log.Any("sqlCount", sqlCount),
			log.Any("args", args),
			log.Err(err),
		)
		return
	}
	count = int(db.RowsAffected)

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
