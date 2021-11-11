package da

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cd *DBContentDA) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator) (lps []*entity.LessonPlanForSchedule, err error) {
	lps = []*entity.LessonPlanForSchedule{}
	sql := `
select 
	cc.id,
	cc.content_name as name, 
	ccp.property_id  as program_id
from cms_contents cc 
left join cms_content_properties ccp on ccp.content_id =cc.id
where 
	cc.content_type=? 
	and cc.publish_status in (?)
	and ccp.property_type =? 
	and cc.id in(
		select  
			content_id 
		from cms_authed_contents 
		where	org_id in (?, ?)			 
		and delete_at = 0
	)
order by cc.create_at 
`
	args := []interface{}{
		entity.ContentTypePlan,
		entity.ContentStatusPublished,
		entity.ContentPropertyTypeProgram,
		op.OrgID,
		constant.ShareToAll,
	}
	err = cd.s.QueryRawSQL(ctx, &lps, sql, args...)
	if err != nil {
		return
	}
	return
}
