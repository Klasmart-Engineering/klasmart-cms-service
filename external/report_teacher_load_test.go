package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

var utoken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjZjMGEwZDc5LTg2YmEtNDYxZS04MzhlLTkxNjEyOWRiNjE2OSIsImVtYWlsIjoidGVhY2hlcjAyXzAwMUB5b3BtYWlsLmNvbSIsImV4cCI6MTYzMjQ2NTg3MCwiaXNzIjoia2lkc2xvb3AifQ.FpIGosh89pjQeGahko64l0_tr7C6ZztOLJOTm9gUV5xf8e271AUCDfQuKWn4bX5_MjUNGEiTRRII6uF6-KbQ5i4okOaRbeF5GWHwC2NT_jDZK4SqH0HQCTI0-tyHt6uyptL3XRQWL2mvKJBGMQ7w8mxbPgtjRgOOFaiTWiEG7Twks2cpqOkng6LPBDNJNDaPSMBvihICAyg3kALESIosPPqnqbZTpjfTNmsfChAo9uHbkeab087QawcRI0LpDGguIe70JpUzsIjhgVFRuC8mDDv2qvMQuUpaAnA4t23l-_rUO_xltj5w9naFrq7o2jrU4LhDIiLjizKDtISfJj-B-Q"
var op = entity.Operator{
	UserID: "6c0a0d79-86ba-461e-838e-916129db6169",
	OrgID:  "f27efd10-000e-4542-bef2-0ccda39b93d3",
	Token:  utoken,
}

func TestAmsTeacherLoadService_BatchGetClassWithStudent(t *testing.T) {
	ctx := context.Background()
	result, err := GetTeacherLoadServiceProvider().BatchGetClassWithStudent(ctx, &op, []string{"6c0a0d79-86ba-461e-838e-916129db6169"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", result)
}
