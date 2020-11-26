package model

import (
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/contentdata"
	"testing"
)

func TestUnmarshalContentData(t *testing.T) {
	data := "{\"segmentId\":\"1\",\"condition\":\"\",\"materialId\":\"5fbdb67de0ad99a97be8681c\",\"material\":null,\"next\":[{\"segmentId\":\"11\",\"condition\":\"\",\"materialId\":\"5fbcd442869fc4f88ee362cd\",\"material\":null,\"next\":[{\"segmentId\":\"111\",\"condition\":\"\",\"materialId\":\"5fbcbc5f4a60a5856c1abbea\",\"material\":null,\"next\":[],\"teacher_manual\":\"\"}],\"teacher_manual\":\"\"}],\"teacher_manual\":\"teacher_manual-5fbf0d0c25b6981b872b7988.pdf\"}"
	ins := contentdata.LessonData{}
	err := json.Unmarshal([]byte(data), &ins)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Done")
}
