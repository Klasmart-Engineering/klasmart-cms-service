package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
)

func TestOrganizationService_BatchGet(t *testing.T) {
	config.LoadEnvConfig()
	orgs, err := GetOrganizationServiceProvider().BatchGet(context.Background(), []string{
		"3f135b91-a616-4c80-914a-e4463104dbac",
		"3f135b91-a616-4c80-914a-e4463104dbad",
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := range orgs {
		if orgs[i] != nil {
			fmt.Println(*(orgs[i]))
		} else {
			fmt.Println(i)
		}
	}
}

func TestClassService_BatchGet(t *testing.T) {
	classes, err := GetClassServiceProvider().BatchGet(context.Background(), []string{
		"f3d3cdf5-9ca8-44cf-a604-482e5d183049",
		"3b8074d5-893c-41c7-942f-d2115cc8bc32",
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := range classes {
		if classes[i] != nil {
			fmt.Println(*(classes[i]))
		} else {
			fmt.Println(i)
		}
	}
}
func TestAmsOrganizationService_GetOrganizationOrSchoolName(t *testing.T) {
	config.LoadEnvConfig()
	ids := []string{
		"f3d3cdf5-9ca8-44cf-a604-482e5d183049",
		"3b8074d5-893c-41c7-942f-d2115cc8bc32"}
	names, err := GetOrganizationServiceProvider().GetOrganizationOrSchoolName(context.Background(), ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		fmt.Println(n)
	}
}

func TestAmsOrganizationService_GetOrganizationsAssociatedWithUserID(t *testing.T) {
	config.LoadEnvConfig()
	id := "a161e13f-f620-5284-8ab5-93445d8064bf"
	orgs, err := GetOrganizationServiceProvider().GetOrganizationsAssociatedWithUserID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range orgs {
		fmt.Println(n)
	}
}

func TestAmsOrganizationService_GetSchoolsAssociatedWithUserID(t *testing.T) {
	config.LoadEnvConfig()
	id := "a161e13f-f620-5284-8ab5-93445d8064bf"
	orgs, err := GetSchoolServiceProvider().GetSchoolsAssociatedWithUserID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range orgs {
		fmt.Println(n)
	}
}

func TestAmsPermissionService_HasPermissions(t *testing.T) {
	config.LoadEnvConfig()
	op := entity.Operator{
		UserID: "690776ed-502e-5b68-9a03-bd7400a37762",
		OrgID: "49dd6bce-18a9-47cb-8da7-5c425c66a0ff",
	}
	permissions := []PermissionName{"archived_content_page_205", "create_lesson_material_220"}
	perms, err := GetPermissionServiceProvider().HasPermissions(context.Background(), &op, permissions)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(perms)
}