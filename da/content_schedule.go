package da

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cd *DBContentDA) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator) (lps []*entity.LessonPlanForSchedule, err error) {

	paths, err := GetFolderDA().GetSharedContentParentPath(ctx, dbo.MustGetDB(ctx), []string{op.OrgID, constant.ShareToAll})
	if err != nil {
		return
	}

	var condition string
	if len(paths) <= 0 {
		condition = "1=0"
	} else {
		conds := make([]string, len(paths))
		for i, v := range paths {
			conds[i] = "dir_path like " + "'" + v + "%" + "'"
		}
		condition = fmt.Sprintf("(%s)", strings.Join(conds, " or "))
	}

	lps = []*entity.LessonPlanForSchedule{}
	sql := fmt.Sprintf(`
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
	and %s 
order by cc.create_at 
`, condition)
	args := []interface{}{
		entity.ContentTypePlan,
		entity.ContentStatusPublished,
		entity.ContentPropertyTypeProgram,
	}
	err = cd.s.QueryRawSQL(ctx, &lps, sql, args...)
	if err != nil {
		return
	}
	return
}
