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

func getAssessmentH5P() *assessmentH5P {
	return &assessmentH5P{}
}

func (m *assessmentH5P) batchGetRoomMap(ctx context.Context, operator *entity.Operator, roomIDs []string, includeComment bool) (map[string]*entity.AssessmentH5PRoom, error) {
	// batch get room score map
	roomScoreMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, operator, roomIDs)
	if err != nil {
		log.Error(ctx, "batch get room map: batch get scores failed",
			log.Strings("room_ids", roomIDs),
			log.Any("operator", operator),
		)
		return nil, err
	}

	// batch get room comment map
	var roomCommentMap map[string]map[string][]string
	if includeComment {
		roomCommentMap, err = m.batchGetRoomCommentMap(ctx, operator, roomIDs)
		if err != nil {
			log.Error(ctx, "batch get room map: batch get comments failed",
				log.Strings("room_ids", roomIDs),
				log.Bool("include_comment", includeComment),
				log.Any("operator", operator),
			)
			return nil, err
		}
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

			// fill comment
			if includeComment &&
				roomCommentMap != nil &&
				assessmentUser.UserID != "" &&
				roomCommentMap[roomID] != nil &&
				len(roomCommentMap[roomID][assessmentUser.UserID]) > 0 {
				comments := roomCommentMap[roomID][assessmentUser.UserID]
				assessmentUser.Comment = comments[len(comments)-1]
			}

			// fill contents
			assessmentContents := make([]*entity.AssessmentH5PContent, 0, len(u.Scores))
			for _, s := range u.Scores {
				assessmentContent := entity.AssessmentH5PContent{
					Scores: s.Score.Scores,
				}
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
	attempted := false
	for _, u := range room.Users {
		for _, c := range u.Contents {
			if len(c.Answers) > 0 || len(c.Scores) > 0 {
				attempted = true
				break
			}
		}
		if attempted {
			break
		}
	}
	return attempted
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
	roomMap, err := m.batchGetRoomMap(ctx, operator, []string{view.RoomID}, true)
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
	aggContentsMap := map[string][]*entity.AssessmentH5PContent{}
	aggUserContentOrderedIDsMap := map[string][]int{}
	for _, u := range room.Users {
		for id, contents := range getAssessmentH5P().getContentsMapByContentID(u) {
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
		user := getAssessmentH5P().getUserMap(room)[s.ID]
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
			contentMapGroupByKey := map[string][]*entity.AssessmentH5PContent{}
			for _, c := range aggContentsMap[lm.ID] {
				key := fmt.Sprintf("%s:%s", c.ContentID, c.SubH5PID)
				contentMapGroupByKey[key] = append(contentMapGroupByKey[key], c)
			}
			var contents []*entity.AssessmentH5PContent
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
						LessonMaterialID:     lm.ID,
						LessonMaterialName:   lm.Name,
						LessonMaterialType:   content.ContentType,
						Answer:               getAssessmentH5P().getAnswer(content),
						MaxScore:             getAssessmentH5P().getMaxPossibleScore(content),
						AchievedScore:        getAssessmentH5P().getAchievedScore(content),
						Attempted:            len(content.Answers) > 0 || len(content.Scores) > 0,
						IsH5P:                lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend,
						OutcomeNames:         lpOutcomeNameMap[lm.ID],
						H5PID:                content.H5PID,
						SubH5PID:             content.SubH5PID,
						NotApplicableScoring: getAssessmentH5P().canScoring(content.ContentType),
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
				return itemI.SubH5PID < itemJ.SubH5PID
			}
			return lmIndexMap[itemI.LessonMaterialID] < lmIndexMap[itemJ.LessonMaterialID]
		})

		if len(newItem.LessonMaterials) == 0 {
			log.Debug(ctx, "get h5p student view items: empty lesson materials",
				log.Any("temp_result", r),
				log.Any("view", view),
			)
		}
		r = append(r, &newItem)
	}

	sort.Sort(entity.AssessmentStudentViewH5PItemsOrder(r))

	return r, nil
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
	if room == nil {
		log.Debug(ctx, "get room complete rate: room is empty",
			log.Any("view", view),
		)
		return 0
	}

	// calc agg template
	aggContentsMap := map[string][]*entity.AssessmentH5PContent{}
	aggUserContentOrderedIDsMap := map[string][]int{}
	for _, u := range room.Users {
		for id, contents := range getAssessmentH5P().getContentsMapByContentID(u) {
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
		user := getAssessmentH5P().getUserMap(room)[s.ID]
		uid := ""
		if user != nil {
			uid = user.UserID
		}
		for _, lm := range view.LessonMaterials {
			if !(lm.Checked && (lm.FileType == entity.FileTypeH5p || lm.FileType == entity.FileTypeH5pExtend)) {
				continue
			}
			aggContents := aggContentsMap[lm.ID]
			contentMapGroupByKey := map[string][]*entity.AssessmentH5PContent{}
			for _, c := range aggContents {
				key := fmt.Sprintf("%s:%s", c.ContentID, c.SubH5PID)
				contentMapGroupByKey[key] = append(contentMapGroupByKey[key], c)
			}
			var contents []*entity.AssessmentH5PContent
			attendContentOrderIDs := aggUserContentOrderedIDsMap[uid]
			for _, keyedContents := range contentMapGroupByKey {
				if len(keyedContents) == 0 {
					continue
				}
				hit := false
				for _, c := range keyedContents {
					if utils.ContainsInt(attendContentOrderIDs, c.OrderedID) {
						hit = true
						contents = append(contents, c)
						break
					}
				}
				if !hit {
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
					if len(contents) > 1 && content.SubH5PID == "" {
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
			log.String("room_id", view.RoomID),
		)
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
	"AudioRecorderBookMaker":       false,
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
	return canScoringMap[contentType]
}
