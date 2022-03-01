package assessmentV2

import (
	"context"
	"database/sql"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	testUtils "gitlab.badanamu.com.cn/calmisland/kidsloop2/test/utils"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func TestAdd(t *testing.T) {
	ctx := context.Background()

	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		tx.DB = tx.Debug()

		err := da.GetScheduleDA().DeleteWithFollowing(ctx, tx, "5fac9892232323230981aafeecf6b48", 0)
		if err != nil {
			t.Log(err)
			return err
		}

		_, err = da.GetScheduleDA().Insert(ctx, &[]*entity.Schedule{
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

func TestGetAssessmentUserResultDBView(t *testing.T) {
	ctx := context.Background()
	testUtils.InitConfig(ctx)
	testUtils.InitDB(ctx)
	total, result, err := GetAssessmentUserResultDA().GetAssessmentUserResultDBView(ctx, &AssessmentUserResultDBViewCondition{
		UserIDs: entity.NullStrings{
			Strings: []string{"aea0e494-e56f-417e-99a7-81774c879bf8"},
		},
		OrgID: sql.NullString{
			String: "f27efd10-000e-4542-bef2-0ccda39b93d3",
			Valid:  true,
		},
		OrderBy: NewAssessmentUserResultOrderBy("-complete_at"),
		Pager: dbo.Pager{
			PageSize: 5,
			Page:     1,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Log(total)
	t.Log(result)
}
