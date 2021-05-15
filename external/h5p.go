package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func GetH5PServiceProvider() H5PServiceProvider {
	return &h5pServiceProvider{}
}

type H5PServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, studentIDs []string)
}

type StudentActivity struct {
	StudentID     string       `json:"student_id"`
	ActivityID    string       `json:"activity_id"`
	ActivityType  ActivityType `json:"activity_type"`
	Answer        string       `json:"answer"`
	MaxScore      float64      `json:"max_score"`
	AchievedScore float64      `json:"achieved_score"`
}

type ActivityType string

const (
	ActivityTypeEssay = "Essay"
)

type h5pServiceProvider struct{}

func (*h5pServiceProvider) BatchGet(ctx context.Context, operator *entity.Operator, studentIDs []string) {
	panic("implement me")
}
