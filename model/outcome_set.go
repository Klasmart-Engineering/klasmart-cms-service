package model

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IOutcomeSetModel interface {
	CreateOutcomeSet(ctx context.Context, op *entity.Operator, name string) (string, error)
	PullOutcomeSet(ctx context.Context, op *entity.Operator, name string) ([]*entity.Set, error)
	BulkBindOutcomeSet(ctx context.Context, op *entity.Operator, outcomeIDs []string, setIDs []string) error
	BindByOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error
}

type OutcomeSetModel struct{}

func (OutcomeSetModel) CreateOutcomeSet(ctx context.Context, op *entity.Operator, name string) (string, error) {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeSetLock, op.OrgID, name)
	if err != nil {
		log.Error(ctx, "CreateOutcomeSet: NewLock failed",
			log.Err(err),
			log.Any("op", op),
			log.String("set", name))
		return "", err
	}
	locker.Lock()
	defer locker.Unlock()

	set := entity.Set{
		ID:             utils.NewID(),
		Name:           name,
		OrganizationID: op.OrgID,
	}
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		exist, err := da.GetOutcomeSetDA().IsSetExist(ctx, op, tx, set.Name)
		if err != nil {
			log.Error(ctx, "CreateSet: IsSetExist failed",
				log.Err(err),
				log.String("name", name))
			return err
		}
		if exist {
			return constant.ErrDuplicateRecord
		}
		err = da.GetOutcomeSetDA().CreateSet(ctx, tx, &set)
		if err != nil {
			log.Error(ctx, "CreateSet: CreateSet failed",
				log.Err(err),
				log.String("name", name))
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return set.ID, nil
}
func (OutcomeSetModel) PullOutcomeSet(ctx context.Context, op *entity.Operator, name string) ([]*entity.Set, error) {
	_, sets, err := da.GetOutcomeSetDA().SearchSet(ctx, dbo.MustGetDB(ctx), &da.SetCondition{
		Name:           sql.NullString{String: name, Valid: true},
		OrganizationID: sql.NullString{String: op.OrgID, Valid: true},
		Pager:          dbo.NoPager,
	})
	if err != nil {
		log.Error(ctx, "PullOutcomeSet: SearchSet failed",
			log.Err(err),
			log.String("name", name),
			log.Any("op", op))
		return nil, err
	}
	return sets, nil
}

func (OutcomeSetModel) BulkBindOutcomeSet(ctx context.Context, op *entity.Operator, outcomeIDs []string, setIDs []string) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixOutcomeSetLock, op.OrgID)
	if err != nil {
		log.Error(ctx, "BulkBindOutcomeSet: NewLock failed",
			log.Err(err),
			log.Any("op", op),
			log.Strings("outcome", outcomeIDs),
			log.Strings("set", setIDs))
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		hasLocked, err := GetOutcomeModel().HasLockedOutcome(ctx, op, tx, outcomeIDs)
		if err != nil {
			log.Error(ctx, "BulkBindOutcomeSet: HasLockedOutcome failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("outcomes", outcomeIDs),
				log.Strings("sets", setIDs))
			return err
		}
		if hasLocked {
			return constant.ErrHasLocked
		}
		_, outcomeTags, err := da.GetOutcomeSetDA().SearchOutcomeSet(ctx, tx, &da.OutcomeSetCondition{
			OutcomeIDs: dbo.NullStrings{Strings: outcomeIDs, Valid: true},
			SetIDs:     dbo.NullStrings{Strings: setIDs, Valid: true},
		})

		duplicate := make(map[string]struct{})
		for i := range outcomeTags {
			key := fmt.Sprintf("%s:%s", outcomeTags[i].OutcomeID, outcomeTags[i].SetID)
			duplicate[key] = struct{}{}
		}

		var outcomeSets []*entity.OutcomeSet
		for i := range outcomeIDs {
			for j := range setIDs {
				key := fmt.Sprintf("%s:%s", outcomeIDs[i], setIDs[j])
				if _, ok := duplicate[key]; !ok {
					outcomeSets = append(outcomeSets, &entity.OutcomeSet{
						OutcomeID: outcomeIDs[i],
						SetID:     setIDs[j],
					})
				}
			}
		}
		if len(outcomeSets) == 0 {
			log.Info(ctx, "BulkBindOutcomeSet: all already bind",
				log.Strings("outcome", outcomeIDs),
				log.Strings("set", setIDs))
			return nil
		}
		err = da.GetOutcomeSetDA().BulkBindOutcomeSet(ctx, op, dbo.MustGetDB(ctx), outcomeSets)
		if err != nil {
			log.Error(ctx, "BulkBindOutcomeSet: BulkBindOutcomeSet failed",
				log.Err(err),
				log.Any("op", op),
				log.Strings("outcomes", outcomeIDs),
				log.Strings("sets", setIDs))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	da.GetOutcomeRedis().CleanOutcomeConditionCache(ctx, op, nil)
	return nil
}

func (OutcomeSetModel) BindByOutcome(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, outcome *entity.Outcome) error {
	if len(outcome.Sets) == 0 {
		return nil
	}
	outcomeSets := make([]*entity.OutcomeSet, len(outcome.Sets))
	for i := range outcome.Sets {
		outcomeSet := entity.OutcomeSet{
			OutcomeID: outcome.ID,
			SetID:     outcome.Sets[i].ID,
		}
		outcomeSets[i] = &outcomeSet
	}
	err := da.GetOutcomeSetDA().BindOutcomeSet(ctx, op, tx, outcomeSets)
	if err != nil {
		log.Error(ctx, "BindByOutcome: BindOutcomeSet failed",
			log.String("op", op.UserID),
			log.Any("outcome", outcome))
		return err
	}
	return nil
}

var (
	_outcomeSetModel     IOutcomeSetModel
	_outcomeSetModelOnce sync.Once
)

func GetOutcomeSetModel() IOutcomeSetModel {
	_outcomeSetModelOnce.Do(func() {
		_outcomeSetModel = new(OutcomeSetModel)
	})
	return _outcomeSetModel
}
