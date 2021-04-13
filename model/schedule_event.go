package model

import (
	"context"
	"database/sql"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IScheduleEventModel interface {
	AddMembersEvent(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error
	DeleteMembersEvent(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error
}

type scheduleEventModel struct {
}

func (s *scheduleEventModel) AddMembersEvent(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error {
	if event.ClassID == "" || len(event.Members) <= 0 {
		log.Info(ctx, "event invalid args", log.Any("event", event))
		return constant.ErrInvalidArgs
	}
	scheduleIDs, err := GetScheduleModel().GetRosterClassNotStartScheduleIDs(ctx, event.ClassID, nil)
	if err != nil {
		log.Error(ctx, "class has schedule error", log.Err(err), log.Any("event", event))
		return err
	}
	if len(scheduleIDs) <= 0 {
		log.Debug(ctx, "class are not scheduled", log.Any("event", event))
		return nil
	}

	log.Debug(ctx, "about scheduleIDs", log.Any("scheduleIDs", scheduleIDs), log.Int("schedule count", len(scheduleIDs)), log.Any("event", event))

	userDistinct := make(map[string]bool)
	relations := make([]*entity.ScheduleRelation, 0, len(event.Members)*len(scheduleIDs))

	for _, user := range event.Members {
		var roleType = user.RoleType.ToScheduleRelationType()
		if roleType == entity.ScheduleRelationTypeInvalid {
			log.Info(ctx, "user role type invalid", log.Any("event", event))
			continue
		}
		for _, scheduleID := range scheduleIDs {
			if _, ok := userDistinct[scheduleID+user.ID]; ok {
				continue
			}

			countCondition := &da.ScheduleRelationCondition{
				ScheduleID: sql.NullString{
					String: scheduleID,
					Valid:  true,
				},
				RelationID: sql.NullString{
					String: user.ID,
					Valid:  true,
				},
				RelationType: sql.NullString{
					String: roleType.String(),
					Valid:  true,
				},
			}
			count, err := GetScheduleRelationModel().Count(ctx, op, countCondition)
			if err != nil {
				log.Error(ctx, "count by schedule condition error", log.Err(err), log.Any("countCondition", countCondition))
				return err
			}
			if count > 0 {
				log.Info(ctx, "User has been scheduled", log.Any("countCondition", countCondition))
				continue
			}
			relations = append(relations, &entity.ScheduleRelation{
				ID:           utils.NewID(),
				ScheduleID:   scheduleID,
				RelationID:   user.ID,
				RelationType: roleType,
			})
			userDistinct[scheduleID+user.ID] = true
		}
	}

	if len(relations) <= 0 {
		log.Info(ctx, "relation data not found", log.Any("event", event))
		return nil
	}

	_, err = da.GetScheduleRelationDA().BatchInsert(ctx, dbo.MustGetDB(ctx), relations)
	if err != nil {
		log.Error(ctx, "count by schedule condition error", log.Err(err), log.Any("event", event), log.Any("relations", relations))
		return err
	}

	log.Debug(ctx, "class add user event end", log.Any("relations", relations), log.Any("event", event))

	return nil
}

func (s *scheduleEventModel) DeleteMembersEvent(ctx context.Context, op *entity.Operator, event *entity.ClassUpdateMembersEvent) error {
	if event.ClassID == "" || len(event.Members) <= 0 {
		log.Info(ctx, "event invalid args", log.Any("event", event))
		return constant.ErrInvalidArgs
	}
	userIDs := make([]string, len(event.Members))
	for i, item := range event.Members {
		userIDs[i] = item.ID
	}
	condition := da.GetNotStartCondition(event.ClassID, userIDs)

	ids, err := GetScheduleRelationModel().GetIDs(ctx, op, condition)
	if err != nil {
		log.Error(ctx, "get ids by condition error", log.Err(err), log.Any("event", event), log.Any("condition", condition))
		return err
	}
	err = da.GetScheduleRelationDA().DeleteByIDs(ctx, dbo.MustGetDB(ctx), ids)
	if err != nil {
		log.Error(ctx, "delete by ids error", log.Err(err), log.Any("event", event), log.Any("condition", condition), log.Strings("ids", ids))
		return err
	}
	log.Debug(ctx, "class delete user event end", log.Any("event", event), log.Any("condition", condition))
	return err
}

var (
	_scheduleEventOnce  sync.Once
	_scheduleEventModel IScheduleEventModel
)

func GetScheduleEventModel() IScheduleEventModel {
	_scheduleEventOnce.Do(func() {
		_scheduleEventModel = &scheduleEventModel{}
	})
	return _scheduleEventModel
}
