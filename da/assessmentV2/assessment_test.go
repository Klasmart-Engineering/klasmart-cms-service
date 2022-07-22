package assessmentV2

import (
	"context"
	"database/sql"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"testing"

	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	testUtils "github.com/KL-Engineering/kidsloop-cms-service/test/utils"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
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
	da.InitMySQL(ctx)
	total, result, err := GetAssessmentUserResultDA().GetAssessmentUserResultDBView(ctx, &AssessmentUserResultDBViewCondition{
		UserIDs: entity.NullStrings{
			Strings: []string{"aea0e494-e56f-417e-99a7-81774c879bf8"},
			Valid:   true,
		},
		OrgID: sql.NullString{
			String: "f27efd10-000e-4542-bef2-0ccda39b93d3",
			Valid:  true,
		},
		CompleteAtGe: sql.NullInt64{
			Int64: 11,
			Valid: true,
		},
		CompleteAtLe: sql.NullInt64{
			Int64: 0,
			Valid: true,
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

func TestQueryAssessment(t *testing.T) {
	ctx := context.Background()
	testUtils.InitConfig(ctx)
	da.InitMySQL(ctx)
	var data []*v2.Assessment
	err := GetAssessmentDA().Query(ctx, &AssessmentCondition{
		Status: entity.NullStrings{
			Strings: []string{
				v2.AssessmentStatusPending.String(),
				v2.AssessmentStatusNotStarted.String(),
				v2.AssessmentStatusStarted.String(),
				v2.AssessmentStatusInDraft.String(),
			},
			Valid: true,
		},
		ClassIDs: entity.NullStrings{
			Strings: []string{"fbadc2ed-4bbb-4c19-9b8f-ddf590513f531"},
			Valid:   true,
		},
		DueAtLe: sql.NullInt64{
			Int64: 1658459598,
			Valid: true,
		},
		Pager: dbo.Pager{
			Page:     1,
			PageSize: 10,
		},
	}, &data)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(data)
}
