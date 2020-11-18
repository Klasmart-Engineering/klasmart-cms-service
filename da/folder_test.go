package da

import (
	_"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestCreateTable(t *testing.T) {
	dsn := "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", dsn)
	if !assert.NoError(t, err) {
		return
	}
	db.LogMode(true)
	db.AutoMigrate(entity.FolderItem{})
}