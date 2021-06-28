package model

import (
	"context"
	"fmt"
	"sort"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type assessmentH5P struct{}

func getAssessmentH5p() *assessmentH5P {
	return &assessmentH5P{}
}

func (m *assessmentH5P) getRoomCompleteRate(ctx context.Context, room *entity.AssessmentH5PRoom, view *entity.AssessmentView) float64 {
	if room == nil {
		log.Debug(ctx, "get room complete rate: room is empty",
			log.Any("view", view),
		)
		return 0
	}

	// calc agg template
	aggContentsMap := map[string][]*entity.AssessmentH5PContentScore{}
	aggUserContentOrderedIDsMap := map[string][]int{}
	for _, u := range room.Users {
		for id, contents := range u.ContentsMapByContentID {
			for _, c := range contents {
				// aggregate user attended contents
				if u.UserID != "" {
					aggUserContentOrderedIDsMap[u.UserID] = append(aggUserContentOrderedIDsMap[u.UserID], c.OrderedID)
				}
				exists := false
				for _, c2 := range aggContentsMap[id] {
					if c2 == c {
						exists = true
						break
					}
				}
				// deduplication, only append not exists item
				if !exists {
					aggContentsMap[id] = append(aggContentsMap[id], c)
				}
			}
		}
	}
	log.Debug(ctx, "get room complete rate: print agg maps",
		log.Any("agg_contents_map", aggContentsMap),
		log.Any("agg_user_content_ordered_ids_map", aggUserContentOrderedIDsMap),
	)

	attempted := 0
	total := 0
	for _, s := range view.Students {
		if !s.Checked {
			continue
		}
		user := room.UserMap[s.ID]
		uid := ""
		if user != nil {
			uid = user.UserID
		}
		for _, lm := range view.LessonMaterials {
			if !(lm.Checked && (lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend)) {
				continue
			}
			contentMapGroupByKey := map[string][]*entity.AssessmentH5PContentScore{}
			for _, c := range aggContentsMap[lm.ID] {
				key := fmt.Sprintf("%s:%s", c.ContentID, c.SubH5PID)
				contentMapGroupByKey[key] = append(contentMapGroupByKey[key], c)
			}
			var contents []*entity.AssessmentH5PContentScore
			attendContentOrderIDs := aggUserContentOrderedIDsMap[uid]
			for _, contents2 := range contentMapGroupByKey {
				if len(contents2) == 0 {
					continue
				}
				hit := false
				for _, c := range contents2 {
					if utils.ContainsInt(attendContentOrderIDs, c.OrderedID) {
						hit = true
						contents = append(contents, c)
						break
					}
				}
				if !hit {
					c := contents2[0]
					newContent := entity.AssessmentH5PContentScore{
						OrderedID:        c.OrderedID,
						H5PID:            c.H5PID,
						ContentID:        c.ContentID,
						ContentName:      c.ContentName,
						SubH5PID:         c.SubH5PID,
						SubContentNumber: c.SubContentNumber,
					}
					contents = append(contents, &newContent)
				}
			}
			if len(contents) > 0 {
				for _, content := range contents {
					if content == nil {
						log.Debug(ctx, "get room complete rate: not found content from h5p room",
							log.String("room_id", view.RoomID),
							log.Any("not_found_content_id", lm.Source),
							log.Any("room", room),
						)
						continue
					}
					if len(content.Answers) > 0 || len(content.Scores) > 0 {
						attempted++
					}
					total++
				}
				continue
			}
			total++
		}
	}

	if total > 0 {
		log.Debug(ctx, "get room complete rate: print attempted and total",
			log.Int("attempted", attempted),
			log.Int("total", total),
		)
		return float64(attempted) / float64(total)
	}

	return 0
}

func (m *assessmentH5P) batchGetRoomScoreMap(ctx context.Context, operator *entity.Operator, roomIDs []string, enableComment bool) (map[string]*entity.AssessmentH5PRoom, error) {
	roomScoreMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "batch get room score map: batch get failed",
			log.Strings("room_ids", roomIDs),
			log.Any("operator", operator),
		)
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
			assessmentContentsMapByH5PID := make(map[string][]*entity.AssessmentH5PContentScore, len(u.Scores))
			assessmentContentMapBySubH5PID := make(map[string]*entity.AssessmentH5PContentScore, len(u.Scores))
			assessmentContentsMapByContentID := make(map[string][]*entity.AssessmentH5PContentScore, len(u.Scores))
			latestContentID := ""
			subContentNumber := 0

			// normalize external contents order
			scoresIndexMap := make(map[string]int, len(u.Scores))
			for i, s := range u.Scores {
				if s.Content != nil {
					scoresIndexMap[s.Content.ContentID] = i
				}
			}
			sort.Slice(u.Scores, func(i, j int) bool {
				itemI := u.Scores[i]
				itemJ := u.Scores[j]
				if itemI.Content.ContentID == itemJ.Content.ContentID {
					if itemI.Content.SubContentID == "" && itemJ.Content.SubContentID != "" {
						return true
					}
					if itemI.Content.SubContentID != "" && itemJ.Content.SubContentID == "" {
						return false
					}
				}
				return scoresIndexMap[itemI.Content.ContentID] < scoresIndexMap[itemJ.Content.ContentID]
			})

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
					assessmentContent.SubH5PID = s.Content.SubContentID
					if s.Content.SubContentID == "" {
						subContentNumber = 0
					} else if s.Content.ContentID != latestContentID {
						subContentNumber = 1 // 0 for not set
					} else {
						subContentNumber++
					}
					assessmentContent.SubContentNumber = subContentNumber
					latestContentID = s.Content.ContentID
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
				assessmentContentsMapByH5PID[assessmentContent.H5PID] = append(assessmentContentsMapByH5PID[assessmentContent.H5PID], &assessmentContent)
				assessmentContentMapBySubH5PID[assessmentContent.SubH5PID] = &assessmentContent
				assessmentContentsMapByContentID[assessmentContent.ContentID] = append(assessmentContentsMapByContentID[assessmentContent.ContentID], &assessmentContent)
			}
			assessmentUser := entity.AssessmentH5PUser{
				Contents:               assessmentContents,
				ContentsMapByH5PID:     assessmentContentsMapByH5PID,
				ContentMapBySubH5PID:   assessmentContentMapBySubH5PID,
				ContentsMapByContentID: assessmentContentsMapByContentID,
			}
			if u.User != nil {
				assessmentUser.UserID = u.User.UserID
			}
			if enableComment &&
				roomCommentMap != nil &&
				roomCommentMap[roomID] != nil &&
				assessmentUser.UserID != "" &&
				len(roomCommentMap[roomID][assessmentUser.UserID]) > 0 {
				comments := roomCommentMap[roomID][assessmentUser.UserID]
				assessmentUser.Comment = comments[len(comments)-1]
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

	latestOrderedID := 1
	for _, item := range result {
		for _, u := range item.Users {
			for _, c := range u.Contents {
				c.OrderedID = latestOrderedID
				latestOrderedID++
			}
		}
	}

	log.Debug(ctx, "get room score map",
		log.Strings("room_ids", roomIDs),
		log.Any("operator", operator),
		log.Any("result", result),
	)

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

func (m *assessmentH5P) getH5PStudentViewItems(ctx context.Context, operator *entity.Operator, tx *dbo.DBContext, view *entity.AssessmentView) ([]*entity.AssessmentStudentViewH5PItem, error) {
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

	// get lesson materials order map
	lmIndexMap := make(map[string]int, len(view.LessonMaterials))
	for i, lm := range view.LessonMaterials {
		lmIndexMap[lm.ID] = i
	}

	// calc agg template
	aggContentsMap := map[string][]*entity.AssessmentH5PContentScore{}
	aggUserContentOrderedIDsMap := map[string][]int{}
	for _, u := range room.Users {
		for id, contents := range u.ContentsMapByContentID {
			for _, c := range contents {
				if u.UserID != "" {
					aggUserContentOrderedIDsMap[u.UserID] = append(aggUserContentOrderedIDsMap[u.UserID], c.OrderedID)
				}
				exists := false
				for _, c2 := range aggContentsMap[id] {
					if c2 == c {
						exists = true
						break
					}
				}
				if !exists {
					aggContentsMap[id] = append(aggContentsMap[id], c)
				}
			}
		}
	}
	log.Debug(ctx, "get h5p student view items: print agg maps",
		log.Any("agg_contents_map", aggContentsMap),
		log.Any("agg_user_content_ordered_ids_map", aggUserContentOrderedIDsMap),
	)

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
			uid := ""
			if user != nil {
				uid = user.UserID
			}
			contentMapGroupByKey := map[string][]*entity.AssessmentH5PContentScore{}
			for _, c := range aggContentsMap[lm.ID] {
				key := fmt.Sprintf("%s:%s", c.ContentID, c.SubH5PID)
				contentMapGroupByKey[key] = append(contentMapGroupByKey[key], c)
			}
			var contents []*entity.AssessmentH5PContentScore
			attendContentOrderIDs := aggUserContentOrderedIDsMap[uid]
			for _, contents2 := range contentMapGroupByKey {
				if len(contents2) == 0 {
					continue
				}
				hit := false
				for _, c := range contents2 {
					if utils.ContainsInt(attendContentOrderIDs, c.OrderedID) {
						hit = true
						contents = append(contents, c)
						break
					}
				}
				if !hit {
					c := contents2[0]
					newContent := entity.AssessmentH5PContentScore{
						OrderedID:        c.OrderedID,
						H5PID:            c.H5PID,
						ContentID:        c.ContentID,
						ContentName:      c.ContentName,
						SubH5PID:         c.SubH5PID,
						SubContentNumber: c.SubContentNumber,
					}
					contents = append(contents, &newContent)
				}
			}
			if len(contents) > 0 {
				for _, content := range contents {
					if content == nil {
						log.Debug(ctx, "get h5p assessment detail: not found content from h5p room",
							log.String("room_id", view.RoomID),
							log.Any("not_found_content_id", lm.Source),
							log.Any("room", room),
						)
						continue
					}
					newLessonMaterial := entity.AssessmentStudentViewH5PLessonMaterial{
						LessonMaterialID:   lm.ID,
						LessonMaterialName: lm.Name,
						LessonMaterialType: content.ContentType,
						Answer:             content.Answer,
						MaxScore:           content.MaxPossibleScore,
						AchievedScore:      content.AchievedScore,
						Attempted:          len(content.Answers) > 0 || len(content.Scores) > 0,
						IsH5P:              lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend,
						OutcomeNames:       lpOutcomeNameMap[lm.ID],
						SubContentNumber:   content.SubContentNumber,
						Number:             content.SubH5PID,
						H5PID:              content.H5PID,
						SubH5PID:           content.SubH5PID,
					}
					newItem.LessonMaterials = append(newItem.LessonMaterials, &newLessonMaterial)
				}
				continue
			}
			newLessMaterial := entity.AssessmentStudentViewH5PLessonMaterial{
				LessonMaterialID:   lm.ID,
				LessonMaterialName: lm.Name,
				IsH5P:              lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend,
				OutcomeNames:       lpOutcomeNameMap[lm.ID],
			}
			newItem.LessonMaterials = append(newItem.LessonMaterials, &newLessMaterial)
		}
		sort.Slice(newItem.LessonMaterials, func(i, j int) bool {
			itemI := newItem.LessonMaterials[i]
			itemJ := newItem.LessonMaterials[j]
			if itemI.LessonMaterialID == itemJ.LessonMaterialID {
				return itemI.SubContentNumber < itemJ.SubContentNumber
			}
			return lmIndexMap[itemI.LessonMaterialID] < lmIndexMap[itemJ.LessonMaterialID]
		})
		lastLessonMaterialID := ""
		number := 0
		subNumber := 0
		for _, lm := range newItem.LessonMaterials {
			if lastLessonMaterialID != lm.LessonMaterialID {
				number++
				lastLessonMaterialID = lm.LessonMaterialID
			}
			if lm.SubH5PID == "" {
				subNumber = 0
			} else {
				subNumber++
			}
			if subNumber > 0 {
				lm.Number = fmt.Sprintf("%d-%d", number, subNumber)
			} else {
				lm.Number = fmt.Sprintf("%d", number)
			}
		}
		if len(newItem.LessonMaterials) == 0 {
			log.Debug(ctx, "get h5p student view items: empty lesson materials",
				log.Any("temp_result", r),
				log.Any("view", view),
			)
		}
		r = append(r, &newItem)
	}

	// filter parent h5p
	for _, item := range r {
		var deletingLMs []*entity.AssessmentStudentViewH5PLessonMaterial
		lmCountMap := make(map[string]int, len(item.LessonMaterials))
		for _, lm := range item.LessonMaterials {
			lmCountMap[lm.LessonMaterialID] = lmCountMap[lm.LessonMaterialID] + 1
		}
		for _, lm := range item.LessonMaterials {
			if lmCountMap[lm.LessonMaterialID] > 1 && lm.SubH5PID == "" {
				deletingLMs = append(deletingLMs, lm)
			}
		}
		for _, deletingLM := range deletingLMs {
			for i, lm := range item.LessonMaterials {
				if deletingLM == lm {
					item.LessonMaterials = append(item.LessonMaterials[:i], item.LessonMaterials[i+1:]...)
					break
				}
			}
		}
	}

	sort.Sort(entity.AssessmentStudentViewH5PItemsOrder(r))

	return r, nil
}
