package model

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

var (
	homeFunStudyModelInstance     IHomeFunStudyModel
	homeFunStudyModelInstanceOnce = sync.Once{}
)

func GetHomeFunStudyModel() IHomeFunStudyModel {
	homeFunStudyModelInstanceOnce.Do(func() {
		homeFunStudyModelInstance = &homeFunStudyModel{}
	})
	return homeFunStudyModelInstance
}

type IHomeFunStudyModel interface {
	SaveHomeFunStudy(ctx context.Context, operator *entity.Operator, cmd entity.SaveHomeFunStudyArgs) error
}

type homeFunStudyModel struct{}

func (m *homeFunStudyModel) SaveHomeFunStudy(ctx context.Context, operator *entity.Operator, cmd entity.SaveHomeFunStudyArgs) error {
	panic("implement me")
}

func (m *homeFunStudyModel) DecodeTeacherIDs(s string) ([]string, error) {
	if s == "[]" {
		return nil, nil
	}
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *homeFunStudyModel) EncodeTeacherIDs(ctx context.Context, teacherIDs []string) (string, error) {
	if len(teacherIDs) == 0 {
		return "[]", nil
	}
	bs, err := json.Marshal(teacherIDs)
	if err != nil {
		log.Error(ctx, "EncodeTeacherIDs: json marshal error",
			log.Strings("teacher_ids", teacherIDs),
			log.Err(err),
		)
		return "", err
	}
	return string(bs), nil
}

func (m *homeFunStudyModel) EncodeTitle(className string, lessonName string) string {
	if className == "" {
		className = constant.HomeFunStudyNoClass
	}
	return fmt.Sprintf("%s-%s", className, lessonName)
}
