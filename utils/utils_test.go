package utils

import "testing"

func TestGetUrlParamStr(t *testing.T) {
	str := "https://res.alpha.kidsloop.net/drawing_feedback/62501ffdcdb897fce54da0b1.png?aaa=bbb"
	t.Log(GetUrlParamStr(str))
}
