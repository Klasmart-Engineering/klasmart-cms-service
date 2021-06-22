package model

import (
	"context"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"sort"
	"strings"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type assessmentBase struct{}

func (m *assessmentBase) getDetail(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetail, error) {
	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "Get: da.GetAssessmentDA().GetExcludeSoftDeleted: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
		return nil, err
	}

	// convert to assessment view
	var (
		views []*entity.AssessmentView
		view  *entity.AssessmentView
	)
	if views, err = m.toViews(ctx, tx, operator, []*entity.Assessment{assessment}, entity.ConvertToViewsOptions{
		EnableProgram:    true,
		EnableSubjects:   true,
		EnableTeachers:   true,
		EnableStudents:   true,
		EnableClass:      true,
		EnableLessonPlan: true,
	}); err != nil {
		log.Error(ctx, "Get: GetAssessmentUtils().toViews: get failed",
			log.Err(err),
			log.String("assessment_id", id),
			log.Any("operator", operator),
		)
		return nil, err
	}
	view = views[0]

	// fill result
	result := entity.AssessmentDetail{
		ID:           view.ID,
		Title:        view.Title,
		Status:       view.Status,
		Schedule:     view.Schedule,
		RoomID:       view.RoomID,
		Class:        view.Class,
		Teachers:     view.Teachers,
		Students:     view.Students,
		Program:      view.Program,
		Subjects:     view.Subjects,
		ClassEndTime: view.ClassEndTime,
		ClassLength:  view.ClassLength,
		CompleteTime: view.CompleteTime,
	}

	// fill lesson plan and lesson materials
	var contentIDs []string
	if view.LessonPlan != nil {
		contentIDs = append(contentIDs, view.LessonPlan.ID)
	}
	for _, lm := range view.LessonMaterials {
		contentIDs = append(contentIDs, lm.ID)
	}
	var currentContentOutcomeMap map[string][]string
	if len(contentIDs) > 0 {
		assessmentContentOutcomeMap, err := m.getAssessmentContentOutcomeMap(ctx, tx, []string{id}, contentIDs)
		if err != nil {
			log.Error(ctx, "Get: m.getAssessmentContentOutcomeMap: get failed",
				log.Err(err),
				log.String("assessment_id", id),
				log.Strings("content_ids", contentIDs),
			)
			return nil, err
		}
		currentContentOutcomeMap = assessmentContentOutcomeMap[id]
		if view.LessonPlan != nil {
			result.LessonPlan = entity.AssessmentDetailContent{
				ID:         view.LessonPlan.ID,
				Name:       view.LessonPlan.Name,
				Checked:    true,
				OutcomeIDs: currentContentOutcomeMap[view.LessonPlan.ID],
			}
		}
		for _, m := range view.LessonMaterials {
			result.LessonMaterials = append(result.LessonMaterials, &entity.AssessmentDetailContent{
				ID:         m.ID,
				Name:       m.Name,
				Comment:    m.Comment,
				Checked:    m.Checked,
				OutcomeIDs: currentContentOutcomeMap[m.ID],
			})
		}
	}

	// fill outcomes
	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().QueryTx(ctx, tx, &da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{id},
			Valid:   true,
		},
	}, &assessmentOutcomes); err != nil {
		log.Error(ctx, "Get: da.GetAssessmentOutcomeDA().GetListByAssessmentID: get list failed",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}
	if len(assessmentOutcomes) > 0 {
		var (
			assessmentOutcomeMap = make(map[string]entity.AssessmentOutcome, len(assessmentOutcomes))
			outcomeIDs           = make([]string, 0, len(assessmentOutcomes))
			outcomes             = make([]*entity.Outcome, 0, len(assessmentOutcomes))
		)
		for _, o := range assessmentOutcomes {
			assessmentOutcomeMap[o.OutcomeID] = *o
			outcomeIDs = append(outcomeIDs, o.OutcomeID)
		}
		if outcomes, err = GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs); err != nil {
			log.Error(ctx, "Get: GetOutcomeModel().GetByIDs: get failed",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
				log.String("assessment_id", id),
				log.Any("operator", operator),
			)
			return nil, err
		}
		sort.Sort(OutcomesOrderByAssumedAndName(outcomes))

		var (
			outcomeAttendances      = make([]*entity.OutcomeAttendance, 0, len(outcomeIDs))
			outcomeAttendanceIDsMap = make(map[string][]string, len(outcomeIDs))
		)
		outcomeAttendances, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, tx, id, outcomeIDs)
		if err != nil {
			log.Error(ctx, "Get: da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs: batch get failed",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
				log.String("assessment_id", id),
				log.Any("operator", operator),
			)
			return nil, err
		}
		for _, item := range outcomeAttendances {
			outcomeAttendanceIDsMap[item.OutcomeID] = append(outcomeAttendanceIDsMap[item.OutcomeID], item.AttendanceID)
		}
		for _, o := range outcomes {
			newOutcome := entity.AssessmentDetailOutcome{
				OutcomeID:     o.ID,
				OutcomeName:   o.Name,
				Assumed:       o.Assumed,
				Skip:          assessmentOutcomeMap[o.ID].Skip,
				NoneAchieved:  assessmentOutcomeMap[o.ID].NoneAchieved,
				AttendanceIDs: outcomeAttendanceIDsMap[o.ID],
				Checked:       assessmentOutcomeMap[o.ID].Checked,
			}
			result.Outcomes = append(result.Outcomes, &newOutcome)
		}
	}

	// fill remaining time
	result.RemainingTime = int64(m.calcRemainingTime(view.Schedule.DueAt, view.CreateAt).Seconds())

	// fill student view items
	if view.Schedule.ClassType != entity.ScheduleClassTypeOfflineClass {
		result.StudentViewItems, err = m.getH5PStudentViewItems(ctx, operator, tx, view)
		if err != nil {
			log.Error(ctx, "get assessment detail: get student view items failed",
				log.Err(err),
				log.Any("operator", operator),
				log.Any("view", view),
			)
			return nil, err
		}
	}

	return &result, nil
}

func (m *assessmentBase) getAssessmentContentOutcomeMap(ctx context.Context, tx *dbo.DBContext, assessmentIDs []string, contentIDs []string) (map[string]map[string][]string, error) {
	var assessmentContentOutcomes []*entity.AssessmentContentOutcome
	cond := da.QueryAssessmentContentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		ContentIDs: entity.NullStrings{
			Strings: contentIDs,
			Valid:   true,
		},
	}
	if err := da.GetAssessmentContentOutcomeDA().QueryTx(ctx, tx, &cond, &assessmentContentOutcomes); err != nil {
		log.Error(ctx, "getAssessmentContentOutcomeMap: da.GetAssessmentContentOutcomeDA().QueryTx: get failed",
			log.Err(err),
			log.Any("cond", cond),
			log.Strings("assessment_ids", assessmentIDs),
			log.Strings("content_ids", contentIDs),
		)
		return nil, err
	}
	result := map[string]map[string][]string{}
	for _, co := range assessmentContentOutcomes {
		if result[co.AssessmentID] == nil {
			result[co.AssessmentID] = map[string][]string{co.ContentID: {co.OutcomeID}}
		} else {
			result[co.AssessmentID][co.ContentID] = append(result[co.AssessmentID][co.ContentID], co.OutcomeID)
		}
	}
	return result, nil
}

func (m *assessmentBase) existsByScheduleID(ctx context.Context, operator *entity.Operator, scheduleID string) (bool, error) {
	var assessments []*entity.Assessment
	cond := da.QueryAssessmentConditions{
		ScheduleIDs: entity.NullStrings{
			Strings: []string{scheduleID},
			Valid:   true,
		},
	}
	if err := da.GetAssessmentDA().Query(ctx, &cond, &assessments); err != nil {
		log.Error(ctx, "existsByScheduleID: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Any("cond", cond),
		)
		return false, err
	}
	return len(assessments) > 0, nil
}

func (m *assessmentBase) calcRemainingTime(dueAt int64, createdAt int64) time.Duration {
	var r int64
	if dueAt != 0 {
		r = dueAt - time.Now().Unix()
	} else {
		r = time.Unix(createdAt, 0).Add(config.Get().Assessment.DefaultRemainingTime).Unix() - time.Now().Unix()
	}
	if r < 0 {
		return 0
	}
	return time.Duration(r) * time.Second
}

func (m *assessmentBase) checkEditPermission(ctx context.Context, operator *entity.Operator, id string) error {
	hasP439, err := NewAssessmentPermissionChecker(operator).HasP439(ctx)
	if err != nil {
		return err
	}
	if !hasP439 {
		log.Error(ctx, "check edit permission: not have permission 439",
			log.String("id", id),
			log.Any("operator", operator),
		)
		return constant.ErrForbidden
	}
	teacherIDs, err := da.GetAssessmentAttendanceDA().GetTeacherIDsByAssessmentID(ctx, dbo.MustGetDB(ctx), id)
	if err != nil {
		log.Error(ctx, "check edit permission: get teacher ids failed by assessment id",
			log.String("assessment_id", id),
			log.Any("operator", operator),
		)
		return err
	}
	hasOperator := false
	for _, tid := range teacherIDs {
		if tid == operator.UserID {
			hasOperator = true
			break
		}
	}
	if !hasOperator {
		log.Error(ctx, "check edit permission: not found operator",
			log.String("id", id),
			log.Any("operator", operator),
		)
		return constant.ErrForbidden
	}
	return nil
}

func (m *assessmentBase) toViews(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessments []*entity.Assessment, options entity.ConvertToViewsOptions) ([]*entity.AssessmentView, error) {
	if len(assessments) == 0 {
		return nil, nil
	}

	var (
		err           error
		assessmentIDs []string
		scheduleIDs   []string
		schedules     []*entity.ScheduleVariable
		scheduleMap   = map[string]*entity.ScheduleVariable{}
	)
	for _, a := range assessments {
		assessmentIDs = append(assessmentIDs, a.ID)
		scheduleIDs = append(scheduleIDs, a.ScheduleID)
	}
	if schedules, err = GetScheduleModel().GetVariableDataByIDs(ctx, operator, scheduleIDs, &entity.ScheduleInclude{Subject: true}); err != nil {
		log.Error(ctx, "toViews: GetScheduleModel().GetVariableDataByIDs: get failed",
			log.Err(err),
			log.Strings("assessment_ids", assessmentIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	for _, s := range schedules {
		scheduleMap[s.ID] = s
	}

	// fill program
	var programNameMap map[string]string
	if options.EnableProgram {
		programIDs := make([]string, 0, len(schedules))
		for _, s := range schedules {
			programIDs = append(programIDs, s.ProgramID)
		}
		programNameMap, err = external.GetProgramServiceProvider().BatchGetNameMap(ctx, operator, programIDs)
		if err != nil {
			log.Error(ctx, "toViews: external.GetProgramServiceProvider().BatchGetNameMap: get failed",
				log.Err(err),
				log.Strings("assessment_ids", assessmentIDs),
				log.Strings("program_ids", programIDs),
				log.Any("operator", operator),
			)
			return nil, err
		}
	}

	// fill teachers
	var (
		teacherNameMap        map[string]string
		assessmentTeachersMap map[string][]*entity.AssessmentAttendance
	)
	if options.EnableTeachers {
		var (
			assessmentAttendances []*entity.AssessmentAttendance
			teacherIDs            []string
		)
		assessmentTeachersMap = map[string][]*entity.AssessmentAttendance{}
		if err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.QueryAssessmentAttendanceConditions{
			AssessmentIDs: entity.NullStrings{
				Strings: assessmentIDs,
				Valid:   true,
			},
			Role: entity.NullAssessmentAttendanceRole{
				Value: entity.AssessmentAttendanceRoleTeacher,
				Valid: true,
			},
		}, &assessmentAttendances); err != nil {
			log.Error(ctx, "toViews: da.GetAssessmentAttendanceDA().QueryTx: query failed",
				log.Err(err),
				log.Strings("assessment_ids", assessmentIDs),
				log.Any("operator", operator),
			)
			return nil, err
		}
		sort.Sort(AssessmentAttendanceOrderByOrigin(assessmentAttendances))
		for _, a := range assessmentAttendances {
			teacherIDs = append(teacherIDs, a.AttendanceID)
			assessmentTeachersMap[a.AssessmentID] = append(assessmentTeachersMap[a.AssessmentID], a)
		}
		if teacherNameMap, err = external.GetTeacherServiceProvider().BatchGetNameMap(ctx, operator, teacherIDs); err != nil {
			log.Error(ctx, "toViews: external.GetTeacherServiceProvider().BatchGetNameMap: get failed",
				log.Err(err),
				log.Strings("teacher_ids", teacherIDs),
				log.Strings("assessment_ids", assessmentIDs),
				log.Any("operator", operator),
			)
			return nil, err
		}
	}

	// fill students
	var (
		studentNameMap        map[string]string
		assessmentStudentsMap map[string][]*entity.AssessmentAttendance
	)
	if options.EnableStudents {
		var (
			assessmentAttendances []*entity.AssessmentAttendance
			studentIDs            []string
		)
		assessmentStudentsMap = map[string][]*entity.AssessmentAttendance{}
		if err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.QueryAssessmentAttendanceConditions{
			AssessmentIDs: entity.NullStrings{
				Strings: assessmentIDs,
				Valid:   true,
			},
			Role: entity.NullAssessmentAttendanceRole{
				Value: entity.AssessmentAttendanceRoleStudent,
				Valid: true,
			},
		}, &assessmentAttendances); err != nil {
			log.Error(ctx, "toViews: da.GetAssessmentAttendanceDA().QueryTx: query failed",
				log.Err(err),
				log.Strings("assessment_ids", assessmentIDs),
				log.Any("operator", operator),
			)
			return nil, err
		}
		sort.Sort(AssessmentAttendanceOrderByOrigin(assessmentAttendances))
		for _, a := range assessmentAttendances {
			if !options.CheckedStudents.Valid || options.CheckedStudents.Bool == a.Checked {
				studentIDs = append(studentIDs, a.AttendanceID)
				assessmentStudentsMap[a.AssessmentID] = append(assessmentStudentsMap[a.AssessmentID], a)
			}
		}
		if studentNameMap, err = external.GetStudentServiceProvider().BatchGetNameMap(ctx, operator, studentIDs); err != nil {
			log.Error(ctx, "toViews: external.GetStudentServiceProvider().BatchGetNameMap: get failed",
				log.Err(err),
				log.Strings("student_ids", studentIDs),
				log.Strings("assessment_ids", assessmentIDs),
				log.Any("operator", operator),
			)
			return nil, err
		}
	}

	// fill class
	var classNameMap map[string]string
	if options.EnableClass {
		var classIDs []string
		for _, s := range schedules {
			classIDs = append(classIDs, s.ClassID)
		}
		classNameMap, err = external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, classIDs)
		if err != nil {
			return nil, err
		}
	}

	// fill lesson plan
	var (
		assessmentLessonPlanMap      map[string]*entity.AssessmentLessonPlan
		assessmentLessonMaterialsMap map[string][]*entity.AssessmentLessonMaterial
		sortedLessonMaterialIDsMap   map[string][]string
	)
	if options.EnableLessonPlan {
		var contents []*entity.AssessmentContent
		err := da.GetAssessmentContentDA().Query(ctx, &da.QueryAssessmentContentConditions{
			AssessmentIDs: entity.NullStrings{
				Strings: assessmentIDs,
				Valid:   true,
			},
		}, &contents)
		if err != nil {
			log.Error(ctx, "convert to views: query assessment content failed",
				log.Err(err),
				log.Strings("assessment_ids", assessmentIDs),
			)
			return nil, err
		}
		assessmentLessonPlanMap = map[string]*entity.AssessmentLessonPlan{}
		assessmentLessonMaterialsMap = map[string][]*entity.AssessmentLessonMaterial{}

		var lessonMaterialIDs []string
		for _, c := range contents {
			switch c.ContentType {
			case entity.ContentTypeMaterial:
				lessonMaterialIDs = append(lessonMaterialIDs, c.ContentID)
			}
		}
		lessonMaterialSourceMap, err := m.batchGetLessonMaterialDataMap(ctx, tx, operator, lessonMaterialIDs)
		if err != nil {
			log.Error(ctx, "to views: get lesson material source map failed",
				log.Err(err),
				log.Strings("lesson_material_ids", lessonMaterialIDs),
			)
			return nil, err
		}

		var lessonPlanIDs []string
		for _, c := range contents {
			switch c.ContentType {
			case entity.ContentTypePlan:
				lessonPlanIDs = append(lessonPlanIDs, c.ContentID)
				assessmentLessonPlanMap[c.AssessmentID] = &entity.AssessmentLessonPlan{
					ID:   c.ContentID,
					Name: c.ContentName,
				}
			case entity.ContentTypeMaterial:
				data := lessonMaterialSourceMap[c.ContentID]
				if data == nil {
					data = &MaterialData{}
				}
				assessmentLessonMaterialsMap[c.AssessmentID] = append(assessmentLessonMaterialsMap[c.AssessmentID], &entity.AssessmentLessonMaterial{
					ID:       c.ContentID,
					Name:     c.ContentName,
					FileType: data.FileType,
					Comment:  c.ContentComment,
					Source:   string(data.Source),
					Checked:  c.Checked,
				})
			}
		}

		sortedLessonMaterialIDsMap, err = m.getSortedLessonMaterialIDsMap(ctx, tx, operator, lessonPlanIDs)
		if err != nil {
			log.Error(ctx, "to assessment views: get sorted lesson material ids map failed",
				log.Err(err),
				log.Strings("lesson_plan_ids", lessonPlanIDs),
			)
			return nil, err
		}
		log.Debug(ctx, "to assessment views: get sorted lesson material ids map",
			log.Strings("lesson_plan_ids", lessonPlanIDs),
			log.Any("sorted_lesson_material_ids_map", sortedLessonMaterialIDsMap),
		)
	}

	var result []*entity.AssessmentView
	for _, a := range assessments {
		var (
			v = entity.AssessmentView{Assessment: a}
			s = scheduleMap[a.ScheduleID]
		)
		if s == nil {
			log.Warn(ctx, "List: not found schedule", log.Any("assessment", a))
			continue
		}
		v.Schedule = s.Schedule
		v.RoomID = s.RoomID
		if options.EnableProgram {
			v.Program = entity.AssessmentProgram{
				ID:   s.ProgramID,
				Name: programNameMap[s.ProgramID],
			}
		}
		if options.EnableSubjects {
			for _, subject := range s.Subjects {
				v.Subjects = append(v.Subjects, &entity.AssessmentSubject{
					ID:   subject.ID,
					Name: subject.Name,
				})
			}
		}
		if options.EnableTeachers {
			for _, t := range assessmentTeachersMap[a.ID] {
				v.Teachers = append(v.Teachers, &entity.AssessmentTeacher{
					ID:   t.AttendanceID,
					Name: teacherNameMap[t.AttendanceID],
				})
			}
		}
		if options.EnableStudents {
			for _, s := range assessmentStudentsMap[a.ID] {
				v.Students = append(v.Students, &entity.AssessmentStudent{
					ID:      s.AttendanceID,
					Name:    studentNameMap[s.AttendanceID],
					Checked: s.Checked,
				})
			}
		}
		if options.EnableClass {
			v.Class = entity.AssessmentClass{
				ID:   s.ClassID,
				Name: classNameMap[s.ClassID],
			}
		}
		if options.EnableLessonPlan {
			lp := assessmentLessonPlanMap[a.ID]
			v.LessonPlan = lp
			var sortLessonMaterialIDs []string
			if lp != nil {
				sortLessonMaterialIDs = sortedLessonMaterialIDsMap[lp.ID]
			}
			lms := assessmentLessonMaterialsMap[a.ID]
			m.sortedByLessonMaterialIDs(lms, sortLessonMaterialIDs)
			v.LessonMaterials = lms
		}
		result = append(result, &v)
	}

	log.Debug(ctx, "convert assessments to views",
		log.Any("result", result),
		log.Any("operator", operator),
		log.Any("assessments", assessments),
		log.Any("options", options),
	)

	return result, nil
}

func (m *assessmentBase) getRoomCompleteRate(room *entity.AssessmentH5PRoom, v *entity.AssessmentView) float64 {
	if room == nil {
		return 0
	}

	// calc total
	checkedUserIDs := make([]string, 0, len(v.Students))
	for _, s := range v.Students {
		if s.Checked {
			checkedUserIDs = append(checkedUserIDs, s.ID)
		}
	}
	checkedH5PIDs := make([]string, 0, len(v.LessonMaterials))
	for _, lm := range v.LessonMaterials {
		if lm.Checked && (lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend) {
			checkedH5PIDs = append(checkedH5PIDs, lm.Source)
		}
	}
	total := len(checkedUserIDs) * len(checkedH5PIDs)

	// calc attempted
	attempted := 0
	for _, uid := range utils.SliceDeduplicationExcludeEmpty(checkedUserIDs) {
		u := room.UserMap[uid]
		if u == nil {
			continue
		}
		for _, h5pID := range utils.SliceDeduplicationExcludeEmpty(checkedH5PIDs) {
			c := u.ContentMap[h5pID]
			if c == nil {
				continue
			}
			if len(c.Answers) > 0 || len(c.Scores) > 0 {
				attempted++
			}
		}
	}

	if total > 0 {
		return float64(attempted) / float64(total)
	}

	return 0
}

func (m *assessmentBase) batchGetRoomScoreMap(ctx context.Context, operator *entity.Operator, roomIDs []string, enableComment bool) (map[string]*entity.AssessmentH5PRoom, error) {
	roomScoreMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		return nil, err
	}

	var roomCommentMap map[string]map[string][]string
	if enableComment {
		roomCommentMap, err = m.batchGetRoomCommentMap(ctx, operator, roomIDs)
		if err != nil {
			return nil, err
		}
	}

	result := make(map[string]*entity.AssessmentH5PRoom, len(roomScoreMap))
	for roomID, users := range roomScoreMap {
		assessmentUsers := make([]*entity.AssessmentH5PUser, 0, len(users))
		assessmentUserMap := make(map[string]*entity.AssessmentH5PUser, len(users))
		attempted, total := 0, 0
		for _, u := range users {
			assessmentContents := make([]*entity.AssessmentH5PContentScore, 0, len(u.Scores))
			assessmentContentMap := make(map[string]*entity.AssessmentH5PContentScore, len(u.Scores))
			for _, s := range u.Scores {
				total++
				assessmentContent := entity.AssessmentH5PContentScore{
					Scores: s.Score.Scores,
				}
				if s.Content != nil {
					assessmentContent.H5PID = s.Content.H5PID
					assessmentContent.ContentID = s.Content.ContentID
					assessmentContent.ContentName = s.Content.Name
					assessmentContent.ContentType = s.Content.Type
				}
				if len(s.Score.Answers) > 0 {
					assessmentContent.MaxPossibleScore = s.Score.Answers[0].MaximumPossibleScore
				}
				for _, a := range s.Score.Answers {
					assessmentContent.Answers = append(assessmentContent.Answers, a.Answer)
				}
				attemptedFlag := false
				if len(assessmentContent.Answers) > 0 {
					assessmentContent.Answer = assessmentContent.Answers[0]
					attemptedFlag = true
				}
				if len(s.TeacherScores) > 0 {
					assessmentContent.AchievedScore = s.TeacherScores[len(s.TeacherScores)-1].Score
					attemptedFlag = true
				} else if len(s.Score.Scores) > 0 {
					assessmentContent.AchievedScore = s.Score.Scores[0]
					attemptedFlag = true
				}
				if attemptedFlag {
					attempted++
				}
				assessmentContents = append(assessmentContents, &assessmentContent)
				assessmentContentMap[assessmentContent.H5PID] = &assessmentContent
			}
			assessmentUser := entity.AssessmentH5PUser{
				Contents:   assessmentContents,
				ContentMap: assessmentContentMap,
			}
			if u.User != nil {
				assessmentUser.UserID = u.User.UserID
			}
			if enableComment &&
				roomCommentMap != nil &&
				roomCommentMap[roomID] != nil &&
				assessmentUser.UserID != "" &&
				len(roomCommentMap[roomID][assessmentUser.UserID]) > 0 {
				cc := roomCommentMap[roomID][assessmentUser.UserID]
				assessmentUser.Comment = cc[len(cc)-1]
			}
			assessmentUsers = append(assessmentUsers, &assessmentUser)
			assessmentUserMap[assessmentUser.UserID] = &assessmentUser
		}
		room := entity.AssessmentH5PRoom{
			AnyoneAttempted: attempted > 0,
			Users:           assessmentUsers,
			UserMap:         assessmentUserMap,
		}
		result[roomID] = &room
	}

	log.Debug(ctx, "get room score map",
		log.Strings("room_ids", roomIDs),
		log.Any("operator", operator),
		log.Any("result", result),
	)

	return result, nil
}

func (m *assessmentBase) batchGetRoomCommentMap(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]map[string][]string, error) {
	commentMap, err := external.GetH5PRoomCommentServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string][]string, len(commentMap))
	for roomID, users := range commentMap {
		result[roomID] = make(map[string][]string, len(users))
		for _, u := range users {
			for _, c := range u.TeacherComments {
				var uid string
				if c.Student != nil {
					uid = c.Student.UserID
				}
				if uid == "" && u.User != nil {
					uid = u.User.UserID
				}
				if uid == "" {
					continue
				}
				result[roomID][uid] = append(result[roomID][uid], c.Comment)
			}
		}
	}
	log.Debug(ctx, "batch get room comment map",
		log.Any("result", result),
		log.Strings("room_ids", roomIDs),
	)
	return result, nil
}

func (m *assessmentBase) getH5PStudentViewItems(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, view *entity.AssessmentView) ([]*entity.AssessmentStudentViewH5PItem, error) {
	roomMap, err := m.batchGetRoomScoreMap(ctx, operator, []string{view.RoomID}, true)
	if err != nil {
		log.Error(ctx, "get assessment detail: get room map failed",
			log.Err(err),
			log.String("room_id", view.RoomID),
		)
		return nil, err
	}
	room := roomMap[view.RoomID]
	if room == nil {
		log.Debug(ctx, "add h5p assessment detail: not found room", log.String("room_id", view.RoomID))
		return nil, nil
	}

	// get outcome names map
	lessonMaterialIDs := make([]string, 0, len(view.LessonMaterials))
	for _, lp := range view.LessonMaterials {
		lessonMaterialIDs = append(lessonMaterialIDs, lp.ID)
	}
	var lpOutcomes []*entity.AssessmentContentOutcome
	if err := da.GetAssessmentContentOutcomeDA().Query(ctx, &da.QueryAssessmentContentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{view.ID},
			Valid:   true,
		},
		ContentIDs: entity.NullStrings{
			Strings: lessonMaterialIDs,
			Valid:   true,
		},
	}, &lpOutcomes); err != nil {
		log.Error(ctx, "get h5p student view items: query assessment content outcome failed",
			log.Err(err),
			log.Any("view", view),
			log.Strings("lesson_materials", lessonMaterialIDs),
		)
		return nil, err
	}
	outcomeIDs := make([]string, 0, len(lpOutcomes))
	for _, o := range lpOutcomes {
		outcomeIDs = append(outcomeIDs, o.OutcomeID)
	}
	outcomeIDs = utils.SliceDeduplicationExcludeEmpty(outcomeIDs)
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
	if err != nil {
		log.Error(ctx, "get h5p student view items: get outcomes failed by id",
			log.Err(err),
			log.Any("view", view),
			log.Strings("lesson_materials", lessonMaterialIDs),
		)
		return nil, err
	}
	outcomeNameMap := make(map[string]string, len(outcomes))
	for _, o := range outcomes {
		outcomeNameMap[o.ID] = o.Name
	}
	lpOutcomeNameMap := make(map[string][]string, len(lpOutcomes))
	for _, o := range lpOutcomes {
		lpOutcomeNameMap[o.ContentID] = append(lpOutcomeNameMap[o.ContentID], outcomeNameMap[o.OutcomeID])
	}

	r := make([]*entity.AssessmentStudentViewH5PItem, 0, len(view.Students))
	for _, s := range view.Students {
		newItem := entity.AssessmentStudentViewH5PItem{
			StudentID:   s.ID,
			StudentName: s.Name,
		}
		user := room.UserMap[s.ID]
		if user != nil {
			newItem.Comment = user.Comment
		} else {
			log.Debug(ctx, "get h5p assessment detail: not found user from h5p room",
				log.String("room_id", view.RoomID),
				log.Any("not_found_student_id", s.ID),
				log.Any("room", room),
			)
		}
		for _, lm := range view.LessonMaterials {
			newLessMaterial := entity.AssessmentStudentViewH5PLessonMaterial{
				LessonMaterialID:   lm.ID,
				LessonMaterialName: lm.Name,
				IsH5P:              lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend,
				OutcomeNames:       lpOutcomeNameMap[lm.ID],
			}
			var content *entity.AssessmentH5PContentScore
			if user != nil {
				content = user.ContentMap[lm.Source]
				if content != nil {
					newLessMaterial.LessonMaterialType = content.ContentType
					newLessMaterial.Answer = content.Answer
					newLessMaterial.MaxScore = content.MaxPossibleScore
					newLessMaterial.AchievedScore = content.AchievedScore
					newLessMaterial.Attempted = len(content.Answers) > 0 || len(content.Scores) > 0
				} else {
					log.Debug(ctx, "get h5p assessment detail: not found content from h5p room",
						log.String("room_id", view.RoomID),
						log.Any("not_found_content_id", lm.Source),
						log.Any("room", room),
					)
				}
			}
			newItem.LessonMaterials = append(newItem.LessonMaterials, &newLessMaterial)
		}
		r = append(r, &newItem)
	}

	sort.Sort(entity.AssessmentStudentViewH5PItemsOrder(r))

	return r, nil
}

func (m *assessmentBase) batchGetLatestLessonPlanMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, lessonPlanIDs []string) (map[string]*entity.AssessmentExternalLessonPlan, error) {
	lessonPlanIDs = utils.SliceDeduplication(lessonPlanIDs)

	lessonPlanIDs, err := GetContentModel().GetLatestContentIDByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "batchGetLatestLessonPlanMap: GetContentModel().GetLatestContentIDByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}
	lessonPlans, err := GetContentModel().GetContentByIDList(ctx, tx, lessonPlanIDs, operator)
	if err != nil {
		log.Error(ctx, "toViews: GetContentModel().GetContentByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}
	result := make(map[string]*entity.AssessmentExternalLessonPlan, len(lessonPlans))
	for _, lp := range lessonPlans {
		result[lp.ID] = &entity.AssessmentExternalLessonPlan{
			ID:         lp.ID,
			Name:       lp.Name,
			OutcomeIDs: lp.Outcomes,
		}
	}

	// fill lesson materials
	m2, err := GetContentModel().GetContentsSubContentsMapByIDList(ctx, dbo.MustGetDB(ctx), lessonPlanIDs, operator)
	if err != nil {
		log.Error(ctx, "List: GetContentModel().GetContentsSubContentsMapByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}
	for id, lp := range result {
		lms := m2[id]
		for _, lm := range lms {
			newMaterial := &entity.AssessmentExternalLessonMaterial{
				ID:   lm.ID,
				Name: lm.Name,
			}
			switch v := lm.Data.(type) {
			case *MaterialData:
				newMaterial.Source = string(v.Source)
			}
			lp.Materials = append(lp.Materials, newMaterial)
		}
	}

	// fill outcomes
	var lessonMaterialIDs []string
	for _, lp := range result {
		for _, lm := range lp.Materials {
			lessonMaterialIDs = append(lessonMaterialIDs, lm.ID)
		}
	}
	lessonMaterials, err := GetContentModel().GetRawContentByIDList(ctx, tx, lessonMaterialIDs)
	lessonMaterialMap := make(map[string]*entity.Content, len(lessonMaterials))
	for _, lm := range lessonMaterials {
		lessonMaterialMap[lm.ID] = lm
	}
	for _, lp := range result {
		lessonMaterialIDs = append(lessonMaterialIDs, lp.ID)
		for _, lm := range lp.Materials {
			lm2 := lessonMaterialMap[lm.ID]
			if lm2 != nil {
				lm.OutcomeIDs = strings.Split(lm2.Outcomes, constant.StringArraySeparator)
			}
		}
	}

	return result, nil
}

func (m *assessmentBase) batchGetLessonMaterialDataMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, ids []string) (map[string]*MaterialData, error) {
	lessonMaterials, err := GetContentModel().GetContentByIDList(ctx, tx, ids, operator)
	if err != nil {
		log.Error(ctx, "get lesson material source map: get contents faield",
			log.Err(err),
			log.Strings("ids", ids),
		)
		return nil, err
	}
	result := make(map[string]*MaterialData, len(lessonMaterials))
	for _, lm := range lessonMaterials {
		data, err := GetContentModel().CreateContentData(ctx, lm.ContentType, lm.Data)
		if err != nil {
			log.Error(ctx, "get lesson material source map: create content data failed",
				log.Err(err),
				log.Any("content", lm),
			)
			return nil, err
		}
		switch v := data.(type) {
		case *MaterialData:
			result[lm.ID] = v
		}
	}
	return result, nil
}

func (m *assessmentBase) existsAssessmentsByScheduleIDs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) (bool, error) {
	count, err := da.GetAssessmentDA().CountTx(ctx, tx, &da.QueryAssessmentConditions{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, entity.Assessment{})
	if err != nil {
		log.Error(ctx, "exists assessments by schedule ids: count failed",
			log.Err(err),
			log.Strings("schedule_id", scheduleIDs),
			log.Any("operator", operator),
		)
		return false, nil
	}
	if count > 0 {
		log.Info(ctx, "exists assessments by schedule ids: assessment already exists",
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return true, nil
	}
	return false, nil
}

func (m *assessmentBase) prepareBatchAddSuperArgs(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args []*entity.AddAssessmentArgs) (*entity.BatchAddAssessmentSuperArgs, error) {
	// get schedule ids
	var scheduleIDs []string
	for _, item := range args {
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}
	scheduleIDs = utils.SliceDeduplicationExcludeEmpty(scheduleIDs)

	// check if assessment already exits
	ok, err := m.existsAssessmentsByScheduleIDs(ctx, tx, operator, scheduleIDs)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errors.New("assessment already existed")
	}

	// get contents
	var lessonPlanIDs []string
	for _, item := range args {
		lessonPlanIDs = append(lessonPlanIDs, item.LessonPlanID)
	}
	lessonPlanMap, err := m.batchGetLatestLessonPlanMap(ctx, tx, operator, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "batch add assessments: batch get latest lesson plan map failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}

	// get outcomes
	var (
		outcomeIDs                []string
		scheduleIDToOutcomeIDsMap = make(map[string][]string, len(args))
	)
	for _, item := range args {
		var itemOutcomeIDs []string
		lp := lessonPlanMap[item.LessonPlanID]
		if lp == nil {
			continue
		}
		itemOutcomeIDs = append(itemOutcomeIDs, lp.OutcomeIDs...)
		for _, lm := range lp.Materials {
			itemOutcomeIDs = append(itemOutcomeIDs, lm.OutcomeIDs...)
		}
		scheduleIDToOutcomeIDsMap[item.ScheduleID] = itemOutcomeIDs
		outcomeIDs = append(outcomeIDs, itemOutcomeIDs...)
	}
	outcomes := make([]*entity.Outcome, 0, len(outcomeIDs))
	if len(outcomeIDs) > 0 {
		outcomeIDs = utils.SliceDeduplication(outcomeIDs)
		if outcomes, err = GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs); err != nil {
			log.Error(ctx, "batch add assessments: batch get outcomes failed",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
				log.Any("args", args),
				log.Any("operator", operator),
			)
			return nil, err
		}
	}
	outcomeMap := make(map[string]*entity.Outcome, len(outcomeIDs))
	for _, o := range outcomes {
		outcomeMap[o.ID] = o
	}

	return &entity.BatchAddAssessmentSuperArgs{
		Raw:                       args,
		ScheduleIDs:               scheduleIDs,
		Outcomes:                  outcomes,
		OutcomeMap:                outcomeMap,
		LessonPlanMap:             lessonPlanMap,
		ScheduleIDToOutcomeIDsMap: scheduleIDToOutcomeIDsMap,
	}, nil
}

func (m *assessmentBase) batchAdd(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.BatchAddAssessmentSuperArgs) ([]string, error) {
	log.Debug(ctx, "batch add assessments: print args", log.Any("args", args), log.Any("operator", operator))

	// add assessment
	newAssessments := make([]*entity.Assessment, 0, len(args.ScheduleIDs))
	now := time.Now().Unix()
	for _, item := range args.Raw {
		newAssessments = append(newAssessments, &entity.Assessment{
			ID:           utils.NewID(),
			ScheduleID:   item.ScheduleID,
			Title:        item.Title,
			Status:       entity.AssessmentStatusInProgress,
			CreateAt:     now,
			UpdateAt:     now,
			ClassLength:  item.ClassLength,
			ClassEndTime: item.ClassEndTime,
		})
	}
	if err := da.GetAssessmentDA().BatchInsert(ctx, tx, newAssessments); err != nil {
		log.Error(ctx, "batch add assessments: batch insert assessments failed",
			log.Err(err),
			log.Strings("schedule_ids", args.ScheduleIDs),
			log.Any("new_assessments", newAssessments),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// add attendances
	if err := m.batchAddAttendances(ctx, tx, newAssessments, args.Raw); err != nil {
		return nil, err
	}

	// parse args to map
	scheduleIDToArgsItemMap := make(map[string]*entity.AddAssessmentArgs, len(args.Raw))
	for _, item := range args.Raw {
		scheduleIDToArgsItemMap[item.ScheduleID] = item
	}

	// add contents
	if err := m.batchAddContents(ctx, tx, operator, newAssessments, scheduleIDToArgsItemMap, args.LessonPlanMap, args.ScheduleIDs); err != nil {
		return nil, err
	}

	// add assessment outcomes
	if err := m.batchAddOutcomes(ctx, tx, operator, args.Outcomes, newAssessments, args.ScheduleIDToOutcomeIDsMap, args.OutcomeMap); err != nil {
		return nil, err
	}

	// add outcome attendances
	if err := m.batchAddOutcomeAttendances(ctx, tx, operator, args.Raw, args.Outcomes, newAssessments, scheduleIDToArgsItemMap, args.ScheduleIDToOutcomeIDsMap, args.OutcomeMap); err != nil {
		return nil, err
	}

	// add assessment content outcomes
	if err := m.batchAddContentOutcomes(ctx, tx, operator, newAssessments, args.LessonPlanMap, scheduleIDToArgsItemMap); err != nil {
		return nil, err
	}

	// collect assessment ids
	var newAssessmentIDs []string
	for _, a := range newAssessments {
		newAssessmentIDs = append(newAssessmentIDs, a.ID)
	}

	return newAssessmentIDs, nil
}

func (m *assessmentBase) batchAddAttendances(ctx context.Context, tx *dbo.DBContext, newAssessments []*entity.Assessment, args []*entity.AddAssessmentArgs) error {
	scheduleIDToAssessmentIDMap := make(map[string]string, len(newAssessments))
	for _, item := range newAssessments {
		scheduleIDToAssessmentIDMap[item.ScheduleID] = item.ID
	}
	var attendances []*entity.AssessmentAttendance
	for _, item := range args {
		for _, attendance := range item.Attendances {
			newAttendance := entity.AssessmentAttendance{
				ID:           utils.NewID(),
				AssessmentID: scheduleIDToAssessmentIDMap[item.ScheduleID],
				AttendanceID: attendance.RelationID,
				Checked:      true,
			}
			switch attendance.RelationType {
			case entity.ScheduleRelationTypeClassRosterStudent:
				newAttendance.Origin = entity.AssessmentAttendanceOriginClassRoaster
				newAttendance.Role = entity.AssessmentAttendanceRoleStudent
			case entity.ScheduleRelationTypeClassRosterTeacher:
				newAttendance.Origin = entity.AssessmentAttendanceOriginClassRoaster
				newAttendance.Role = entity.AssessmentAttendanceRoleTeacher
			case entity.ScheduleRelationTypeParticipantStudent:
				newAttendance.Origin = entity.AssessmentAttendanceOriginParticipants
				newAttendance.Role = entity.AssessmentAttendanceRoleStudent
			case entity.ScheduleRelationTypeParticipantTeacher:
				newAttendance.Origin = entity.AssessmentAttendanceOriginParticipants
				newAttendance.Role = entity.AssessmentAttendanceRoleTeacher
			default:
				continue
			}
			attendances = append(attendances, &newAttendance)
		}
	}
	if err := da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, attendances); err != nil {
		log.Error(ctx, "batch add assessments: batch insert attendance failed",
			log.Err(err),
			log.Any("attendances", attendances),
			log.Any("args", args),
		)
		return err
	}
	return nil
}

func (m *assessmentBase) batchAddContents(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, newAssessments []*entity.Assessment, scheduleIDToArgsItemMap map[string]*entity.AddAssessmentArgs, lessonPlanMap map[string]*entity.AssessmentExternalLessonPlan, scheduleIDs []string) error {
	var assessmentContents []*entity.AssessmentContent
	assessmentContentKeys := map[[2]string]bool{}
	for _, a := range newAssessments {
		schedule := scheduleIDToArgsItemMap[a.ScheduleID]
		if schedule == nil {
			continue
		}
		lp := lessonPlanMap[schedule.LessonPlanID]
		if lp == nil {
			continue
		}
		assessmentContents = append(assessmentContents, &entity.AssessmentContent{
			ID:           utils.NewID(),
			AssessmentID: a.ID,
			ContentID:    lp.ID,
			ContentName:  lp.Name,
			ContentType:  entity.ContentTypePlan,
			Checked:      true,
		})
		for _, lm := range lp.Materials {
			key := [2]string{a.ID, lm.ID}
			if assessmentContentKeys[key] {
				continue
			}
			assessmentContentKeys[key] = true
			assessmentContents = append(assessmentContents, &entity.AssessmentContent{
				ID:           utils.NewID(),
				AssessmentID: a.ID,
				ContentID:    lm.ID,
				ContentName:  lm.Name,
				ContentType:  entity.ContentTypeMaterial,
				Checked:      true,
			})
		}
	}
	if err := da.GetAssessmentContentDA().BatchInsert(ctx, tx, assessmentContents); err != nil {
		log.Error(ctx, "batch add assessments: batch insert assessment content failed",
			log.Err(err),
			log.Any("schedule_ids", scheduleIDs),
			log.Any("assessment_contents", assessmentContents),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (m *assessmentBase) batchAddOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, outcomes []*entity.Outcome, newAssessments []*entity.Assessment, scheduleIDToOutcomeIDsMap map[string][]string, outcomeMap map[string]*entity.Outcome) error {
	if len(outcomes) > 0 {
		var assessmentOutcomes []*entity.AssessmentOutcome
		for _, a := range newAssessments {
			outcomeIDs := scheduleIDToOutcomeIDsMap[a.ScheduleID]
			for _, outcomeID := range outcomeIDs {
				o := outcomeMap[outcomeID]
				assumed := false
				if o != nil {
					assumed = o.Assumed
				}
				assessmentOutcomes = append(assessmentOutcomes, &entity.AssessmentOutcome{
					ID:           utils.NewID(),
					AssessmentID: a.ID,
					OutcomeID:    outcomeID,
					Skip:         false,
					NoneAchieved: assumed,
					Checked:      true,
				})
			}
		}
		if len(assessmentOutcomes) > 0 {
			if err := da.GetAssessmentOutcomeDA().BatchInsert(ctx, tx, assessmentOutcomes); err != nil {
				log.Error(ctx, "batch add assessments: batch insert assessment outcome failed",
					log.Err(err),
					log.Any("assessment_outcomes", assessmentOutcomes),
					log.Any("outcomes", outcomes),
					log.Any("operator", operator),
				)
				return err
			}
		}
	}
	return nil
}

func (m *assessmentBase) batchAddOutcomeAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args []*entity.AddAssessmentArgs, outcomes []*entity.Outcome, newAssessments []*entity.Assessment, scheduleIDToArgsItemMap map[string]*entity.AddAssessmentArgs, scheduleIDToOutcomeIDsMap map[string][]string, outcomeMap map[string]*entity.Outcome) error {
	if len(outcomes) > 0 {
		var outcomeAttendances []*entity.OutcomeAttendance
		for _, a := range newAssessments {
			argsItem := scheduleIDToArgsItemMap[a.ScheduleID]
			if argsItem == nil {
				continue
			}
			studentIDs := make([]string, 0, len(argsItem.Attendances))
			for _, a2 := range argsItem.Attendances {
				if a2.RelationType == entity.ScheduleRelationTypeClassRosterStudent ||
					a2.RelationType == entity.ScheduleRelationTypeParticipantStudent {
					studentIDs = append(studentIDs, a2.RelationID)
				}
			}
			outcomeIDs := scheduleIDToOutcomeIDsMap[a.ScheduleID]
			for _, outcomeID := range outcomeIDs {
				outcome := outcomeMap[outcomeID]
				if outcome == nil {
					continue
				}
				if !outcome.Assumed {
					continue
				}
				for _, sid := range studentIDs {
					outcomeAttendances = append(outcomeAttendances, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: a.ID,
						OutcomeID:    outcome.ID,
						AttendanceID: sid,
					})
				}
			}
		}
		if len(outcomeAttendances) > 0 {
			if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, outcomeAttendances); err != nil {
				log.Error(ctx, "batch add assessments: batch insert outcome attendances failed",
					log.Err(err),
					log.Any("args", args),
					log.Any("outcome_attendances", outcomeAttendances),
					log.Any("operator", operator),
				)
				return err
			}
		}
	}
	return nil
}

func (m *assessmentBase) batchAddContentOutcomes(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, newAssessments []*entity.Assessment, lessonPlanMap map[string]*entity.AssessmentExternalLessonPlan, scheduleIDToArgsItemMap map[string]*entity.AddAssessmentArgs) error {
	var assessmentContentOutcomes []*entity.AssessmentContentOutcome
	for _, a := range newAssessments {
		argsItem := scheduleIDToArgsItemMap[a.ScheduleID]
		if argsItem == nil {
			continue
		}
		lp := lessonPlanMap[argsItem.LessonPlanID]
		if lp == nil {
			continue
		}
		for _, oid := range lp.OutcomeIDs {
			assessmentContentOutcomes = append(assessmentContentOutcomes, &entity.AssessmentContentOutcome{
				ID:           utils.NewID(),
				AssessmentID: a.ID,
				ContentID:    lp.ID,
				OutcomeID:    oid,
			})
		}
		for _, lm := range lp.Materials {
			for _, oid := range lm.OutcomeIDs {
				assessmentContentOutcomes = append(assessmentContentOutcomes, &entity.AssessmentContentOutcome{
					ID:           utils.NewID(),
					AssessmentID: a.ID,
					ContentID:    lm.ID,
					OutcomeID:    oid,
				})
			}
		}
	}
	if len(assessmentContentOutcomes) > 0 {
		if err := da.GetAssessmentContentOutcomeDA().BatchInsert(ctx, tx, assessmentContentOutcomes); err != nil {
			log.Error(ctx, "batch add assessments: batch insert assessment content outcomes failed",
				log.Err(err),
				log.Any("assessment_content_outcomes", assessmentContentOutcomes),
				log.Any("operator", operator),
			)
			return err
		}
	}
	return nil
}

func (m *assessmentBase) update(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args *entity.UpdateAssessmentArgs) error {
	// validate args
	if !args.Action.Valid() {
		log.Error(ctx, "update assessment: invalid action", log.Any("args", args))
		return constant.ErrInvalidArgs
	}
	if args.Outcomes != nil {
		for _, item := range args.Outcomes {
			if item.Skip && item.NoneAchieved {
				log.Error(ctx, "update assessment: check skip and none achieved combination", log.Any("args", args))
				return constant.ErrInvalidArgs
			}
			if (item.Skip || item.NoneAchieved) && len(item.AttendanceIDs) > 0 {
				log.Error(ctx, "update assessment: check skip and none achieved combination with attendance ids", log.Any("args", args))
				return constant.ErrInvalidArgs
			}
		}
	}

	// check assessment status
	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, dbo.MustGetDB(ctx), args.ID)
	if err != nil {
		log.Error(ctx, "update assessment: get assessment exclude soft deleted failed",
			log.Err(err),
			log.Any("args", args),
		)
		return err
	}
	if assessment.Status == entity.AssessmentStatusComplete {
		log.Info(ctx, "update assessment: assessment has completed, not allow update",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return ErrAssessmentHasCompleted
	}

	// permission check
	if err := m.checkEditPermission(ctx, operator, args.ID); err != nil {
		return err
	}

	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// update assessment attendances
		if args.StudentIDs != nil {
			if err := da.GetAssessmentAttendanceDA().UncheckStudents(ctx, tx, args.ID); err != nil {
				log.Error(ctx, "update: da.GetAssessmentAttendanceDA().UncheckStudents: uncheck failed",
					log.Err(err),
					log.Any("args", args),
				)
				return err
			}
			if args.StudentIDs != nil && len(args.StudentIDs) > 0 {
				if err := da.GetAssessmentAttendanceDA().BatchCheck(ctx, tx, args.ID, args.StudentIDs); err != nil {
					log.Error(ctx, "update: da.GetAssessmentAttendanceDA().BatchCheck: check failed",
						log.Err(err),
						log.Any("args", args),
					)
					return err
				}
			}
		}

		if args.Outcomes != nil {
			// update assessment outcomes
			if err := da.GetAssessmentOutcomeDA().UncheckByAssessmentID(ctx, tx, args.ID); err != nil {
				log.Error(ctx, "Update: da.GetAssessmentOutcomeDA().UncheckStudents: uncheck assessment outcome failed by assessment id",
					log.Err(err),
					log.Any("args", args),
					log.String("id", args.ID),
				)
				return err
			}
			for _, oa := range args.Outcomes {
				newAssessmentOutcome := entity.AssessmentOutcome{
					AssessmentID: args.ID,
					OutcomeID:    oa.OutcomeID,
					Skip:         oa.Skip,
					NoneAchieved: oa.NoneAchieved,
					Checked:      true,
				}
				if err := da.GetAssessmentOutcomeDA().UpdateByAssessmentIDAndOutcomeID(ctx, tx, &newAssessmentOutcome); err != nil {
					log.Error(ctx, "update assessment: batch update assessment outcome failed",
						log.Err(err),
						log.Any("new_assessment_outcome", newAssessmentOutcome),
						log.Any("args", args),
						log.String("assessment_id", args.ID),
					)
					return err
				}
			}

			// update outcome attendances
			var (
				outcomeIDs         []string
				outcomeAttendances []*entity.OutcomeAttendance
			)
			for _, oa := range args.Outcomes {
				outcomeIDs = append(outcomeIDs, oa.OutcomeID)
				if oa.Skip {
					continue
				}
				for _, attendanceID := range oa.AttendanceIDs {
					outcomeAttendances = append(outcomeAttendances, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: args.ID,
						OutcomeID:    oa.OutcomeID,
						AttendanceID: attendanceID,
					})
				}
			}
			if err := da.GetOutcomeAttendanceDA().BatchDeleteByAssessmentIDAndOutcomeIDs(ctx, tx, args.ID, outcomeIDs); err != nil {
				log.Error(ctx, "update assessment: batch delete outcome attendance map failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
					log.Any("args", args),
				)
				return err
			}
			if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, outcomeAttendances); err != nil {
				log.Error(ctx, "update assessment: batch insert outcome attendance map failed",
					log.Err(err),
					log.Any("outcome_attendances", outcomeAttendances),
					log.Any("args", args),
				)
				return err
			}
		}

		// update assessment contents
		for _, ma := range args.LessonMaterials {
			updateArgs := da.UpdatePartialAssessmentContentArgs{
				AssessmentID:   args.ID,
				ContentID:      ma.ID,
				ContentComment: ma.Comment,
				Checked:        ma.Checked,
			}
			if err = da.GetAssessmentContentDA().UpdatePartial(ctx, tx, &updateArgs); err != nil {
				log.Error(ctx, "Update: da.GetAssessmentContentDA().UpdatePartial: update failed",
					log.Err(err),
					log.Any("args", args),
					log.Any("update_args", updateArgs),
					log.Any("operator", operator),
				)
				return err
			}
		}

		// get schedule
		schedules, err := GetScheduleModel().GetVariableDataByIDs(ctx, operator, []string{assessment.ScheduleID}, nil)
		if err != nil {
			log.Error(ctx, "update class and live assessment: get plain schedule failed",
				log.Err(err),
				log.String("schedule_id", assessment.ScheduleID),
				log.Any("args", args),
			)
			return err
		}
		if len(schedules) == 0 {
			errMsg := "update class and live assessment: not found schedule"
			log.Error(ctx, errMsg,
				log.String("schedule_id", assessment.ScheduleID),
				log.Any("args", args),
			)
			return errors.New(errMsg)
		}
		schedule := schedules[0]

		// set scores and comments
		if schedule.ClassType != entity.ScheduleClassTypeOfflineClass {
			if err := m.updateStudentViewItems(ctx, tx, operator, schedule.RoomID, args.StudentViewItems); err != nil {
				log.Error(ctx, "update assessment: update student view items failed",
					log.Err(err),
					log.Any("args", args),
					log.Any("schedule", schedule),
					log.Any("operator", operator),
				)
				return err
			}
		}

		// update assessment status
		if args.Action == entity.UpdateAssessmentActionComplete {
			if err := da.GetAssessmentDA().UpdateStatus(ctx, tx, args.ID, entity.AssessmentStatusComplete); err != nil {
				log.Error(ctx, "Update: da.GetAssessmentDA().UpdateStatus: update failed",
					log.Err(err),
					log.Any("args", args),
					log.Any("operator", operator),
				)
				return err
			}
		}

		return nil
	}); err != nil {
		log.Error(ctx, "Update: tx failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return err
	}

	return nil
}

func (m *assessmentBase) updateStudentViewItems(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, roomID string, items []*entity.UpdateAssessmentH5PStudent) error {
	// set scores
	var lmIDs []string
	for _, item := range items {
		for _, lm := range item.LessonMaterials {
			lmIDs = append(lmIDs, lm.LessonMaterialID)
		}
	}
	lms, err := GetContentModel().GetRawContentByIDList(ctx, tx, lmIDs)
	if err != nil {
		log.Error(ctx, "update assessment: batch get contents failed",
			log.Err(err),
			log.Any("items", items),
			log.Strings("lm_ids", lmIDs),
		)
		return err
	}
	lmDataMap := make(map[string]*MaterialData, len(lms))
	for _, lm := range lms {
		data, err := GetContentModel().CreateContentData(ctx, lm.ContentType, lm.Data)
		if err != nil {
			return err
		}
		lmData, ok := data.(*MaterialData)
		if ok {
			lmDataMap[lm.ID] = lmData
		}
	}
	var newScores []*external.H5PSetScoreRequest
	for _, item := range items {
		for _, lm := range item.LessonMaterials {
			lmData := lmDataMap[lm.LessonMaterialID]
			if lmData == nil {
				log.Debug(ctx, "not found lesson material id in data map",
					log.String("lesson_material_id", lm.LessonMaterialID),
				)
				continue
			}
			if lmData.FileType != entity.FileTypeH5p && lmData.FileType != entity.FileTypeH5pExtend {
				continue
			}
			if lmData.Source.IsNil() {
				log.Debug(ctx, "lesson material source is nil",
					log.String("lesson_material_id", lm.LessonMaterialID),
					log.Any("data", lmData),
				)
				continue
			}
			newScore := external.H5PSetScoreRequest{
				RoomID:    roomID,
				ContentID: string(lmData.Source),
				StudentID: item.StudentID,
				Score:     lm.AchievedScore,
			}
			newScores = append(newScores, &newScore)
		}
	}
	if _, err := external.GetH5PRoomScoreServiceProvider().BatchSet(ctx, operator, newScores); err != nil {
		return err
	}

	// set comments
	var newComments []*external.H5PAddRoomCommentRequest
	for _, item := range items {
		newComment := external.H5PAddRoomCommentRequest{
			RoomID:    roomID,
			StudentID: item.StudentID,
			Comment:   item.Comment,
		}
		newComments = append(newComments, &newComment)
	}
	if _, err := external.GetH5PRoomCommentServiceProvider().BatchAdd(ctx, operator, newComments); err != nil {
		return err
	}

	return nil
}

func (m *assessmentBase) getSortedLessonMaterialIDsMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, lessonPlanIDs []string) (map[string][]string, error) {
	if len(lessonPlanIDs) == 0 {
		return map[string][]string{}, nil
	}
	contentMap, err := GetContentModel().GetContentsSubContentsMapByIDList(ctx, tx, lessonPlanIDs, operator)
	if err != nil {
		log.Error(ctx, "get sorted content ids: get content map failed",
			log.Err(err),
			log.Strings("ids", lessonPlanIDs),
		)
		return nil, err
	}
	r := make(map[string][]string, len(contentMap))
	for aid, cc := range contentMap {
		for _, c := range cc {
			r[aid] = append(r[aid], c.ID)
		}
	}
	return r, nil
}

func (m *assessmentBase) sortedByLessonMaterialIDs(items []*entity.AssessmentLessonMaterial, lessonMaterialIDs []string) {
	if len(items) == 0 || len(lessonMaterialIDs) == 0 {
		return
	}
	idMap := make(map[string]int, len(lessonMaterialIDs))
	for i, id := range lessonMaterialIDs {
		idMap[id] = i + 1
	}
	sort.Slice(items, func(i, j int) bool {
		idI := idMap[items[i].ID]
		idJ := idMap[items[j].ID]
		if idI == 0 {
			return false
		}
		if idJ == 0 {
			return true
		}
		return idI < idJ
	})
}
