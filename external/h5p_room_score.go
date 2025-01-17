package external

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/KL-Engineering/chlorine"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

// 	answers {
// 	  answer
// 	  score
// 	  # date
// 	  minimumPossibleScore
// 	  maximumPossibleScore
// 	}
type H5PScoreAnswer struct {
	Answer               string  `json:"answer"`
	Score                float64 `json:"score"`
	Date                 int64   `json:"date"`
	MinimumPossibleScore float64 `json:"minimumPossibleScore"`
	MaximumPossibleScore float64 `json:"maximumPossibleScore"`
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
//	seen: Boolean!
// 	teacherScores: [TeacherScore!]!
// }
type H5PUserContentScore struct {
	//User          *H5PUser           `json:"user"`
	Content       *H5PContent        `json:"content"`
	Score         *H5PScore          `json:"score"`
	Seen          bool               `json:"seen"`
	TeacherScores []*H5PTeacherScore `json:"teacherScores"`
}

func (rc *H5PContent) GetInternalID() string {
	if rc.ParentID == "" {
		return rc.ContentID
	} else {
		if rc.SubContentID != "" {
			return rc.SubContentID
		} else {
			return rc.H5PID
		}
	}
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
	RoomID       string
	StudentID    string
	ContentID    string
	SubContentID string
	Score        float64
}

type H5PRoomScoreServiceProvider interface {
	Get(ctx context.Context, operator *entity.Operator, roomID string) ([]*H5PUserScores, error)
	BatchGet(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string][]*H5PUserScores, error)
	Set(ctx context.Context, operator *entity.Operator, request *H5PSetScoreRequest) (*H5PTeacherScore, error)
	BatchSet(ctx context.Context, operator *entity.Operator, requests []*H5PSetScoreRequest) ([]*H5PTeacherScore, error)
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

func (s H5PRoomScoreService) Set(ctx context.Context, operator *entity.Operator, request *H5PSetScoreRequest) (*H5PTeacherScore, error) {
	scores, err := s.BatchSet(ctx, operator, []*H5PSetScoreRequest{request})
	if err != nil {
		return nil, err
	}

	if len(scores) != 1 {
		log.Error(ctx, "h5p room set score result not found",
			log.Any("request", request),
			log.Any("scores", scores))
		return nil, constant.ErrRecordNotFound
	}

	return scores[0], nil
}

func (s H5PRoomScoreService) BatchSet(ctx context.Context, operator *entity.Operator, requests []*H5PSetScoreRequest) ([]*H5PTeacherScore, error) {
	if len(requests) == 0 {
		return []*H5PTeacherScore{}, nil
	}

	mutation := `
mutation {
	{{range $i, $e := .}}
	q{{$i}}: setScore(
		room_id: "{{$e.RoomID}}"
		student_id: "{{$e.StudentID}}"
		content_id: "{{$e.ContentID}}"
		subcontent_id: "{{$e.SubContentID}}"
		score: {{$e.Score}}
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
	{{end}}
}`

	temp, err := template.New("").Parse(mutation)
	if err != nil {
		log.Error(ctx, "init template failed", log.Err(err))
		return nil, err
	}

	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, requests)
	if err != nil {
		log.Error(ctx, "execute template failed", log.Err(err), log.Any("requests", requests))
		return nil, err
	}

	_request := chlorine.NewRequest(buffer.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*H5PTeacherScore{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetH5PClient().Run(ctx, _request, response)
	if err != nil {
		log.Error(ctx, "set room scores failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("requests", requests))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "set room scores failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("requests", requests))
		return nil, response.Errors
	}

	scores := make([]*H5PTeacherScore, 0, len(data))
	for index := range requests {
		score := data[fmt.Sprintf("q%d", index)]
		if score == nil {
			log.Error(ctx, "h5p set room score result not found", log.Any("request", requests[index]))
			return nil, constant.ErrRecordNotFound
		}

		// date is saved in milliseconds, we are more used to processing by seconds
		score.Date = score.Date / 1000
		scores = append(scores, score)
	}

	log.Info(ctx, "set room score success",
		log.Any("requests", requests),
		log.Any("scores", scores))

	return scores, nil
}
