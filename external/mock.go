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
	Program            []*Program       `json:"program"`
	Subject            []*Subject       `json:"subject"`
	Developmental      []*Developmental `json:"developmental"`
	Skills             []*Skill         `json:"skills"`
	Age                []*Age           `json:"age"`
	Grade              []*Grade         `json:"grade"`
	VisibilitySettings []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"visibility_settings"`
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
