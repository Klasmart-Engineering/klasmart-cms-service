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

type AgeServiceProvider interface {
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Age, error)
	GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Age, error)
	GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Age, error)
}

type Age struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Status APStatus `json:"status"`
	System bool     `json:"system"`
}

func GetAgeServiceProvider() AgeServiceProvider {
	return &AmsAgeService{}
}

type AmsAgeService struct{}

func (s AmsAgeService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Age, error) {
	if len(ids) == 0 {
		return []*Age{}, nil
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)
	sb.WriteString("query {")
	for index, id := range _ids {
		fmt.Fprintf(sb, "q%d: age_range(id: \"%s\") {id name status system}\n", index, id)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))

	data := map[string]*Age{}
	response := &chlorine.Response{
		Data: &data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get ages by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get ages by ids failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Strings("ids", ids))
		return nil, response.Errors
	}

	ages := make([]*Age, 0, len(data))
	for index := range ids {
		age := data[fmt.Sprintf("q%d", indexMapping[index])]
		if age == nil {
			log.Error(ctx, "age not found", log.String("id", ids[index]))
			return nil, constant.ErrRecordNotFound
		}
		ages = append(ages, age)
	}

	log.Info(ctx, "get ages by ids success",
		log.Strings("ids", ids),
		log.Any("ages", ages))

	return ages, nil
}

func (s AmsAgeService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Age, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			age_ranges {
				id
				name
				status
				system
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("program_id", programID)

	data := &struct {
		Program struct {
			Ages []*Age `json:"age_ranges"`
		} `json:"program"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query ages by operator failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("programID", programID),
			log.Any("condition", condition))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "get ages by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.String("programID", programID),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	ages := make([]*Age, 0, len(data.Program.Ages))
	for _, age := range data.Program.Ages {
		if condition.Status.Valid {
			if condition.Status.Status != age.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if age.Status != Active {
				continue
			}
		}

		if condition.System.Valid && age.System != condition.System.Bool {
			continue
		}

		ages = append(ages, age)
	}

	log.Info(ctx, "get ages by program success",
		log.Any("operator", operator),
		log.String("programID", programID),
		log.Any("condition", condition),
		log.Any("ages", ages))

	return ages, nil
}

func (s AmsAgeService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Age, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			ageRanges {
				id
				name
				status
				system
			}			
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("organization_id", operator.OrgID)

	data := &struct {
		Organization struct {
			Ages []*Age `json:"ageRanges"`
		} `json:"organization"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "query ages by operator failed",
			log.Err(err),
			log.Any("operator", operator))
		return nil, err
	}

	if len(response.Errors) > 0 {
		log.Error(ctx, "query ages by operator failed",
			log.Err(response.Errors),
			log.Any("operator", operator),
			log.Any("condition", condition))
		return nil, response.Errors
	}

	ages := make([]*Age, 0, len(data.Organization.Ages))
	for _, age := range data.Organization.Ages {
		if condition.Status.Valid {
			if condition.Status.Status != age.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if age.Status != Active {
				continue
			}
		}

		if condition.System.Valid && age.System != condition.System.Bool {
			continue
		}

		ages = append(ages, age)
	}

	log.Info(ctx, "get ages by operator success",
		log.Any("operator", operator),
		log.Any("condition", condition),
		log.Any("ages", ages))

	return ages, nil
}