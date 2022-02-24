package entity

type User struct {
	UserID   string `gorm:"column:user_id;PRIMARY_KEY"`
	UserName string `gorm:"column:user_name"`
	Email    string `gorm:"column:email"`
	Phone    string `gorm:"column:phone"`

	Secret string `gorm:"column:secret"`
	Salt   string `gorm:"salt"`

	Avatar   string `gorm:"avatar"`
	Gender   string `gorm:"gender"`
	Birthday int64  `gorm:"birthday"`

	CreateAt int64 `gorm:"column:create_at"`
	UpdateAt int64 `gorm:"column:update_at"`
	DeleteAt int64 `gorm:"column:delete_at"`

	CreateID int64  `gorm:"column:create_id"`
	UpdateID int64  `gorm:"column:update_id"`
	DeleteID int64  `gorm:"column:delete_id"`
	AmsID    string `gorm:"column:ams_id"`
}

func (User) TableName() string {
	return "users"
}

func (user User) GetID() interface{} {
	return user.UserID
}

type GetUserCountCondition struct {
	OrgID     NullString
	RoleID    NullString
	SchoolIDs NullStrings
	ClassIDs  NullStrings
	StudentID NullString
}
