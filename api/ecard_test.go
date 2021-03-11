package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreatECard(t *testing.T) {
	req := CreatECardRequest{
		ProvinceCode: "01",
		CityCode:     "ac",
		Duration:     365,
		CardType:     "D",
		Counts:       10,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/ecards", string(data))
	//request, err := http.NewRequest("POST", "https://hub-test.kidsloop.cn/v1/ecards", bytes.NewBuffer(data))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//res, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestLinkECard(t *testing.T) {
	req := LinkECardRequest{
		CardPassword: "4NUW0FQSD4NG37OJ",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	res := DoHttp(http.MethodPost, prefix+"/ecards/link", string(data))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestQueryECard(t *testing.T) {
	res := DoHttp(http.MethodGet, prefix+"/ecards", string(""))
	fmt.Println(res)
}

