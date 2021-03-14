package external

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type SubjectServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Subject, error)
	GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Subject, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*Subject, error)
}

type Subject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetSubjectServiceProvider() SubjectServiceProvider {
	return &AmsSubjectService{}
}

type AmsSubjectService struct{}

func (s AmsSubjectService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Subject, error) {
	if len(ids) == 0 {
		return []*Subject{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: subject(id: \"%s\") {id name}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*Subject{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get subjects by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	subjects := make([]*Subject, 0, len(data))
	for index := range ids {
		subject := data[fmt.Sprintf("q%d", indexMapping[index])]
		if subject == nil {
			log.Error(ctx, "subject not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		subjects = append(subjects, subject)
	}

	log.Info(ctx, "get subjects by ids success",
		log.Strings("ids", ids),
		log.Any("subjects", subjects))

	return subjects, nil
}

func (s AmsSubjectService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string) ([]*Subject, error) {
	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			subjects {
				id
				name
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("program_id", programID)

	data := &struct {
		Program struct {
			Subjects []*Subject `json:"subjects"`
		} `json:"program"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query subjects by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("programID", programID))
		return nil, err
	}

	subjects := data.Program.Subjects

	log.Info(ctx, "get subjects by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("subjects", subjects))

	return subjects, nil
}

func (s AmsSubjectService) GetByOrganization(ctx context.Context, operator *entity.Operator) ([]*Subject, error) {
	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			subjects {
				id
				name
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", operator.OrgID)

	data := &struct {
		Organization struct {
			Subjects []*Subject `json:"subjects"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query subjects by operator failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	subjects := data.Organization.Subjects

	log.Info(ctx, "get subjects by operator success",
		log.Any("operator", operator),
		log.Any("subjects", subjects))

	return subjects, nil
}
