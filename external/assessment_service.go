package external

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/KL-Engineering/chlorine"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type AssessmentServiceProvider interface {
	Get(ctx context.Context, operator *entity.Operator, roomIDs []string, options ...AssessmentServiceOption) (map[string]*RoomInfo, error)
}

type AssessmentServiceOption func(option *AssessmentServiceGetOption)

type AssessmentServiceGetOption struct {
	Score          AssessmentServiceGetQuery
	TeacherComment AssessmentServiceGetQuery
}

type AssessmentServiceGetQuery struct {
	Field    string
	Fragment string
}

func WithAssessmentGetScore(short bool) AssessmentServiceOption {
	if short {
		return AssessmentGetShortScore()
	}

	return func(option *AssessmentServiceGetOption) {
		option.Score = AssessmentServiceGetQuery{}

		option.Score.Field = "...scoresByUser"
		option.Score.Fragment = `
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
}`
	}

}

func AssessmentGetShortScore() AssessmentServiceOption {
	return func(option *AssessmentServiceGetOption) {
		option.Score = AssessmentServiceGetQuery{}

		option.Score.Field = "...scoresByUser"
		option.Score.Fragment = `
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
}`
	}
}

func WithAssessmentGetTeacherComment(teacherComment bool) AssessmentServiceOption {
	return func(option *AssessmentServiceGetOption) {
		option.TeacherComment = AssessmentServiceGetQuery{}

		if !teacherComment {
			return
		}

		option.TeacherComment.Field = "...teacherCommentsByStudent"
		option.TeacherComment.Fragment = `
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
}`
	}
}

func GetAssessmentServiceProvider() AssessmentServiceProvider {
	return &AssessmentService{}
}

type AssessmentService struct {
}

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

type RoomInfo struct {
	ScoresByUser             []*H5PUserScores               `json:"scoresByUser"`
	TeacherCommentsByStudent []*H5PTeacherCommentsByStudent `json:"teacherCommentsByStudent"`
}

func (s *AssessmentService) Get(ctx context.Context, operator *entity.Operator, roomIDs []string, options ...AssessmentServiceOption) (map[string]*RoomInfo, error) {
	result := make(map[string]*RoomInfo)

	if len(roomIDs) == 0 {
		return result, nil
	}

	getOption := &AssessmentServiceGetOption{}
	if len(options) <= 0 {
		options = []AssessmentServiceOption{WithAssessmentGetScore(true)}
	}

	for _, op := range options {
		op(getOption)
	}

	query := fmt.Sprintf(`
query {
	{{range $i, $e := .}}
	q{{$i}}: Room(room_id: "{{$e}}") {
	%s
	%s
}
{{end}}
}

%s
%s
`,
		getOption.Score.Field,
		getOption.TeacherComment.Field,

		getOption.Score.Fragment,
		getOption.TeacherComment.Fragment)

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
