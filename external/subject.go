package external

import (
    "context"
    "fmt"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/external/gql"
    "strings"

    "gitlab.badanamu.com.cn/calmisland/chlorine"
    "gitlab.badanamu.com.cn/calmisland/common-log/log"
    "gitlab.badanamu.com.cn/calmisland/kidsloop-cache/cache"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
    "gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type SubjectServiceProvider interface {
    cache.IDataSource
    BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Subject, error)
    BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Subject, error)
    BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
    GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Subject, error)
    GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Subject, error)
}

type Subject struct {
    ID     string   `json:"id"`
    Name   string   `json:"name"`
    Status APStatus `json:"status"`
    System bool     `json:"system"`
}

func (n *Subject) StringID() string {
    return n.ID
}
func (n *Subject) RelatedIDs() []*cache.RelatedEntity {
    return nil
}

func GetSubjectServiceProvider() SubjectServiceProvider {
    return &AmsSubjectService{}
}

type AmsSubjectService struct{}

func (s AmsSubjectService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
    if len(ids) == 0 {
        return nil, nil
    }
    operator, err := optionsWithOperator(ctx, options...)
    if err != nil {
        return nil, err
    }

    _ids, indexMapping := utils.SliceDeduplicationMap(ids)

    sb := new(strings.Builder)
    fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$subject_id_", ": ID!", len(_ids)))
    for index := range _ids {
        fmt.Fprintf(sb, "q%d: subject(id: $subject_id_%d) {id name status system}\n", index, index)
    }
    sb.WriteString("}")

    request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
    for index, id := range _ids {
        request.Var(fmt.Sprintf("subject_id_%d", index), id)
    }

    data := map[string]*Subject{}
    response := &chlorine.Response{
        Data: &data,
    }

    _, err = GetAmsClient().Run(ctx, request, response)
    if err != nil {
        log.Error(ctx, "get subjects by ids failed",
            log.Err(err),
            log.Strings("ids", ids))
        return nil, err
    }

    if len(response.Errors) > 0 {
        log.Error(ctx, "get subjects by ids failed",
            log.Err(response.Errors),
            log.Any("operator", operator),
            log.Strings("ids", ids))
        return nil, response.Errors
    }

    subjects := make([]cache.Object, 0, len(data))
    for index := range ids {
        subject := data[fmt.Sprintf("q%d", indexMapping[index])]
        if subject == nil {
            log.Debug(ctx, "subject not found", log.String("id", ids[index]))
            continue
        }
        subjects = append(subjects, subject)
    }

    log.Info(ctx, "get subjects by ids success",
        log.Strings("ids", ids),
        log.Any("subjects", subjects))

    return subjects, nil
}

func (s AmsSubjectService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*Subject, error) {
    if len(ids) == 0 {
        return []*Subject{}, nil
    }

    uuids := make([]string, 0, len(ids))
    for _, id := range ids {
        if utils.IsValidUUID(id) {
            uuids = append(uuids, id)
        } else {
            log.Warn(ctx, "invalid uuid type", log.String("id", id))
        }
    }

    res := make([]*Subject, 0, len(uuids))
    err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), uuids, &res, operator)
    if err != nil {
        return nil, err
    }

    return res, nil
}

func (s AmsSubjectService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*Subject, error) {
    subjects, err := s.BatchGet(ctx, operator, ids)
    if err != nil {
        return map[string]*Subject{}, err
    }

    dict := make(map[string]*Subject, len(subjects))
    for _, subject := range subjects {
        dict[subject.ID] = subject
    }

    return dict, nil
}

func (s AmsSubjectService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
    subjects, err := s.BatchGet(ctx, operator, ids)
    if err != nil {
        return map[string]string{}, err
    }

    dict := make(map[string]string, len(subjects))
    for _, subject := range subjects {
        dict[subject.ID] = subject.Name
    }

    return dict, nil
}

func (s AmsSubjectService) getWithProgram(ctx context.Context, operator *entity.Operator, programID string, condition *APCondition) ([]*Subject, error) {
    filter := gql.SubjectFilter{
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
    var subjects []*Subject
    var pages []gql.SubjectsConnectionResponse
    err := gql.Query(ctx, operator, filter.FilterType(), filter, &pages)
    if err != nil {
        log.Error(ctx, "get subject by program failed",
            log.Err(err),
            log.Any("operator", operator),
            log.Any("filter", filter))
        return nil, err
    }
    for _, p := range pages {
        for _, v := range p.Edges {
            obj := &Subject{
                ID:     v.Node.ID,
                Name:   v.Node.Name,
                Status: APStatus(v.Node.Status),
                System: v.Node.System,
            }
            subjects = append(subjects, obj)
        }
    }
    return subjects, nil
}
func (s AmsSubjectService) getByProgram(ctx context.Context, operator *entity.Operator, programID string, condition *APCondition) ([]*Subject, error) {
    request := chlorine.NewRequest(`
	query($program_id: ID!) {
		program(id: $program_id) {
			subjects {
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

    if len(response.Errors) > 0 {
        log.Error(ctx, "get ages by ids failed",
            log.Err(response.Errors),
            log.Any("operator", operator),
            log.String("programID", programID))
        return nil, response.Errors
    }

    subjects := make([]*Subject, 0, len(data.Program.Subjects))
    for _, subject := range data.Program.Subjects {
        if condition.Status.Valid {
            if condition.Status.Status != subject.Status {
                continue
            }
        } else {
            // only status = "Active" data is returned by default
            if subject.Status != Active {
                continue
            }
        }

        if condition.System.Valid && subject.System != condition.System.Bool {
            continue
        }

        subjects = append(subjects, subject)
    }

    log.Info(ctx, "get subjects by program success",
        log.Any("operator", operator),
        log.String("programID", programID),
        log.Any("subjects", subjects))

    return subjects, nil
}
func (s AmsSubjectService) GetByProgram(ctx context.Context, operator *entity.Operator, programID string, options ...APOption) ([]*Subject, error) {
    condition := NewCondition(options...)
    if config.Get().AMS.ReplaceWithConnection {
        return s.getWithProgram(ctx, operator, programID, condition)
    }
    return s.getByProgram(ctx, operator, programID, condition)
}

func (s AmsSubjectService) getWithOrganization(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*Subject, error) {
    filter := gql.SubjectFilter{
        OrganizationID: &gql.UUIDFilter{
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
    var subjects []*Subject
    var pages []gql.SubjectsConnectionResponse
    err := gql.Query(ctx, operator, filter.FilterType(), filter, &pages)
    if err != nil {
        log.Error(ctx, "get subject by organization failed",
            log.Err(err),
            log.Any("operator", operator),
            log.Any("filter", filter))
        return nil, err
    }
    for _, p := range pages {
        for _, v := range p.Edges {
            obj := &Subject{
                ID:     v.Node.ID,
                Name:   v.Node.Name,
                Status: APStatus(v.Node.Status),
                System: v.Node.System,
            }
            subjects = append(subjects, obj)
        }
    }
    return subjects, nil
}
func (s AmsSubjectService) getByOrganization(ctx context.Context, operator *entity.Operator, id string, condition *APCondition) ([]*Subject, error) {
    request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			subjects {
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

    if len(response.Errors) > 0 {
        log.Error(ctx, "query subjects by operator failed",
            log.Err(response.Errors),
            log.Any("operator", operator))
        return nil, response.Errors
    }

    subjects := make([]*Subject, 0, len(data.Organization.Subjects))
    for _, subject := range data.Organization.Subjects {
        if condition.Status.Valid {
            if condition.Status.Status != subject.Status {
                continue
            }
        } else {
            // only status = "Active" data is returned by default
            if subject.Status != Active {
                continue
            }
        }

        if condition.System.Valid && subject.System != condition.System.Bool {
            continue
        }

        subjects = append(subjects, subject)
    }

    log.Info(ctx, "get subjects by operator success",
        log.Any("operator", operator),
        log.Any("subjects", subjects))

    return subjects, nil
}
func (s AmsSubjectService) GetByOrganization(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*Subject, error) {
    condition := NewCondition(options...)

    if config.Get().AMS.ReplaceWithConnection {
        return s.getWithOrganization(ctx, operator, operator.OrgID, condition)
    }
    return s.getByOrganization(ctx, operator, operator.OrgID, condition)
}

func (s AmsSubjectService) Name() string {
    return "ams_subject_service"
}
