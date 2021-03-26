package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IOrganizationRegionDA interface{
	GetOrganizationByHeadquarter(ctx context.Context, db *dbo.DBContext, headquarterID string) ([]*entity.OrganizationRegion, error)
}

type OrganizationRegionDA struct {
	s dbo.BaseDA
}

func (cd *OrganizationRegionDA) GetOrganizationByHeadquarter(ctx context.Context, db *dbo.DBContext, headquarterID string) ([]*entity.OrganizationRegion, error) {
	objs := make([]*entity.OrganizationRegion, 0)
	condition := &OrganizationRegionCondition{
		HeadquarterIDs: []string{headquarterID},
	}
	err := cd.s.QueryTx(ctx, db, condition, &objs)
	if err != nil {
		return nil, err
	}
	return objs, nil
}

type OrganizationRegionCondition struct {
	IDS           		[]string `json:"ids"`
	HeadquarterIDs   	[]string `json:"headquarter_ids"`

	Pager       		utils.Pager
}

func (s *OrganizationRegionCondition) GetConditions() ([]string, []interface{}) {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if len(s.IDS) > 0 {
		condition := " id in (?) "
		conditions = append(conditions, condition)
		params = append(params, s.IDS)
	}

	if len(s.HeadquarterIDs) > 0 {
		conditions = append(conditions, "headquarter IN (?)")
		params = append(params, s.HeadquarterIDs)
	}

	conditions = append(conditions, " delete_at = 0")
	return conditions, params
}
func (s *OrganizationRegionCondition) GetPager() *dbo.Pager {
	return &dbo.Pager{
		Page:     int(s.Pager.PageIndex),
		PageSize: int(s.Pager.PageSize),
	}
}
func (s *OrganizationRegionCondition) GetOrderBy() string {
	return "create_at desc"
}


var (
	orgRegionDA    IOrganizationRegionDA
	_orgRegionOnce sync.Once
)

func GetOrganizationRegionDA() IOrganizationRegionDA {
	_orgRegionOnce.Do(func() {
		orgRegionDA = new(OrganizationRegionDA)
	})

	return orgRegionDA
}