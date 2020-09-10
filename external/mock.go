package external

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

type mockData struct {
	Programs           []*Program       `json:"program"`
	Subjects           []*Subject       `json:"subject"`
	Developmentals     []*Developmental `json:"developmental"`
	Skills             []*Skill         `json:"skills"`
	Ages               []*Age           `json:"age"`
	Grades             []*Grade         `json:"grade"`
	VisibilitySettings []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"visibility_settings"`
	Classes       []*Class        `json:"classes"`
	ClassTypes    []*ClassType    `json:"class_types"`
	Organizations []*Organization `json:"organizations"`
	Teachers      []*Teacher      `json:"teachers"`
	Students      []*Student      `json:"students"`
	Users         []*UserInfo     `json:"users"`
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
