package model

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

const operatorOrgID = "92db7ddd-1f23-4f64-bd47-94f6d34a50c0"

func TestContentModel_CreateContent(t *testing.T) {
	ctx := context.Background()
	req := entity.CreateContentRequest{
		ContentType: entity.ContentTypeAssets, // ContentType
		//SourceType:  //string      `json:"source_type"`
		Name:        "test",                                           //string      `json:"name"`
		Program:     "7565ae11-8130-4b7d-ac24-1d9dd6f792f2",           //string      `json:"program"`
		Subject:     []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71"}, //[]string    `json:"subject"`
		Category:    []string{"2d5ea951-836c-471e-996e-76823a992689"}, //[]string    `json:"developmental"`
		SubCategory: []string{},                                       //[]string    `json:"skills"`
		Age:         []string{},                                       //[]string    `json:"age"`
		Grade:       []string{},                                       //[]string    `json:"grade"`
		Keywords:    []string{},                                       //[]string    `json:"keywords"`
		Description: "test description",                               //string      `json:"description"`
		Thumbnail:   "",                                               //string      `json:"thumbnail"`
		SuggestTime: 100,                                              //int         `json:"suggest_time"`

		SelfStudy:    false, //TinyIntBool `json:"self_study"`
		DrawActivity: false, //TinyIntBool `json:"draw_activity"`
		LessonType:   "1",   //string      `json:"lesson_type"`

		Outcomes: []string{}, //[]string `json:"outcomes"`

		PublishScope: []string{operatorOrgID}, //[]string `json:"publish_scope"`

		Data:  `{"source": "assets-FOUWJDNDNEYRGAUT4QSW2TCBMHBOA4TRSOG2JJDCBCUIB3K5DRKA====.mp4", "file_type": 2, "input_source": 2}`, //string `json:"data"`
		Extra: "",                                                                                                                     //string `json:"extra"`

		ParentFolder: "5fc8466ea1207f8c137118c9", //string `json:"parent_folder"`
	}
	operator := init2Operator(operatorOrgID, "brilliant.yang@badanamu.com.cn", "Try1try123")
	content, err := GetContentModel().CreateContent(ctx, req, operator)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(content)
}

func init2Operator(orgID string, authTo string, authCode string) *entity.Operator {
	if authTo == "" {
		authTo = os.Getenv("auth_to")
	}
	if authCode == "" {
		authCode = os.Getenv("auth_code")
	}
	loginUrl := os.Getenv("ams_auth_endpoint")
	transferUrl := os.Getenv("auth_endpoint")
	data := struct {
		DeviceID   string `json:"deviceId"`
		DeviceName string `json:"deviceName"`
		Email      string `json:"email"`
		Password   string `json:"pw"`
	}{
		DeviceID:   "webpage",
		DeviceName: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36",
		Email:      authTo,
		Password:   authCode,
	}
	body, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", loginUrl+"/v1/login", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	loginResponse := struct {
		AccessToken string `json:"accessToken"`
	}{}
	err = json.Unmarshal(response, &loginResponse)
	if err != nil {
		panic(err)
	}

	data2 := struct {
		Token string `json:"token"`
	}{loginResponse.AccessToken}
	body, err = json.Marshal(data2)
	if err != nil {
		panic(err)
	}
	req, err = http.NewRequest("POST", transferUrl+"/transfer", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	//req.Header.Set("origin", "https://auth.kidsloop.net")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authority", "auth.kidsloop.net")
	req.Header.Set("accept", "application/json")
	//req.Header.Set("path", "/transfer")
	//req.Header.Set("scheme", "https")
	req.Header.Set("method", "POST")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36")
	req.Header.Set("referer", "https://auth.kidsloop.net/?continue=https%3A%2F%2Fbeta-hub.kidsloop.net%2F%23%2Fadmin%2Forganizations")
	req.Header.Set("cookie", "_ga=GA1.2.489381037.1617355818; locale=en; privacy=true")
	resp, err = (&http.Client{}).Do(req)
	op := &entity.Operator{
		OrgID: orgID,
	}
	if err != nil {
		op.Token = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6Ijc0MmIxNjI1LWVmYjctNWM1YS1iZDI3LWRjYzIwMmQ2YTEzNiIsImVtYWlsIjoib3JnMTIyMTAyQHlvcG1haWwuY29tIiwiZXhwIjoxNjE4OTA5OTE1LCJpc3MiOiJraWRzbG9vcCJ9.IRcPjpqH3AQfL_0i3rEPKXXLHbcjGvegv4iVseqSENzhr7X_iCckP2gLth4plN_mX-dNphQqJvV0-L5enTn1u8g3jbmXpR5VALV5Bf_5G-A6xZWUAwtxCyxVKlTqtOM5Pi-WEg8gPHgS9sGL2vT7eviOlcG3S3W0LV5QzYBBC55okNtHZLwS0N-eXzVT8oKOwyMTU8ftqTQ5f9slCUV7ennZrJ6FJX8oozlHixIg4NcTMpo_S0al4GTw2--BJU_DrEQZ80dgtBse1TE8QxY0_R8tbW6SNUJkSKkOZqVCUAcmUG_sY5rN5HGFzeuniNJpe179xdF8OUXCiH-9YmIBaw"
		infos := strings.Split(op.Token, ".")
		info, err := base64.RawStdEncoding.DecodeString(infos[1])
		if err != nil {
			panic(err)
		}
		var user struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(info, &user)
		op.UserID = user.ID
		return op
	}
	for i := range resp.Cookies() {
		cookie := resp.Cookies()[i]
		if cookie.Name == "access" {
			op.Token = cookie.Value
			infos := strings.Split(op.Token, ".")
			info, err := base64.RawStdEncoding.DecodeString(infos[1])
			if err != nil {
				panic(err)
			}
			var user struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(info, &user)
			op.UserID = user.ID
			return op
		}
	}
	return nil
}
