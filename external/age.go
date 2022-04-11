package external

import (
    "context"
    "fmt"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/external/gql"
    "strings"

    "gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"

    "gitlab.badanamu.com.cn/calmisland/chlorine"
    "gitlab.badanamu.com.cn/calmisland/common-log/log"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type AgeServiceProvider interface {
    cache.IDataSource
    BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Age, error)
    BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Age, error)
    BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
    GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Age, error)
    GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Age, error)
}

type Age struct {
    ID     string   `json:"id"`
    Name   string   `json:"name"`
    Status APStatus `json:"status"`
    System bool     `json:"system"`
}

func (n *Age) StringID() string {
    return n.ID
}
func (n *Age) RelatedIDs() []*cache.RelatedEntity {
    return nil
}
func GetAgeServiceProvider() AgeServiceProvider {
    return &AmsAgeService{}
}

type AmsAgeService struct{}

func (s AmsAgeService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Age, error) {
    if len(ids) == 0 {
        return []*Age{}, nil
    }

    uuids := make([]string, 0, len(ids))
    for _, id := range ids {
        if utils.IsValidUUID(id) {
            uuids = append(uuids, id)
        } else {
            log.Warn(ctx, "invalid uuid type", log.String("id", id))
        }
    }

    res := make([]*Age, 0, len(uuids))
    err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), uuids, &res, operator)
    if err != nil {
        return nil, err
    }

    return res, nil
}
func (s AmsAgeService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
    if len(ids) == 0 {
        return nil, nil
    }
    operator, err := optionsWithOperator(ctx, options...)
    if err != nil {
        return nil, err
    }

    _ids, indexMapping := utils.SliceDeduplicationMap(ids)

    sb := new(strings.Builder)
    fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$age_id_", ": ID!", len(_ids)))
    for index := range _ids {
        fmt.Fprintf(sb, "q%d: age_range(id: $age_id_%d) {id name status system}\n", index, index)
    }
    sb.WriteString("}")

    request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
    for index, id := range _ids {
        request.Var(fmt.Sprintf("age_id_%d", index), id)
    }

    data := map[string]*Age{}
    response := &chlorine.Response{
        Data: &data,
    }

    _, err = GetAmsClient().Run(ctx, request, response)
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

    ages := make([]cache.Object, 0, len(data))
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

func (s AmsAgeService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Age, error) {
    ages, err := s.BatchGet(ctx, operator, ids)
    if err != nil {
        return map[string]*Age{}, err
    }

    dict := make(map[string]*Age, len(ages))
    for _, age := range ages {
        dict[age.ID] = age
    }

    return dict, nil
}

func (s AmsAgeService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
    ages, err := s.BatchGet(ctx, operator, ids)
    if err != nil {
        return map[string]string{}, err
    }

    dict := make(map[string]string, len(ages))
    for _, age := range ages {
        dict[age.ID] = age.Name
    }

    return dict, nil
}

func (s AmsAgeService) getWithProgram(ctx context.Context, operator *entity.Operator, programID string, condition *APCondition) ([]*Age, error) {
    filter := gql.AgeRangeFilter{
        ProgramID: &gql.UUIDFilter{
            Operator: gql.OperatorTypeEq,
            Value:    gql.UUID(programID),
        },
        Status: &gql.StringFilter{
            Operator: gql.OperatorTypeEq,
            Value:    Active.String(),
        },
    }

    if condition.Status.Valid {
        filter.Status.Value = condition.Status.Status.String()
    }
    if condition.System.Valid {
        filter.System = &gql.BooleanFilter{
            Operator: gql.OperatorTypeEq,
            Value:    condition.System.Valid,
        }
    }
    var pages []gql.AgesConnectionResponse
    err := gql.Query(ctx, operator, filter.FilterType(), filter, &pages)
    if err != nil {
        log.Error(ctx, "get age by program failed",
            log.Err(err),
            log.Any("operator", operator),
            log.Any("filter", filter))
        return nil, err
    }
    var ages []*Age
    for _, page := range pages {
        for _, v := range page.Edges {
            age := Age{
                ID:     v.Node.ID,
                Name:   v.Node.Name,
                Status: APStatus(v.Node.Status),
                System: v.Node.System,
            }
            ages = append(ages, &age)
        }
    }
    return ages, nil
}

func (s AmsAgeService) getByProgram(ctx context.Context, operator *entity.Operator, programID string, condition *APCondition) ([]*Age, error) {
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

func (s AmsAgeService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Age, error) {
    condition := NewCondition(options...)
    if config.Get().AMS.ReplaceWithConnection {
        return s.getWithProgram(ctx, operator, programID, condition)
    }
    return s.getByProgram(ctx, operator, programID, condition)
}

func (s AmsAgeService) getWithOrganization(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*Age, error) {
    filter := gql.AgeRangeFilter{
        ProgramID: &gql.UUIDFilter{
            Operator: gql.OperatorTypeEq,
            Value:    gql.UUID(id),
        },
        Status: &gql.StringFilter{
            Operator: gql.OperatorTypeEq,
            Value:    Active.String(),
        },
    }

    if condition.Status.Valid {
        filter.Status.Value = condition.Status.Status.String()
    }
    if condition.System.Valid {
        filter.System = &gql.BooleanFilter{
            Operator: gql.OperatorTypeEq,
            Value:    condition.System.Valid,
        }
    }

    var pages []gql.AgesConnectionResponse
    err := gql.Query(ctx, operator, filter.FilterType(), filter, &pages)
    if err != nil {
        log.Error(ctx, "get age by organization failed",
            log.Err(err),
            log.Any("operator", operator),
            log.Any("filter", filter))
        return nil, err
    }
    var ages []*Age
    for _, page := range pages {
        for _, v := range page.Edges {
            age := Age{
                ID:     v.Node.ID,
                Name:   v.Node.Name,
                Status: APStatus(v.Node.Status),
                System: v.Node.System,
            }
            ages = append(ages, &age)
        }
    }
    return ages, nil
}

func (s AmsAgeService) getByOrganization(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*Age, error) {
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
    request.Var("organization_id", id)

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
func (s AmsAgeService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Age, error) {
    condition := NewCondition(options...)

    if config.Get().AMS.ReplaceWithConnection {
        return s.getWithOrganization(ctx, operator, operator.OrgID, condition)
    }
    return s.getByOrganization(ctx, operator, operator.OrgID, condition)
}
func (s AmsAgeService) Name() string {
    return "ams_age_service"
}
