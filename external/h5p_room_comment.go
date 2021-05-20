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

// type TeacherComment {
// 	student: User!
// 	teacher: User!
// 	date: Timestamp!
// 	comment: String!
// }

type H5PTeacherComment struct {
	Teacher *H5PUser `json:"teacher"`
	Student *H5PUser `json:"student"`
	Date    int64    `json:"date"`
	Comment string   `json:"comment"`
}

// type TeacherCommentsByStudent {
// 	student: User!
// 	teacherComments: [TeacherComment!]!
// }
type H5PTeacherCommentsByStudent struct {
	User            *H5PUser             `json:"user"`
	TeacherComments []*H5PTeacherComment `json:"teacherComments"`
}

type H5PRoomCommentServiceProvider interface {
	Get(ctx context.Context, operator *entity.Operator, roomID string) ([]*H5PTeacherCommentsByStudent, error)
	BatchGet(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string][]*H5PTeacherCommentsByStudent, error)
	Add(ctx context.Context, operator *entity.Operator, request *H5PAddRoomCommentRequest) (*H5PTeacherComment, error)
	BatchAdd(ctx context.Context, operator *entity.Operator, requests []*H5PAddRoomCommentRequest) ([]*H5PTeacherComment, error)
}

func GetH5PRoomCommentServiceProvider() H5PRoomCommentServiceProvider {
	return &H5PRoomCommentService{}
}

type H5PRoomCommentService struct{}

func (s H5PRoomCommentService) Get(ctx context.Context, operator *entity.Operator, roomID string) ([]*H5PTeacherCommentsByStudent, error) {
	comments, err := s.BatchGet(ctx, operator, []string{roomID})
	if err != nil {
		return nil, err
	}

	comment, found := comments[roomID]
	if !found {
		log.Error(ctx, "h5p room comment not found",
			log.String("roomID", roomID),
			log.Any("comments", comments))
		return nil, constant.ErrRecordNotFound
	}

	return comment, nil
}

func (s H5PRoomCommentService) BatchGet(ctx context.Context, operator *entity.Operator, roomIDs []string) (map[string][]*H5PTeacherCommentsByStudent, error) {
	if len(roomIDs) == 0 {
		return map[string][]*H5PTeacherCommentsByStudent{}, nil
	}

	query := `
query {
	{{range $i, $e := .}}
	q{{$i}}: Room(room_id: "{{$e}}") {
		teacherCommentsByStudent {
			student {
				user_id
				given_name
				family_name
			}
			teacherComments {
				student {
					user_id
					given_name
					family_name
				}
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
		TeacherCommentsByStudent []*H5PTeacherCommentsByStudent `json:"teacherCommentsByStudent"`
	}{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetH5PClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get room comments failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Strings("roomIDs", roomIDs))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get room comments failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Strings("roomIDs", roomIDs))
		return nil, response.Errors
	}

	for _, studentComments := range data {
		for _, teacherComments := range studentComments.TeacherCommentsByStudent {
			for _, comment := range teacherComments.TeacherComments {
				// date is saved in milliseconds, we are more used to processing by seconds
				comment.Date = comment.Date / 1000
			}
		}
	}

	comments := make(map[string][]*H5PTeacherCommentsByStudent, len(data))
	for index := range roomIDs {
		comment := data[fmt.Sprintf("q%d", indexMapping[index])]
		if comment == nil {
			log.Error(ctx, "user content comment not found", log.String("roomID", roomIDs[index]))
			return nil, constant.ErrRecordNotFound
		}

		comments[roomIDs[index]] = comment.TeacherCommentsByStudent
	}

	log.Info(ctx, "get room comments success",
		log.Strings("roomIDs", roomIDs),
		log.Any("comments", comments))

	return comments, nil
}

type H5PAddRoomCommentRequest struct {
	RoomID    string
	StudentID string
	Comment   string
}

func (s H5PRoomCommentService) Add(ctx context.Context, operator *entity.Operator, request *H5PAddRoomCommentRequest) (*H5PTeacherComment, error) {
	comments, err := s.BatchAdd(ctx, operator, []*H5PAddRoomCommentRequest{request})
	if err != nil {
		return nil, err
	}

	if len(comments) != 1 {
		log.Error(ctx, "h5p room set comment result not found",
			log.Any("request", request),
			log.Any("comments", comments))
		return nil, constant.ErrRecordNotFound
	}

	return comments[0], nil
}

func (s H5PRoomCommentService) BatchAdd(ctx context.Context, operator *entity.Operator, requests []*H5PAddRoomCommentRequest) ([]*H5PTeacherComment, error) {
	if len(requests) == 0 {
		return []*H5PTeacherComment{}, nil
	}

	mutation := `
mutation {
	{{range $i, $e := .}}
	q{{$i}}: addComment(
		comment: "{{$e.Comment}}"
		student_id: "{{$e.StudentID}}"
		room_id: "{{$e.RoomID}}"
	) {
		student {
			user_id
			given_name
			family_name
		}
		teacher {
			user_id
			given_name
			family_name
		}
		date
		comment
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

	data := map[string]*H5PTeacherComment{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetH5PClient().Run(ctx, _request, response)
	if err != nil {
		log.Error(ctx, "add room comments failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("requests", requests))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "add room comments failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("query", buffer.String()),
			log.Any("requests", requests))
		return nil, response.Errors
	}

	comments := make([]*H5PTeacherComment, 0, len(data))
	for index := range requests {
		comment := data[fmt.Sprintf("q%d", index)]
		if comment == nil {
			log.Error(ctx, "h5p room comment result not found", log.Any("request", requests[index]))
			return nil, constant.ErrRecordNotFound
		}

		// date is saved in milliseconds, we are more used to processing by seconds
		comment.Date = comment.Date / 1000
		comments = append(comments, comment)
	}

	log.Info(ctx, "add room comment success",
		log.Any("requests", requests),
		log.Any("comments", comments))

	return comments, nil
}
