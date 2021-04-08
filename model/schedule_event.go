package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type IScheduleEventModel interface {
	AddUserEvent(ctx context.Context, op *entity.Operator, event *entity.ScheduleClassEvent) error
	DeleteUserEvent(ctx context.Context, op *entity.Operator, event *entity.ScheduleClassEvent) error
}

type scheduleEventModel struct {
}

func (s *scheduleEventModel) AddUserEvent(ctx context.Context, op *entity.Operator, event *entity.ScheduleClassEvent) error {
	if event.ClassID == "" || len(event.Users) <= 0 {
		log.Info(ctx, "event invalid args", log.Any("event", event))
		return constant.ErrInvalidArgs
	}
	scheduleIDs, err := GetScheduleModel().GetRosterClassNotStartScheduleIDs(ctx, op, event.ClassID)
	if err != nil {
		log.Error(ctx, "class has schedule error", log.Err(err), log.Any("event", event))
		return err
	}
	if len(scheduleIDs) <= 0 {
		log.Debug(ctx, "class are not scheduled", log.Any("event", event))
		return nil
	}
	userDistinct := make(map[string]bool)
	relations := make([]*entity.ScheduleRelation, 0, len(scheduleIDs)*len(event.Users))
	for _, scheduleID := range scheduleIDs {
		for _, user := range event.Users {
			if _, ok := userDistinct[user.ID]; ok {
				continue
			}
			var roleType entity.ScheduleRelationType
			switch user.RoleType {
			case entity.ClassUserRoleTypeEventTeacher:
				roleType = entity.ScheduleRelationTypeClassRosterTeacher
			case entity.ClassUserRoleTypeEventStudent:
				roleType = entity.ScheduleRelationTypeClassRosterStudent
			default:
				log.Info(ctx, "user role type is invalid", log.Any("event", event))
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
			userDistinct[user.ID] = true
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

	return nil
}

func (s *scheduleEventModel) DeleteUserEvent(ctx context.Context, op *entity.Operator, event *entity.ScheduleClassEvent) error {
	if event.ClassID == "" || len(event.Users) <= 0 {
		log.Info(ctx, "event invalid args", log.Any("event", event))
		return constant.ErrInvalidArgs
	}
	scheduleIDs, err := GetScheduleModel().GetRosterClassNotStartScheduleIDs(ctx, op, event.ClassID)
	if err != nil {
		log.Error(ctx, "class has schedule error", log.Err(err), log.Any("event", event))
		return err
	}
	if len(scheduleIDs) <= 0 {
		log.Debug(ctx, "class are not scheduled", log.Any("event", event))
		return nil
	}
	userIDs := make([]string, len(event.Users))
	for i, item := range event.Users {
		userIDs[i] = item.ID
	}
	err = da.GetScheduleRelationDA().DeleteByRelationIDs(ctx, dbo.MustGetDB(ctx), scheduleIDs, userIDs)
	if err != nil {
		log.Error(ctx, "delete error", log.Err(err), log.Any("event", event), log.Strings("scheduleIDs", scheduleIDs))
		return err
	}
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
