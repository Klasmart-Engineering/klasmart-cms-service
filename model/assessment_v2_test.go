package model

import (
	"context"
	"testing"
	"time"

	"github.com/KL-Engineering/dbo"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

const assessmentOpToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1NjM5NDE3MywiaXNzIjoia2lkc2xvb3AifQ.MFG44Dl2d67JIgC2m3-f-2aqPSLKQtkvXp0RswwJhXRWiDrFPNJRM6FL6F2iUZ0aKdWOFqz4LQyV8n10U0WO-jO_AFc9ovwCyVEwkOei1xD9oZXEgZrsD9JF3gXhcNBfPkKfFEUaBEgVtUW31fHIDujyoaTqs_eDbn52FB7gV1MyTAxII-4U1685_-vsPK7DNaoI85PPspac-6LkBaAzfXlmQGoy-9aiBTexQ31OGNbtPBi1hL-XZ4tfomB_jh4a5L3f8YaGbRZ9AEGH2lHZ9rtQdYD7y5o0ML4g1mZ5LO76r-S3Xsi280t6Qlfk5YqJcCo0W7tE7OP39-7emsTXUg"

func TestAssessmentModel_GetByID(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = assessmentOpToken
	result, err := GetAssessmentModelV2().GetByID(ctx, op, "617f9fa688932e7cb3892fca")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestAssessmentModel_Query(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = assessmentOpToken
	result, err := GetAssessmentModelV2().Page(ctx, op, &v2.AssessmentQueryReq{
		//QueryKey:       "org mi",
		//QueryType:      v2.QueryTypeTeacherName,
		AssessmentType: v2.AssessmentTypeOnlineStudy,
		OrderBy:        "-create_at",
		Status:         "Draft,Complete",
		PageIndex:      1,
		PageSize:       20,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestAssessmentModel_StatisticsCount(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0NTc3ODIyNiwiaXNzIjoia2lkc2xvb3AifQ.pAQe9Iu0k7GCX_YW26rCqRHPpdBAEKRzL23qkVjdbpJzVLBSn7brep3JzIjqioA3OEx53JZ7JzaVnv7dAvabr4CIPtJwxdIvtM6RB0UfzcDTI0qSfEpAr-TVLvw2oomxwnt7YOEd3xRr-V7B-T9l0auGOdStJwWNG60q1gdwpg9t6q9KIqAlAuyUDIOthsUi7-sT-jPoZtpXV9Riog0pilEqqejo5y3wYE6U5Xu5tIupYbikpAPdsA1DCY4T5KC06j4ao1YEdumjGEbC2YUOS__THbEq-69R5Fgv1RiuL98nQESAmrGE0TItNEk0Bf1rhRNcC0xzxTukr-WgIP4Zqw"

	result, err := GetAssessmentModelV2().StatisticsCount(ctx, op, &v2.StatisticsCountReq{Status: ""})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestAssessmentModel_DeleteByScheduleIDs(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
	}
	err := GetAssessmentInternalModel().DeleteByScheduleIDsTx(ctx, op, dbo.MustGetDB(ctx), []string{"6099c496e05f6e940027387c"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAssessmentModel_AddWhenCreateSchedules(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
	}

	scheduleID := "6099c496e05f6e940027387c"

	schedule, err := GetScheduleModel().GetPlainByID(ctx, scheduleID)
	if err != nil {
		t.Fatal(err)
	}
	relations, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, op, scheduleID)
	if err != nil {
		t.Fatal(err)
	}
	users := make([]*v2.AssessmentUserReq, 0, len(relations))
	for _, item := range relations {
		if item.RelationType == entity.ScheduleRelationTypeClassRosterTeacher ||
			item.RelationType == entity.ScheduleRelationTypeParticipantTeacher {
			users = append(users, &v2.AssessmentUserReq{
				UserID:   item.RelationID,
				UserType: v2.AssessmentUserTypeTeacher,
			})
		}
		if item.RelationType == entity.ScheduleRelationTypeClassRosterStudent ||
			item.RelationType == entity.ScheduleRelationTypeParticipantStudent {
			users = append(users, &v2.AssessmentUserReq{
				UserID:   item.RelationID,
				UserType: v2.AssessmentUserTypeStudent,
			})
		}
	}

	assessmentType, err := v2.GetAssessmentTypeByScheduleType(ctx, v2.GetAssessmentTypeByScheduleTypeInput{})
	if err != nil {
		t.Fatal(err)
	}
	err = GetAssessmentInternalModel().AddWhenCreateSchedules(ctx, dbo.MustGetDB(ctx), op, &v2.AssessmentAddWhenCreateSchedulesReq{
		RepeatScheduleIDs:    []string{"6099c496e05f6e940027387c"},
		Users:                users,
		AssessmentType:       assessmentType,
		ClassRosterClassName: "className",
		ScheduleTitle:        schedule.Title,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAssessmentModel_ScheduleEndClassCallback(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0MDA3OTQ2OSwiaXNzIjoia2lkc2xvb3AifQ.dRe0xdDBeoBzdbOzwzOTB6KNRAj-bS1Jd5MMxsEH5ZCK0ss9cLQpgHuf6nwKS1X2Qf-metgEzzdmjSVyWyudWC5J8Ick4MC4pYcc28_4_c8cy9vcQnMUczIDEw4aZal5o7gbOMJmAb9v-lR9Oa8gRTrojI132spEr-5vZ7MwUMbpsGP0Eqocah0ZiMzCpkQDX3xzi2m3_G511OkUgnaKYrA0OW0BQM67iRASDR32gUuvQuqPUX-loDkDACsuDTT99u1fPqpVVzc3i9JdOvSJYu90cn8eXWTtOznvmb48f5lnJ68Zxg_9duGmsuBTKZ8gbait7OBilhZ43QJucD2DBg"

	scheduleID := "6099c496e05f6e940027387c"

	err := GetAssessmentInternalModel().ScheduleEndClassCallback(ctx, op, &v2.ScheduleEndClassCallBackReq{
		ScheduleID: scheduleID,
		AttendanceIDs: []string{
			"b0ffe4a2-94fb-41b0-9e7a-8e2e51686003",
			"ff1d68d3-dc01-472f-906c-f9f578bd0936",
			"82c33e43-42e5-432a-8199-65beb38f5449",
		},
		ClassLength: 100,
		ClassEndAt:  time.Now().Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLiveRoomEventBus_PubEndClass(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0MzI0OTY4OSwiaXNzIjoia2lkc2xvb3AifQ.MI47b5wd5QCoVESS3w7fD5iEfDMNcyBp2-fJQtGa-aMa38QOtiRFFAJl6EB7UFx7mkyIiK7FfkpTvEkWGck5aONuLFhw4_UF8mKETeooWoLDzpb_sOHl6saysf23ONx-xUc4DOXVE8A5bc3ZkiA-k7WqiiKGXYH_fdZQwlTgZGm4D2U8iqAjm4EvUNrsefq_AcrhF4mGcdgAUWWRcLiMhu44kxLaNbGLZsgpiXDE8y_29P38oK4R-jFYOfBeGZQSJcwu-QQgZnreXuuMrUVMP0NPZItsUkjZMLRmuhB2cCPZZRTRUCdrsraXJjnhSbGrvpGLXuO-K1nFP6NPDJhHeg"
	args := &v2.ScheduleEndClassCallBackReq{
		ScheduleID:    "6099c496e05f6e940027387c",
		AttendanceIDs: []string{"80affb9c-94f3-4d32-9ec8-ddcf8d70b9ec", "82c33e43-42e5-432a-8199-65beb38f5449"},
		ClassLength:   10,
		ClassEndAt:    20,
	}

	err := GetLiveRoomEventBusModel().PubEndClass(ctx, op, args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPageForHomePage(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0NTc3OTEyNywiaXNzIjoia2lkc2xvb3AifQ.JiGubGqmjyfnYSCfIGtuHC8UK6j_BARb3NiltW_LrA5wBchjtysscn0aiECFPDD_r7eTQkxJaT-Ywb8_NQJ1YpRB2ShhyWOd2e-mrPmPaJ9_LJjq1uURAjJVi21wFuh4F1gl8qhi03UnHRBw8Lwo8I1oQmrpuy4JrCGJHaPK8o4nZMNOA0BeKV61kCWmqI3INPjACSboavYsYKfBnt0PlN6IukleccT_Ho_E9uzoLvRzGLbsGJIF7dSQ7KExBF3-WPLEz9jpJwDqFNiVctfP79FWlGZFITZyyopCx2DsG3pFrrL835JA3S1MfGNy9MzLCq6Gz9qrHSnT6ER-SvhDBA"

	result, err := GetAssessmentModelV2().PageForHomePage(ctx, op, &v2.AssessmentQueryReq{
		//QueryKey:       "org mi",
		//QueryType:      v2.QueryTypeTeacherName,
		//AssessmentType: v2.AssessmentTypeOnlineClass,
		//OrderBy:        "-create_at",
		Status:    "Complete",
		PageIndex: 1,
		PageSize:  5,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestQueryTeacherFeedback(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "aea0e494-e56f-417e-99a7-81774c879bf8",
		OrgID:  "f27efd10-000e-4542-bef2-0ccda39b93d3",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFlYTBlNDk0LWU1NmYtNDE3ZS05OWE3LTgxNzc0Yzg3OWJmOCIsImVtYWlsIjoic3R1MDJfMDAxQHlvcG1haWwuY29tIiwiZXhwIjoxNjUwOTYyNzA4LCJpc3MiOiJraWRzbG9vcCJ9.ZXwqpJV_0wDk_AmSOsQnwSTfP8V7X9lT1EXbNYf8JBBPKwgTQ412LNwtrBmDpNNL5hUq6YOt4iW1zA5IB3vVCX90x0AYzknU39Q2HPuONzaIxkHPVZI3tYLn00-rYdncDYK1XiJbv_APp4D0qC9JuSEn7aGF6OzvU31uCkSk9ab_ZE2tsJogA7ueqhNzlzWCOZ_vsx5ggZULT2_YBL2JHojYHfsPoCDrphkJm3gT3b6w7FIQozgL89fOLmTMPWabQ5zaJo95VZRj3GFESiFTr6Ld6SFfuMlSfxRvkkrCAHuSCDjoNVYBwR4iweCvs31i9ul3bAm02hC5OYQN5SGrWg"

	total, result, err := GetAssessmentModelV2().QueryStudentAssessment(ctx, op, &v2.StudentQueryAssessmentConditions{
		OrgID:     "f27efd10-000e-4542-bef2-0ccda39b93d3",
		StudentID: "aea0e494-e56f-417e-99a7-81774c879bf8",
		//ScheduleIDs: []string{"61249b0b8217ac62053ede9a", "61249d928217ac62053edfc2", "6125b30c4196280adf6424a4"},
		ClassType: "OfflineStudy",
		OrderBy:   "-complete_at",
		Page:      1,
		PageSize:  5,
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Log(total)
	t.Log(result)
}
