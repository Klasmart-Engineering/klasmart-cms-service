package entity_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func TestStudentUsageRecordInJwt(t *testing.T) {
	bf, _ := json.Marshal(entity.StudentUsageRecordInJwt{
		StudentUsageRecord: entity.StudentUsageRecord{
			ClassType:         "live",
			RoomID:            "room_id",
			LessonMaterialUrl: "",
			ContentType:       "",
			ActionType:        "",
			Timestamp:         0,
			Students: []*entity.Student{
				{
					UserID: "b1af180f-b2ca-48a5-93b7-10641aebb2ed",
					Email:  "qa+stress_t1@calmid.com",
					Name:   "xxxxxxxxxx",
				}, {
					UserID: "b1af180f-b2ca-48a5-93b7-10641aebb2ed",
					Email:  "qa+stress_t1@calmid.com",
					Name:   "xxxxxxxxxx",
				},
			},
		},
		StandardClaims: &jwt.StandardClaims{
			Audience:  "xxxxxxxxxx",
			ExpiresAt: time.Now().Unix(),
			Id:        "xxxxxxxxxx",
			IssuedAt:  time.Now().Unix(),
			Issuer:    "xxxxxxxxxx",
			NotBefore: time.Now().Unix(),
			Subject:   "xxxxxxxxxx",
		},
	})
	fmt.Println(string(bf))
}

type User struct {
	Name string `json:"name"`
}
