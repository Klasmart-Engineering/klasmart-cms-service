package model

import (
	"context"
	"database/sql"
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

var (
	assessmentUtilsInstance     *assessmentUtils
	assessmentUtilsInstanceOnce = sync.Once{}
)

func GetAssessmentUtils() *assessmentUtils {
	assessmentUtilsInstanceOnce.Do(func() {
		assessmentUtilsInstance = &assessmentUtils{}
	})
	return assessmentUtilsInstance
}

type assessmentUtils struct{}

func (m *assessmentUtils) ExistsByScheduleID(ctx context.Context, operator *entity.Operator, scheduleID string) (bool, error) {
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

func (m *assessmentUtils) CalcRemainingTime(dueAt int64, createdAt int64) time.Duration {
	var r int64
	if dueAt != 0 {
		r = dueAt - time.Now().Unix()
	} else {
		r = time.Unix(createdAt, 0).Add(constant.AssessmentDefaultRemainingTime).Unix() - time.Now().Unix()
	}
	if r < 0 {
		return 0
	}
	return time.Duration(r) * time.Second
}

func (m *assessmentUtils) ToViews(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, assessments []*entity.Assessment, options entity.ConvertToViewsOptions) ([]*entity.AssessmentView, error) {
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
		log.Error(ctx, "ToViews: GetScheduleModel().GetVariableDataByIDs: get failed",
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
			log.Error(ctx, "ToViews: external.GetProgramServiceProvider().BatchGetNameMap: get failed",
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
			log.Error(ctx, "ToViews: da.GetAssessmentAttendanceDA().QueryTx: query failed",
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
			log.Error(ctx, "ToViews: external.GetTeacherServiceProvider().BatchGetNameMap: get failed",
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
			log.Error(ctx, "ToViews: da.GetAssessmentAttendanceDA().QueryTx: query failed",
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
			log.Error(ctx, "ToViews: external.GetStudentServiceProvider().BatchGetNameMap: get failed",
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
				lessonMaterialIDs = append(lessonMaterialIDs, c.ID)
			}
		}
		lessonMaterialSourceMap, err := m.BatchGetLessonMaterialDataMap(ctx, tx, operator, lessonMaterialIDs)
		if err != nil {
			log.Error(ctx, "to views: get lesson material source map failed",
				log.Err(err),
				log.Strings("lesson_material_ids", lessonMaterialIDs),
			)
			return nil, err
		}

		for _, c := range contents {
			switch c.ContentType {
			case entity.ContentTypePlan:
				assessmentLessonPlanMap[c.AssessmentID] = &entity.AssessmentLessonPlan{
					ID:   c.ContentID,
					Name: c.ContentName,
				}
			case entity.ContentTypeMaterial:
				data := lessonMaterialSourceMap[c.ID]
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
			v.LessonPlan = assessmentLessonPlanMap[a.ID]
			v.LessonMaterials = assessmentLessonMaterialsMap[a.ID]
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

func (m *assessmentUtils) GetRoomCompleteRate(room *entity.AssessmentH5PRoom, userIDs []string, h5pIDs []string) float64 {
	if room == nil {
		return 0
	}
	userIDExistsMap := map[string]bool{}
	for _, uid := range userIDs {
		userIDExistsMap[uid] = true
	}
	h5pIDExistsMap := map[string]bool{}
	for _, id := range h5pIDs {
		h5pIDExistsMap[id] = true
	}
	total := len(userIDs) * len(h5pIDs)
	attempted := 0
	for _, u := range room.Users {
		if !userIDExistsMap[u.UserID] {
			continue
		}
		for _, c := range u.Contents {
			if !h5pIDExistsMap[c.H5PID] {
				continue
			}
			if len(c.Answers) > 0 {
				attempted++
			}
		}
	}
	if total > 0 {
		return float64(attempted) / float64(total)
	}
	return 0
}

func (m *assessmentUtils) BatchGetRoomScoreMap(ctx context.Context, operator *entity.Operator, roomIDs []string, enableComment bool) (map[string]*entity.AssessmentH5PRoom, error) {
	roomScoreMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		return nil, err
	}

	var roomCommentMap map[string]map[string][]string
	if enableComment {
		roomCommentMap, err = m.BatchGetRoomCommentMap(ctx, operator, roomIDs)
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
				if len(assessmentContent.Answers) > 0 {
					assessmentContent.Answer = assessmentContent.Answers[0]
					attempted++
				}
				if len(s.TeacherScores) > 0 {
					assessmentContent.AchievedScore = s.TeacherScores[0].Score
				} else if len(s.Score.Scores) > 0 {
					assessmentContent.AchievedScore = s.Score.Scores[len(s.Score.Scores)-1]
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
				assessmentUser.Comment = roomCommentMap[roomID][assessmentUser.UserID][0]
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

func (m *assessmentUtils) BatchGetRoomCommentMap(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]map[string][]string, error) {
	commentMap, err := external.GetH5PRoomCommentServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string][]string, len(commentMap))
	for roomID, users := range commentMap {
		result[roomID] = make(map[string][]string, len(users))
		for _, u := range users {
			if u.User == nil {
				log.Debug(ctx, "get room comment map: user is nil",
					log.Strings("room_ids", roomIDs),
					log.Any("comment_map", commentMap),
					log.Any("operator", operator),
				)
				continue
			}
			for _, c := range u.TeacherComments {
				result[roomID][u.User.UserID] = append(result[roomID][u.User.UserID], c.Comment)
			}
		}
	}
	return result, nil
}

func (m *assessmentUtils) GetH5PStudentViewItems(ctx context.Context, operator *entity.Operator, view *entity.AssessmentView) ([]*entity.AssessmentStudentViewH5PItem, error) {
	roomMap, err := m.BatchGetRoomScoreMap(ctx, operator, []string{view.RoomID}, true)
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
				OutcomeNames:       nil,
			}
			var content *entity.AssessmentH5PContentScore
			if user != nil {
				content = user.ContentMap[lm.Source]
				if content != nil {
					newLessMaterial.LessonMaterialType = content.ContentType
					newLessMaterial.Answer = content.Answer
					newLessMaterial.MaxScore = content.MaxPossibleScore
					newLessMaterial.AchievedScore = content.AchievedScore
					newLessMaterial.Attempted = len(content.Answers) > 0
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

func (m *assessmentUtils) BatchGetLatestLessonPlanMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, lessonPlanIDs []string) (map[string]*entity.AssessmentExternalLessonPlan, error) {
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
		log.Error(ctx, "ToViews: GetContentModel().GetContentByIDList: get failed",
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

func (m *assessmentUtils) BatchGetLessonMaterialDataMap(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, ids []string) (map[string]*MaterialData, error) {
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

func (m *assessmentUtils) AddAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input entity.AddAttendancesInput) error {
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