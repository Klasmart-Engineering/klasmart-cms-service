package external

import (
	"context"
	"fmt"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"strings"
	"sync"

	"github.com/KL-Engineering/chlorine"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cache/cache"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"github.com/KL-Engineering/kidsloop-cms-service/utils/kl2cache"
)

type SchoolServiceProvider interface {
	cache.IDataSource
	BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableSchool, error)
	BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableSchool, error)
	BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error)
	GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string][]*School, error)
	GetByOrganizationID(ctx context.Context, operator *entity.Operator, organizationID string, options ...APOption) ([]*School, error)
	GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName, options ...APOption) ([]*School, error)
	GetByOperator(ctx context.Context, operator *entity.Operator, options ...APOption) ([]*School, error)
	GetByUsers(ctx context.Context, operator *entity.Operator, orgID string, userIDs []string, options ...APOption) (map[string][]*School, error)
}

type School struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Status         APStatus `json:"status"`
	OrganizationId string   `json:"organizationId" gqls:"organizationId"`
}

type NullableSchool struct {
	Valid bool   `json:"valid"`
	StrID string `json:"str_id"`
	*School
}

func (n *NullableSchool) StringID() string {
	return n.StrID
}
func (n *NullableSchool) RelatedIDs() []*cache.RelatedEntity {
	return nil
}

var (
	_amsSchoolService SchoolServiceProvider
	_amsSchoolOnce    sync.Once
)

func GetSchoolServiceProvider() SchoolServiceProvider {
	_amsSchoolOnce.Do(func() {
		if config.Get().AMS.UseDeprecatedQuery {
			_amsSchoolService = &AmsSchoolService{
				BaseCacheKey: kl2cache.KeyByStrings{
					"kl2cache",
					"AmsSchoolService",
				},
			}
		} else {
			_amsSchoolService = &AmsSchoolConnectionService{}
		}
	})

	return _amsSchoolService
}

type AmsSchoolService struct {
	BaseCacheKey kl2cache.KeyByStrings
}

func (s AmsSchoolService) BatchGet(ctx context.Context, operator *entity.Operator, ids []string) ([]*NullableSchool, error) {
	if len(ids) == 0 {
		return []*NullableSchool{}, nil
	}
	res := make([]*NullableSchool, 0, len(ids))
	err := cache.GetPassiveCacheRefresher().BatchGet(ctx, s.Name(), ids, &res, operator)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s AmsSchoolService) QueryByIDs(ctx context.Context, ids []string, options ...interface{}) ([]cache.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	operator, err := optionsWithOperator(ctx, options...)
	if err != nil {
		return nil, err
	}

	_ids, indexMapping := utils.SliceDeduplicationMap(ids)

	sb := new(strings.Builder)

	fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$school_id_", ": ID!", len(_ids)))
	for index := range _ids {
		fmt.Fprintf(sb, "q%d: school(school_id: $school_id_%d) {id:school_id name:school_name status}\n", index, index)
	}
	sb.WriteString("}")

	request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
	for index, id := range _ids {
		request.Var(fmt.Sprintf("school_id_%d", index), id)
	}

	data := map[string]*School{}

	response := &chlorine.Response{
		Data: &data,
	}

	_, err = GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get schools by ids failed",
			log.Err(err),
			log.Strings("ids", ids))
		return nil, err
	}

	schools := make([]cache.Object, 0, len(data))
	for index := range ids {
		school := data[fmt.Sprintf("q%d", indexMapping[index])]
		schools = append(schools, &NullableSchool{
			Valid:  school != nil,
			School: school,
			StrID:  ids[index],
		})
	}

	log.Info(ctx, "get schools by ids success",
		log.Strings("ids", ids),
		log.Any("schools", schools))

	return schools, nil
}

func (s AmsSchoolService) BatchGetMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]*NullableSchool, error) {
	schools, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]*NullableSchool{}, err
	}

	dict := make(map[string]*NullableSchool, len(schools))
	for _, school := range schools {
		if school.School == nil || !school.Valid {
			continue
		}
		dict[school.ID] = school
	}

	return dict, nil
}

func (s AmsSchoolService) BatchGetNameMap(ctx context.Context, operator *entity.Operator, ids []string) (map[string]string, error) {
	schools, err := s.BatchGet(ctx, operator, ids)
	if err != nil {
		return map[string]string{}, err
	}

	dict := make(map[string]string, len(schools))
	for _, school := range schools {
		if school.School == nil || !school.Valid {
			continue
		}
		dict[school.ID] = school.Name
	}

	return dict, nil
}

//TODO:Test failed
func (s AmsSchoolService) GetByClasses(ctx context.Context, operator *entity.Operator, classIDs []string, options ...APOption) (map[string][]*School, error) {
	_classIDs := utils.SliceDeduplicationExcludeEmpty(classIDs)

	if len(classIDs) == 0 {
		return map[string][]*School{}, nil
	}

	schools := make(map[string][]*School, len(classIDs))
	var mapLock sync.RWMutex

	total := len(_classIDs)
	pageSize := constant.AMSRequestUserClassPageSize
	pageCount := (total + pageSize - 1) / pageSize

	condition := NewCondition(options...)
	cerr := make(chan error, pageCount)

	for i := 0; i < pageCount; i++ {
		go func(j int) {
			start := j * pageSize
			end := (j + 1) * pageSize
			if end >= total {
				end = total
			}
			pageClassIDs := _classIDs[start:end]

			sb := new(strings.Builder)
			fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$class_id_", ": ID!", len(pageClassIDs)))
			for index := range pageClassIDs {
				fmt.Fprintf(sb, `q%d: class(class_id: $class_id_%d) {schools{id:school_id name:school_name status}}`, index, index)
			}
			sb.WriteString("}")

			request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
			for index, id := range pageClassIDs {
				request.Var(fmt.Sprintf("class_id_%d", index), id)
			}

			data := map[string]*struct {
				Schools []*School `json:"schools"`
			}{}

			response := &chlorine.Response{
				Data: &data,
			}

			_, err := GetAmsClient().Run(ctx, request, response)
			if err != nil {
				log.Error(ctx, "get schools by classes failed",
					log.Err(err),
					log.Any("operator", operator),
					log.Strings("pageClassIDs", pageClassIDs))
				cerr <- err
				return
			}

			for index, classID := range pageClassIDs {
				class := data[fmt.Sprintf("q%d", index)]
				if class == nil {
					continue
				}
				mapLock.Lock()
				schools[classID] = make([]*School, 0, len(class.Schools))
				mapLock.Unlock()
				for _, school := range class.Schools {
					if condition.Status.Valid {
						if condition.Status.Status != school.Status {
							continue
						}
					} else {
						// only status = "Active" data is returned by default
						if school.Status != Active {
							continue
						}
					}
					mapLock.RLock()
					schools[classID] = append(schools[classID], school)
					mapLock.RUnlock()
				}
			}
			cerr <- nil
		}(i)
	}

	for i := 0; i < pageCount; i++ {
		if err := <-cerr; err != nil {
			return nil, err
		}
	}

	log.Info(ctx, "get schools by classes success",
		log.Any("operator", operator),
		log.Strings("classIDs", classIDs),
		log.Any("schools", schools))

	return schools, nil
}

func (s AmsSchoolService) GetByOrganizationID(ctx context.Context, operator *entity.Operator, organizationID string, options ...APOption) (schools []*School, err error) {
	schools = []*School{}
	key := append(s.BaseCacheKey,
		"GetByOrganizationID",
		operator.UserID,
		organizationID)
	fGetData := func(ctx context.Context, key kl2cache.Key) (val interface{}, err error) {
		request := chlorine.NewRequest(`
	query($organization_id: ID!) {
		organization(organization_id: $organization_id) {
			schools{
				school_id
				school_name
				status
			}
		}
	}`, chlorine.ReqToken(operator.Token))
		request.Var("organization_id", organizationID)

		data := &struct {
			Organization struct {
				Schools []struct {
					SchoolID   string   `json:"school_id"`
					SchoolName string   `json:"school_name"`
					Status     APStatus `json:"status"`
				} `json:"schools"`
			} `json:"organization"`
		}{}

		response := &chlorine.Response{
			Data: data,
		}

		_, err = GetAmsClient().Run(ctx, request, response)
		if err != nil {
			log.Error(ctx, "query schools by organization failed",
				log.Err(err),
				log.String("organizationID", organizationID))
			return
		}

		schools := make([]*School, 0, len(data.Organization.Schools))
		for _, school := range data.Organization.Schools {
			schools = append(schools, &School{
				ID:     school.SchoolID,
				Name:   school.SchoolName,
				Status: school.Status,
			})
		}
		val = schools

		log.Info(ctx, "query schools by organization success",
			log.String("organizationID", organizationID),
			log.Any("schools", schools))
		return
	}
	err = kl2cache.DefaultProvider.Get(ctx, key, &schools, fGetData)
	if err != nil {
		return
	}

	condition := NewCondition(options...)
	stat := Active
	if condition.Status.Valid {
		stat = condition.Status.Status
	}
	schools1 := make([]*School, 0, len(schools))
	for _, s := range schools {
		if s.Status != stat {
			continue
		}
		schools1 = append(schools1, s)
	}
	schools = schools1
	return
}

func (s AmsSchoolService) GetByPermission(ctx context.Context, operator *entity.Operator, permissionName PermissionName, options ...APOption) ([]*School, error) {
	condition := NewCondition(options...)

	request := chlorine.NewRequest(`
	query(
		$user_id: ID!
		$permission_name: String!
	) {
		user(user_id: $user_id) {
			schoolsWithPermission(permission_name: $permission_name) {
				school {
					school_id
					school_name
					status
					organization {
						organization_id
					}
				}
			}
		}
	}`, chlorine.ReqToken(operator.Token))
	request.Var("user_id", operator.UserID)
	request.Var("permission_name", permissionName.String())

	data := &struct {
		User struct {
			SchoolsWithPermission []struct {
				School struct {
					SchoolID     string   `json:"school_id"`
					SchoolName   string   `json:"school_name"`
					Status       APStatus `json:"status"`
					Organization struct {
						OrganizationID string `json:"organization_id"`
					} `json:"organization"`
				} `json:"school"`
			} `json:"schoolsWithPermission"`
		} `json:"user"`
	}{}

	response := &chlorine.Response{
		Data: data,
	}

	_, err := GetAmsClient().Run(ctx, request, response)
	if err != nil {
		log.Error(ctx, "get schools by permission failed",
			log.Err(err),
			log.Any("operator", operator),
			log.String("permissionName", permissionName.String()))
		return nil, err
	}

	schools := make([]*School, 0, len(data.User.SchoolsWithPermission))
	for _, membership := range data.User.SchoolsWithPermission {
		// filtering by operator's org id
		if membership.School.Organization.OrganizationID != operator.OrgID {
			continue
		}

		if condition.Status.Valid {
			if condition.Status.Status != membership.School.Status {
				continue
			}
		} else {
			// only status = "Active" data is returned by default
			if membership.School.Status != Active {
				continue
			}
		}

		schools = append(schools, &School{
			ID:     membership.School.SchoolID,
			Name:   membership.School.SchoolName,
			Status: membership.School.Status,
		})
	}

	log.Info(ctx, "get schools by permission",
		log.Any("operator", operator),
		log.String("permissionName", permissionName.String()),
		log.Any("schools", schools))

	return schools, nil
}

func (s AmsSchoolService) GetByOperator(ctx context.Context, operator *entity.Operator, options ...APOption) (schools []*School, err error) {
	schools = []*School{}
	key := append(s.BaseCacheKey,
		"GetByOperator",
		operator.OrgID,
		operator.UserID)
	fGetData := func(ctx context.Context, key kl2cache.Key) (val interface{}, err error) {
		request := chlorine.NewRequest(`
	query($user_id: ID!) {
		user(user_id: $user_id) {
			school_memberships{
				school {
					school_id
					school_name
					status
					organization {
						organization_id
					}
				}
			}
		}
	}`, chlorine.ReqToken(operator.Token))
		request.Var("user_id", operator.UserID)

		data := &struct {
			User struct {
				SchoolMemberships []struct {
					School struct {
						SchoolID     string   `json:"school_id"`
						SchoolName   string   `json:"school_name"`
						Status       APStatus `json:"status"`
						Organization struct {
							OrganizationID string `json:"organization_id"`
						} `json:"organization"`
					} `json:"school"`
				} `json:"school_memberships"`
			} `json:"user"`
		}{}

		response := &chlorine.Response{
			Data: data,
		}

		_, err = GetAmsClient().Run(ctx, request, response)
		if err != nil {
			log.Error(ctx, "get schools by operator failed",
				log.Err(err),
				log.Any("operator", operator))
			return nil, err
		}

		schools := make([]*School, 0, len(data.User.SchoolMemberships))
		for _, membership := range data.User.SchoolMemberships {
			// filtering by operator's org id
			if membership.School.Organization.OrganizationID != operator.OrgID {
				continue
			}
			schools = append(schools, &School{
				ID:     membership.School.SchoolID,
				Name:   membership.School.SchoolName,
				Status: membership.School.Status,
			})
		}
		val = schools

		log.Info(ctx, "get schools by operator success",
			log.Any("operator", operator),
			log.Any("schools", schools))
		return
	}
	err = kl2cache.DefaultProvider.Get(ctx, key, &schools, fGetData)
	if err != nil {
		return
	}
	condition := NewCondition(options...)
	stat := Active
	if condition.Status.Valid {
		stat = condition.Status.Status
	}
	schools1 := make([]*School, 0, len(schools))
	for _, school := range schools {
		if school.Status != stat {
			continue
		}
		schools1 = append(schools1, school)
	}
	schools = schools1
	return
}

func (s AmsSchoolService) GetByUsers(ctx context.Context, operator *entity.Operator, orgID string, userIDs []string, options ...APOption) (map[string][]*School, error) {
	_userIDs := utils.SliceDeduplicationExcludeEmpty(userIDs)

	if len(_userIDs) == 0 {
		return map[string][]*School{}, nil
	}

	schools := make(map[string][]*School, len(_userIDs))
	var mapLock sync.RWMutex

	total := len(_userIDs)
	pageSize := constant.AMSRequestUserSchoolPageSize
	pageCount := (total + pageSize - 1) / pageSize

	condition := NewCondition(options...)
	cerr := make(chan error, pageCount)

	for i := 0; i < pageCount; i++ {
		go func(j int) {
			start := j * pageSize
			end := (j + 1) * pageSize
			if end >= total {
				end = total
			}
			pageUserIDs := _userIDs[start:end]

			sb := new(strings.Builder)

			fmt.Fprintf(sb, "query (%s) {", utils.StringCountRange(ctx, "$user_id_", ": ID!", len(pageUserIDs)))
			for index := range pageUserIDs {
				fmt.Fprintf(sb, `q%d: user(user_id: $user_id_%d) {school_memberships {school {school_id school_name status organization {organization_id}}}}`, index, index)
			}
			sb.WriteString("}")

			request := chlorine.NewRequest(sb.String(), chlorine.ReqToken(operator.Token))
			for index, id := range pageUserIDs {
				request.Var(fmt.Sprintf("user_id_%d", index), id)
			}

			data := map[string]*struct {
				SchoolMemberships []struct {
					School struct {
						SchoolID     string   `json:"school_id"`
						SchoolName   string   `json:"school_name"`
						Status       APStatus `json:"status"`
						Organization struct {
							OrganizationID string `json:"organization_id"`
						} `json:"organization"`
					} `json:"school"`
				} `json:"school_memberships"`
			}{}
			response := &chlorine.Response{
				Data: &data,
			}

			_, err := GetAmsClient().Run(ctx, request, response)
			if err != nil {
				log.Error(ctx, "get user schools filter by org and user ids failed",
					log.Err(err),
					log.Any("operator", operator),
					log.String("orgID", orgID),
					log.Strings("pageUserIDs", pageUserIDs))
				cerr <- err
				return
			}

			for index, userID := range pageUserIDs {
				user := data[fmt.Sprintf("q%d", index)]
				if user == nil {
					continue
				}
				mapLock.Lock()
				schools[userID] = make([]*School, 0)
				mapLock.Unlock()

				for _, membership := range user.SchoolMemberships {
					// filtering by operator's org id
					if membership.School.Organization.OrganizationID != orgID {
						continue
					}

					if condition.Status.Valid {
						if condition.Status.Status != membership.School.Status {
							continue
						}
					} else {
						// only status = "Active" data is returned by default
						if membership.School.Status != Active {
							continue
						}
					}
					mapLock.RLock()
					schools[userID] = append(schools[userID], &School{
						ID:     membership.School.SchoolID,
						Name:   membership.School.SchoolName,
						Status: membership.School.Status,
					})
					mapLock.RUnlock()
				}
			}
			cerr <- nil
		}(i)
	}

	for i := 0; i < pageCount; i++ {
		if err := <-cerr; err != nil {
			return nil, err
		}
	}

	log.Info(ctx, "get user school filter by org and user ids success",
		log.Any("operator", operator),
		log.String("orgID", orgID),
		log.Strings("userIDs", userIDs),
		log.Any("schools", schools))

	return schools, nil
}
func (s AmsSchoolService) Name() string {
	return "ams_school_service"
}
