package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"testing"
)

func TestShortcodeValidate(t *testing.T) {
	for i := 0; i < constant.ShortcodeSpace; i++ {
		value, err := utils.NumToBHex(context.TODO(), i, constant.ShortcodeBaseCustom, constant.ShortcodeShowLength)
		if err != nil {
			t.Fatal(err)
		}
		if !shortcode5Validate.MatchString(value) {
			t.Fatal(value)
		}
	}
	for i := 0; i < constant.ShortcodeBaseCustom*constant.ShortcodeBaseCustom*constant.ShortcodeBaseCustom; i++ {
		value, err := utils.NumToBHex(context.TODO(), i, constant.ShortcodeBaseCustom, 3)
		if err != nil {
			t.Fatal(err)
		}
		if !shortcode3Validate.MatchString(value) {
			t.Fatal(value)
		}
	}
	//value := "0A"
	//if !shortcode3Validate.MatchString(value) && !shortcode5Validate.MatchString(value) {
	//	t.Fatal("mismatch")
	//}
}
