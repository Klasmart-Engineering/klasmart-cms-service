package model

import (
	"context"
	"testing"
	"time"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
)

func TestAssessmentModel_GetByID(t *testing.T) {
	ctx := context.Background()
	op := &entity.Operator{
		UserID: "c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
		Token:  "",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0NTQzNzYzMiwiaXNzIjoia2lkc2xvb3AifQ.j7UGDogT1KYeOU--OnB37PsakQ44gcYBDt5D3fBSyBz9ipIzGvtP4wm5Jnv05SJv5ZF8dbmK7WKmGRBtFqCPlO_HHmrdooE562I4Y4IBVwolKsWzET7N3j_fcJUNvUhC6qzBcI3acrbmJBwaYR3DGjeODziR9ENu-Afhz8boSQvgNmiETjhPlxVg5c1N1nHFbjiO6qDRXFG469_jYZZNZBGGYxeOhAJH4NyTCvOwDiyW7SwAk3GZJj0S73ImOmhtiqz4bLFtvnaRR3uukFw6EqoWAPSGsw6RI1eu3kgTJXMQsYitOP72vZA23RGFQTJ68eDWDU9Lr5NFEIl2-ingUA"
	result, err := GetAssessmentModelV2().GetByID(ctx, op, "6114e9c1c83d392dc61e14ad")
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
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTY0MzA5NjQ3NCwiaXNzIjoia2lkc2xvb3AifQ.DSaYgvLCgIt_jwKfyyTFKk7JaJCzrK5yATmxhWtDIX9GJ0XS-zKhhqJfhjoij7GkzHWYJX9gF5iEb1UF4oviow8rndfMD1zOB5JkWa9SHEbmG7Q-kxmv1huBi0UmAdQN7Auvqrmrqc69rMRT_ckgmGKlckfunm_eFyaFjXN44re5tO5XPOWNw2g5kBF9oj-T1VHvhfbaueIWUTqysvvAnT_LtiENaCnyAWEJTnjj5kxttR6lTQJ_6REKWC4901OL6_VcjtuavffQDSc3uLturtst_JA9sgk1bUeQmAkGrzq3x2r1I5QRXmG7JMhFDqJDj7l_AXXf3NIhUzBOt_PVxQ"
	t.Log(op)
	result, err := GetAssessmentModelV2().Page(ctx, op, &v2.AssessmentQueryReq{
		//QueryKey:       "org mi",
		//QueryType:      v2.QueryTypeTeacherName,
		AssessmentType: v2.AssessmentTypeOnlineClass,
		OrderBy:        "-create_at",
		Status:         "NotStared,Started,Draft,Complete",
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
		UserID: "17a28338-3b88-4bac-ab15-cce3887af357", //"c57ef68d-a635-451d-b997-aebc3c29b99a",
		OrgID:  "6300b3c5-8936-497e-ba1f-d67164b59c65",
	}
	op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE3YTI4MzM4LTNiODgtNGJhYy1hYjE1LWNjZTM4ODdhZjM1NyIsImVtYWlsIjoib3JnbWlAeW9wbWFpbC5jb20iLCJleHAiOjE2Mzk5ODczODcsImlzcyI6ImtpZHNsb29wIn0.KZmDx445-P1g2YL7goVhTB1eXef-w1AWGwptVuroe2zy3-xoeVRHZ197vp2Yq_JMRKZ5PVbsd6clsgj6mg92FkUNxvFowhcy4EIB-UwO_6OG-ZO3yt5O9maGsHag7RovN9hRa1dBamAX9SgxSdBtCq7w4y6nEaS9IljN6AOWXLxP0Ued3v2dLoKHdAxSn1hzwdlh1e-baYvor_Cvne7CpRWoP8l7mGY85QofskS5UqYsbP-cvoJesh_HFh_Wq2p0r-YQCsx1PZAICSElNh5-5rt41_XfbVf5GntWkqk899CRN6QtGRmORxVBTTHcUD_KnhDyr2u0x2sFB-3mkavVWw"

	result, err := GetAssessmentModelV2().StatisticsCount(ctx, op, &v2.StatisticsCountReq{Status: "NotStared,Started,Draft,Complete"})
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
	err := GetAssessmentModelV2().DeleteByScheduleIDsTx(ctx, op, dbo.MustGetDB(ctx), []string{"6099c496e05f6e940027387c"})
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

	assessmentType, err := v2.GetAssessmentTypeByScheduleType(ctx, schedule.ClassType, schedule.IsHomeFun)
	if err != nil {
		t.Fatal(err)
	}
	err = GetAssessmentModelV2().AddWhenCreateSchedules(ctx, dbo.MustGetDB(ctx), op, &v2.AssessmentAddWhenCreateSchedulesReq{
		RepeatScheduleIDs:    []string{"6099c496e05f6e940027387c"},
		Users:                users,
		AssessmentType:       assessmentType,
		LessPlanID:           schedule.LessonPlanID,
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

	err := GetAssessmentModelV2().ScheduleEndClassCallback(ctx, op, &v2.ScheduleEndClassCallBackReq{
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