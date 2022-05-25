package api

import (
	"context"
	"fmt"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/model"
	"log"
	"testing"
)

func TestGetTree(t *testing.T) {
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: "root:123456@tcp(127.0.0.1:3306)/cms2?parseTime=true&charset=utf8mb4",
		},
		AMS: config.AMSConfig{},
	}
	var err error
	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)

	ctx := context.Background()
	op := &entity.Operator{UserID: "afdfc0d9-ada9-4e66-b225-20f956d1a399", OrgID: "6300b3c5-8936-497e-ba1f-d67164b59c65"}
	var request entity.TreeRequest
	request.Key = "123"
	request.Role = "all"
	request.Type = "all"
	condition := entity.ContentConditionRequest{}
	if request.Role == constant.TreeQueryForMe.String() {
		condition.Author = constant.Self
	}
	if request.Type == constant.TreeQueryTypeAll.String() {
		condition.Name = request.Key
	}
	if request.Type == constant.TreeQueryTypeName.String() {
		condition.ContentName = request.Key
	}
	//if query is not self, filter conditions
	if request.Role != constant.TreeQueryForMe.String() {
		err := model.GetContentFilterModel().FilterPublishContent(ctx, &condition, op)
		//no available content visibility settings, return nil
		if err == model.ErrNoAvailableVisibilitySettings {
			//no available visibility settings
			return
		}
		if err != nil {
			return
		}
	}
	var result *entity.TreeResponse
	if request.Role == constant.TreeQueryForMe.String() {
		result, err = model.GetFolderModel().GetPrivateTree(ctx, &condition, op)
	} else {
		result, err = model.GetFolderModel().GetAllTree(ctx, &condition, op)
	}
	fmt.Println(result, err)
}
