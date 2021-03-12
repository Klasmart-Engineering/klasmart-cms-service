package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
)

type IOrganizationRegion interface{
	GetOrganizationByHeadquarter(ctx context.Context, db *dbo.DBContext, headquarterID string) ([]string, error)
}

type OrganizationRegion struct {
}
func (o *OrganizationRegion) GetOrganizationByHeadquarter(ctx context.Context, db *dbo.DBContext, headquarterID string) ([]string, error){
	records, err := da.GetOrganizationRegionDA().GetOrganizationByHeadquarter(ctx, db, headquarterID)
	if err != nil{
		log.Error(ctx, "GetOrganizationByHeadquarter failed",
			log.Err(err),
			log.Any("headquarterID", headquarterID))
		return nil, err
	}
	orgIDs := make([]string ,len(records))
	for i := range records {
		orgIDs[i] = records[i].OrganizationID
	}
	return orgIDs, nil
}

