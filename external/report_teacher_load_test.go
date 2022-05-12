package external

import (
	"context"
	"fmt"
	"testing"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

var utoken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjI0OTc5OCwiaXNzIjoia2lkc2xvb3AifQ.D4kvqsPQe3ZaUr073_V76u4n5qtBTKECD3NYxh0Dxgu8PjQcbXlgDl-9vTkj92aUrovO8riFKoODoqFBvm9PzHeGl0yRHgNjJV0TPsCwLWqNqE4YOK9ecprEKDGjt_jfMNFju4Fmk5dFsXrqfxO8BrSqGymFw6k9gDqvHQePHaIpFxsWqk4z3Ut41wOaDAS3lRDGmYD3iftEQo7JSdQEu1ytQXiY1hdU0XgUmv1Knh3cPOb0t1KO-cqSXU-w0McFmR55WefmbFWdKDy5KA18NF6MKqpw67Rbxx9uZ_9qe7iCETPZl0eusYvynC5-JmDJbZ0LSG6ffVAKASh8DlbZ-A"
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
