package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	Score                 AssessmentServiceGetQuery
	TeacherComment        AssessmentServiceGetQuery
	CompletionPercentages AssessmentServiceGetQuery
}

type AssessmentServiceGetQuery struct {
	Enable   bool
	Field    string
	Fragment string
}

func WithAssessmentGetScore() AssessmentServiceOption {
	return func(option *AssessmentServiceGetOption) {
		option.Score = AssessmentServiceGetQuery{
			Enable: true,
		}

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

func WithAssessmentGetTeacherComment() AssessmentServiceOption {
	return func(option *AssessmentServiceGetOption) {
		option.TeacherComment = AssessmentServiceGetQuery{
			Enable: true,
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

func WithAssessmentGetCompletionPercentages() AssessmentServiceOption {
	return func(option *AssessmentServiceGetOption) {
		option.CompletionPercentages = AssessmentServiceGetQuery{
			Enable: true,
		}
		option.CompletionPercentages.Field = ""
		option.CompletionPercentages.Fragment = ""
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
	CompletionPercentage     float64
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
		return result, nil
	}

	for _, op := range options {
		op(getOption)
	}

	if getOption.CompletionPercentages.Enable {
		var sb strings.Builder
		length := len(roomIDs)
		for index, item := range roomIDs {
			fmt.Fprint(&sb, "\""+item+"\"")
			if index > length-1 {
				fmt.Fprint(&sb, ",")
			}
		}
		getOption.CompletionPercentages.Field = fmt.Sprintf("completionPercentages(room_ids:[%s])", sb.String())
	}

	scoreAndCommentQuery := ""
	if getOption.Score.Enable || getOption.TeacherComment.Enable {
		scoreAndCommentQuery = fmt.Sprintf(`
	{{range $i, $e := .}}
	q{{$i}}: Room(room_id: "{{$e}}") {
	%s
	%s
}
{{end}}
`,
			getOption.Score.Field,
			getOption.TeacherComment.Field)
	}

	query := fmt.Sprintf(`
query {
	%s
	%s
}

%s
%s
`,
		getOption.CompletionPercentages.Field,
		scoreAndCommentQuery,

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
	data := map[string]interface{}{}
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

	// key:roomId
	var completionPercentages []float64
	if item, ok := data["completionPercentages"]; ok {
		if items, ok := item.([]interface{}); ok {
			completionPercentages = make([]float64, 0, len(items))
			for _, completion := range items {
				completionFloat, ok := completion.(float64)
				if !ok {
					log.Warn(ctx, "completion data is not float64", log.Any("completion", completion))
				}
				// The conversion type failed, also added to the array as a placeholder
				completionPercentages = append(completionPercentages, 0)
			}
		} else {
			log.Warn(ctx, "completionPercentages data is not array", log.Any("completionPercentages", item))
		}
	}

	delete(data, "completionPercentages")

	jb, _ := json.Marshal(data)
	var roomDataMap map[string]*RoomInfo
	err = json.Unmarshal(jb, &roomDataMap)
	if err != nil {
		log.Error(ctx, "unmarshal room data info error", log.Err(err), log.Strings("roomIDs", roomIDs))
		return result, nil
	}

	for index := range roomIDs {
		roomInfo := roomDataMap[fmt.Sprintf("q%d", indexMapping[index])]
		if roomInfo == nil {
			roomInfo = new(RoomInfo)
		}

		if len(completionPercentages) > index {
			roomInfo.CompletionPercentage = completionPercentages[index]
		}

		result[roomIDs[index]] = roomInfo
	}

	return result, nil
}
