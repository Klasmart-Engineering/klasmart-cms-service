package external

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

var clsToken = "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImFmZGZjMGQ5LWFkYTktNGU2Ni1iMjI1LTIwZjk1NmQxYTM5OSIsImVtYWlsIjoib3JnMTExOUB5b3BtYWlsLmNvbSIsImV4cCI6MTY1MjA2ODgzOCwiaXNzIjoia2lkc2xvb3AifQ.uZaNkzIbtPBvg0QgYyQZgfA1rTkWFZ0RmgAPNgofKCaGYxJH1EifzMDdTIa7wIArzXfM_InOOwkYmuwqdvfO43AOlnl5YjZwCxSO0TEtu2pZuRooSF_Wgx95fx-GA8ygTBdNOuMs6kRzN2GZhxOS7w1U1-FR_fHo_NOxewgI4TiVlpLyM_DSGUa7ytdaGYUD3xJrtv4-jUlJ5yNXp6DYVM9GiCE8oBMA0oOGqFNYgkJdGR3BF6iyD6nuS-k8S9Kqdzxf8_TcBHEND7Ax1lP-Uisar-B7EmL7SZX126I-KdxLC-3pqy7eob4ruSkfalWsWAe3k47ntE1dtC1hE0_8bA"
var usrIDs = []string{
	"000d653d-7961-447c-8d66-ad8c4a40eae6",
}

func TestAmsClassService_GetByUserIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	provider := AmsClassConnectionService{}
	classes1, err := provider.AmsClassService.GetByUserIDs(ctx, testOperator, usrIDs, WithStatus(Active))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByUserIDs() error = %v", err)
		return
	}
	classes2, err := provider.GetByUserIDs(context.TODO(), testOperator, usrIDs, WithStatus(Active))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByUserIDs() error = %v", err)
		return
	}

	fmt.Println("len:", len(classes1) == len(classes2))
}

func TestAmsClassService_GetByOrganizationIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	provider := AmsClassConnectionService{}
	classes1, err := provider.AmsClassService.GetByOrganizationIDs(ctx, testOperator, orgIDs, WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}

	classes2, err := provider.GetByOrganizationIDs(ctx, testOperator, orgIDs, WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}
	fmt.Println("len:", len(classes1) == len(classes2))
}

func TestAmsClassService_GetBySchoolIDs(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	provider := AmsClassConnectionService{}
	classes1, err := provider.AmsClassService.GetBySchoolIDs(ctx, testOperator, schIDs, WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}

	classes2, err := provider.GetBySchoolIDs(ctx, testOperator, orgIDs, WithStatus(Inactive))
	if err != nil {
		t.Errorf("GetClassServiceProvider().GetByOrganizationIDs() error = %v", err)
		return
	}
	fmt.Println("len:", len(classes1) == len(classes2))
}

func TestAmsClassService_GetOnlyUnderOrgClasses(t *testing.T) {
	ctx := context.Background()
	testOperator.Token = clsToken
	result, err := GetClassServiceProvider().GetOnlyUnderOrgClasses(ctx, testOperator, orgID)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(bs)
}
