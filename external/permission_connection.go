package external

type PermissionFilter struct {
	RoleID *UUIDFilter        `json:"roleId,omitempty" gqls:"roleId,omitempty"`
	Name   *StringFilter      `json:"name,omitempty" gqls:"name,omitempty"`
	Allow  *BooleanFilter     `json:"allow,omitempty" gqls:"allow,omitempty"`
	AND    []PermissionFilter `json:"AND,omitempty" gqls:"AND,omitempty"`
	OR     []PermissionFilter `json:"OR,omitempty" gqls:"OR,omitempty"`
}

func (PermissionFilter) FilterName() FilterType {
	return PermissionFilterType
}

func (PermissionFilter) ConnectionName() ConnectionType {
	return PermissionsConnectionType
}

type AmsPermissionConnectionService struct {
	AmsPermissionService
}

type PermissionConnectionNode struct {
	ID          string `json:"id" gqls:"id"`
	Name        string `json:"name" gqls:"name"`
	Category    string `json:"category" gqls:"category"`
	Group       string `json:"group" gqls:"group"`
	Level       string `json:"level" gqls:"level"`
	Description string `json:"description" gqls:"description"`
	Allow       bool   `json:"allow" gqls:"allow"`
}

type PermissionsConnectionEdge struct {
	Cursor string                   `json:"cursor" gqls:"cursor"`
	Node   PermissionConnectionNode `json:"node" gqls:"node"`
}

type PermissionsConnectionResponse struct {
	TotalCount int                         `json:"totalCount" gqls:"totalCount"`
	PageInfo   ConnectionPageInfo          `json:"pageInfo" gqls:"pageInfo"`
	Edges      []PermissionsConnectionEdge `json:"edges" gqls:"edges"`
}

func (pcs PermissionsConnectionResponse) GetPageInfo() *ConnectionPageInfo {
	return &pcs.PageInfo
}
