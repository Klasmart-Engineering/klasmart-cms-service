package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func GetH5PServiceProvider() H5PServiceProvider {
	return &h5pServiceProvider{}
}

type H5PServiceProvider interface {
	BatchGetMap(ctx context.Context, operator *entity.Operator, studentIDs []string) (map[string]map[string]*H5PStudentScore, error)
}

type H5PStudentScore struct {
	StudentID        string       `json:"student_id"`
	ActivityID       string       `json:"activity_id"`
	ActivityType     ActivityType `json:"activity_type"`
	Answer           string       `json:"answer"`
	MaxPossibleScore float64      `json:"max_possible_score"`
	AchievedScore    float64      `json:"achieved_score"`
	Attempted        bool         `json:"attempted"`
}

type ActivityType string

const (
	ActivityTypeEssay = "Essay"
)

type h5pServiceProvider struct{}

func (*h5pServiceProvider) BatchGetMap(ctx context.Context, operator *entity.Operator, studentIDs []string) (map[string]map[string]*H5PStudentScore, error) {
	panic("implement me")
}
