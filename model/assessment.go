package model

import (
	"context"
	"database/sql"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sort"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IAssessmentModel interface {
	AddClassAndLive(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) ([]string, error)
	ExistsByScheduleID(ctx context.Context, operator *entity.Operator, scheduleID string) (bool, error)
	ConvertToViews(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessments []*entity.Assessment, options entity.ConvertToViewsOptions) ([]*entity.AssessmentView, error)
	AddAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input entity.AddAttendancesInput) error
	BatchAddAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input entity.BatchAddAttendancesInput) error
	AddContents(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, contents []*entity.ContentInfoWithDetails) error
	BatchGetLatestLessonPlanMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, lessonPlanIDs []string) (map[string]*entity.AssessmentExternalLessonPlan, error)
}

var (
	ErrNotFoundAttendance = errors.New("not found attendance")
	ErrNotAttemptedAssessment = errors.New("not attempted assessment")

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

func (*assessmentModel) AddClassAndLive(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) ([]string, error) {
	newAssessmentIDs := make([]string, 0, 2)
	args.Type = entity.AssessmentTypeClassAndLiveOutcome
	id, err := GetOutcomeAssessmentModel().Add(ctx, tx, operator, args)
	if err != nil {
		return nil, err
	}
	newAssessmentIDs = append(newAssessmentIDs, id)
	args.Type = entity.AssessmentTypeClassAndLiveH5P
	id2, err := GetH5PAssessmentModel().AddClassAndLive(ctx, tx, operator, args)
	if err != nil {
		return nil, err
	}
	newAssessmentIDs = append(newAssessmentIDs, id2)
	return newAssessmentIDs, nil
}

func (m *assessmentModel) ExistsByScheduleID(ctx context.Context, operator *entity.Operator, scheduleID string) (bool, error) {
	var assessments []*entity.Assessment
	cond := da.QueryAssessmentConditions{
		ScheduleIDs: entity.NullStrings{
			Strings: []string{scheduleID},
			Valid:   true,
		},
	}
	if err := da.GetAssessmentDA().Query(ctx, &cond, &assessments); err != nil {
		log.Error(ctx, "ExistsByScheduleID: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Any("cond", cond),
		)
		return false, err
	}
	return len(assessments) > 0, nil
}

func (m *assessmentModel) ConvertToViews(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessments []*entity.Assessment, options entity.ConvertToViewsOptions) ([]*entity.AssessmentView, error) {
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
		log.Error(ctx, "ConvertToViews: GetScheduleModel().GetVariableDataByIDs: get failed",
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
			log.Error(ctx, "ConvertToViews: external.GetProgramServiceProvider().BatchGetNameMap: get failed",
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
			log.Error(ctx, "ConvertToViews: da.GetAssessmentAttendanceDA().QueryTx: query failed",
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
			log.Error(ctx, "ConvertToViews: external.GetTeacherServiceProvider().BatchGetNameMap: get failed",
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
			log.Error(ctx, "ConvertToViews: da.GetAssessmentAttendanceDA().QueryTx: query failed",
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
			log.Error(ctx, "ConvertToViews: external.GetStudentServiceProvider().BatchGetNameMap: get failed",
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
		assessmentLessonPlanMap map[string]*entity.AssessmentLessonPlan
		assessmentLessonMaterialsMap map[string][]*entity.AssessmentLessonMaterial
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
		for _, c := range contents {
			switch c.ContentType {
			case entity.ContentTypePlan:
				assessmentLessonPlanMap[c.AssessmentID] = &entity.AssessmentLessonPlan{
					ID:        c.ContentID,
					Name:      c.ContentName,
				}
			case entity.ContentTypeMaterial:
				assessmentLessonMaterialsMap[c.AssessmentID] = append(assessmentLessonMaterialsMap[c.AssessmentID], &entity.AssessmentLessonMaterial{
					ID:      c.ContentID,
					Name:    c.ContentName,
					Comment: c.ContentComment,
					Source:  c.ContentSource,
					Checked: c.Checked,
				})
			}
		}
	}

	var result []*entity.AssessmentView
	for _, a := range assessments {
		var (
			s = scheduleMap[a.ScheduleID]
			v = entity.AssessmentView{Assessment: a, Schedule: s.Schedule}
		)
		if s == nil || v.Schedule == nil {
			log.Warn(ctx, "List: not found schedule", log.Any("assessment", a))
			continue
		}
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
			v.LessonPlan = assessmentLessonPlanMap[a.ID]
			v.LessonMaterials = assessmentLessonMaterialsMap[a.ID]
		}
		result = append(result, &v)
	}

	return result, nil
}

func (m *assessmentModel) AddAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input entity.AddAttendancesInput) error {
	var (
		err               error
		scheduleRelations []*entity.ScheduleRelation
	)
	if input.ScheduleRelations != nil {
		scheduleRelations = input.ScheduleRelations
	} else {
		cond := &da.ScheduleRelationCondition{
			ScheduleID: sql.NullString{
				String: input.ScheduleID,
				Valid:  true,
			},
			RelationIDs: entity.NullStrings{
				Strings: input.AttendanceIDs,
				Valid:   true,
			},
		}
		if scheduleRelations, err = GetScheduleRelationModel().Query(ctx, operator, cond); err != nil {
			log.Error(ctx, "AddAttendances: GetScheduleRelationModel().Query: get failed",
				log.Err(err),
				log.Any("input", input),
				log.Any("operator", operator),
			)
			return err
		}
	}
	if len(scheduleRelations) == 0 {
		log.Error(ctx, "AddAttendances: not found any schedule relations",
			log.Err(err),
			log.Any("input", input),
			log.Any("operator", operator),
		)
		return ErrNotFoundAttendance
	}

	var assessmentAttendances []*entity.AssessmentAttendance
	for _, relation := range scheduleRelations {
		newAttendance := entity.AssessmentAttendance{
			ID:           utils.NewID(),
			AssessmentID: input.AssessmentID,
			AttendanceID: relation.RelationID,
			Checked:      true,
		}
		switch relation.RelationType {
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
		assessmentAttendances = append(assessmentAttendances, &newAttendance)
	}
	if err = da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, assessmentAttendances); err != nil {
		log.Error(ctx, "AddAttendances: da.GetAssessmentAttendanceDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("input", input),
			log.Any("scheduleRelations", scheduleRelations),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (m *assessmentModel) BatchAddAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input entity.BatchAddAttendancesInput) error {
	if len(input.Items) == 0 {
		return nil
	}
	var (
		err                         error
		cond                        = da.ScheduleRelationCondition{}
		scheduleRelations           []*entity.ScheduleRelation
		scheduleIDToAssessmentIDMap = make(map[string]string, len(input.Items))
	)
	for _, item := range input.Items {
		cond.ScheduleAndRelations = append(cond.ScheduleAndRelations, &da.ScheduleAndRelations{
			ScheduleID:  item.ScheduleID,
			RelationIDs: item.AttendanceIDs,
		})
		scheduleIDToAssessmentIDMap[item.ScheduleID] = item.AssessmentID
	}
	if scheduleRelations, err = GetScheduleRelationModel().Query(ctx, operator, &cond); err != nil {
		log.Error(ctx, "AddAttendances: GetScheduleRelationModel().Query: get failed",
			log.Err(err),
			log.Any("input", input),
			log.Any("operator", operator),
		)
		return err
	}

	var assessmentAttendances []*entity.AssessmentAttendance
	for _, relation := range scheduleRelations {
		newAttendance := entity.AssessmentAttendance{
			ID:           utils.NewID(),
			AssessmentID: scheduleIDToAssessmentIDMap[relation.ScheduleID],
			AttendanceID: relation.RelationID,
			Checked:      true,
		}
		switch relation.RelationType {
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
		assessmentAttendances = append(assessmentAttendances, &newAttendance)
	}
	if err = da.GetAssessmentAttendanceDA().BatchInsert(ctx, tx, assessmentAttendances); err != nil {
		log.Error(ctx, "AddAttendances: da.GetAssessmentAttendanceDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("input", input),
			log.Any("scheduleRelations", scheduleRelations),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (m *assessmentModel) AddContents(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessmentID string, contents []*entity.ContentInfoWithDetails) error {
	if len(contents) == 0 {
		return nil
	}
	var (
		assessmentContents []*entity.AssessmentContent
	)
	for _, c := range contents {
		assessmentContents = append(assessmentContents, &entity.AssessmentContent{
			ID:           utils.NewID(),
			AssessmentID: assessmentID,
			ContentID:    c.ID,
			ContentName:  c.Name,
			ContentType:  c.ContentType,
			Checked:      true,
		})
	}
	if len(assessmentContents) == 0 {
		return nil
	}
	if err := da.GetAssessmentContentDA().BatchInsert(ctx, tx, assessmentContents); err != nil {
		log.Error(ctx, "addAssessmentContentsAndOutcomes: da.GetAssessmentContentDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("assessment_contents", assessmentContents),
			log.String("assessment_id", assessmentID),
			log.Any("contents", contents),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (m *assessmentModel) BatchGetLatestLessonPlanMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, lessonPlanIDs []string) (map[string]*entity.AssessmentExternalLessonPlan, error) {
	lessonPlanIDs = utils.SliceDeduplication(lessonPlanIDs)
	var (
		err         error
		lessonPlans = make([]*entity.ContentInfoWithDetails, 0, len(lessonPlanIDs))
	)
	lessonPlanIDs, err = GetContentModel().GetLatestContentIDByIDList(ctx, tx, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "BatchGetLatestLessonPlanMap: GetContentModel().GetLatestContentIDByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}
	lessonPlans, err = GetContentModel().GetContentByIDList(ctx, tx, lessonPlanIDs, operator)
	if err != nil {
		log.Error(ctx, "ConvertToViews: GetContentModel().GetContentByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}
	lessonPlanMap := make(map[string]*entity.AssessmentExternalLessonPlan, len(lessonPlans))
	for _, lp := range lessonPlans {
		lessonPlanMap[lp.ID] = &entity.AssessmentExternalLessonPlan{
			ID:   lp.ID,
			Name: lp.Name,
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
	for id, lp := range lessonPlanMap {
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
	return lessonPlanMap, nil
}
