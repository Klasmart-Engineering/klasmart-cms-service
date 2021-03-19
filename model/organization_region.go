package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"sync"
)

type IOrganizationRegion interface{
	GetOrganizationByHeadquarter(ctx context.Context, db *dbo.DBContext, headquarterID string) ([]string, error)
	GetOrganizationByHeadquarterForDetails(ctx context.Context, db *dbo.DBContext, operator *entity.Operator) ([]*entity.RegionOrganizationInfo, error)
}

type OrganizationRegion struct {
}
func (o *OrganizationRegion) GetOrganizationByHeadquarter(ctx context.Context, db *dbo.DBContext, headquarterID string) ([]string, error){
	records, err := da.GetOrganizationRegionDA().GetOrganizationByHeadquarter(ctx, db, headquarterID)
	if err != nil{
		log.Error(ctx, "GetOrganizationByHeadquarter failed",
			log.Err(err),
			log.String("headquarterID", headquarterID))
		return nil, err
	}
	orgIDs := make([]string ,len(records))
	for i := range records {
		orgIDs[i] = records[i].OrganizationID
	}
	return orgIDs, nil
}
func (o *OrganizationRegion) GetOrganizationByHeadquarterForDetails(ctx context.Context, db *dbo.DBContext, operator *entity.Operator) ([]*entity.RegionOrganizationInfo, error) {
	oids, err := o.GetOrganizationByHeadquarter(ctx, db, operator.OrgID)
	if err != nil{
		return nil, err
	}
	orgs, err := external.GetOrganizationServiceProvider().BatchGet(ctx, operator, oids)
	if err != nil {
		log.Error(ctx, "BatchGet org info failed",
			log.Err(err),
			log.Strings("oids", oids))
		return nil, err
	}
	list := make([]*entity.RegionOrganizationInfo, len(orgs))
	for i := range orgs {
		list[i] = &entity.RegionOrganizationInfo{
			ID:   orgs[i].ID,
			Name: orgs[i].Name,
		}
	}
	return list, nil
}

var (
	_organizationRegionOnce  sync.Once
	_organizationRegionModel IOrganizationRegion
)

func GetOrganizationRegionModel() IOrganizationRegion {
	_organizationRegionOnce.Do(func() {
		_organizationRegionModel = &OrganizationRegion{}
	})
	return _organizationRegionModel
}