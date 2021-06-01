package model

import (
	"context"
	"database/sql"
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

type IH5PAssessmentModel interface {
	List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListH5PAssessmentsArgs) (*entity.ListH5PAssessmentsResult, error)
	GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.GetH5PAssessmentDetailResult, error)
	Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs) error
	AddClassAndLive(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error)
	DeleteStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error
	AddStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input []*entity.AddStudyInput) ([]string, error)
	BatchCheckAnyoneAttempted(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, roomIDs []string) (map[string]bool, error)
}

var (
	ErrAssessmentNotFoundSchedule = errors.New("assessment: not found schedule")

	h5pAssessmentModelInstance     IH5PAssessmentModel
	h5pAssessmentModelInstanceOnce = sync.Once{}
)

func GetH5PAssessmentModel() IH5PAssessmentModel {
	h5pAssessmentModelInstanceOnce.Do(func() {
		h5pAssessmentModelInstance = &h5pAssessmentModel{}
	})
	return h5pAssessmentModelInstance
}

type h5pAssessmentModel struct{}

func (m *h5pAssessmentModel) List(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.ListH5PAssessmentsArgs) (*entity.ListH5PAssessmentsResult, error) {
	// check args
	if !args.Type.Valid() {
		errMsg := "list h5p assessments: require assessment type"
		log.Error(ctx, errMsg, log.Any("args", args), log.Any("operator", operator))
		return nil, constant.ErrInvalidArgs
	}

	// check permission
	var (
		checker = NewAssessmentPermissionChecker(operator)
		err     error
	)
	if err = checker.SearchAllPermissions(ctx); err != nil {
		log.Error(ctx, "List: checker.SearchAllPermissions: search failed",
			log.Err(err),
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, err
	}
	if !checker.CheckStatus(args.Status.Value) {
		log.Error(ctx, "list h5p assessments: check status failed",
			log.Any("operator", operator),
			log.Any("args", args),
		)
		return nil, constant.ErrForbidden
	}

	// get assessment list
	var (
		assessments []*entity.Assessment
		cond        = da.QueryAssessmentConditions{
			Type: entity.NullAssessmentType{
				Value: args.Type,
				Valid: true,
			},
			OrgID: entity.NullString{
				String: operator.OrgID,
				Valid:  true,
			},
			Status: args.Status,
			AllowTeacherIDs: entity.NullStrings{
				Strings: checker.allowTeacherIDs,
				Valid:   true,
			},
			AllowTeacherIDAndStatusPairs: entity.NullAssessmentAllowTeacherIDAndStatusPairs{
				Values: checker.AllowPairs(),
				Valid:  len(checker.AllowPairs()) > 0,
			},
			OrderBy: args.OrderBy,
			Pager:   args.Pager,
		}
	)
	if args.Query != "" {
		switch args.QueryType {
		case entity.ListH5PAssessmentsQueryTypeTeacherName:
			cond.TeacherIDs.Valid = true
			teachers, err := external.GetTeacherServiceProvider().Query(ctx, operator, operator.OrgID, args.Query)
			if err != nil {
				return nil, err
			}
			for _, item := range teachers {
				cond.TeacherIDs.Strings = append(cond.TeacherIDs.Strings, item.ID)
			}
		}
	}
	log.Debug(ctx, "List: print query cond", log.Any("cond", cond))
	total, err := da.GetAssessmentDA().PageTx(ctx, tx, &cond, &assessments)
	if err != nil {
		log.Error(ctx, "List: da.GetAssessmentDA().QueryTx: query failed",
			log.Err(err),
			log.Any("cond", cond),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// convert to assessment view
	var views []*entity.AssessmentView
	if views, err = GetAssessmentModel().ToViews(ctx, tx, operator, assessments, entity.ConvertToViewsOptions{
		CheckedStudents:  sql.NullBool{Bool: true, Valid: true},
		EnableProgram:    true,
		EnableSubjects:   true,
		EnableTeachers:   true,
		EnableStudents:   true,
		EnableClass:      true,
		EnableLessonPlan: true,
	}); err != nil {
		log.Error(ctx, "List: GetAssessmentModel().ToViews: get failed",
			log.Err(err),
			log.Any("assessments", assessments),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// get scores
	var roomIDs []string
	for _, v := range views {
		roomIDs = append(roomIDs, v.RoomID)
	}
	roomMap, err := m.getRoomScoreMap(ctx, operator, roomIDs, false)
	if err != nil {
		log.Error(ctx, "list h5p assessments: get room user scores map failed",
			log.Err(err),
			log.Strings("room_ids", roomIDs),
		)
		return nil, err
	}

	// construct result
	var result = entity.ListH5PAssessmentsResult{Total: total}
	for _, v := range views {
		teacherNames := make([]string, 0, len(v.Teachers))
		for _, t := range v.Teachers {
			teacherNames = append(teacherNames, t.Name)
		}
		var remainingTime int64
		if v.Schedule.DueAt != 0 {
			remainingTime = v.Schedule.DueAt - time.Now().Unix()
		} else {
			remainingTime = time.Unix(v.CreateAt, 0).Add(constant.AssessmentDefaultRemainingTime).Unix() - time.Now().Unix()
		}
		if remainingTime < 0 {
			remainingTime = 0
		}

		userIDs := make([]string, 0, len(v.Students))
		for _, s := range v.Students {
			userIDs = append(userIDs, s.ID)
		}
		contentIDs := make([]string, 0, len(v.LessonMaterials))
		for _, lm := range v.LessonMaterials {
			contentIDs = append(contentIDs, lm.ID)
		}

		newItem := entity.ListH5PAssessmentsResultItem{
			ID:            v.ID,
			Title:         v.Title,
			TeacherNames:  teacherNames,
			ClassName:     v.Class.Name,
			DueAt:         v.Schedule.DueAt,
			CompleteRate:  m.getRoomCompleteRate(roomMap[v.RoomID], userIDs, contentIDs),
			RemainingTime: remainingTime,
			CompleteAt:    v.CompleteTime,
			ScheduleID:    v.ScheduleID,
			CreateAt:      v.CreateAt,
		}
		result.Items = append(result.Items, &newItem)
	}

	return &result, nil
}

func (m *h5pAssessmentModel) GetDetail(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, id string) (*entity.GetH5PAssessmentDetailResult, error) {
	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "get detail: get assessment failed",
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
	if views, err = GetAssessmentModel().ToViews(ctx, tx, operator, []*entity.Assessment{assessment}, entity.ConvertToViewsOptions{
		EnableProgram:    true,
		EnableSubjects:   true,
		EnableTeachers:   true,
		EnableStudents:   true,
		EnableClass:      true,
		EnableLessonPlan: true,
	}); err != nil {
		log.Error(ctx, "Get: GetAssessmentModel().ToViews: get failed",
			log.Err(err),
			log.String("assessment_id", id),
			log.Any("operator", operator),
		)
		return nil, err
	}
	view = views[0]

	// construct result
	result := entity.GetH5PAssessmentDetailResult{
		ID:               view.ID,
		Title:            view.Title,
		ClassName:        view.Class.Name,
		Teachers:         view.Teachers,
		Students:         view.Students,
		DueAt:            view.Schedule.DueAt,
		LessonPlan:       entity.H5PAssessmentLessonPlan{},
		LessonMaterials:  nil,
		CompleteAt:       view.CompleteTime,
		RemainingTime:    0,
		StudentViewItems: nil,
		ScheduleID:       view.ScheduleID,
		Status:           view.Status,
	}

	// remaining time
	if view.Schedule.DueAt != 0 {
		result.RemainingTime = view.Schedule.DueAt - time.Now().Unix()
	} else {
		result.RemainingTime = time.Unix(view.CreateAt, 0).Add(constant.AssessmentDefaultRemainingTime).Unix() - time.Now().Unix()
	}
	if result.RemainingTime < 0 {
		result.RemainingTime = 0
	}

	// fill lesson plan and lesson materials
	plan, err := da.GetAssessmentContentDA().GetPlan(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "Get: da.GetAssessmentContentDA().GetPlan: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
	}
	result.LessonPlan = entity.H5PAssessmentLessonPlan{
		ID:   plan.ContentID,
		Name: plan.ContentName,
	}
	materials, err := da.GetAssessmentContentDA().GetMaterials(ctx, tx, id)
	if err != nil {
		log.Error(ctx, "Get: da.GetAssessmentContentDA().GetMaterials: get failed",
			log.Err(err),
			log.String("assessment_id", id),
		)
	}
	for _, m := range materials {
		result.LessonMaterials = append(result.LessonMaterials, &entity.H5PAssessmentLessonMaterial{
			ID:      m.ContentID,
			Name:    m.ContentName,
			Comment: m.ContentComment,
			Checked: m.Checked,
		})
	}

	// get h5p room scores
	roomMap, err := m.getRoomScoreMap(ctx, operator, []string{view.RoomID}, true)
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
		return &result, nil
	}

	// student view items
	for _, s := range view.Students {
		newItem := entity.H5PAssessmentStudentViewItem{
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
			newLessMaterial := entity.H5PAssessmentStudentViewLessonMaterial{
				LessonMaterialID:   lm.ID,
				LessonMaterialName: lm.Name,
				IsH5P:              lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend,
			}
			var content *entity.AssessmentH5PContentScore
			if user != nil {
				content = user.ContentMap[lm.ID]
				if content != nil {
					newLessMaterial.LessonMaterialType = content.ContentType
					newLessMaterial.Answer = content.Answer
					newLessMaterial.MaxScore = content.MaxPossibleScore
					newLessMaterial.AchievedScore = content.AchievedScore
					newLessMaterial.Attempted = len(content.Answers) > 0
				} else {
					log.Debug(ctx, "get h5p assessment detail: not found content from h5p room",
						log.String("room_id", view.RoomID),
						log.Any("not_found_content_id", lm.ID),
						log.Any("room", room),
					)
				}
			}
			newItem.LessonMaterials = append(newItem.LessonMaterials, &newLessMaterial)
		}
		result.StudentViewItems = append(result.StudentViewItems, &newItem)
	}

	// order students
	sort.Sort(entity.H5PAssessmentStudentViewItemsOrder(result.StudentViewItems))

	return &result, nil
}

func (m *h5pAssessmentModel) getRoomScoreMap(ctx context.Context, operator *entity.Operator, roomIDs []string, enableComment bool) (map[string]*entity.AssessmentH5PRoom, error) {
	roomScoreMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		return nil, err
	}

	var roomCommentMap map[string]map[string][]string
	if enableComment {
		roomCommentMap, err = m.getRoomCommentMap(ctx, operator, roomIDs)
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
				assessmentContentMap[assessmentContent.ContentID] = &assessmentContent
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

func (m *h5pAssessmentModel) getRoomCompleteRate(room *entity.AssessmentH5PRoom, userIDs []string, contentIDs []string) float64 {
	if room == nil {
		return 0
	}
	userIDExistsMap := map[string]bool{}
	for _, uid := range userIDs {
		userIDExistsMap[uid] = true
	}
	contentIDExistsMap := map[string]bool{}
	for _, cid := range contentIDs {
		contentIDExistsMap[cid] = true
	}
	total, attempted := 0, 0
	for _, u := range room.Users {
		if !userIDExistsMap[u.UserID] {
			continue
		}
		for _, c := range u.Contents {
			if !contentIDExistsMap[c.ContentID] {
				continue
			}
			total++
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

func (m *h5pAssessmentModel) getRoomCommentMap(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]map[string][]string, error) {
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

func (m *h5pAssessmentModel) Update(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, args entity.UpdateH5PAssessmentArgs) error {
	// validate
	if !args.Action.Valid() {
		log.Error(ctx, "update h5p assessment: invalid action", log.Any("args", args))
		return constant.ErrInvalidArgs
	}

	assessment, err := da.GetAssessmentDA().GetExcludeSoftDeleted(ctx, dbo.MustGetDB(ctx), args.ID)
	if err != nil {
		log.Error(ctx, "update h5p assessment: get assessment failed",
			log.Err(err),
			log.Any("args", args),
		)
		return err
	}

	// permission check
	hasP439, err := NewAssessmentPermissionChecker(operator).HasP439(ctx)
	if err != nil {
		return err
	}
	if !hasP439 {
		log.Error(ctx, "update assessment: not have permission 439",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return constant.ErrForbidden
	}
	teacherIDs, err := da.GetAssessmentAttendanceDA().GetTeacherIDsByAssessmentID(ctx, tx, args.ID)
	if err != nil {
		log.Error(ctx, "update study assessment: get teacher ids failed by assessment id ",
			log.Err(err),
			log.String("assessment_id", args.ID),
			log.Any("args", args),
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
		log.Error(ctx, "update h5p assessment: teacher not int assessment",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return constant.ErrForbidden
	}
	if assessment.Status == entity.AssessmentStatusComplete {
		log.Error(ctx, "update h5p assessment: assessment has completed, not allow update",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return ErrAssessmentHasCompleted
	}

	// update assessment students check property
	if args.StudentIDs != nil {
		if err := da.GetAssessmentAttendanceDA().UncheckStudents(ctx, tx, args.ID); err != nil {
			log.Error(ctx, "update h5p assessment: uncheck student failed",
				log.Err(err),
				log.Any("args", args),
			)
			return err
		}
		if args.StudentIDs != nil && len(args.StudentIDs) > 0 {
			if err := da.GetAssessmentAttendanceDA().BatchCheck(ctx, tx, args.ID, args.StudentIDs); err != nil {
				log.Error(ctx, "update h5p assessment: batch check student failed",
					log.Err(err),
					log.Any("args", args),
				)
				return err
			}
		}
	}

	/// update contents
	for _, lm := range args.LessonMaterials {
		updateArgs := da.UpdatePartialAssessmentContentArgs{
			AssessmentID:   args.ID,
			ContentID:      lm.ID,
			ContentComment: lm.Comment,
			Checked:        lm.Checked,
		}
		if err = da.GetAssessmentContentDA().UpdatePartial(ctx, tx, updateArgs); err != nil {
			log.Error(ctx, "update h5p assessment: update assessment content failed",
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
		log.Error(ctx, "update h5p assessment: get plain schedule failed",
			log.Err(err),
			log.String("schedule_id", assessment.ScheduleID),
			log.Any("args", args),
		)
		return err
	}
	if len(schedules) == 0 {
		errMsg := "update h5p assessment: not found schedule"
		log.Error(ctx, errMsg,
			log.String("schedule_id", assessment.ScheduleID),
			log.Any("args", args),
		)
		return errors.New(errMsg)
	}
	schedule := schedules[0]

	// set scores
	var newScores []*external.H5PSetScoreRequest
	for _, item := range args.StudentViewItems {
		for _, lm := range item.LessonMaterials {
			newScore := external.H5PSetScoreRequest{
				RoomID:    schedule.RoomID,
				ContentID: lm.LessonMaterialID,
				StudentID: item.StudentID,
				Score:     lm.AchievedScore,
			}
			newScores = append(newScores, &newScore)
		}
	}
	if _, err := external.GetH5PRoomScoreServiceProvider().BatchSet(ctx, operator, newScores); err != nil {
		return err
	}

	// add comments
	var newComments []*external.H5PAddRoomCommentRequest
	for _, item := range args.StudentViewItems {
		newComment := external.H5PAddRoomCommentRequest{
			RoomID:    schedule.RoomID,
			StudentID: item.StudentID,
			Comment:   item.Comment,
		}
		newComments = append(newComments, &newComment)
	}
	if _, err := external.GetH5PRoomCommentServiceProvider().BatchAdd(ctx, operator, newComments); err != nil {
		return err
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
}

func (m *h5pAssessmentModel) AddClassAndLive(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, args entity.AddAssessmentArgs) (string, error) {
	log.Debug(ctx, "AddClassAndLive: add assessment args", log.Any("args", args), log.Any("operator", operator))

	// clean data
	args.AttendanceIDs = utils.SliceDeduplicationExcludeEmpty(args.AttendanceIDs)

	// check if assessment already exits
	count, err := da.GetAssessmentDA().CountTx(ctx, tx, &da.QueryAssessmentConditions{
		Type: entity.NullAssessmentType{
			Value: entity.AssessmentTypeClassAndLiveH5P,
			Valid: true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: []string{args.ScheduleID},
			Valid:   true,
		},
	}, entity.Assessment{})
	if err != nil {
		log.Error(ctx, "add class and live h5p study assessment: count failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}
	if count > 0 {
		log.Info(ctx, "add class and live h5p study assessment:: assessment already exists",
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", nil
	}

	// get schedule and check class type
	var schedule *entity.SchedulePlain
	if schedule, err = GetScheduleModel().GetPlainByID(ctx, args.ScheduleID); err != nil {
		log.Error(ctx, "AddClassAndLive: GetScheduleModel().GetPlainByID: get failed",
			log.Err(err),
			log.Any("schedule_id", args.ScheduleID),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		switch err {
		case constant.ErrRecordNotFound, dbo.ErrRecordNotFound:
			return "", constant.ErrInvalidArgs
		default:
			return "", err
		}
	}
	if schedule.ClassType == entity.ScheduleClassTypeHomework || schedule.ClassType == entity.ScheduleClassTypeTask {
		log.Info(ctx, "AddClassAndLive: invalid schedule class type",
			log.String("class_type", string(schedule.ClassType)),
			log.Any("schedule", schedule),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", nil
	}

	// fix: permission
	operator.OrgID = schedule.OrgID

	// get contents
	var (
		lessonPlan      *entity.ContentInfoWithDetails
		materialIDs     []string
		materials       []*SubContentsWithName
		materialDetails []*entity.ContentInfoWithDetails
		contents        []*entity.ContentInfoWithDetails
	)
	if lessonPlan, err = GetContentModel().GetVisibleContentByID(ctx, dbo.MustGetDB(ctx), schedule.LessonPlanID, operator); err != nil {
		log.Warn(ctx, "AddClassAndLive: GetContentModel().GetVisibleContentByID: get latest content failed",
			log.Err(err),
			log.Any("args", args),
			log.String("lesson_plan_id", schedule.LessonPlanID),
			log.Any("schedule", schedule),
			log.Any("operator", operator),
		)
	} else {
		contents = append(contents, lessonPlan)
		if materials, err = GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), lessonPlan.ID, operator); err != nil {
			log.Warn(ctx, "AddClassAndLive: GetContentModel().GetContentSubContentsByID: get materials failed",
				log.Err(err),
				log.Any("args", args),
				log.String("latest_lesson_plan_id", lessonPlan.ID),
				log.Any("latest_content", lessonPlan),
				log.Any("operator", operator),
				log.Any("schedule", schedule),
			)
		} else {
			for _, m := range materials {
				materialIDs = append(materialIDs, m.ID)
			}
			materialIDs = utils.SliceDeduplicationExcludeEmpty(materialIDs)
			if materialDetails, err = GetContentModel().GetContentByIDList(ctx, dbo.MustGetDB(ctx), materialIDs, operator); err != nil {
				log.Warn(ctx, "AddClassAndLive: GetContentModel().GetContentByIDList: get contents failed",
					log.Err(err),
					log.Strings("material_ids", materialIDs),
					log.Any("latest_content", lessonPlan),
					log.Any("schedule", schedule),
					log.Any("args", args),
					log.Any("operator", operator),
				)
			} else {
				contents = append(contents, materialDetails...)
			}
		}
	}

	// generate new assessment id
	var newAssessmentID = utils.NewID()

	// add assessment
	var (
		classNameMap  map[string]string
		newAssessment = entity.Assessment{
			ID:           newAssessmentID,
			ScheduleID:   args.ScheduleID,
			Type:         entity.AssessmentTypeClassAndLiveH5P,
			ClassLength:  args.ClassLength,
			ClassEndTime: args.ClassEndTime,
			Status:       entity.AssessmentStatusInProgress,
		}
	)

	if classNameMap, err = external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, []string{schedule.ClassID}); err != nil {
		log.Error(ctx, "Add: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", []string{schedule.ClassID}),
			log.Any("args", args),
		)
		return "", err
	}
	className := classNameMap[schedule.ClassID]
	if className == "" {
		className = constant.AssessmentNoClass
	}
	newAssessment.Title = m.generateTitle(className, schedule.Title)
	if _, err := da.GetAssessmentDA().InsertTx(ctx, tx, &newAssessment); err != nil {
		log.Error(ctx, "add assessment: add failed",
			log.Err(err),
			log.Any("args", args),
			log.Any("new_item", newAssessment),
		)
		return "", err
	}

	// add attendances
	if err := m.addAttendances(ctx, tx, operator, newAssessmentID, schedule, args.AttendanceIDs); err != nil {
		return "", err
	}

	// add contents
	if err = GetAssessmentModel().AddContents(ctx, tx, operator, newAssessmentID, contents); err != nil {
		log.Error(ctx, "Add: GetAssessmentModel().AddContents: add failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Any("contents", contents),
			log.Any("args", args),
			log.Any("operator", operator),
		)
		return "", err
	}

	return newAssessmentID, nil
}

func (m *h5pAssessmentModel) generateTitle(className, lessonName string) string {
	return fmt.Sprintf("%s-%s", className, lessonName)
}

func (m *h5pAssessmentModel) addAttendances(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, newAssessmentID string, schedule *entity.SchedulePlain, attendanceIDs []string) error {
	var finalAttendanceIDs []string
	switch schedule.ClassType {
	case entity.ScheduleClassTypeOfflineClass:
		users, err := GetScheduleRelationModel().GetUsersByScheduleID(ctx, operator, schedule.ID)
		if err != nil {
			log.Error(ctx, "add attendances: get users failed by schedule id",
				log.Err(err),
				log.Any("schedule", schedule),
				log.String("assessment_id", newAssessmentID),
			)
			return err
		}
		for _, u := range users {
			finalAttendanceIDs = append(finalAttendanceIDs, u.RelationID)
		}
	default:
		finalAttendanceIDs = attendanceIDs
	}
	if err := GetAssessmentModel().AddAttendances(ctx, tx, operator, entity.AddAttendancesInput{
		AssessmentID:  newAssessmentID,
		ScheduleID:    schedule.ID,
		AttendanceIDs: finalAttendanceIDs,
	}); err != nil {
		log.Error(ctx, "Add: GetAssessmentModel().AddAttendances: add failed",
			log.Err(err),
			log.String("assessment_id", newAssessmentID),
			log.Strings("attendance_ids", finalAttendanceIDs),
			log.Any("operator", operator),
		)
		return err
	}
	return nil
}

func (m *h5pAssessmentModel) AddStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, input []*entity.AddStudyInput) ([]string, error) {
	log.Debug(ctx, "add studies args", log.Any("input", input), log.Any("operator", operator))

	// check if assessment already exits
	scheduleIDs := make([]string, 0, len(input))
	for _, item := range input {
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}
	count, err := da.GetAssessmentDA().CountTx(ctx, tx, &da.QueryAssessmentConditions{
		Type: entity.NullAssessmentType{
			Value: entity.AssessmentTypeClassAndLiveH5P,
			Valid: true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, entity.Assessment{})
	if err != nil {
		log.Error(ctx, "AddStudies: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Strings("schedule_id", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	if count > 0 {
		log.Info(ctx, "AddStudies: assessment already exists",
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return nil, nil
	}

	// get class name map
	var classIDs []string
	for _, item := range input {
		classIDs = append(classIDs, item.ClassID)
	}
	classNameMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, operator, classIDs)
	if err != nil {
		log.Error(ctx, "AddStudies: external.GetClassServiceProvider().BatchGetNameMap: get failed",
			log.Err(err),
			log.Strings("class_ids", classIDs),
			log.Any("schedule_ids", scheduleIDs),
		)
		return nil, err
	}

	// get contents
	var lessonPlanIDs []string
	for _, item := range input {
		lessonPlanIDs = append(lessonPlanIDs, item.LessonPlanID)
	}
	lessonPlanMap, err := GetAssessmentModel().BatchGetLatestLessonPlanMap(ctx, tx, operator, lessonPlanIDs)
	if err != nil {
		log.Error(ctx, "AddStudies: GetAssessmentModel().BatchGetLatestLessonPlanMap: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", lessonPlanIDs),
		)
		return nil, err
	}

	// add assessment
	newAssessments := make([]*entity.Assessment, 0, len(scheduleIDs))
	now := time.Now().Unix()
	for _, item := range input {
		className := classNameMap[item.ClassID]
		if className == "" {
			className = constant.AssessmentNoClass
		}
		newAssessments = append(newAssessments, &entity.Assessment{
			ID:         utils.NewID(),
			ScheduleID: item.ScheduleID,
			Type:       entity.AssessmentTypeStudyH5P,
			Title:      m.generateTitle(className, item.ScheduleTitle),
			Status:     entity.AssessmentStatusInProgress,
			CreateAt:   now,
			UpdateAt:   now,
		})
	}

	if err := da.GetAssessmentDA().BatchInsert(ctx, tx, newAssessments); err != nil {
		log.Error(ctx, "add studies: add failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("new_assessments", newAssessments),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// add attendances
	scheduleIDToAssessmentIDMap := make(map[string]string, len(newAssessments))
	for _, item := range newAssessments {
		scheduleIDToAssessmentIDMap[item.ScheduleID] = item.ID
	}
	var attendances []*entity.AssessmentAttendance
	for _, item := range input {
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
		log.Error(ctx, "add studies: batch insert attendance failed",
			log.Err(err),
			log.Any("attendances", attendances),
			log.Any("input", input),
		)
		return nil, err
	}

	// add contents
	var (
		scheduleMap        = make(map[string]*entity.AddStudyInput, len(input))
		assessmentContents []*entity.AssessmentContent
	)
	for _, item := range input {
		scheduleMap[item.ScheduleID] = item
	}
	for _, a := range newAssessments {
		schedule := scheduleMap[a.ScheduleID]
		if schedule == nil {
			log.Error(ctx, "add study assessment: not found schedule by id", log.Any("input", input))
			return nil, ErrAssessmentNotFoundSchedule
		}
		lp := lessonPlanMap[schedule.LessonPlanID]
		assessmentContents = append(assessmentContents, &entity.AssessmentContent{
			ID:           utils.NewID(),
			AssessmentID: a.ID,
			ContentID:    lp.ID,
			ContentName:  lp.Name,
			ContentType:  entity.ContentTypePlan,
			Checked:      true,
		})
		for _, lm := range lp.Materials {
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
		log.Error(ctx, "addAssessmentContentsAndOutcomes: da.GetAssessmentContentDA().BatchInsert: batch insert failed",
			log.Err(err),
			log.Any("schedule_ids", scheduleIDs),
			log.Any("assessment_contents", assessmentContents),
			log.Any("operator", operator),
		)
		return nil, err
	}

	var newAssessmentIDs []string
	for _, a := range newAssessments {
		newAssessmentIDs = append(newAssessmentIDs, a.ID)
	}

	return newAssessmentIDs, nil
}

func (m *h5pAssessmentModel) DeleteStudies(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, scheduleIDs []string) error {
	if len(scheduleIDs) == 0 {
		return nil
	}
	var assessments []entity.Assessment
	if err := da.GetAssessmentDA().Query(ctx, &da.QueryAssessmentConditions{
		Type: entity.NullAssessmentType{
			Value: entity.AssessmentTypeClassAndLiveH5P,
			Valid: true,
		},
		OrgID: entity.NullString{
			String: operator.OrgID,
			Valid:  true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, &assessments); err != nil {
		log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().Query: query failed",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Any("operator", operator),
		)
		return err
	}
	assessmentIDs := make([]string, 0, len(assessments))
	for _, a := range assessments {
		assessmentIDs = append(assessmentIDs, a.ID)
	}
	if err := da.GetAssessmentDA().BatchSoftDelete(ctx, tx, assessmentIDs); err != nil {
		log.Error(ctx, "DeleteStudy: da.GetAssessmentDA().BatchSoftDelete",
			log.Err(err),
			log.Strings("schedule_ids", scheduleIDs),
			log.Strings("assessment_ids", assessmentIDs),
		)
		return err
	}
	return nil
}

func (m *h5pAssessmentModel) BatchCheckAnyoneAttempted(ctx context.Context, tx *dbo.DBContext, operator *entity.Operator, roomIDs []string) (map[string]bool, error) {
	if len(roomIDs) == 0 {
		return map[string]bool{}, nil
	}
	roomMap, err := m.getRoomScoreMap(ctx, operator, roomIDs, false)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(roomIDs))
	for _, id := range roomIDs {
		if v := roomMap[id]; v != nil {
			result[id] = v.AnyoneAttempted
		} else {
			result[id] = false
		}
	}
	return result, nil
}

type H5pAssessmentItemsOrder struct {
	Items   []*entity.ListH5PAssessmentsResultItem
	OrderBy entity.AssessmentOrderBy
}

func NewH5pAssessmentItemsOrder(items []*entity.ListH5PAssessmentsResultItem, orderBy entity.AssessmentOrderBy) *H5pAssessmentItemsOrder {
	return &H5pAssessmentItemsOrder{Items: items, OrderBy: orderBy}
}

func (h *H5pAssessmentItemsOrder) Len() int {
	return len(h.Items)
}

func (h *H5pAssessmentItemsOrder) Less(i, j int) bool {
	switch h.OrderBy {
	case entity.AssessmentOrderByCompleteTime:
		if h.Items[i].CompleteAt == 0 && h.Items[j].CompleteAt > 0 {
			return true
		}
		if h.Items[i].CompleteAt != 0 && h.Items[j].CompleteAt == 0 {
			return false
		}
		if h.Items[i].CompleteAt == 0 && h.Items[j].CompleteAt == 0 {
			return h.Items[i].CreateAt < h.Items[j].CreateAt
		}
	case entity.AssessmentOrderByCompleteTimeDesc:
		if h.Items[i].CompleteAt == 0 && h.Items[j].CompleteAt > 0 {
			return false
		}
		if h.Items[i].CompleteAt != 0 && h.Items[j].CompleteAt == 0 {
			return true
		}
		if h.Items[i].CompleteAt == 0 && h.Items[j].CompleteAt == 0 {
			return h.Items[i].CreateAt > h.Items[j].CreateAt
		}
	case entity.AssessmentOrderByCreateAt:
		return h.Items[i].CreateAt < h.Items[j].CreateAt
	case entity.AssessmentOrderByCreateAtDesc:
		return h.Items[i].CreateAt > h.Items[j].CreateAt
	}
	return false
}

func (h *H5pAssessmentItemsOrder) Swap(i, j int) {
	h.Items[i], h.Items[j] = h.Items[j], h.Items[i]
}
