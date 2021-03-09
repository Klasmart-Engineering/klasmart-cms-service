package model

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type IAssessmentModel interface {
	GetAssessment(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetailView, error)
	ListAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, cmd entity.ListAssessmentsQuery) (*entity.ListAssessmentsResult, error)
	AddAssessment(ctx context.Context, operator *entity.Operator, cmd entity.AddAssessmentCommand) (string, error)
	UpdateAssessment(ctx context.Context, operator *entity.Operator, cmd entity.UpdateAssessmentCommand) error
}

var (
	assessmentModelInstance     IAssessmentModel
	assessmentModelInstanceOnce = sync.Once{}
)

func GetAssessmentModel() IAssessmentModel {
	assessmentModelInstanceOnce.Do(func() {
		assessmentModelInstance = &assessmentModel{}
	})
	return assessmentModelInstance
}

type assessmentModel struct{}

type outcomeSliceSortByAssumedAndName []*entity.Outcome

func (s outcomeSliceSortByAssumedAndName) Len() int {
	return len(s)
}

func (s outcomeSliceSortByAssumedAndName) Less(i, j int) bool {
	if s[i].Assumed && !s[j].Assumed {
		return true
	} else if !s[i].Assumed && s[j].Assumed {
		return false
	} else {
		return s[i].Name < s[j].Name
	}
}

func (s outcomeSliceSortByAssumedAndName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (a *assessmentModel) GetAssessment(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, id string) (*entity.AssessmentDetailView, error) {
	var result entity.AssessmentDetailView

	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "get assessment detail: get from da failed",
			log.Err(err),
			log.String("id", "id"),
		)
		return nil, err
	}
	result = entity.AssessmentDetailView{
		ID:           assessment.ID,
		Title:        assessment.Title,
		ClassEndTime: assessment.ClassEndTime,
		ClassLength:  assessment.ClassLength,
		CompleteTime: assessment.CompleteTime,
		Status:       assessment.Status,
	}

	// fill attendances
	{
		var assessmentAttendances []*entity.AssessmentAttendance
		if err := da.GetAssessmentAttendanceDA().QueryTx(ctx, tx, &da.AssessmentAttendanceCondition{
			AssessmentIDs: []string{id},
			Checked:       nil,
		}, &assessmentAttendances); err != nil {
			log.Error(ctx, "get assessment detail: query assessment ids",
				log.Err(err),
				log.String("id", "id"),
			)
			return nil, err
		}
		var attendanceIDs []string
		for _, item := range assessmentAttendances {
			attendanceIDs = append(attendanceIDs, item.AttendanceID)
		}

		studentService := external.GetStudentServiceProvider()
		students, err := studentService.BatchGet(ctx, operator, attendanceIDs)
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get student failed",
				log.Err(err),
				log.String("id", id),
				log.Strings("attendance_ids", attendanceIDs),
			)
			return nil, err
		}
		studentMap := map[string]string{}
		for _, student := range students {
			studentMap[student.ID] = student.Name
		}

		for _, item := range assessmentAttendances {
			result.Attendances = append(result.Attendances, &entity.AssessmentAttendanceView{
				ID:      item.AttendanceID,
				Name:    studentMap[item.AttendanceID],
				Checked: item.Checked,
			})
		}
	}

	// fill subject
	{
		nameMap, err := a.getSubjectNameMap(ctx, []string{assessment.SubjectID})
		if err != nil {
			log.Error(ctx, "get assessment detail: get subject name map failed",
				log.Err(err),
				log.String("id", id),
				log.String("subject_id", assessment.SubjectID),
			)
			return nil, err
		}
		result.Subject = entity.AssessmentSubject{
			ID:   assessment.SubjectID,
			Name: nameMap[assessment.SubjectID],
		}
	}

	// fill teacher
	{
		teacherIDs, err := assessment.DecodeTeacherIDs()
		if err != nil {
			log.Error(ctx, "get assessment detail: decode teacher ids failed",
				log.Err(err),
				log.String("id", id),
				log.Any("assessment", assessment),
			)
			return nil, err
		}
		teacherNameService := external.GetTeacherServiceProvider()
		items, err := teacherNameService.BatchGet(ctx, operator, teacherIDs)
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get teacher failed",
				log.Err(err),
				log.Strings("teacher_ids", teacherIDs),
				log.Any("id", id),
			)
			return nil, err
		}
		for _, item := range items {
			result.Teachers = append(result.Teachers, &entity.AssessmentTeacher{
				ID:   item.ID,
				Name: item.Name,
			})
		}
	}

	// fill number of outcomes and activities
	{
		schedule, err := GetScheduleModel().GetPlainByID(ctx, assessment.ScheduleID)
		if err != nil {
			log.Error(ctx, "add assessment: get schedule failed by id",
				log.Err(err),
				log.Any("id", id),
				log.Any("schedule_id", assessment.ScheduleID),
			)
			return nil, err
		}
		counts, err := GetContentModel().ContentDataCount(ctx, tx, schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "add assessment: get number of activities and outcomes failed",
				log.Err(err),
				log.Any("id", id),
				log.Any("lesson_plan_id", schedule.LessonPlanID),
			)
		}
		result.NumberOfActivities = counts.SubContentCount
		//result.NumberOfOutcomes = counts.OutcomesCount
	}

	// fill outcome attendance maps
	{
		outcomeMap := map[string]*entity.Outcome{}
		outcomeIDs, err := da.GetAssessmentOutcomeDA().GetOutcomeIDsByAssessmentID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "get assessment detail: get outcome ids failed by assessment id",
				log.Err(err),
				log.String("id", id),
			)
			return nil, err
		}
		result.NumberOfOutcomes = len(outcomeIDs)
		if len(outcomeIDs) > 0 {
			outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, &entity.Operator{})
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcomes failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			for _, outcome := range outcomes {
				outcomeMap[outcome.ID] = outcome
			}
			sort.Sort(outcomeSliceSortByAssumedAndName(outcomes))
			outcomeAttendanceItems, err := da.GetOutcomeAttendanceDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, tx, id, outcomeIDs)
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get outcome attendances failed by assessment id and outcome ids",
					log.Err(err),
					log.String("id", id),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			outcomeAttendanceMap := map[string][]string{}
			for _, item := range outcomeAttendanceItems {
				outcomeAttendanceMap[item.OutcomeID] = append(outcomeAttendanceMap[item.OutcomeID], item.AttendanceID)
			}
			assessmentOutcomeItems, err := da.GetAssessmentOutcomeDA().BatchGetByAssessmentIDAndOutcomeIDs(ctx, tx, id, outcomeIDs)
			if err != nil {
				log.Error(ctx, "get assessment detail: batch get assessment outcomes failed by assessment id and outcome ids",
					log.Err(err),
					log.String("id", id),
					log.Strings("outcome_ids", outcomeIDs),
				)
				return nil, err
			}
			assessmentOutcomeMap := map[string]entity.AssessmentOutcome{}
			for _, item := range assessmentOutcomeItems {
				assessmentOutcomeMap[item.OutcomeID] = *item
			}
			for _, outcome := range outcomes {
				newItem := entity.OutcomeAttendanceMapView{
					OutcomeID:     outcome.ID,
					OutcomeName:   outcome.Name,
					Assumed:       outcome.Assumed,
					AttendanceIDs: outcomeAttendanceMap[outcome.ID],
					Skip:          assessmentOutcomeMap[outcome.ID].Skip,
					NoneAchieved:  assessmentOutcomeMap[outcome.ID].NoneAchieved,
				}
				result.OutcomeAttendanceMaps = append(result.OutcomeAttendanceMaps, newItem)
			}
		}
	}
	return &result, nil
}

func (a *assessmentModel) ListAssessments(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, cmd entity.ListAssessmentsQuery) (*entity.ListAssessmentsResult, error) {
	filter := NewAssessmentTeacherAndStatusFilter(operator)
	allowed, err := filter.FilterAll(ctx)
	if err != nil {
		log.Error(ctx, "list assessments: check and filter failed by permissions",
			log.Any("operator", operator),
			log.Any("cmd", cmd),
		)
		return nil, err
	}
	if !allowed {
		return nil, constant.ErrForbidden
	}
	if !filter.CheckStatus(cmd.Status) {
		return nil, constant.ErrForbidden
	}
	cmd.TeacherIDs = append(cmd.TeacherIDs, filter.AllowTeacherIDs()...)
	cmd.AssessmentTeacherAndStatusPairs = append(cmd.AssessmentTeacherAndStatusPairs, filter.AllowPairs()...)

	cond := da.QueryAssessmentsCondition{
		Status:                          cmd.Status,
		TeacherIDs:                      cmd.TeacherIDs,
		AssessmentTeacherAndStatusPairs: cmd.AssessmentTeacherAndStatusPairs,
		OrderBy:                         cmd.OrderBy,
		Page:                            cmd.Page,
		PageSize:                        cmd.PageSize,
	}
	{
		if cmd.TeacherName != nil {
			teacherService := external.GetTeacherServiceProvider()
			items, err := teacherService.Query(ctx, operator, operator.OrgID, *cmd.TeacherName)
			if err != nil {
				log.Error(ctx, "list assessments: query teachers by name failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return nil, err
			}
			log.Debug(ctx, "list assessments: query teachers by name success",
				log.Any("cmd", cmd),
				log.Any("items", items),
			)
			if len(items) > 0 {
				for _, item := range items {
					cond.TeacherIDs = append(cond.TeacherIDs, item.ID)
				}
			} else {
				cond.TeacherIDs = []string{}
			}
		}
		scheduleIDs, err := GetScheduleModel().GetScheduleIDsByOrgID(ctx, tx, operator, operator.OrgID)
		if err != nil {
			log.Error(ctx, "list assessments: get schedule ids failed by org id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("operator", operator),
			)
			return nil, err
		}
		if cond.ScheduleIDs == nil {
			cond.ScheduleIDs = []string{}
		}
		cond.ScheduleIDs = append(cond.ScheduleIDs, scheduleIDs...)
	}

	var items []*entity.Assessment
	if err := da.GetAssessmentDA().QueryTx(ctx, tx, &cond, &items); err != nil {
		log.Error(ctx, "list assessments: query tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("cond", cond),
		)
		return nil, err
	}

	total, err := da.GetAssessmentDA().CountTx(ctx, tx, &cond, &entity.Assessment{})
	if err != nil {
		log.Error(ctx, "list assessments: count tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("cond", cond),
		)
		return nil, err
	}

	var subjectIDs, programIDs, teacherIDs []string

	for _, item := range items {
		subjectIDs = append(subjectIDs, item.SubjectID)
		programIDs = append(programIDs, item.ProgramID)
		itemTeacherIDs, err := item.DecodeTeacherIDs()
		if err != nil {
			log.Error(ctx, "list assessment: decode teacher ids failed")
			return nil, err
		}
		teacherIDs = append(teacherIDs, itemTeacherIDs...)
	}

	subjectNameMap, err := a.getSubjectNameMap(ctx, subjectIDs)
	if err != nil {
		log.Error(ctx, "detail: get subject name map failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
		)
		return nil, err
	}

	programNameMap, err := a.getProgramNameMap(ctx, programIDs)
	if err != nil {
		log.Error(ctx, "detail: get program name map failed",
			log.Err(err),
			log.Strings("program_ids", programIDs),
		)
		return nil, err
	}

	teacherNameMap, err := a.getTeacherNameMap(ctx, operator, teacherIDs)
	if err != nil {
		log.Error(ctx, "detail: get teacher name map failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}

	result := entity.ListAssessmentsResult{Total: total}
	for _, item := range items {
		newItem := entity.AssessmentListView{
			ID:    item.ID,
			Title: item.Title,
			Subject: entity.AssessmentSubject{
				ID:   item.SubjectID,
				Name: subjectNameMap[item.SubjectID],
			},
			Program: entity.AssessmentProgram{
				ID:   item.ProgramID,
				Name: programNameMap[item.ProgramID],
			},
			ClassEndTime: item.ClassEndTime,
			CompleteTime: item.CompleteTime,
			Status:       item.Status,
		}
		teacherIDs, err := item.DecodeTeacherIDs()
		if err != nil {
			log.Error(ctx, "list assessment: decode teacher ids failed",
				log.Err(err),
				log.Any("item", item),
				log.Any("cmd", cmd),
			)
			return nil, err
		}
		for _, teacherID := range teacherIDs {
			newItem.Teachers = append(newItem.Teachers, entity.AssessmentTeacher{
				ID:   teacherID,
				Name: teacherNameMap[teacherID],
			})
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, err
}

func (a *assessmentModel) getProgramNameMap(ctx context.Context, programIDs []string) (map[string]string, error) {
	programNameMap := map[string]string{}
	items, err := GetProgramModel().Query(ctx, &da.ProgramCondition{
		IDs: entity.NullStrings{
			Strings: programIDs,
			Valid:   len(programIDs) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "list assessments: batch get program failed",
			log.Err(err),
			log.Strings("program_ids", programIDs),
		)
		return nil, err
	}
	for _, item := range items {
		programNameMap[item.ID] = item.Name
	}
	return programNameMap, nil
}

func (a *assessmentModel) getSubjectNameMap(ctx context.Context, subjectIDs []string) (map[string]string, error) {
	subjectNameMap := map[string]string{}
	items, err := GetSubjectModel().Query(ctx, &da.SubjectCondition{
		IDs: entity.NullStrings{
			Strings: subjectIDs,
			Valid:   len(subjectIDs) != 0,
		},
	})
	if err != nil {
		log.Error(ctx, "list assessments: batch get subject failed",
			log.Err(err),
			log.Strings("subject_ids", subjectIDs),
		)
		return nil, err
	}
	for _, item := range items {
		subjectNameMap[item.ID] = item.Name
	}
	return subjectNameMap, nil
}

func (a *assessmentModel) getTeacherNameMap(ctx context.Context, operator *entity.Operator, teacherIDs []string) (map[string]string, error) {
	teacherNameMap := map[string]string{}
	teacherNameService := external.GetTeacherServiceProvider()
	items, err := teacherNameService.BatchGet(ctx, operator, teacherIDs)
	if err != nil {
		log.Error(ctx, "list assessments: batch get teacher failed",
			log.Err(err),
			log.Strings("teacher_ids", teacherIDs),
		)
		return nil, err
	}
	for _, item := range items {
		teacherNameMap[item.ID] = item.Name
	}
	return teacherNameMap, nil
}

func (a *assessmentModel) getClassNameMap(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string]string, error) {
	classNameMap := map[string]string{}
	classService := external.GetClassServiceProvider()
	items, err := classService.BatchGet(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "get class name map: batch get class failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
		)
		return nil, err
	}
	for i, item := range items {
		if item.Valid {
			classNameMap[item.ID] = item.Name
		} else {
			log.Warn(ctx, "invalid item", log.Strings("class_ids", classIDs), log.Int("index", i))
		}
	}
	return classNameMap, nil
}

func (a *assessmentModel) AddAssessment(ctx context.Context, operator *entity.Operator, cmd entity.AddAssessmentCommand) (string, error) {
	var (
		outcomeIDs []string
		schedule   *entity.SchedulePlain
	)
	{
		var err error
		schedule, err = GetScheduleModel().GetPlainByID(ctx, cmd.ScheduleID)
		if err != nil {
			log.Error(ctx, "add assessment: get schedule failed by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("schedule_id", cmd.ScheduleID),
			)
			switch err {
			case constant.ErrRecordNotFound, dbo.ErrRecordNotFound:
				return "", constant.ErrInvalidArgs
			default:
				return "", err
			}
		}
		var assessments []entity.Assessment
		if err = da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentsCondition{
			ScheduleIDs: []string{cmd.ScheduleID},
		}, &assessments); err != nil {
			log.Error(ctx, "add assessment: query assessment by schedule ids failed",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("operator", operator),
				log.Any("schedule", schedule),
			)
			return "", err
		}
		if len(assessments) > 0 {
			log.Info(ctx, "add assessment: schedule has created assessment",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("operator", operator),
				log.Any("schedule", schedule),
			)
			return "", nil
		}
		//if schedule.Status == entity.ScheduleStatusClosed {
		//	log.Info(ctx, "add assessment: schedule status closed",
		//		log.Any("cmd", cmd),
		//		log.Any("operator", operator),
		//		log.Any("schedule", schedule),
		//	)
		//	return "", nil
		//}
		if schedule.ClassType == entity.ScheduleClassTypeHomework || schedule.ClassType == entity.ScheduleClassTypeTask {
			log.Info(ctx, "add assessment: invalid class type",
				log.Any("cmd", cmd),
				log.Any("operator", operator),
				log.Any("schedule", schedule),
			)
			return "", nil
		}
		outcomeIDs, err = GetContentModel().GetVisibleContentOutcomeByID(ctx, dbo.MustGetDB(ctx), schedule.LessonPlanID)
		if err != nil {
			log.Error(ctx, "add assessment: get outcome failed by id",
				log.Err(err),
				log.Any("cmd", cmd),
				log.Any("lesson_plan_id", schedule.LessonPlanID),
			)
			return "", err
		}
	}

	var newID = utils.NewID()
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		return a.addTx(ctx, operator, tx, cmd, newID, outcomeIDs, schedule)
	}); err != nil {
		return "", err
	}
	return newID, nil
}

func (a *assessmentModel) addTx(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, cmd entity.AddAssessmentCommand, newID string, outcomeIDs []string, schedule *entity.SchedulePlain) error {
	nowUnix := time.Now().Unix()
	newItem := entity.Assessment{
		ID:           newID,
		ScheduleID:   cmd.ScheduleID,
		ProgramID:    schedule.ProgramID,
		SubjectID:    schedule.SubjectID,
		ClassLength:  cmd.ClassLength,
		ClassEndTime: cmd.ClassEndTime,
		CreateAt:     nowUnix,
		UpdateAt:     nowUnix,
	}
	classNameMap, err := a.getClassNameMap(ctx, operator, []string{schedule.ClassID})
	if err != nil {
		log.Error(ctx, "add assessment: get class name map failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		return err
	}
	newItem.Title = a.title(newItem.ClassEndTime, classNameMap[schedule.ClassID], schedule.Title)

	var teacherIDs []string
	{
		if schedule.ClassID != "" {
			classID2TeachersMap, err := external.GetTeacherServiceProvider().GetByClasses(ctx, operator, []string{schedule.ClassID})
			if err != nil {
				log.Error(ctx, "add assessment: get class teachers failed",
					log.Err(err),
					log.Any("schedule", schedule),
					log.Any("cmd", cmd),
				)
				return err
			}
			teachers := classID2TeachersMap[schedule.ClassID]
			teacherIDs = make([]string, 0, len(teachers))
			for _, teacher := range teachers {
				teacherIDs = append(teacherIDs, teacher.ID)
			}
		}
	}

	// fill teacher ids
	if err := newItem.EncodeAndSetTeacherIDs(teacherIDs); err != nil {
		log.Error(ctx, "add assessment: encode and set teacher ids failed",
			log.Err(err),
			log.Any("schedule", schedule),
			log.Any("cmd", cmd),
		)
		return err
	}

	var studentIDs []string
	{
		if schedule.ClassID != "" {
			students, err := external.GetStudentServiceProvider().GetByClassID(ctx, operator, schedule.ClassID)
			if err != nil {
				log.Error(ctx, "add assessment: get students by class id failed",
					log.Err(err),
					log.Any("schedule", schedule),
					log.Any("cmd", cmd),
				)
				return err
			}
			studentIDs = make([]string, 0, len(students))
			for _, student := range students {
				studentIDs = append(studentIDs, student.ID)
			}
		}
	}

	// filter attendance ids
	if schedule.ClassType == entity.ScheduleClassTypeOfflineClass {
		cmd.AttendanceIDs = studentIDs
	} else {
		cmd.AttendanceIDs = utils.FilterStrings(cmd.AttendanceIDs, studentIDs, teacherIDs)
	}

	if len(outcomeIDs) == 0 {
		newItem.Status = entity.AssessmentStatusComplete
		newItem.CompleteTime = time.Now().Unix()
	} else {
		newItem.Status = entity.AssessmentStatusInProgress
	}
	if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newItem); err != nil {
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("cmd", cmd),
			log.Any("new_item", newItem),
		)
		return err
	}

	if cmd.AttendanceIDs != nil {
		var items []*entity.AssessmentAttendance
		for _, attendanceID := range cmd.AttendanceIDs {
			items = append(items, &entity.AssessmentAttendance{
				ID:           utils.NewID(),
				AssessmentID: newID,
				AttendanceID: attendanceID,
				Checked:      true,
			})
		}
		if err := da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
			log.Error(ctx, "add assessment: batch insert assessment attendance map failed",
				log.Err(err),
				log.Any("items", items),
				log.Any("cmd", cmd),
			)
			return err
		}
	}

	var outcomeMap map[string]*entity.Outcome
	{
		outcomes, err := GetOutcomeModel().GetLearningOutcomesByIDs(ctx, tx, outcomeIDs, &entity.Operator{})
		if err != nil {
			log.Error(ctx, "get assessment detail: batch get outcomes failed by outcome ids",
				log.Err(err),
				log.Strings("outcome_ids", outcomeIDs),
			)
			return err
		}
		outcomeMap = make(map[string]*entity.Outcome, len(outcomes))
		for _, outcome := range outcomes {
			outcomeMap[outcome.ID] = outcome
		}
	}

	{
		var items []*entity.AssessmentOutcome
		for _, outcomeID := range outcomeIDs {
			noneAchieved := false
			if outcome := outcomeMap[outcomeID]; outcome != nil {
				noneAchieved = !outcome.Assumed
			}
			items = append(items, &entity.AssessmentOutcome{
				ID:           utils.NewID(),
				AssessmentID: newID,
				OutcomeID:    outcomeID,
				Skip:         false,
				NoneAchieved: noneAchieved,
			})
		}
		if len(items) > 0 {
			if err := da.GetAssessmentOutcomeDA().BatchInsert(ctx, tx, items); err != nil {
				log.Error(ctx, "add assessment: batch insert assessment outcome map failed",
					log.Err(err),
					log.Any("cmd", cmd),
					log.Any("items", items),
				)
				return err
			}
		}
	}

	{
		var items []*entity.OutcomeAttendance
		for _, outcomeID := range outcomeIDs {
			if outcomeMap[outcomeID] == nil {
				continue
			}
			if !outcomeMap[outcomeID].Assumed {
				continue
			}
			for _, attendanceID := range cmd.AttendanceIDs {
				items = append(items, &entity.OutcomeAttendance{
					ID:           utils.NewID(),
					AssessmentID: newID,
					OutcomeID:    outcomeID,
					AttendanceID: attendanceID,
				})
			}
		}
		if len(items) > 0 {
			if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
				log.Error(ctx, "add assessment: batch insert outcome attendance map failed",
					log.Err(err),
					log.Any("cmd", cmd),
					log.Any("items", items),
				)
				return err
			}
		}
	}

	//{
	//	if err := GetScheduleModel().UpdateScheduleStatus(ctx, tx, schedule.ID, entity.ScheduleStatusClosed); err != nil {
	//		log.Error(ctx, "add assessment: update schedule status to closed",
	//			log.Err(err),
	//			log.Any("cmd", cmd),
	//			log.Any("id", schedule.ID),
	//		)
	//		return err
	//	}
	//}
	return nil
}

func (a *assessmentModel) title(classEndTime int64, className string, lessonName string) string {
	return fmt.Sprintf("%s-%s-%s", time.Unix(classEndTime, 0).Format("20060102"), className, lessonName)
}

func (a *assessmentModel) UpdateAssessment(ctx context.Context, operator *entity.Operator, cmd entity.UpdateAssessmentCommand) error {
	// prepend check
	if !cmd.Action.Valid() {
		log.Error(ctx, "update assessment: invalid action", log.Any("cmd", cmd))
		return constant.ErrInvalidArgs
	}
	if cmd.OutcomeAttendanceMaps != nil {
		for _, item := range *cmd.OutcomeAttendanceMaps {
			if item.Skip && item.NoneAchieved {
				log.Error(ctx, "update assessment: check skip and none achieved combination", log.Any("cmd", cmd))
				return constant.ErrInvalidArgs
			}
			if (item.Skip || item.NoneAchieved) && len(item.AttendanceIDs) > 0 {
				log.Error(ctx, "update assessment: check skip and none achieved combination with attendance ids", log.Any("cmd", cmd))
				return constant.ErrInvalidArgs
			}
		}
	}
	if err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, cmd.ID)
		if err != nil {
			log.Error(ctx, "update assessment: get assessment exclude soft deleted failed",
				log.Err(err),
				log.Any("cmd", cmd),
			)
			return err
		}
		// permission check
		{
			hasP439, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, operator, external.AssessmentEditInProgressAssessment439)
			if err != nil {
				log.Error(ctx, "update assessment: check permission 439 failed",
					log.Any("cmd", cmd),
					log.Any("operator", operator),
				)
				return err
			}
			if !hasP439 {
				log.Error(ctx, "update assessment: not have permission 439",
					log.Any("cmd", cmd),
					log.Any("operator", operator),
				)
				return constant.ErrForbidden
			}
			teacherIDs, err := assessment.DecodeTeacherIDs()
			if err != nil {
				log.Error(ctx, "update assessment: decode teacher ids failed",
					log.Any("cmd", cmd),
					log.Any("operator", operator),
				)
				return err
			}
			find := false
			for _, item := range teacherIDs {
				if item == operator.UserID {
					find = true
					break
				}
			}
			if !find {
				log.Error(ctx, "update assessment: not find my assessment",
					log.Any("cmd", cmd),
					log.Any("operator", operator),
				)
				return constant.ErrForbidden
			}
		}
		if assessment.Status == entity.AssessmentStatusComplete {
			log.Info(ctx, "update assessment: assessment has completed, not allow update",
				log.Any("cmd", cmd),
				log.Any("operator", operator),
			)
			return errors.New("update assessment: assessment has completed, not allow update")
		}
		if cmd.Action == entity.UpdateAssessmentActionComplete {
			if err := da.GetAssessmentDA().UpdateStatus(ctx, tx, cmd.ID, entity.AssessmentStatusComplete); err != nil {
				log.Error(ctx, "update assessment: update status to complete failed",
					log.Err(err),
					log.Any("cmd", cmd),
					log.Any("operator", operator),
				)
				return err
			}
		}
		if cmd.AttendanceIDs != nil {
			if err := da.GetAssessmentAttendanceDA().Uncheck(ctx, tx, cmd.ID); err != nil {
				log.Error(ctx, "update assessment: uncheck assessment attendance failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return err
			}
			if err := da.GetAssessmentAttendanceDA().Check(ctx, tx, cmd.ID, *cmd.AttendanceIDs); err != nil {
				log.Error(ctx, "update assessment: check assessment attendance failed",
					log.Err(err),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		if cmd.OutcomeAttendanceMaps != nil {
			var outcomeIDs []string
			for _, item := range *cmd.OutcomeAttendanceMaps {
				outcomeIDs = append(outcomeIDs, item.OutcomeID)
			}
			if err := da.GetOutcomeAttendanceDA().BatchDeleteByAssessmentIDAndOutcomeIDs(ctx, tx, cmd.ID, outcomeIDs); err != nil {
				log.Error(ctx, "update assessment: batch delete outcome attendance map failed by outcome ids",
					log.Err(err),
					log.Strings("outcome_ids", outcomeIDs),
					log.Any("cmd", cmd),
				)
				return err
			}
			var (
				items              []*entity.OutcomeAttendance
				deletingOutcomeIDs []string
			)
			for _, item := range *cmd.OutcomeAttendanceMaps {
				for _, attendanceID := range item.AttendanceIDs {
					items = append(items, &entity.OutcomeAttendance{
						ID:           utils.NewID(),
						AssessmentID: cmd.ID,
						OutcomeID:    item.OutcomeID,
						AttendanceID: attendanceID,
					})
				}
				if err := da.GetAssessmentOutcomeDA().UpdateByAssessmentIDAndOutcomeID(ctx, tx, entity.AssessmentOutcome{
					AssessmentID: cmd.ID,
					OutcomeID:    item.OutcomeID,
					Skip:         item.Skip,
					NoneAchieved: item.NoneAchieved,
				}); err != nil {
					log.Error(ctx, "update assessment: batch update assessment outcome failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.String("id", cmd.ID),
						log.String("outcome_id", item.OutcomeID),
						log.Bool("skip", item.Skip),
						log.Bool("none_achieved", item.NoneAchieved),
					)
					return err
				}
				if item.Skip {
					deletingOutcomeIDs = append(deletingOutcomeIDs, item.OutcomeID)
				}
			}
			if len(deletingOutcomeIDs) > 0 {
				if err := da.GetOutcomeAttendanceDA().BatchDeleteByAssessmentIDAndOutcomeIDs(ctx, tx, cmd.ID, deletingOutcomeIDs); err != nil {
					log.Error(ctx, "update assessment: batch update assessment outcome failed",
						log.Err(err),
						log.Any("cmd", cmd),
						log.Strings("deleting_outcome_ids", deletingOutcomeIDs),
					)
					return err
				}
			}
			if err := da.GetOutcomeAttendanceDA().BatchInsert(ctx, tx, items); err != nil {
				log.Error(ctx, "update assessment: batch insert outcome attendance map failed",
					log.Err(err),
					log.Any("items", items),
					log.Any("cmd", cmd),
				)
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error(ctx, "update assessment: tx failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		return err
	}
	return nil
}

func (a *assessmentModel) existsTeachersByIDs(ctx context.Context, ids []string, operator *entity.Operator) (bool, error) {
	teacherService := external.GetTeacherServiceProvider()
	if _, err := teacherService.BatchGet(ctx, operator, ids); err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check teacher exists: not found teachers",
				log.Err(err),
				log.Strings("ids", ids),
			)
			return false, nil
		default:
			log.Error(ctx, "check teacher exists: batch get teachers failed",
				log.Err(err),
				log.Strings("ids", ids),
			)
			return false, err
		}
	}
	return true, nil
}

func (a *assessmentModel) existsSubjectByID(ctx context.Context, id string) (bool, error) {
	ids := []string{id}
	_, err := GetSubjectModel().Query(ctx, &da.SubjectCondition{
		IDs: entity.NullStrings{
			Strings: ids,
			Valid:   len(ids) != 0,
		},
	})
	if err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check subject exists: not found subjects",
				log.Err(err),
				log.String("id", id),
			)
			return false, nil
		default:
			log.Error(ctx, "check subject exists: batch get subjects failed",
				log.Err(err),
				log.String("id", id),
			)
			return false, err
		}
	}
	return true, nil
}

func (a *assessmentModel) existsProgramByID(ctx context.Context, id string) (bool, error) {
	ids := []string{id}
	_, err := GetProgramModel().Query(ctx, &da.ProgramCondition{
		IDs: entity.NullStrings{
			Strings: ids,
			Valid:   len(ids) != 0,
		},
	})
	if err != nil {
		switch err {
		case dbo.ErrRecordNotFound, constant.ErrRecordNotFound:
			log.Info(ctx, "check program exists: not found programs",
				log.Err(err),
				log.String("id", id),
			)
			return false, nil
		default:
			log.Error(ctx, "check program exists: batch get programs failed",
				log.Err(err),
				log.String("id", id),
			)
			return false, err
		}
	}
	return true, nil
}

type AssessmentTeacherAndStatusFilter struct {
	operator              *entity.Operator
	allowStatusComplete   bool
	allowStatusInProgress bool
	allowTeacherIDs       []string
	allowPairs            []*entity.AssessmentTeacherAndStatusPair
}

func NewAssessmentTeacherAndStatusFilter(operator *entity.Operator) *AssessmentTeacherAndStatusFilter {
	return &AssessmentTeacherAndStatusFilter{operator: operator}
}

func (f *AssessmentTeacherAndStatusFilter) FilterAll(ctx context.Context) (bool, error) {
	ok, err := f.FilterOrg(ctx)
	if err != nil {
		log.Error(ctx, "FilterAll: FilterOrg: check org permission failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}
	if !ok {
		return false, nil
	}

	ok, err = f.FilterSchool(ctx)
	if err != nil {
		log.Error(ctx, "FilterAll: FilterOrg: check school permission failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}
	if !ok {
		return false, nil
	}

	ok, err = f.FilterSelf(ctx)
	if err != nil {
		log.Error(ctx, "FilterAll: FilterSelf: check self permission failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

func (f *AssessmentTeacherAndStatusFilter) FilterOrg(ctx context.Context) (bool, error) {
	hasP424, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, f.operator, external.AssessmentViewOrgCompletedAssessments424)
	if err != nil {
		log.Error(ctx, "FilterOrg: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 424 failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}

	hasP425, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, f.operator, external.AssessmentViewOrgInProgressAssessments425)
	if err != nil {
		log.Error(ctx, "FilterOrg: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 425 failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}

	if hasP424 || hasP425 {
		teachers, err := external.GetTeacherServiceProvider().GetByOrganization(ctx, f.operator, f.operator.OrgID)
		if err != nil {
			log.Error(ctx, "FilterOrg: external.GetTeacherServiceProvider().GetByOrganization: get teachers failed",
				log.Err(err),
				log.Any("f", f),
			)
			return false, err
		}
		var teacherIDs []string
		for _, teacher := range teachers {
			teacherIDs = append(teacherIDs, teacher.ID)
		}
		f.allowTeacherIDs = append(f.allowTeacherIDs, teacherIDs...)

		if hasP424 {
			f.allowStatusComplete = true
			for _, teacherID := range teacherIDs {
				f.allowPairs = append(f.allowPairs, &entity.AssessmentTeacherAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusComplete,
				})
			}
		}

		if hasP425 {
			f.allowStatusInProgress = true
			for _, teacherID := range teacherIDs {
				f.allowPairs = append(f.allowPairs, &entity.AssessmentTeacherAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusInProgress,
				})
			}
		}
	}

	return true, nil
}

func (f *AssessmentTeacherAndStatusFilter) FilterSchool(ctx context.Context) (bool, error) {
	hasP426, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, f.operator, external.AssessmentViewSchoolCompletedAssessments426)
	if err != nil {
		log.Error(ctx, "FilterSchool: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 426 failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}

	hasP427, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, f.operator, external.AssessmentViewSchoolInProgressAssessments427)
	if err != nil {
		log.Error(ctx, "FilterSchool: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 427 failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}

	if hasP426 || hasP427 {
		schools, err := external.GetSchoolServiceProvider().GetByOperator(ctx, f.operator)
		if err != nil {
			log.Error(ctx, "FilterSchool: external.GetSchoolServiceProvider().GetByOperator: get schools failed",
				log.Err(err),
				log.Any("f", f),
			)
			return false, err
		}
		var schoolIDs []string
		for _, school := range schools {
			schoolIDs = append(schoolIDs, school.ID)
		}
		schoolID2TeachersMap, err := external.GetTeacherServiceProvider().GetBySchools(ctx, f.operator, schoolIDs)
		if err != nil {
			log.Error(ctx, "FilterSchool: external.GetTeacherServiceProvider().GetBySchools: get teachers failed",
				log.Err(err),
				log.Any("f", f),
				log.Any("school_ids", schoolIDs),
			)
			return false, err
		}

		var teacherIDs []string
		for _, teachers := range schoolID2TeachersMap {
			for _, teacher := range teachers {
				teacherIDs = append(teacherIDs, teacher.ID)
			}
		}
		f.allowTeacherIDs = append(f.allowTeacherIDs, teacherIDs...)

		if hasP426 {
			f.allowStatusComplete = true
			for _, teacherID := range f.allowTeacherIDs {
				f.allowPairs = append(f.allowPairs, &entity.AssessmentTeacherAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusComplete,
				})
			}
		}

		if hasP427 {
			f.allowStatusInProgress = true
			for _, teacherID := range teacherIDs {
				f.allowPairs = append(f.allowPairs, &entity.AssessmentTeacherAndStatusPair{
					TeacherID: teacherID,
					Status:    entity.AssessmentStatusInProgress,
				})
			}
		}
	}

	return true, nil
}

func (f *AssessmentTeacherAndStatusFilter) FilterSelf(ctx context.Context) (bool, error) {
	hasP414, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, f.operator, external.AssessmentViewCompletedAssessments414)
	if err != nil {
		log.Error(ctx, "FilterSelf: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 414 failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}
	hasP415, err := external.GetPermissionServiceProvider().HasOrganizationPermission(ctx, f.operator, external.AssessmentViewInProgressAssessments415)
	if err != nil {
		log.Error(ctx, "FilterSelf: external.GetPermissionServiceProvider().HasOrganizationPermission: check permission 415 failed",
			log.Err(err),
			log.Any("f", f),
		)
		return false, err
	}

	if hasP414 || hasP415 {
		if hasP414 {
			f.allowStatusComplete = true
			f.allowPairs = append(f.allowPairs, &entity.AssessmentTeacherAndStatusPair{
				TeacherID: f.operator.UserID,
				Status:    entity.AssessmentStatusComplete,
			})
		}
		if hasP415 {
			f.allowStatusInProgress = true
			f.allowPairs = append(f.allowPairs, &entity.AssessmentTeacherAndStatusPair{
				TeacherID: f.operator.UserID,
				Status:    entity.AssessmentStatusInProgress,
			})
		}
		f.allowTeacherIDs = append(f.allowTeacherIDs, f.operator.UserID)
	}

	return true, nil
}

func (f *AssessmentTeacherAndStatusFilter) AllowTeacherIDs() []string {
	return f.allowTeacherIDs
}

func (f *AssessmentTeacherAndStatusFilter) AllowPairs() []*entity.AssessmentTeacherAndStatusPair {
	var result []*entity.AssessmentTeacherAndStatusPair

	m := map[string]*struct {
		allowStatusInProgress bool
		allowStatusComplete   bool
	}{}
	for _, pair := range f.allowPairs {
		if m[pair.TeacherID] == nil {
			m[pair.TeacherID] = &struct {
				allowStatusInProgress bool
				allowStatusComplete   bool
			}{}
		}
		switch pair.Status {
		case entity.AssessmentStatusComplete:
			m[pair.TeacherID].allowStatusComplete = true
		case entity.AssessmentStatusInProgress:
			m[pair.TeacherID].allowStatusInProgress = true
		}
	}

	for teacherID, statuses := range m {
		if statuses.allowStatusInProgress && statuses.allowStatusComplete {
			continue
		}
		if !statuses.allowStatusInProgress && !statuses.allowStatusComplete {
			continue
		}
		if statuses.allowStatusComplete {
			result = append(result, &entity.AssessmentTeacherAndStatusPair{
				TeacherID: teacherID,
				Status:    entity.AssessmentStatusComplete,
			})
		}
		if statuses.allowStatusInProgress {
			result = append(result, &entity.AssessmentTeacherAndStatusPair{
				TeacherID: teacherID,
				Status:    entity.AssessmentStatusInProgress,
			})
		}
	}

	return result
}

func (f *AssessmentTeacherAndStatusFilter) AllowStatuses() []entity.AssessmentStatus {
	var result []entity.AssessmentStatus
	if f.allowStatusInProgress {
		result = append(result, entity.AssessmentStatusInProgress)
	}
	if f.allowStatusComplete {
		result = append(result, entity.AssessmentStatusComplete)
	}
	return result
}

func (f *AssessmentTeacherAndStatusFilter) CheckStatus(t *entity.AssessmentStatus) bool {
	// if t equal nil that says user give up check.
	if t == nil {
		return true
	}
	ss := f.AllowStatuses()
	for _, s := range ss {
		if s == *t {
			return true
		}
	}
	return false
}
