package model

import (
	"context"
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
	AddClassAndLive(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) ([]string, error)
	ExistsByScheduleID(ctx context.Context, operator *entity.Operator, scheduleID string) (map[entity.AssessmentType]bool, error)
	ConvertToViews(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessments []*entity.Assessment, options entity.ConvertToViewsOptions) ([]*entity.AssessmentView, error)
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

func (*assessmentModel) AddClassAndLive(ctx context.Context, operator *entity.Operator, args entity.AddAssessmentArgs) ([]string, error) {
	newAssessmentIDs := make([]string, 0, 2)
	args.Type = entity.AssessmentTypeClassAndLiveOutcome
	id, err := GetOutcomeAssessmentModel().Add(ctx, operator, args)
	if err != nil {
		return nil, err
	}
	newAssessmentIDs = append(newAssessmentIDs, id)
	args.Type = entity.AssessmentTypeClassAndLiveH5P
	id2, err := GetContentAssessmentModel().AddClassAndLive(ctx, operator, args)
	if err != nil {
		return nil, err
	}
	newAssessmentIDs = append(newAssessmentIDs, id2)
	return newAssessmentIDs, nil
}

func (m *assessmentModel) ExistsByScheduleID(ctx context.Context, operator *entity.Operator, scheduleID string) (map[entity.AssessmentType]bool, error) {
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
		return nil, err
	}
	typeMap := make(map[entity.AssessmentType]bool, len(assessments))
	for _, a := range assessments {
		typeMap[a.Type] = true
	}
	return typeMap, nil
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
		lessonPlanMap map[string]*entity.AssessmentViewLessonPlan
		lessonPlanIDs []string
	)
	if options.EnableLessonPlan {
		var lessonPlans = make([]*entity.ContentInfoWithDetails, 0, len(schedules))
		for _, s := range schedules {
			lessonPlanIDs = append(lessonPlanIDs, s.LessonPlanID)
		}
		lessonPlanIDs = utils.SliceDeduplication(lessonPlanIDs)
		lessonPlanIDs, err = GetContentModel().GetLatestContentIDByIDList(ctx, tx, lessonPlanIDs)
		if err != nil {
			log.Error(ctx, "ConvertToViews: GetContentModel().GetLatestContentIDByIDList: get failed",
				log.Err(err),
				log.Strings("lesson_plan_ids", lessonPlanIDs),
				log.Any("options", options),
			)
			return nil, err
		}
		lessonPlans, err = GetContentModel().GetContentByIDList(ctx, tx, lessonPlanIDs, operator)
		if err != nil {
			log.Error(ctx, "ConvertToViews: GetContentModel().GetContentByIDList: get failed",
				log.Err(err),
				log.Strings("lesson_plan_ids", lessonPlanIDs),
				log.Any("options", options),
			)
			return nil, err
		}
		lessonPlanMap = make(map[string]*entity.AssessmentViewLessonPlan, len(lessonPlans))
		for _, lp := range lessonPlans {
			lessonPlanMap[lp.ID] = &entity.AssessmentViewLessonPlan{
				ID:   lp.ID,
				Name: lp.Name,
			}
		}
	}

	// fill lesson materials
	var lessonPlanToMaterialsMap map[string][]*entity.AssessmentViewLessonMaterial
	if options.EnableLessonPlan && options.EnableLessonMaterials {
		lessonPlanToMaterialsMap = make(map[string][]*entity.AssessmentViewLessonMaterial, len(lessonPlanMap))
		m2, err := GetContentModel().GetContentsSubContentsMapByIDList(ctx, dbo.MustGetDB(ctx), lessonPlanIDs, operator)
		if err != nil {
			log.Error(ctx, "List: GetContentModel().GetContentsSubContentsMapByIDList: get failed",
				log.Err(err),
				log.Strings("lesson_plan_ids", lessonPlanIDs),
				log.Any("options", options),
			)
			return nil, err
		}
		for lessonPlanID, lessonMaterials := range m2 {
			for _, lm := range lessonMaterials {
				newItem := entity.AssessmentViewLessonMaterial{
					ID:     lm.ID,
					Name:   lm.Name,
					Source: "",
				}
				switch v := lm.Data.(type) {
				case *MaterialData:
					newItem.Source = string(v.Source)
				}
				lessonPlanToMaterialsMap[lessonPlanID] = append(lessonPlanToMaterialsMap[lessonPlanID], &newItem)
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
			v.LessonPlan = lessonPlanMap[s.LessonPlanID]
		}
		if options.EnableLessonPlan && options.EnableLessonMaterials {
			v.LessonMaterials = lessonPlanToMaterialsMap[s.LessonPlanID]
		}
		result = append(result, &v)
	}

	return result, nil
}
