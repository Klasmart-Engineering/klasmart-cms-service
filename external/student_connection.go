package external

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type AmsStudentConnectionService struct {
	AmsStudentService
}

func (ascs AmsStudentConnectionService) GetByClassID(ctx context.Context, operator *entity.Operator, classID string) ([]*Student, error) {
	studentMaps, err := ascs.GetByClassIDs(ctx, operator, []string{classID})
	if err != nil {
		log.Error(ctx, "GetByClassID: GetByClassIDs failed",
			log.Err(err),
			log.String("class_id", classID))
		return nil, err
	}
	return studentMaps[classID], nil
}

func (ascs AmsStudentConnectionService) GetByClassIDs(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string][]*Student, error) {
	result := make(map[string][]UsersConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(classIDs)
	err := subPageQuery(ctx, operator, "classNode", "studentsConnection", IDs, result)
	if err != nil {
		log.Error(ctx, "GetByClassIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("class_ids", classIDs))
		return nil, err
	}
	studentsMap := make(map[string][]*Student)
	for k, pages := range result {
		for _, page := range pages {
			for _, edge := range page.Edges {
				student := &Student{
					ID:   edge.Node.ID,
					Name: edge.Node.GivenName,
				}
				studentsMap[k] = append(studentsMap[k], student)
			}
		}
	}
	return studentsMap, nil
}
