package external

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

// 	answers {
// 	  answer
// 	  score
// 	  # date
// 	}
type H5PScoreAnswer struct {
	Answer string  `json:"answer"`
	Score  float64 `json:"score"`
	Date   int64   `json:"date"`
}

// score {
// 	min
// 	max
// 	sum
// 	scoreFrequency
// 	mean
// 	scores
// 	answers {
// 	  answer
// 	  score
// 	  # date
// 	}
// 	median
// 	medians
//   }
type H5PScore struct {
	Min            float64           `json:"min"`
	Max            float64           `json:"max"`
	Sum            float64           `json:"sum"`
	ScoreFrequency float64           `json:"scoreFrequency"`
	Mean           float64           `json:"mean"`
	Scores         []float64         `json:"scores"`
	Answers        []*H5PScoreAnswer `json:"answers"`
	Median         float64           `json:"median"`
	Medians        []float64         `json:"medians"`
}

// type TeacherScore {
// 	teacher: User!
// 	student: User!
// 	content: Content!
// 	score: Float!
// 	date: Timestamp!
// }
type H5PTeacherScore struct {
	Teacher *H5PUser    `json:"teacher"`
	Student *H5PUser    `json:"student"`
	Content *H5PContent `json:"content"`
	Score   float64     `json:"score"`
	Date    int64       `json:"date"`
}

// type UserContentScore {
// 	user: User!
// 	content: Content!
// 	score: Score!
// 	teacherScores: [TeacherScore!]!
// 	minimumPossibleScore: Float!
// 	maximumPossibleScore: Float!
// }
type H5PUserContentScore struct {
	User                 *H5PUser           `json:"user"`
	Content              *H5PContent        `json:"content"`
	Score                *H5PScore          `json:"score"`
	TeacherScores        []*H5PTeacherScore `json:"teacherScores"`
	MinimumPossibleScore float64            `json:"minimumPossibleScore"`
	MaximumPossibleScore float64            `json:"maximumPossibleScore"`
}

// type UserScores {
// 	user: User!
// 	scores: [UserContentScore!]!
// }
type H5PUserScores struct {
	User   *H5PUser               `json:"user"`
	Scores []*H5PUserContentScore `json:"scores"`
}

type H5PSetScoreRequest struct {
	RoomID    string
	ContentID string
	StudentID string
	Score     float64
}

type H5PRoomScoreServiceProvider interface {
	Get(ctx context.Context, operator *entity.Operator, roomID string) ([]*H5PUserScores, error)
	BatchGet(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string][]*H5PUserScores, error)
	Set(ctx context.Context, operator *entity.Operator, request *H5PSetScoreRequest) (*H5PTeacherScore, error)
}

func GetH5PRoomScoreServiceProvider() H5PRoomScoreServiceProvider {
	return &H5PRoomScoreService{}
}

type H5PRoomScoreService struct{}

func (s H5PRoomScoreService) Get(ctx context.Context, operator *entity.Operator, roomID string) ([]*H5PUserScores, error) {
	scores, err := s.BatchGet(ctx, operator, []string{roomID})
	if err != nil {
		return nil, err
	}

	score, found := scores[roomID]
	if !found {
		log.Error(ctx, "h5p room score not found",
			log.String("roomID", roomID),
			log.Any("scores", scores))
		return nil, constant.ErrRecordNotFound
	}

	return score, nil
}

func (s H5PRoomScoreService) BatchGet(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string][]*H5PUserScores, error) {
	if len(roomIDs) == 0 {
		return map[string][]*H5PUserScores{}, nil
	}

	query := `
query {
	{{range $i, $e := .}}
	q{{$i}}: Room(room_id: "{{$e}}") {
		scoresByUser {
			user {
				user_id
				given_name
				family_name
			}
			scores {
				content {
					content_id
					name
					type
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
						# date
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
					}
					score
					date
				}
				minimumPossibleScore
				maximumPossibleScore
			}
		}
  	}
	{{end}}
}`

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
		log.Error(ctx, "get room scores failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Strings("roomIDs", roomIDs))
		return nil, response.Errors
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

func (s H5PRoomScoreService) Set(ctx context.Context, operator *entity.Operator, request *H5PSetScoreRequest) (*H5PTeacherScore, error) {
	query := `
mutation {
	setScore(
		score: {{.Score}}
		content_id: "{{.ContentID}}"
		student_id: "{{.StudentID}}"
		room_id: "{{.RoomID}}"
	) {
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
		}
		score
		date
	}
}`

	temp, err := template.New("").Parse(query)
	if err != nil {
		log.Error(ctx, "init template failed", log.Err(err))
		return nil, err
	}

	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, request)
	if err != nil {
		log.Error(ctx, "execute template failed", log.Err(err), log.Any("request", request))
		return nil, err
	}

	_request := chlorine.NewRequest(buffer.String(), chlorine.ReqToken(operator.Token))

	data := struct {
		SetScore *H5PTeacherScore `json:"setScore"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetH5PClient().Run(ctx, _request, response)
	if err != nil {
		log.Error(ctx, "set room score failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("request", request))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "set room score failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("request", request))
		return nil, response.Errors
	}

	score := data.SetScore
	if score == nil {
		log.Error(ctx, "room score not found", log.Any("request", request))
		return nil, constant.ErrRecordNotFound
	}

	log.Info(ctx, "set room score success",
		log.Any("request", request),
		log.Any("score", score))

	return score, nil
}
