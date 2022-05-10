package da

import (
	"context"
	"testing"

	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

func TestAdd(t *testing.T) {
	ctx := context.Background()

	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		tx.DB = tx.Debug()

		err := GetScheduleDA().DeleteWithFollowing(ctx, tx, "5fac9892232323230981aafeecf6b48", 0)
		if err != nil {
			t.Log(err)
			return err
		}

		_, err = GetScheduleDA().Insert(ctx, &[]*entity.Schedule{
			&entity.Schedule{
				ID:              utils.NewID(),
				Title:           "test title",
				ClassID:         "111111",
				LessonPlanID:    "1111",
				OrgID:           "111111",
				StartAt:         0,
				EndAt:           0,
				Status:          "",
				IsAllDay:        false,
				SubjectID:       "",
				ProgramID:       "",
				ClassType:       "",
				DueAt:           0,
				Description:     "",
				Attachment:      "",
				ScheduleVersion: 0,
				RepeatID:        "",
				RepeatJson:      "{}",
				IsHidden:        false,
				IsHomeFun:       false,
				CreatedID:       "",
				UpdatedID:       "",
				DeletedID:       "",
				CreatedAt:       0,
				UpdatedAt:       0,
				DeleteAt:        0,
			},
		})
		if err != nil {
			t.Log(err)
			return err
		}
		return nil
	})
	if err != nil {
		t.Log(err)
		return
	}
}
