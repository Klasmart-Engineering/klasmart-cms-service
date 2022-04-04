package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type AssessmentServiceProvider interface {
	GetRoomInfoByRoomID(ctx context.Context, operator *entity.Operator, roomID string) (*RoomInfo, error)
}

func GetAssessmentServiceProvider() AssessmentServiceProvider {
	return &AssessmentService{}
}

type AssessmentService struct{}

type ScoresByContent struct {
	Content *H5PContent `json:"content"`
}

type RoomInfo struct {
	ScoresByContent          []*ScoresByContent             `json:"scoresByContent"`
	ScoresByUser             []*H5PUserScores               `json:"scoresByUser"`
	TeacherCommentsByStudent []*H5PTeacherCommentsByStudent `json:"teacherCommentsByStudent"`
}

func (s *AssessmentService) GetRoomInfoByRoomID(ctx context.Context, operator *entity.Operator, roomID string) (*RoomInfo, error) {
	if roomID == "" {
		return new(RoomInfo), nil
	}

	query := `
query(
$room_id: String! 
) {
	Room(room_id: $room_id) {
		...scoresByUser
		...scoresByContent
		...teacherCommentsByStudent
  	}
}

fragment scoresByContent on Room{
	 scoresByContent{
		content{
		  parent_id
		  content_id
		  subcontent_id
		  h5p_id
		  type
		  fileType
		}
	  }
}
fragment teacherCommentsByStudent on Room {
	teacherCommentsByStudent {
	  student {
		  user_id
	  }
	  teacherComments {
		  teacher {
			  user_id
			  given_name
			  family_name
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

	request := chlorine.NewRequest(query, chlorine.ReqToken(operator.Token))
	request.Var("room_id", roomID)

	data := map[string]*RoomInfo{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetH5PClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get room scores failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", query),
			log.String("roomID", roomID))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Warn(ctx, "get room scores failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", query),
			log.String("roomID", roomID))
		//return nil, response.Errors
	}

	result := new(RoomInfo)

	if item, ok := data["Room"]; ok && item != nil {
		result = item
	}

	return result, nil
}
