package intergrate_academic_profile

import (
	"context"
	"log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
)

func (s MapperImpl) Grade(ctx context.Context, organizationID, programID, gradeID string) (string, error) {
	oldGrade, err := model.GetGradeModel().GetByID(ctx, gradeID)
	if err != nil {
		return "", err
	}

	newProgramID, err := s.Program(ctx, organizationID, programID)
	newGrades, err := external.GetAgeServiceProvider().GetByProgram(ctx, &entity.Operator{
		OrgID: organizationID,
	}, newProgramID)
	if err != nil {
		return "", err
	}

	var result string
	var count int
	for _, item := range newGrades {
		if oldGrade.Name == item.Name {
			count++
			if count > 1 {
				break
			}
			result = item.ID
		}
	}
	if count <= 0 {
		log.Printf("age not found;orgID:%s,newProgramID:%s,gradeName:%s", organizationID, newProgramID, oldGrade.Name)
		return "", constant.ErrRecordNotFound
	}
	if count > 1 {
		log.Printf("age has many;orgID:%s,newProgramID:%s,gradeName:%s", organizationID, newProgramID, oldGrade.Name)
		return "", constant.ErrConflict
	}

	return result, nil
}
