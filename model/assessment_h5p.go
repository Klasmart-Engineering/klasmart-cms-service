package model

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type assessmentH5P struct{}

func getAssessmentH5P() *assessmentH5P {
	return &assessmentH5P{}
}

func (m *assessmentH5P) batchGetRoomMap(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]*entity.AssessmentH5PRoom, error) {
	// batch get room score map
	roomScoreMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "batch get room map: batch get scores failed",
			log.Err(err),
			log.Strings("room_ids", roomIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// mapping
	result := make(map[string]*entity.AssessmentH5PRoom, len(roomScoreMap))
	for roomID, users := range roomScoreMap {
		room := entity.AssessmentH5PRoom{}

		// fill users
		assessmentUsers := make([]*entity.AssessmentH5PUser, 0, len(users))
		for _, u := range users {
			assessmentUser := entity.AssessmentH5PUser{}

			// fill user id
			if u.User != nil {
				assessmentUser.UserID = u.User.UserID
			}

			// fill contents
			assessmentContents := make([]*entity.AssessmentH5PContent, 0, len(u.Scores))
			for _, s := range u.Scores {
				assessmentContent := entity.AssessmentH5PContent{
					Seen: s.Seen,
				}
				if s.Content != nil {
					assessmentContent.ParentID = s.Content.ParentID
					assessmentContent.H5PID = s.Content.H5PID
					assessmentContent.SubH5PID = s.Content.SubContentID
					assessmentContent.ContentID = s.Content.ContentID
					assessmentContent.ContentName = s.Content.Name
					assessmentContent.ContentType = s.Content.Type
				}
				if s.Score != nil {
					assessmentContent.Scores = s.Score.Scores
					for _, a := range s.Score.Answers {
						if a == nil {
							continue
						}
						assessmentContent.Answers = append(assessmentContent.Answers, &entity.AssessmentH5PAnswer{
							Answer:               a.Answer,
							Score:                a.Score,
							MinimumPossibleScore: a.MinimumPossibleScore,
							MaximumPossibleScore: a.MaximumPossibleScore,
							Date:                 a.Date,
						})
					}
				}
				for _, ts := range s.TeacherScores {
					if ts == nil {
						continue
					}
					item := entity.AssessmentH5PTeacherScore{
						TeacherID: ts.Teacher.UserID,
						Score:     ts.Score,
						Date:      ts.Date,
					}
					if ts.Teacher != nil {
						item.TeacherID = ts.Teacher.UserID
					}
					assessmentContent.TeacherScores = append(assessmentContent.TeacherScores, &item)
				}

				assessmentContents = append(assessmentContents, &assessmentContent)
			}
			assessmentUser.Contents = assessmentContents

			// append user
			assessmentUsers = append(assessmentUsers, &assessmentUser)
		}
		room.Users = assessmentUsers

		// fill room
		result[roomID] = &room
	}

	// fill ordered id
	latestOrderedID := 1
	for _, r := range result {
		for _, u := range r.Users {
			for _, c := range u.Contents {
				c.OrderedID = latestOrderedID
				latestOrderedID++
			}
		}
	}

	log.Debug(ctx, "batch get room map",
		log.Strings("room_ids", roomIDs),
		log.Any("operator", operator),
		log.Any("result", result),
	)

	return result, nil
}

func (m *assessmentH5P) sortContentsByOrderedID(contents []*entity.AssessmentH5PContent) {
	sort.Slice(contents, func(i, j int) bool {
		return contents[i].OrderedID < contents[j].OrderedID
	})
	for _, c := range contents {
		if len(c.Children) > 0 {
			m.sortContentsByOrderedID(c.Children)
		}
	}
}

func (m *assessmentH5P) getAnswer(content *entity.AssessmentH5PContent) string {
	if len(content.Answers) > 0 {
		return content.Answers[0].Answer
	}
	return ""
}

func (m *assessmentH5P) getAchievedScore(content *entity.AssessmentH5PContent) float64 {
	if len(content.TeacherScores) > 0 {
		return content.TeacherScores[len(content.TeacherScores)-1].Score
	}
	if len(content.Scores) > 0 {
		return content.Scores[0]
	}
	return 0
}

func (m *assessmentH5P) getMaxPossibleScore(content *entity.AssessmentH5PContent) float64 {
	if len(content.Answers) > 0 {
		return content.Answers[0].MaximumPossibleScore
	}
	return 0
}

func (m *assessmentH5P) isAnyoneAttempted(room *entity.AssessmentH5PRoom) bool {
	for _, u := range room.Users {
		for _, c := range u.Contents {
			if len(c.Answers) > 0 || len(c.Scores) > 0 {
				return true
			}
		}
	}
	return false
}

func (m *assessmentH5P) getUserMap(room *entity.AssessmentH5PRoom) map[string]*entity.AssessmentH5PUser {
	if room == nil {
		return map[string]*entity.AssessmentH5PUser{}
	}
	result := make(map[string]*entity.AssessmentH5PUser, len(room.Users))
	for _, u := range room.Users {
		if u == nil || u.UserID == "" {
			continue
		}
		result[u.UserID] = u
	}
	return result
}

func (m *assessmentH5P) getContentsMapByContentID(user *entity.AssessmentH5PUser) map[string][]*entity.AssessmentH5PContent {
	if user == nil {
		return map[string][]*entity.AssessmentH5PContent{}
	}
	result := make(map[string][]*entity.AssessmentH5PContent, len(user.Contents))
	for _, c := range user.Contents {
		if c == nil || c.ContentID == "" {
			continue
		}
		result[c.ContentID] = append(result[c.ContentID], c)
	}
	return result
}

func (m *assessmentH5P) getStudentViewItems(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, view *entity.AssessmentView) ([]*entity.AssessmentStudentViewH5PItem, error) {
	var roomIDs = []string{view.RoomID}
	// get room
	roomMap, err := m.batchGetRoomMap(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "get student view items: batch get room map failed",
			log.Err(err),
			log.String("room_id", view.RoomID),
			log.Any("view", view),
		)
		return nil, err
	}
	room := roomMap[view.RoomID]
	if room == nil {
		log.Debug(ctx, "get student view items: not found room", log.String("room_id", view.RoomID))
		return nil, nil
	}

	// batch get students lesson materials map
	studentLessonMaterialsMap, err := m.batchGetStudentViewH5PLessonMaterialsMap(ctx, operator, tx, view, room)

	// get room comments
	roomCommentMap, err := m.batchGetRoomCommentMap(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "batch get comments failed",
			log.Err(err),
			log.Strings("room_ids", roomIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// assembly result
	result := make([]*entity.AssessmentStudentViewH5PItem, 0, len(view.Students))
	for _, s := range view.Students {
		newItem := entity.AssessmentStudentViewH5PItem{
			StudentID:   s.ID,
			StudentName: s.Name,
		}

		// fill comment
		if roomComment, ok := roomCommentMap[view.RoomID]; ok {
			if comments, ok := roomComment[s.ID]; ok {
				if len(comments) > 0 {
					newItem.Comment = comments[len(comments)-1]
				}
			}
		}

		// fill lesson materials
		newItem.LessonMaterials = studentLessonMaterialsMap[s.ID]

		// append item
		result = append(result, &newItem)
	}

	// sort result by student name
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].StudentName) < strings.ToLower(result[j].StudentName)
	})

	return result, nil
}

func (m *assessmentH5P) getKeyedH5PContentsTemplateMap(room *entity.AssessmentH5PRoom, contentID string) map[string][]*entity.AssessmentH5PContent {
	keyedH5PContentsMap := map[string][]*entity.AssessmentH5PContent{}
	for _, u := range room.Users {
		for _, c := range u.Contents {
			if c == nil || c.ContentID == "" {
				continue
			}
			if c.ContentID != contentID {
				continue
			}
			key := m.generateH5PContentKey(c.ContentID, c.SubH5PID)
			keyedH5PContentsMap[key] = append(keyedH5PContentsMap[key], c)
		}
	}
	return keyedH5PContentsMap
}

func (m *assessmentH5P) generateH5PContentKey(contentID string, subH5PID string) string {
	return strings.Join([]string{contentID, subH5PID}, ":")
}

func (m *assessmentH5P) getKeyedUserH5PContentsMap(room *entity.AssessmentH5PRoom) map[string]*entity.AssessmentH5PContent {
	keyedUserH5PContentsMap := map[string]*entity.AssessmentH5PContent{}
	for _, u := range room.Users {
		for _, c := range u.Contents {
			if c == nil || c.ContentID == "" {
				continue
			}
			key := m.generateUserH5PContentKey(c.ContentID, c.SubH5PID, u.UserID)
			keyedUserH5PContentsMap[key] = c
		}
	}
	return keyedUserH5PContentsMap
}

func (m *assessmentH5P) generateUserH5PContentKey(contentID string, subH5PID string, userID string) string {
	return strings.Join([]string{contentID, subH5PID, userID}, ":")
}

func (m *assessmentH5P) batchGetStudentViewH5PLessonMaterialsMap(
	ctx context.Context,
	operator *entity.Operator,
	tx *dbo.DBContext,
	view *entity.AssessmentView,
	room *entity.AssessmentH5PRoom,
) (map[string][]*entity.AssessmentStudentViewH5PLessonMaterial, error) {
	// get assessment outcomes
	assessmentOutcomes, err := m.getAssessmentOutcomes(ctx, view.ID)
	if err != nil {
		log.Error(ctx, "batch get student view h5p lesson materials map: get assessment outcomes map failed",
			log.Err(err),
			log.Any("assessment_id", view.ID),
		)
		return nil, err
	}

	// get assessment outcomes map
	assessmentOutcomeMap := map[string]*entity.AssessmentOutcome{}
	for _, ao := range assessmentOutcomes {
		assessmentOutcomeMap[ao.OutcomeID] = ao
	}

	// get assessment content outcomes
	assessmentContentOutcomes, err := m.getAssessmentContentOutcomes(ctx, view.ID)
	if err != nil {
		log.Error(ctx, "batch get student view h5p lesson materials map: get assessment content outcomes map failed",
			log.Err(err),
			log.Any("assessment_id", view.ID),
		)
		return nil, err
	}

	// assessment content outcomes map
	assessmentContentOutcomesMap := map[string][]*entity.AssessmentContentOutcome{}
	for _, item := range assessmentContentOutcomes {
		assessmentContentOutcomesMap[item.ContentID] = append(assessmentContentOutcomesMap[item.ContentID], item)
	}

	// get outcomes map
	outcomeIDs := make([]string, 0, len(assessmentOutcomes))
	for _, ao := range assessmentOutcomes {
		outcomeIDs = append(outcomeIDs, ao.OutcomeID)
	}
	outcomeMap, err := m.getOutcomeMap(ctx, operator, tx, outcomeIDs)
	if err != nil {
		log.Error(ctx, "batch get student view h5p lesson materials map: get outcomes map failed",
			log.Err(err),
			log.Any("outcome_ids", outcomeIDs),
		)
		return nil, err
	}

	// get attendance ids
	attendanceIDs := make([]string, len(view.Students))
	for _, s := range view.Students {
		attendanceIDs = append(attendanceIDs, s.ID)
	}
	attendanceIDs = utils.SliceDeduplicationExcludeEmpty(attendanceIDs)

	// get attendance content outcome exists map
	contentOutcomeAttendanceExistsMap, err := m.getContentOutcomeAttendanceExistsMap(ctx, view.ID, attendanceIDs, assessmentContentOutcomes)
	if err != nil {
		log.Error(ctx, "batch get student view h5p lesson materials map: get outcomes map failed",
			log.Err(err),
			log.Strings("attendance_ids", attendanceIDs),
			log.Any("assessment_content_outcomes", assessmentContentOutcomes),
		)
		return nil, err
	}

	attendanceContentOutcomesMap, err := m.getAttendanceContentOutcomesMap(
		ctx,
		attendanceIDs,
		assessmentContentOutcomes,
		assessmentOutcomeMap,
		outcomeMap,
		contentOutcomeAttendanceExistsMap,
	)
	if err != nil {
		log.Error(ctx, "batch get student view h5p lesson materials map: get outcomes map failed",
			log.Err(err),
			log.Strings("attendance_ids", attendanceIDs),
			log.Any("assessment_content_outcomes", assessmentContentOutcomes),
			log.Any("assessment_outcomes_map", assessmentOutcomeMap),
			log.Any("outcome_map", outcomeMap),
			log.Any("content_outcome_attendance_exists_map", contentOutcomeAttendanceExistsMap),
		)
		return nil, err
	}

	// get keyed user h5p contents map
	keyedUserH5PContentsMap := m.getKeyedUserH5PContentsMap(room)

	// assembly result
	result := map[string][]*entity.AssessmentStudentViewH5PLessonMaterial{}
	for _, s := range view.Students {
		for lmIndex, lm := range view.LessonMaterials {
			keyedH5PContentsTemplateMap := m.getKeyedH5PContentsTemplateMap(room, lm.ID)
			var contents []*entity.AssessmentH5PContent
			for _, keyedContents := range keyedH5PContentsTemplateMap {
				if len(keyedContents) == 0 {
					continue
				}
				findUserContent := false
				for _, c := range keyedContents {
					userContent := keyedUserH5PContentsMap[m.generateUserH5PContentKey(c.ContentID, c.SubH5PID, s.ID)]
					if userContent != nil {
						findUserContent = true
						contents = append(contents, userContent)
						break
					}
				}
				if !findUserContent {
					c := keyedContents[0]
					newContent := entity.AssessmentH5PContent{
						OrderedID:   c.OrderedID,
						ParentID:    c.ParentID,
						H5PID:       c.H5PID,
						SubH5PID:    c.SubH5PID,
						ContentID:   c.ContentID,
						ContentName: c.ContentName,
					}
					contents = append(contents, &newContent)
				}
			}
			if len(contents) == 0 {
				newLessonMaterial := entity.AssessmentStudentViewH5PLessonMaterial{
					LessonMaterialOrderedNumber: lmIndex,
					LessonMaterialID:            lm.ID,
					LessonMaterialName:          lm.Name,
					IsH5P:                       lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend,
				}
				if attendanceContentOutcomesMap[s.ID] != nil {
					newLessonMaterial.Outcomes = attendanceContentOutcomesMap[s.ID][lm.ID]
				}
				result[s.ID] = append(result[s.ID], &newLessonMaterial)
				continue
			}
			for _, content := range contents {
				if content == nil {
					log.Debug(ctx, "batch get student view h5p lesson materials map: content is nil",
						log.Any("lesson_material", lm),
						log.Any("room", room),
					)
					continue
				}
				newLessonMaterial := entity.AssessmentStudentViewH5PLessonMaterial{
					OrderedID:                   content.OrderedID,
					ParentID:                    content.ParentID,
					H5PID:                       content.H5PID,
					SubH5PID:                    content.SubH5PID,
					LessonMaterialOrderedNumber: lmIndex,
					LessonMaterialID:            lm.ID,
					LessonMaterialName:          lm.Name,
					LessonMaterialType:          content.ContentType,
					Answer:                      getAssessmentH5P().getAnswer(content),
					MaxScore:                    getAssessmentH5P().getMaxPossibleScore(content),
					AchievedScore:               getAssessmentH5P().getAchievedScore(content),
					Attempted:                   content.Seen,
					IsH5P:                       lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend,
					NotApplicableScoring:        !getAssessmentH5P().canScoring(content.ContentType),
				}
				if attendanceContentOutcomesMap[s.ID] != nil {
					newLessonMaterial.Outcomes = attendanceContentOutcomesMap[s.ID][lm.ID]
				}
				result[s.ID] = append(result[s.ID], &newLessonMaterial)
			}
		}
	}

	// number lesson materials
	for _, lessonMaterials := range result {
		m.numberAndFlagStudentViewH5PLessonMaterials(ctx, view, lessonMaterials)
		m.sortNumberedStudentViewH5PLessonMaterials(ctx, lessonMaterials)
	}

	return result, nil
}

func (m *assessmentH5P) getAttendanceContentOutcomesMap(ctx context.Context,
	attendanceIDs []string,
	assessmentContentOutcomes []*entity.AssessmentContentOutcome,
	assessmentOutcomeMap map[string]*entity.AssessmentOutcome,
	outcomeMap map[string]*entity.Outcome,
	contentOutcomeAttendanceExistsMap map[[3]string]bool,
) (map[string]map[string][]*entity.AssessmentDetailContentOutcome, error) {
	result := map[string]map[string][]*entity.AssessmentDetailContentOutcome{}
	for _, attendanceID := range attendanceIDs {
		if result[attendanceID] == nil {
			result[attendanceID] = map[string][]*entity.AssessmentDetailContentOutcome{}
		}
		for _, co := range assessmentContentOutcomes {
			o := outcomeMap[co.OutcomeID]
			if o == nil {
				continue
			}
			ao := assessmentOutcomeMap[co.OutcomeID]
			if ao == nil {
				continue
			}
			result[attendanceID][co.ContentID] = append(result[attendanceID][co.ContentID], &entity.AssessmentDetailContentOutcome{
				ContentID:    co.ContentID,
				OutcomeID:    co.OutcomeID,
				OutcomeName:  o.Name,
				Assumed:      o.Assumed,
				NoneAchieved: co.NoneAchieved,
				Checked:      contentOutcomeAttendanceExistsMap[[3]string{co.ContentID, co.OutcomeID, attendanceID}],
			})
		}
	}
	log.Debug(ctx, "get attendance content outcomes map result",
		log.Any("result", result),
		log.Strings("attendance_ids", attendanceIDs),
		log.Any("assessment_content_outcomes", assessmentContentOutcomes),
		log.Any("assessment_outcome_map", assessmentOutcomeMap),
		log.Any("outcome_map", outcomeMap),
		log.Any("content_outcome_attendance_exists_map", contentOutcomeAttendanceExistsMap),
	)
	return result, nil
}

func (m *assessmentH5P) getContentOutcomeAttendanceExistsMap(ctx context.Context, assessmentID string, attendanceIDs []string, assessmentContentOutcomes []*entity.AssessmentContentOutcome) (map[[3]string]bool, error) {
	if len(assessmentContentOutcomes) == 0 {
		log.Debug(ctx, "get content outcome attendance exists map: empty assessment content outcomes")
		return map[[3]string]bool{}, nil
	}
	keys := make([]*da.ContentIDAndOutcomeIDKey, 0, len(assessmentContentOutcomes))
	for _, co := range assessmentContentOutcomes {
		keys = append(keys, &da.ContentIDAndOutcomeIDKey{
			ContentID: co.ContentID,
			OutcomeID: co.OutcomeID,
		})
	}
	queryAssessmentContentOutcomeAttendanceCond := da.QueryAssessmentContentOutcomeAttendanceCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{assessmentID},
			Valid:   true,
		},
		ContentIDAndOutcomeIDPairs: da.NullContentIDAndOutcomeIDKeys{
			Value: keys,
			Valid: true,
		},
		AttendanceIDs: entity.NullStrings{
			Strings: attendanceIDs,
			Valid:   true,
		},
	}
	var contentOutcomeAttendances []*entity.AssessmentContentOutcomeAttendance
	if err := da.GetAssessmentContentOutcomeAttendanceDA().Query(ctx, &queryAssessmentContentOutcomeAttendanceCond, &contentOutcomeAttendances); err != nil {
		log.Error(ctx, "get content outcomes attendance exists map: query assessment content outcome attendance failed",
			log.Err(err),
			log.Any("cond", queryAssessmentContentOutcomeAttendanceCond),
		)
		return nil, err
	}
	contentOutcomeAttendanceExistsMap := map[[3]string]bool{}
	for _, coa := range contentOutcomeAttendances {
		key := [3]string{coa.ContentID, coa.OutcomeID, coa.AttendanceID}
		contentOutcomeAttendanceExistsMap[key] = true
	}
	return contentOutcomeAttendanceExistsMap, nil
}

func (m *assessmentH5P) numberAndFlagStudentViewH5PLessonMaterials(ctx context.Context, view *entity.AssessmentView, lessonMaterials []*entity.AssessmentStudentViewH5PLessonMaterial) {
	// sort by cms lesson materials
	lmIndexMap := make(map[string]int, len(view.LessonMaterials))
	for i, lm := range view.LessonMaterials {
		lmIndexMap[lm.ID] = i
	}
	sort.Slice(lessonMaterials, func(i, j int) bool {
		itemI := lessonMaterials[i]
		itemJ := lessonMaterials[j]
		if itemI.LessonMaterialID != itemJ.LessonMaterialID {
			return lmIndexMap[itemI.LessonMaterialID] < lmIndexMap[itemJ.LessonMaterialID]
		}
		return itemI.OrderedID < itemJ.OrderedID
	})

	// sort by tree level
	treedLessonMaterials := m.treeingStudentViewLessonMaterials(lessonMaterials)
	m.sortTreedStudentViewH5PLessonMaterials(treedLessonMaterials)
	m.doNumberStudentViewH5PLessonMaterials(treedLessonMaterials, "")
	m.flagHasSubItems(treedLessonMaterials)
}

func (m *assessmentH5P) flagHasSubItems(treedLessonMaterials []*entity.AssessmentStudentViewH5PLessonMaterial) {
	for _, lm := range treedLessonMaterials {
		if lm.ParentID == "" && len(lm.Children) > 0 {
			lm.HasSubItems = true
		}
	}
}

func (m *assessmentH5P) sortNumberedStudentViewH5PLessonMaterials(ctx context.Context, lessonMaterials []*entity.AssessmentStudentViewH5PLessonMaterial) {
	sort.Slice(lessonMaterials, func(i, j int) bool {
		a := strings.Split(lessonMaterials[i].Number, "-")
		b := strings.Split(lessonMaterials[j].Number, "-")
		min := int(math.Min(float64(len(a)), float64(len(b))))
		for i := 0; i < min; i++ {
			s1 := fmt.Sprintf("%06s", a[i])
			s2 := fmt.Sprintf("%06s", b[i])
			if s1 != s2 {
				return s1 < s2
			}
		}
		return len(a) < len(b)
	})
}

func (m *assessmentH5P) doNumberStudentViewH5PLessonMaterials(treedLessonMaterials []*entity.AssessmentStudentViewH5PLessonMaterial, prefix string) {
	for i, lm := range treedLessonMaterials {
		if len(prefix) > 0 {
			lm.Number = fmt.Sprintf("%s-%d", prefix, i+1)
		} else {
			lm.Number = fmt.Sprintf("%d", i+1)
		}
		if len(lm.Children) > 0 {
			m.doNumberStudentViewH5PLessonMaterials(lm.Children, lm.Number)
		}
	}
}

func (m *assessmentH5P) sortTreedStudentViewH5PLessonMaterials(treedLessonMaterials []*entity.AssessmentStudentViewH5PLessonMaterial) {
	sort.Slice(treedLessonMaterials, func(i, j int) bool {
		if treedLessonMaterials[i].LessonMaterialOrderedNumber != treedLessonMaterials[j].LessonMaterialOrderedNumber {
			return treedLessonMaterials[i].LessonMaterialOrderedNumber < treedLessonMaterials[j].LessonMaterialOrderedNumber
		}
		return treedLessonMaterials[i].OrderedID < treedLessonMaterials[j].OrderedID
	})
	for _, lm := range treedLessonMaterials {
		if len(lm.Children) > 0 {
			m.sortTreedStudentViewH5PLessonMaterials(lm.Children)
		}
	}
}

func (m *assessmentH5P) treeingStudentViewLessonMaterials(contents []*entity.AssessmentStudentViewH5PLessonMaterial) []*entity.AssessmentStudentViewH5PLessonMaterial {
	var rootContents []*entity.AssessmentStudentViewH5PLessonMaterial
	for _, c := range contents {
		if c.ParentID == "" {
			rootContents = append(rootContents, c)
		}
	}

	var level2Contents []*entity.AssessmentStudentViewH5PLessonMaterial
	for _, root := range rootContents {
		for _, c := range contents {
			if c.ParentID == root.H5PID && c.ParentID != "" {
				root.Children = append(root.Children, c)
				level2Contents = append(level2Contents, c)
			}
		}
	}

	m.treeingRemainingStudentViewLessonMaterials(contents, level2Contents)

	return rootContents
}

//  treeingRemainingStudentViewLessonMaterials only apply to level 2
func (m *assessmentH5P) treeingRemainingStudentViewLessonMaterials(contents []*entity.AssessmentStudentViewH5PLessonMaterial, parentContents []*entity.AssessmentStudentViewH5PLessonMaterial) {
	for _, parent := range parentContents {
		var subContents []*entity.AssessmentStudentViewH5PLessonMaterial
		for _, c := range contents {
			if c.ParentID == parent.SubH5PID && c.ParentID != "" {
				parent.Children = append(parent.Children, c)
				subContents = append(subContents, c)
			}
		}
		if len(subContents) > 0 {
			m.treeingRemainingStudentViewLessonMaterials(contents, subContents)
		}
	}
}

func (m *assessmentH5P) getOutcomeMap(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, outcomeIDs []string) (map[string]*entity.Outcome, error) {
	outcomeIDs = utils.SliceDeduplicationExcludeEmpty(outcomeIDs)
	if len(outcomeIDs) == 0 {
		return map[string]*entity.Outcome{}, nil
	}
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, operator, tx, outcomeIDs)
	if err != nil {
		log.Error(ctx, "get outcome map: batch get outcomes failed",
			log.Err(err),
			log.Strings("outcome_ids", outcomeIDs),
		)
		return nil, err
	}
	outcomeMap := make(map[string]*entity.Outcome, len(outcomes))
	for _, outcome := range outcomes {
		outcomeMap[outcome.ID] = outcome
	}
	return outcomeMap, nil
}

func (m *assessmentH5P) getAssessmentOutcomes(ctx context.Context, assessmentID string) ([]*entity.AssessmentOutcome, error) {
	cond := da.QueryAssessmentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{assessmentID},
			Valid:   true,
		},
	}
	var assessmentOutcomes []*entity.AssessmentOutcome
	if err := da.GetAssessmentOutcomeDA().Query(ctx, &cond, &assessmentOutcomes); err != nil {
		log.Error(ctx, "get assessment outcomes failed",
			log.String("assessment_id", assessmentID),
		)
		return nil, err
	}
	return assessmentOutcomes, nil
}

func (m *assessmentH5P) getAssessmentContentOutcomes(ctx context.Context, assessmentID string) ([]*entity.AssessmentContentOutcome, error) {
	if assessmentID == "" {
		return nil, nil
	}
	var contentOutcomes []*entity.AssessmentContentOutcome
	if err := da.GetAssessmentContentOutcomeDA().Query(ctx, &da.QueryAssessmentContentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{assessmentID},
			Valid:   true,
		},
	}, &contentOutcomes); err != nil {
		log.Error(ctx, "get assessment content outcomes map: query assessment content outcomes failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return nil, err
	}
	return contentOutcomes, nil
}

func (m *assessmentH5P) getAssessmentContentOutcomesMap(ctx context.Context, assessmentID string) (map[string][]*entity.AssessmentContentOutcome, error) {
	if assessmentID == "" {
		return map[string][]*entity.AssessmentContentOutcome{}, nil
	}
	var contentOutcomes []*entity.AssessmentContentOutcome
	if err := da.GetAssessmentContentOutcomeDA().Query(ctx, &da.QueryAssessmentContentOutcomeConditions{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{assessmentID},
			Valid:   true,
		},
	}, &contentOutcomes); err != nil {
		log.Error(ctx, "get assessment content outcomes map: query assessment content outcomes failed",
			log.Err(err),
			log.String("assessment_id", assessmentID),
		)
		return nil, err
	}
	contentOutcomesMap := map[string][]*entity.AssessmentContentOutcome{}
	for _, co := range contentOutcomes {
		contentOutcomesMap[co.ContentID] = append(contentOutcomesMap[co.ContentID], co)
	}
	return contentOutcomesMap, nil
}

func (m *assessmentH5P) getContentOutcomesMap(
	outcomeMap map[string]*entity.Outcome,
	assessmentContentOutcomesMap map[string][]*entity.AssessmentContentOutcome,
) (map[string][]*entity.Outcome, error) {
	result := map[string][]*entity.Outcome{}
	for contentID, contentOutcomes := range assessmentContentOutcomesMap {
		for _, co := range contentOutcomes {
			o := outcomeMap[co.OutcomeID]
			if o != nil {
				result[contentID] = append(result[contentID], o)
			}
		}
	}
	return result, nil
}

func (m *assessmentH5P) batchGetRoomCommentMap(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]map[string][]string, error) {
	commentMap, err := external.GetH5PRoomCommentServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "batch get room comment map failed",
			log.Strings("room_ids", roomIDs),
			log.Any("operator", operator),
		)
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

func (m *assessmentH5P) batchGetRoomCommentObjectMap(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]map[string][]*entity.H5PRoomComment, error) {
	commentMap, err := external.GetH5PRoomCommentServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "batch get room comment object map failed",
			log.Strings("room_ids", roomIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}
	result := make(map[string]map[string][]*entity.H5PRoomComment, len(commentMap))
	for roomID, users := range commentMap {
		result[roomID] = make(map[string][]*entity.H5PRoomComment, len(users))
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
				comment := entity.H5PRoomComment{
					Comment: c.Comment,
				}
				if c.Teacher != nil {
					comment.TeacherID = c.Teacher.UserID
					comment.GivenName = c.Teacher.GivenName
					comment.FamilyName = c.Teacher.FamilyName
				}
				result[roomID][uid] = append(result[roomID][uid], &comment)
			}
		}
	}
	log.Debug(ctx, "batch get room comment object map",
		log.Any("result", result),
		log.Strings("room_ids", roomIDs),
	)
	return result, nil
}

func (m *assessmentH5P) calcRoomCompleteRate(ctx context.Context, room *entity.AssessmentH5PRoom, view *entity.AssessmentView) float64 {
	if room == nil || view == nil {
		log.Debug(ctx, "calc room complete rate: invalid args",
			log.Any("view", view),
			log.Any("room", room),
		)
		return 0
	}

	keyedUserH5PContentsMap := m.getKeyedUserH5PContentsMap(room)
	attempted := 0
	total := 0
	for _, s := range view.Students {
		if !s.Checked {
			continue
		}
		for _, lm := range view.LessonMaterials {
			if !(lm.Checked && (lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend)) {
				continue
			}
			var contents []*entity.AssessmentH5PContent
			for _, keyedContents := range m.getKeyedH5PContentsTemplateMap(room, lm.ID) {
				if len(keyedContents) == 0 {
					continue
				}
				findUserContent := false
				for _, c := range keyedContents {
					userContent := keyedUserH5PContentsMap[m.generateUserH5PContentKey(c.ContentID, c.SubH5PID, s.ID)]
					if userContent != nil {
						findUserContent = true
						contents = append(contents, userContent)
						break
					}
				}
				if !findUserContent {
					c := keyedContents[0]
					newContent := entity.AssessmentH5PContent{
						OrderedID:   c.OrderedID,
						H5PID:       c.H5PID,
						ContentID:   c.ContentID,
						ContentName: c.ContentName,
						SubH5PID:    c.SubH5PID,
					}
					contents = append(contents, &newContent)
				}
			}
			if len(contents) == 0 {
				total++
				continue
			}
			for _, content := range contents {
				if content == nil {
					log.Debug(ctx, "calc room complete rate: not found content",
						log.String("room_id", view.RoomID),
						log.Any("view", view),
						log.Any("room", room),
					)
					continue
				}
				if len(content.Answers) > 0 || len(content.Scores) > 0 {
					attempted++
				}
				total++
			}
		}
	}

	log.Debug(ctx, "calc room complete rate",
		log.Int("attempted", attempted),
		log.Int("total", total),
		log.Any("room", room),
		log.Any("view", view),
	)

	if total > 0 {
		return float64(attempted) / float64(total)
	}

	return 0
}

var canScoringMap = map[string]bool{
	"Accordion":                    false,
	"AdvancedBlanks":               true,
	"AdventCalendar":               false,
	"Agamotto":                     false,
	"ArithmeticQuiz":               true,
	"Audio":                        false,
	"AudioRecorder":                false,
	"BookMaker":                    false,
	"BranchingScenario":            true,
	"Chart":                        false,
	"Collage":                      false,
	"Column":                       true,
	"CoursePresentationKID":        true,
	"Dialogcards":                  false,
	"Dictation":                    true,
	"DocumentationTool":            false,
	"DragQuestion":                 true,
	"DragText":                     true,
	"Essay":                        true,
	"Blanks":                       true,
	"ImageMultipleHotspotQuestion": true,
	"ImageHotspotQuestion":         true,
	"FindTheWords":                 true,
	"Flashcards":                   true,
	"GuessTheAnswer":               false,
	"IFrameEmbed":                  false,
	"ImageHotspots":                false,
	"ImagePair":                    true,
	"ImageSequencing":              true,
	"ImageJuxtaposition":           false,
	"ImageSlider":                  false,
	"ImpressPresentation":          false,
	"InteractiveBook":              true,
	"InteractiveVideo":             true,
	"JigsawPuzzleKID":              true,
	"KewArCode":                    false,
	"MarkTheWords":                 true,
	"MemoryGame":                   true,
	"MultipleChoice":               true,
	"PersonalityQuiz":              false,
	"Questionnaire":                false,
	"QuestionSet":                  true,
	"Shape":                        false,
	"SingleChoiceSet":              true,
	"SpeakTheWords":                true,
	"SpeqkTheWordsSet":             true,
	"Summary":                      true,
	"Timeline":                     false,
	"TrueFalse":                    true,
	"TwitterUserFeed":              false,
	"ThreeImage":                   false,
}

func (m *assessmentH5P) canScoring(contentType string) bool {
	if v, ok := canScoringMap[contentType]; ok {
		return v
	}
	return true
}
