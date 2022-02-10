package v2

import (
	"context"
	"fmt"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type AssessmentType string

const (
	AssessmentTypeOfflineClass AssessmentType = "OfflineClass"
	AssessmentTypeOnlineClass  AssessmentType = "OnlineClass"
	AssessmentTypeOnlineStudy  AssessmentType = "OnlineStudy"
	AssessmentTypeOfflineStudy AssessmentType = "OfflineStudy"
)

func GetAssessmentStatusByReq() map[AssessmentStatusForApiCompliant][]string {
	return map[AssessmentStatusForApiCompliant][]string{
		AssessmentStatusCompliantNotCompleted: []string{
			AssessmentStatusNotStarted.String(),
			AssessmentStatusStarted.String(),
			AssessmentStatusInDraft.String(),
		},
		AssessmentStatusCompliantCompleted: []string{
			AssessmentStatusComplete.String(),
		},
	}
}

//func (a AssessmentType) Status(req AssessmentStatusForApiCompliant) []string {
//	switch a {
//	case AssessmentTypeOnlineClass,
//		AssessmentTypeOfflineClass:
//		if req == "" {
//			return []string{
//				AssessmentStatusStarted.String(),
//				AssessmentStatusInDraft.String(),
//				AssessmentStatusComplete.String(),
//			}
//		} else if req == AssessmentStatusCompliantNotCompleted {
//			return []string{
//				AssessmentStatusStarted.String(),
//				AssessmentStatusInDraft.String(),
//			}
//		} else {
//			return []string{AssessmentStatusComplete.String()}
//		}
//	case AssessmentTypeOnlineStudy:
//		if req == "" {
//			return []string{
//				AssessmentStatusNotStarted.String(),
//				AssessmentStatusStarted.String(),
//				AssessmentStatusInDraft.String(),
//				AssessmentStatusComplete.String(),
//			}
//		} else if req == AssessmentStatusCompliantNotCompleted {
//			return []string{
//				AssessmentStatusNotStarted.String(),
//				AssessmentStatusStarted.String(),
//				AssessmentStatusInDraft.String(),
//			}
//		} else {
//			return []string{AssessmentStatusComplete.String()}
//		}
//	case AssessmentTypeOfflineStudy:
//		if req == "" {
//			return []string{
//				UserResultProcessStatusStarted.String(),
//				UserResultProcessStatusDraft.String(),
//				UserResultProcessStatusComplete.String(),
//			}
//		} else if req == AssessmentStatusCompliantNotCompleted {
//			return []string{
//				UserResultProcessStatusStarted.String(),
//				UserResultProcessStatusDraft.String(),
//			}
//		} else {
//			return []string{UserResultProcessStatusComplete.String()}
//		}
//	}
//
//	return []string{}
//}

func (a AssessmentType) String() string {
	return string(a)
}

type GenerateAssessmentTitleInput struct {
	ClassName    string
	ScheduleName string
	ClassEndAt   int64
}

func (a AssessmentType) Title(ctx context.Context, input GenerateAssessmentTitleInput) (string, error) {
	var title string
	if input.ClassName == "" {
		input.ClassName = constant.AssessmentNoClass
	}
	var timeStr string
	if input.ClassEndAt > 0 {
		timeStr = time.Unix(input.ClassEndAt, 0).Format("20060102")
	}

	switch a {
	case AssessmentTypeOfflineClass, AssessmentTypeOnlineClass:
		title = fmt.Sprintf("%s-%s-%s", timeStr, input.ClassName, input.ScheduleName)
	case AssessmentTypeOnlineStudy, AssessmentTypeOfflineStudy:
		title = fmt.Sprintf("%s-%s", input.ClassName, input.ScheduleName)
	default:
		log.Error(ctx, "get assessment title error", log.Any("input", input))
		return "", constant.ErrInvalidArgs
	}
	return title, nil
}

func GetAssessmentTypeByScheduleType(ctx context.Context, scheduleType entity.ScheduleClassType, isHomeFun bool) (AssessmentType, error) {
	var result AssessmentType

	switch scheduleType {
	case entity.ScheduleClassTypeHomework:
		if isHomeFun {
			result = AssessmentTypeOfflineStudy
		} else {
			result = AssessmentTypeOnlineStudy
		}
	case entity.ScheduleClassTypeOfflineClass:
		result = AssessmentTypeOfflineClass
	case entity.ScheduleClassTypeOnlineClass:
		result = AssessmentTypeOnlineClass
	default:
		log.Error(ctx, "ConvertScheduleTypeToAssessmentType error", log.String("scheduleType", scheduleType.String()), log.Bool("isHomeFun", isHomeFun))
		return "", constant.ErrInvalidArgs
	}

	return result, nil
}

type AssessmentQueryType string

const (
	QueryTypeTeacherName AssessmentQueryType = "TeacherName"
)

type AssessmentStatus string

const (
	// home fun study
	AssessmentStatusNotApplicable AssessmentStatus = "NA"
	// when create schedule
	AssessmentStatusNotStarted AssessmentStatus = "NotStarted"
	// when user started work
	AssessmentStatusStarted AssessmentStatus = "Started"
	// when teacher click save
	AssessmentStatusInDraft AssessmentStatus = "Draft"
	// when teacher click complete
	AssessmentStatusComplete AssessmentStatus = "Complete"
)

func (a AssessmentStatus) String() string {
	return string(a)
}
func (a AssessmentStatus) Valid() bool {
	return true
}

func (a AssessmentStatus) ToReply() AssessmentStatusForApiCompliant {
	switch a {
	case AssessmentStatusComplete:
		return AssessmentStatusCompliantCompleted
	default:
		return AssessmentStatusCompliantNotCompleted
	}
}

type UserResultProcessStatus string

const (
	UserResultProcessStatusStarted  UserResultProcessStatus = "Started"
	UserResultProcessStatusDraft    UserResultProcessStatus = "Draft"
	UserResultProcessStatusComplete UserResultProcessStatus = "Complete"
)

func (a UserResultProcessStatus) String() string {
	return string(a)
}

func (a UserResultProcessStatus) Valid() bool {
	switch a {
	case UserResultProcessStatusStarted, UserResultProcessStatusDraft, UserResultProcessStatusComplete:
		return true
	}

	return false
}

type AssessmentUserType string

const (
	AssessmentUserTypeTeacher AssessmentUserType = "Teacher"
	AssessmentUserTypeStudent AssessmentUserType = "Student"
)

func (a AssessmentUserType) String() string {
	return string(a)
}
func GetUserTypeByScheduleRelationType(relationType entity.ScheduleRelationType) AssessmentUserType {
	switch relationType {
	case entity.ScheduleRelationTypeClassRosterStudent, entity.ScheduleRelationTypeParticipantStudent:
		return AssessmentUserTypeStudent
	case entity.ScheduleRelationTypeClassRosterTeacher, entity.ScheduleRelationTypeParticipantTeacher:
		return AssessmentUserTypeTeacher
	default:
		return ""
	}
}

type QuestionSource string

const (
	QuestionSourceAttachment QuestionSource = "Attachment"
	QuestionSourceContent    QuestionSource = "AssessmentContent"
)

type AssessmentUserOutcomeStatus string

const (
	AssessmentUserOutcomeStatusUnknown     AssessmentUserOutcomeStatus = "Unknown"
	AssessmentUserOutcomeStatusNotCovered  AssessmentUserOutcomeStatus = "NotCovered"
	AssessmentUserOutcomeStatusNotAchieved AssessmentUserOutcomeStatus = "NotAchieved"
	AssessmentUserOutcomeStatusAchieved    AssessmentUserOutcomeStatus = "Achieved"
)

func (a AssessmentUserOutcomeStatus) String() string {
	return string(a)
}
func (a AssessmentUserOutcomeStatus) Valid() bool {
	switch a {
	case AssessmentUserOutcomeStatusUnknown,
		AssessmentUserOutcomeStatusNotCovered,
		AssessmentUserOutcomeStatusNotAchieved,
		AssessmentUserOutcomeStatusAchieved:
		return true
	}

	return false
}

type AssessmentUserStatus string

const (
	AssessmentUserStatusParticipate    AssessmentUserStatus = "Participate"
	AssessmentUserStatusNotParticipate AssessmentUserStatus = "NotParticipate"
)

func (a AssessmentUserStatus) String() string {
	return string(a)
}

func (a AssessmentUserStatus) Valid() bool {
	switch a {
	case AssessmentUserStatusParticipate, AssessmentUserStatusNotParticipate:
		return true
	default:
		return false
	}
}

type AssessmentContentStatus string

const (
	AssessmentContentStatusCovered    AssessmentContentStatus = "Covered"
	AssessmentContentStatusNotCovered AssessmentContentStatus = "NotCovered"
)

func (a AssessmentContentStatus) Valid() bool {
	switch a {
	case AssessmentContentStatusCovered, AssessmentContentStatusNotCovered:
		return true
	default:
		return false
	}
}

type AssessmentOutcomeAssignType string

const (
	AssessmentOutcomeAssignTypeLessonPlan     AssessmentOutcomeAssignType = "LessonPlan"
	AssessmentOutcomeAssignTypeLessonMaterial AssessmentOutcomeAssignType = "LessonMaterial"
)

type AssessmentContentType string

const (
	AssessmentContentTypeUnknown AssessmentContentType = "Unknown"

	AssessmentContentTypeLessonPlan     AssessmentContentType = "LessonPlan"
	AssessmentContentTypeLessonMaterial AssessmentContentType = "LessonMaterial"
)

func (a AssessmentContentType) String() string {
	return string(a)
}

type AssessmentFileArchived string

const (
	AssessmentFileTypeHasChildContainer         AssessmentFileArchived = "HasChildContainer"
	AssessmentFileTypeNotChildSubContainer      AssessmentFileArchived = "NotChildContainer"
	AssessmentFileTypeSupportScoreStandAlone    AssessmentFileArchived = "SupportScoreStandAlone"
	AssessmentFileTypeNotSupportScoreStandAlone AssessmentFileArchived = "NotSupportScoreStandAlone"
	AssessmentFileTypeNotUnknown                AssessmentFileArchived = "Unknown"
)

type AssessmentUserAssess int

const (
	AssessmentUserAssessPoor AssessmentUserAssess = iota + 1
	AssessmentUserAssessFair
	AssessmentUserAssessAverage
	AssessmentUserAssessGood
	AssessmentUserAssessExcellent
)

func (s AssessmentUserAssess) Valid() bool {
	switch s {
	case AssessmentUserAssessPoor,
		AssessmentUserAssessFair,
		AssessmentUserAssessAverage,
		AssessmentUserAssessGood,
		AssessmentUserAssessExcellent:
		return true
	}
	return false
}

type AssessmentAction string

const (
	AssessmentActionDraft    AssessmentAction = "Draft"
	AssessmentActionComplete AssessmentAction = "Complete"
)
