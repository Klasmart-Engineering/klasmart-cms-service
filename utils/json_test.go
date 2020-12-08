package utils

import (
	"testing"
)

type JsonContent struct {
	CMSPageSize int    `json:"cms_page_size"`
	Str         string `json:"str"`
}

func TestJsonMerge(t *testing.T) {
	defaultSetting := &JsonContent{
		CMSPageSize: 10,
		Str:         "2",
	}
	var userSetting *JsonContent = new(JsonContent)
	userSetting.CMSPageSize = 2

	err := JsonMerge(userSetting, defaultSetting)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(userSetting.CMSPageSize, userSetting.Str)
}
