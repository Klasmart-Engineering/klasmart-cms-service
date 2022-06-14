package model

import (
	"context"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

func TestShortcodeValidate(t *testing.T) {
	for i := 0; i < constant.ShortcodeSpace; i++ {
		value, err := utils.NumToBHex(context.TODO(), i, constant.ShortcodeBaseCustom, constant.ShortcodeShowLength)
		if err != nil {
			t.Fatal(err)
		}
		if !entity.Shortcode5Validate.MatchString(value) {
			t.Fatal(value)
		}
	}
	for i := 0; i < constant.ShortcodeBaseCustom*constant.ShortcodeBaseCustom*constant.ShortcodeBaseCustom; i++ {
		value, err := utils.NumToBHex(context.TODO(), i, constant.ShortcodeBaseCustom, 3)
		if err != nil {
			t.Fatal(err)
		}
		if !entity.Shortcode3Validate.MatchString(value) {
			t.Fatal(value)
		}
	}
	//value := "0A"
	//if !shortcode3Validate.MatchString(value) && !shortcode5Validate.MatchString(value) {
	//	t.Fatal("mismatch")
	//}
}
