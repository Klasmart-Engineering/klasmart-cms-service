package external

import (
	"context"
	"fmt"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

//var utoken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjZjMGEwZDc5LTg2YmEtNDYxZS04MzhlLTkxNjEyOWRiNjE2OSIsImVtYWlsIjoidGVhY2hlcjAyXzAwMUB5b3BtYWlsLmNvbSIsImV4cCI6MTYzMjQ2NTg3MCwiaXNzIjoia2lkc2xvb3AifQ.FpIGosh89pjQeGahko64l0_tr7C6ZztOLJOTm9gUV5xf8e271AUCDfQuKWn4bX5_MjUNGEiTRRII6uF6-KbQ5i4okOaRbeF5GWHwC2NT_jDZK4SqH0HQCTI0-tyHt6uyptL3XRQWL2mvKJBGMQ7w8mxbPgtjRgOOFaiTWiEG7Twks2cpqOkng6LPBDNJNDaPSMBvihICAyg3kALESIosPPqnqbZTpjfTNmsfChAo9uHbkeab087QawcRI0LpDGguIe70JpUzsIjhgVFRuC8mDDv2qvMQuUpaAnA4t23l-_rUO_xltj5w9naFrq7o2jrU4LhDIiLjizKDtISfJj-B-Q"
//var utoken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTYzNzIxNjExOCwiaXNzIjoia2lkc2xvb3AifQ.hoUu0dtOrHhcQ6j7eiNAOK5dJI-GhTF7AMgKHx6oJSqV8_b51PCT28JEojsn11G88Hg5ZXt0mC5YVxk6qF8b6RVgFaOoIYQXH_vd2DzSpYuGQ4AFTP8QFEwv7oq12PwzJXQ0LU7Ot0QNUZt7AVF-pBRQ4hX91v-_o4zB7Pwb8AfUyECvmB-Ln0CM-j4p-n5k4mijXA_m67H8W7UCTW7nfWwO540F85NciI64KzGHHQ4o5T-YxYhcbNgsSjpJ06SzenyDdNqOhxITu830q0eDHWi4WSokFH2rpx9o8o33fW-Jvi3nCTtlz6oN_OZ1pe9CeZ-hksqphDlBQJEedktSSw"
//var utoken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImM1N2VmNjhkLWE2MzUtNDUxZC1iOTk3LWFlYmMzYzI5Yjk5YSIsImVtYWlsIjoib3JnYmFkYUB5b3BtYWlsLmNvbSIsImV4cCI6MTYzNzIxODgyNCwiaXNzIjoia2lkc2xvb3AifQ.gW3XdI5FvZ3Y8DPalwWjMqoOlEe1apYjGgf2rABQ07R3htUuAD0x4qYAl8eXpmywqQacGwIThEU7mGKmZHiGpZWHuISqs_L6FRQLRvtOPoBLFofKojmF-xmdAq31yOSJfTdRKqTx_5_-7i3CkqBu9tdSEoi-vUAVq5KWd4f9jlE3V1Xrs777AF5ya3VdeDSoMEc4L8jYAl8mC0ZLL8cr22WOzPcvhm9OmT3qee7jJgSSx2ZVT6QGGIIAEPl4KZySSZTevgMCrEfH6wo6F_wHPI4kjTR8MZjF_74NSyccMpXQvxud9SHi6CCPzmxlzMHcA3DXUF_toRJ9xwlWNIOevw"
var utoken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTYzNzU3MTgwNSwiaXNzIjoia2lkc2xvb3AifQ.kmDfZVWiV-oXg_V9wHXGvXu-IvlVLnKA4lyo6dDXDRC7X5MDKKs4zlM54pX0TbleT2nqxgcJTTV-La3GeyspI3VAXKL7K7iji5z5ok0RAP6XK7g8AWjUIPgwGLS2MB-AAOw-lQImpYz6eO93WRTVnouPHmk4B9m5pyQi9xOSWeLg6XPWZufUUDQyZysvPdWXOV7icHx-0VE1VR_1_n1KHNq362-k231NgeSd7lheRgd-j-88r4-WTIbLETaUYYjhiemg9jbKp4pEPX3SDCWkAqzJhrT3RWEamvtVLX3IA7kPE-TUHg_PS91Mp2xvl5E-GybBfEBWv_CarmpNBzeRvQ"
var op = entity.Operator{
	UserID: "8cd9f417-0812-44fb-9c50-1b78217ee76f",
	OrgID:  "60c064cc-bbd8-4724-b3f6-b886dce4774f",
	Token:  utoken,
}

func TestAmsTeacherLoadService_BatchGetClassWithStudent(t *testing.T) {
	ctx := context.Background()
	result, err := GetTeacherLoadServiceProvider().BatchGetActiveClassWithStudent(ctx, &op, []string{"8cd9f417-0812-44fb-9c50-1b78217ee76f"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", result)
}
