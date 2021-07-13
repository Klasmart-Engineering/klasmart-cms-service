package entity

type Operator struct {
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
	Token  string `json:"-"` // hide token in log
}

type ExternalOptions struct {
	OrgIDs     []string
	UsrIDs     []string
	ProgIDs    []string
	SubjectIDs []string
	CatIDs     []string
	SubcatIDs  []string
	GradeIDs   []string
	AgeIDs     []string
}

type ExternalNameMap struct {
	OrgIDMap     map[string]string
	UsrIDMap     map[string]string
	ProgIDMap    map[string]string
	SubjectIDMap map[string]string
	CatIDMap     map[string]string
	SubcatIDMap  map[string]string
	GradeIDMap   map[string]string
	AgeIDMap     map[string]string
}
