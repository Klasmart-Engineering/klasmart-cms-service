package external

import (
	"context"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type StudentFilter UserFilter

func (StudentFilter) FilterName() FilterType {
	return UserFilterType
}

func (StudentFilter) ConnectionName() ConnectionType {
	return StudentsConnectionType
}

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

func (ascs AmsStudentConnectionService) pageNodes(ctx context.Context, operator *entity.Operator, pages []UsersConnectionResponse) []*Student {
	if len(pages) == 0 {
		log.Warn(ctx, "pageNodes is empty",
			log.Any("operator", operator))
		return []*Student{}
	}
	students := make([]*Student, 0, pages[0].TotalCount)
	exists := make(map[string]bool)
	for _, page := range pages {
		for _, edge := range page.Edges {
			if _, ok := exists[edge.Node.ID]; ok {
				log.Warn(ctx, "pageNodes: student exist",
					log.Any("student", edge.Node),
					log.Any("operator", operator))
				continue
			}
			exists[edge.Node.ID] = true
			student := &Student{
				ID:         edge.Node.ID,
				GivenName:  edge.Node.GivenName,
				FamilyName: edge.Node.FamilyName,
			}
			students = append(students, student)
		}
	}
	return students
}

func (ascs AmsStudentConnectionService) GetByClassIDs(ctx context.Context, operator *entity.Operator, classIDs []string) (map[string][]*Student, error) {
	result := make(map[string][]UsersConnectionResponse)
	IDs := utils.SliceDeduplicationExcludeEmpty(classIDs)
	err := subPageQuery(ctx, operator, "classNode", StudentFilter{}, IDs, result)
	if err != nil {
		log.Error(ctx, "GetByClassIDs: subPageQuery failed",
			log.Err(err),
			log.Strings("class_ids", classIDs))
		return nil, err
	}
	studentsMap := make(map[string][]*Student)
	for k, pages := range result {
		log.Error(ctx, "GetByClassIDs", log.String("class_id", k))
		studentsMap[k] = ascs.pageNodes(ctx, operator, pages)
	}
	return studentsMap, nil
}
