package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IOrganizationPropertyModel interface {
	MustGet(ctx context.Context, id string) (*entity.OrganizationProperty, error)
	MustGetTx(ctx context.Context, tx *dbo.DBContext, id string) (*entity.OrganizationProperty, error)
	GetOrDefault(ctx context.Context, id string) (*entity.OrganizationProperty, error)
	GetTxOrDefault(ctx context.Context, tx *dbo.DBContext, id string) (*entity.OrganizationProperty, error)
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

func (s organizationPropertyModel) MustGet(ctx context.Context, id string) (*entity.OrganizationProperty, error) {
	organizationProperty := new(entity.OrganizationProperty)
	err := da.GetOrganizationPropertyDA().Get(ctx, id, organizationProperty)
	if err != nil {
		return nil, err
	}

	return organizationProperty, nil
}

func (s organizationPropertyModel) MustGetTx(ctx context.Context, tx *dbo.DBContext, id string) (*entity.OrganizationProperty, error) {
	organizationProperty := new(entity.OrganizationProperty)
	err := da.GetOrganizationPropertyDA().GetTx(ctx, tx, id, organizationProperty)
	if err != nil {
		return nil, err
	}

	return organizationProperty, nil
}

func (s organizationPropertyModel) GetOrDefault(ctx context.Context, id string) (*entity.OrganizationProperty, error) {
	organizationProperty, err := s.MustGet(ctx, id)
	if err == nil {
		return organizationProperty, nil
	}

	if err == dbo.ErrRecordNotFound {
		return s.createDefault(id), nil
	}

	return nil, err
}

func (s organizationPropertyModel) GetTxOrDefault(ctx context.Context, tx *dbo.DBContext, id string) (*entity.OrganizationProperty, error) {
	organizationProperty, err := s.MustGetTx(ctx, tx, id)
	if err == nil {
		return organizationProperty, nil
	}

	if err == dbo.ErrRecordNotFound {
		return s.createDefault(id), nil
	}

	return nil, err
}

func (s organizationPropertyModel) createDefault(id string) *entity.OrganizationProperty {
	return &entity.OrganizationProperty{
		ID:   id,
		Type: entity.OrganizationTypeNormal,
	}
}
