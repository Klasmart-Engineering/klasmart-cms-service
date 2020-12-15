package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IOrganizationPropertyModel interface {
	Get(ctx context.Context, id string) (*entity.OrganizationProperty, error)
	GetTx(ctx context.Context, tx *dbo.DBContext, id string) (*entity.OrganizationProperty, error)
}

var (
	_organizationPropertyOnce  sync.Once
	_organizationPropertyModel IOrganizationPropertyModel
)

func GetOrganizationPropertyModel() IOrganizationPropertyModel {
	_organizationPropertyOnce.Do(func() {
		_organizationPropertyModel = &organizationPropertyModel{}
	})
	return _organizationPropertyModel
}

type organizationPropertyModel struct{}

func (s organizationPropertyModel) Get(ctx context.Context, id string) (*entity.OrganizationProperty, error) {
	organizationProperty := new(entity.OrganizationProperty)
	err := da.GetOrganizationPropertyDA().Get(ctx, id, organizationProperty)
	if err != nil {
		return nil, err
	}

	return organizationProperty, nil
}

func (s organizationPropertyModel) GetTx(ctx context.Context, tx *dbo.DBContext, id string) (*entity.OrganizationProperty, error) {
	organizationProperty := new(entity.OrganizationProperty)
	err := da.GetOrganizationPropertyDA().GetTx(ctx, tx, id, organizationProperty)
	if err != nil {
		return nil, err
	}

	return organizationProperty, nil
}
