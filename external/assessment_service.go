package external

import (
	"bytes"
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"text/template"
)

type AssessmentServiceProvider interface {
	GetScoresWithCommentsByRoomIDs(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]*RoomInfo, error)
	SetScoreAndComment(ctx context.Context, operator *entity.Operator, scores []*SetScoreAndComment) error
}

func GetAssessmentServiceProvider() AssessmentServiceProvider {
	return &AssessmentService{}
}

type AssessmentService struct{}

type SetScoreAndComment struct {
	RoomID    string
	StudentID string
	Comment   string
	Scores    []*SetScore
}
type SetScore struct {
	ContentID    string
	SubContentID string
	Score        float64
}

func (s *AssessmentService) SetScoreAndComment(ctx context.Context, operator *entity.Operator, scores []*SetScoreAndComment) error {
	mutation := `
mutation {
	{{range $i, $e := .}}
	c{{$i}}: setComment(
		comment: """{{$e.Comment}}"""
		student_id: "{{$e.StudentID}}"
		room_id: "{{$e.RoomID}}"
	) {
		student {
			user_id
		}
	}
	{{range $i2, $e2 := $e.Scores}}
	s{{$i2}}: setScore(
		room_id: "{{$e.RoomID}}"
		student_id: "{{$e.StudentID}}"
		content_id: "{{$e2.ContentID}}"
		subcontent_id: "{{$e2.SubContentID}}"
		score: {{$e2.Score}}
	) {
		student {
			user_id
		}
		content {
			content_id
		}
	}
	{{end}}
	{{end}}
}`

	temp, err := template.New("").Parse(mutation)
	if err != nil {
		log.Error(ctx, "init template failed", log.Err(err))
		return err
	}

	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, scores)
	if err != nil {
		log.Error(ctx, "execute template failed", log.Err(err), log.Any("scores", scores))
		return err
	}

	_request := chlorine.NewRequest(buffer.String(), chlorine.ReqToken(operator.Token))

	data := map[string]interface{}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetH5PClient().Run(ctx, _request, response)
	if err != nil {
		log.Error(ctx, "set room scores failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("requests", scores))
		return err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "set room scores failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("requests", scores))
		return response.Errors
	}

	log.Debug(ctx, "set room score success", log.Any("requests", scores))

	return nil
}

func (s *AssessmentService) GetScoresByRoomIDs(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string][]*H5PUserScores, error) {
	if len(roomIDs) == 0 {
		return map[string][]*H5PUserScores{}, nil
	}

	query := `
query {
	{{range $i, $e := .}}
	q{{$i}}: Room(room_id: "{{$e}}") {
		...scoresByUser
  	}
	{{end}}
}
fragment scoresByUser on Room {
	scoresByUser {
		user {
			user_id
			given_name
			family_name
		}
		scores {
			seen
			content {
				parent_id
				content_id
				name
				type
				fileType
				h5p_id
				subcontent_id
			}
			score {
				min
				max
				sum
				scoreFrequency
				mean
				scores
				answers {
					answer
					score
					date
					minimumPossibleScore
					maximumPossibleScore
				}
				median
				medians
			}
			teacherScores {
				teacher {
					user_id
					given_name
					family_name
				}
				student {
					user_id
					given_name
					family_name
				}
				content {
					content_id
					name
					type
					fileType
					h5p_id
					subcontent_id
				}
				score
				date
			}
		}
	}
}
`

	temp, err := template.New("").Parse(query)
	if err != nil {
		log.Error(ctx, "init template failed", log.Err(err))
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(roomIDs)

	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, _ids)
	if err != nil {
		log.Error(ctx, "execute template failed", log.Err(err), log.Strings("roomIDs", roomIDs))
		return nil, err
	}

	request := chlorine.NewRequest(buffer.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*struct {
		ScoresByUser []*H5PUserScores `json:"scoresByUser"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetH5PClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get room scores failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Strings("roomIDs", roomIDs))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Warn(ctx, "get room scores failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Strings("roomIDs", roomIDs))
		//return nil, response.Errors
	}

	for _, studentScores := range data {
		for _, scoreByUser := range studentScores.ScoresByUser {
			for _, score := range scoreByUser.Scores {
				for _, teacherScore := range score.TeacherScores {
					// date is saved in milliseconds, we are more used to processing by seconds
					teacherScore.Date = teacherScore.Date / 1000
				}
			}
		}
	}

	scores := make(map[string][]*H5PUserScores, len(data))
	for index := range roomIDs {
		score := data[fmt.Sprintf("q%d", indexMapping[index])]
		if score == nil {
			log.Error(ctx, "user content score not found", log.String("roomID", roomIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		scores[roomIDs[index]] = score.ScoresByUser
	}

	log.Info(ctx, "get room scores success",
		log.Strings("roomIDs", roomIDs),
		log.Any("scores", scores))

	return scores, nil
}

//
//type ScoresByContent struct {
//	Content *H5PContent `json:"content"`
//}

type RoomInfo struct {
	ScoresByUser             []*H5PUserScores               `json:"scoresByUser"`
	TeacherCommentsByStudent []*H5PTeacherCommentsByStudent `json:"teacherCommentsByStudent"`
}

func (s *AssessmentService) GetScoresWithCommentsByRoomIDs(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string]*RoomInfo, error) {
	result := make(map[string]*RoomInfo)

	if len(roomIDs) == 0 {
		return result, nil
	}

	query := `
query {
	{{range $i, $e := .}}
	q{{$i}}: Room(room_id: "{{$e}}") {
	...scoresByUser
	...teacherCommentsByStudent
}
{{end}}
}

fragment teacherCommentsByStudent on Room {
	teacherCommentsByStudent {
	  student {
		  user_id
	  }
	  teacherComments {
		  teacher {
			  user_id
		  }
		  date
		  comment
	  }
  	}
}
fragment scoresByUser on Room {
	scoresByUser {
		user {
			user_id
		}
		scores {
			seen
			content {
				parent_id
				content_id
				name
				type
				fileType
				h5p_id
				subcontent_id
			}
			score {
				min
				max
				sum
				scoreFrequency
				mean
				scores
				answers {
					answer
					score
					date
					minimumPossibleScore
					maximumPossibleScore
				}
			}
			teacherScores {
				teacher {
					user_id
				}
				student {
					user_id
				}
				content {
					content_id
					name
					type
					fileType
					h5p_id
					subcontent_id
				}
				score
				date
			}
		}
	}
}
`

	temp, err := template.New("").Parse(query)
	if err != nil {
		log.Error(ctx, "init template failed", log.Err(err))
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(roomIDs)

	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, _ids)
	if err != nil {
		log.Error(ctx, "execute template failed", log.Err(err), log.Strings("roomIDs", roomIDs))
		return nil, err
	}

	request := chlorine.NewRequest(buffer.String(), chlorine.ReqToken(operator.Token))
	data := map[string]*RoomInfo{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetH5PClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get room scores failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", query),
			log.Strings("roomIDs", roomIDs))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Warn(ctx, "get room scores failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", query),
			log.Strings("roomIDs", roomIDs))
		//return nil, response.Errors
	}

	//scores := make(map[string][]*H5PUserScores, len(data))
	for index := range roomIDs {
		dataItem := data[fmt.Sprintf("q%d", indexMapping[index])]
		if dataItem == nil {
			log.Warn(ctx, "user content score not found", log.String("roomID", roomIDs[index]))
			//return nil, constant.ErrRecordNotFound
			dataItem = new(RoomInfo)
		}

		result[roomIDs[index]] = dataItem
	}

	return result, nil
}
