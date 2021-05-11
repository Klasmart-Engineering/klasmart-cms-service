package utils

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"testing"
)

func TestURLSig(t *testing.T) {
	token, err := GenerateH5pJWT(context.Background(), "edit", "5fa3f41cb65ebc00122a74a6")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(token)
}

func TestNumToBHex(t *testing.T) {
	for i := 0; i < constant.ShortcodeSpace; i++ {
		res, err := NumToBHex(context.TODO(), i, constant.ShortcodeBaseCustom, constant.ShortcodeShowLength)
		if err != nil {
			t.Fatal(err)
		}
		ind, err := BHexToNum(context.TODO(), res)
		if err != nil {
			t.Fatal(err)
		}
		if ind != i {
			t.Fatalf("i:%d, ind:%d\n", i, ind)
		}
	}
}

func TestBHexToNum(t *testing.T) {
	result, err := NumToBHex(context.TODO(), 60466175, constant.ShortcodeBaseCustom, constant.ShortcodeShowLength)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(result)
	ind, err := BHexToNum(context.TODO(), result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ind)
}
