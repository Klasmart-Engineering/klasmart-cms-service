package external

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type H5PScoreService struct{}

// type Score {
// 	min: Float
// 	max: Float
// 	sum: Float!
// 	frequency: Float!
// 	mean: Float
// 	scores: [Int!]!
// 	median: Float
// 	medians: [Float!]
// }

type H5PScore struct {
	Min       float64   `json:"min"`
	Max       float64   `json:"max"`
	Frequency float64   `json:"frequency"`
	Mean      float64   `json:"mean"`
	Scores    []float64 `json:"scores"`
	Median    float64   `json:"median"`
	Medians   []float64 `json:"medians"`
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

type H5PScoreServiceProvider interface {
	GetByRoom(ctx context.Context, operator *entity.Operator, roomID string) ([]*H5PUserScores, error)
}

func GetH5PScoreServiceProvider() H5PScoreServiceProvider {
	return &H5PScoreService{}
}

func (s H5PScoreService) GetByRoom(ctx context.Context, operator *entity.Operator, roomID string) ([]*H5PUserScores, error) {
	return nil, nil
}
