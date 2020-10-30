package external

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"io/ioutil"
	"net/http"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

type mockData struct {
	Options            []*mockOption `json:"options"`
	VisibilitySettings []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"visibility_settings"`
	Classes       []*Class             `json:"classes"`
	LessonTypes   []*entity.LessonType `json:"lesson_types"`
	ClassTypes    []*entity.ClassType  `json:"class_types"`
	Organizations []*Organization      `json:"organizations"`
	Teachers      []*Teacher           `json:"teachers"`
	Students      []*Student           `json:"students"`
	Users         []*UserInfo          `json:"users"`
}

type mockOption struct {
	Program       *entity.Program         `json:"program"`
	Subject       []*entity.Subject       `json:"subject"`
	Developmental []*entity.Developmental `json:"developmental"`
	Age           []*entity.Age           `json:"age"`
	Grade         []*entity.Grade         `json:"grade"`
}

var (
	_mockData *mockData
	_mockOnce sync.Once
)

func GetMockData() *mockData {
	_mockOnce.Do(func() {
		_mockData = &mockData{}
		response, err := http.DefaultClient.Get("https://launch.kidsloop.cn/static/mock-korea-data/select-options.json")
		if err != nil {
			log.Error(context.Background(), "read mock json failed", log.Err(err))
			return
		}

		defer response.Body.Close()

		buffer, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Error(context.Background(), "read response failed", log.Err(err))
			return
		}

		err = json.Unmarshal(buffer, &_mockData)
		if err != nil {
			log.Error(context.Background(), "unmarshal response failed", log.Err(err), log.ByteString("json", buffer))
			return
		}
	})

	return _mockData
}
