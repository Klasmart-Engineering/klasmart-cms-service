package model

import (
	"context"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

const assessmentOpToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ2NTNmODgwLWI5MjAtNDM1Zi04ZjJkLTk4YjVkNDYyMWViOCIsImVtYWlsIjoic2Nob29sXzAzMDMyOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0OTQwOTgxMywiaXNzIjoia2lkc2xvb3AifQ.GjJXnYds3Tk-qCA8jaJeDsj64FTOAeqUKyFgFLBwjHHz7kIa6zSRtqsJfSyH4EBradrYoxOtYRXsXCdm7wEI0vkkgJ8ktsyeNUo9nmYL80I7XFMM7ASdWQYKulunj4oOB3-BOaRx8i2aG9gx3uU-hxIGbqT44MBoTiulw8NAF3vxr5jAP8btZLbomCkyWm6AxVOcMbsPZo23Mxo2a742fUR8DmuBl1lffT3RdZkg7Tn6upBKT5Ec0ldh5CEU0pPaDPSFrB47uVvwELsjTzjZXo7snI2pFO_hoB0587XYy_2HA2qgDpjyK7a3HrqVqT-uqB11K5wvhH80DIclxac8wg"

func TestAssessmentModel_GetByID(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "f27efd10-000e-4542-bef2-0ccda39b93d3",
		Token:  "",
	}
	op.Token = assessmentOpToken
	result, err := GetAssessmentModelV2().GetByID(ctx, op, "60a36e7fde590052a3c5dd96")
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
	t.Log(op)
	result, err := GetAssessmentModelV2().Page(ctx, op, &v2.AssessmentQueryReq{
		//QueryKey:       "org mi",
		//QueryType:      v2.QueryTypeTeacherName,
		AssessmentType: v2.AssessmentTypeReviewStudy,
		OrderBy:        "-create_at",
		Status:         "Started,Draft,Complete",
		PageIndex:      1,
		PageSize:       10,
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
		UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399",
		OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFlYTBlNDk0LWU1NmYtNDE3ZS05OWE3LTgxNzc0Yzg3OWJmOCIsImVtYWlsIjoic3R1MDJfMDAxQHlvcG1haWwuY29tIiwiZXhwIjoxNjQ2MTIxNDQ2LCJpc3MiOiJraWRzbG9vcCJ9.IMXxYAIB-eH1es1dS4xD5WOTLG6lkhcJGo6Kp3qEYaeO6ivcHOjNiq7bZwnK-QuU-kv2qeWUwZ0uQHFTnJ7CupL8KGB-fye2nn1I1sPZ4VL_eSWiyYG8rV4zXEukTEm3EmIGMN_TvCIkloRYFqq9_PWfYxc1pu8wBRPHbzU0hZwtDcUweuLLvkCev4LAJoaHI0DlvfrJ6NK0GVGI-p2ROf2219kPuHwd-1RVR91jHNwzRejFi1Y3eft-olU04aPg5mwEvGk3E-2SC8zzuXi9EE_FrcAzPpU5kuhTXMZh1LHdFGlI2Ws9slX-LOKga-rt5-Qsk-xUaE9vj2eQBTWtqA"

	total, result, err := GetAssessmentModelV2().QueryTeacherFeedback(ctx, op, &v2.StudentQueryAssessmentConditions{
		OrgID:     "f27efd10-000e-4542-bef2-0ccda39b93d3",
		StudentID: "aea0e494-e56f-417e-99a7-81774c879bf8",
		ClassType: "home_fun_study",
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
