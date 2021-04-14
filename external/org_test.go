package external

import (
	"context"
	"testing"
)

func TestOrganizationService_BatchGet(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().BatchGet(context.TODO(),
		testOperator,
		[]string{"3f135b91-a616-4c80-914a-e4463104dbac", "3f135b91-a616-4c80-914a-e4463104dbad"})
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().BatchGet() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestAmsOrganizationService_GetNameByOrganizationOrSchool(t *testing.T) {
	ids := []string{"9e285fc9-50fd-4cf2-ba5b-3f191c3338b4", "ac341b5f-a5f8-44a4-8b43-c0ff21d337b2"}
	names, err := GetOrganizationServiceProvider().GetNameByOrganizationOrSchool(context.TODO(), testOperator, ids)
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() error = %v", err)
		return
	}

	if len(names) != len(ids) {
		t.Error("GetOrganizationServiceProvider().GetNameByOrganizationOrSchool() get empty slice")
		return
	}
}

func TestAmsOrganizationService_GetByPermission(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().GetByPermission(context.TODO(),
		testOperator,
		CreateContentPage201,
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetByPermission() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().GetByPermission() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().GetByPermission() get null")
			return
		}
	}
}

func TestAmsOrganizationService_GetByUserID(t *testing.T) {
	orgs, err := GetOrganizationServiceProvider().GetByUserID(context.TODO(),
		testOperator,
		"335e0577-99cb-5d88-b5e1-dfdb14d5d4c2",
		WithStatus(Active))
	if err != nil {
		t.Errorf("GetOrganizationServiceProvider().GetByUserID() error = %v", err)
		return
	}

	if len(orgs) == 0 {
		t.Error("GetOrganizationServiceProvider().GetByUserID() get empty slice")
		return
	}

	for _, org := range orgs {
		if org == nil {
			t.Error("GetOrganizationServiceProvider().GetByUserID() get null")
			return
		}
	}
}

func TestAmsOrganizationService_OrgAthPrgSjtCtgSubCtgGrdAge(t *testing.T) {
	op := initOperator("8a31ebab-b879-4790-af99-ee4941a778b3", "", "")
	orgIDs := []string{"8a31ebab-b879-4790-af99-ee4941a778b3"}
	athIDs := []string{"2013e53e-52dd-5e1c-af0b-b503e31c8a59"}
	prgIDs := []string{"7565ae11-8230-4b7d-ac24-1d9dd6f792f2", "75004121-0c0d-486c-ba65-4c57deacb44b"}
	sbjIDs := []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "7cf8d3a3-5493-46c9-93eb-12f220d101d0"}
	catIDs := []string{"bcfd9d76-cf05-4ccd-9a41-6b886da661be", "6933de3e-a568-4d56-8c01-e110bda22926"}
	sbcIDs := []string{"49e73e4f-8ffc-47e3-9b87-0f9686d361d7", "b2cc7a69-4e64-4e97-9587-0078dccd845a"}
	grdIDs := []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a", "0ecb8fa9-d77e-4dd3-b220-7e79704f1b03"}
	ageIDs := []string{"fe0b81a4-5b02-4548-8fb0-d49cd4a4604a", "21f1da64-b6c8-4e74-9fef-09d08cfd8e6c"}
	organizations, authors, programs, subjects, categories, subCategories, grades, ages, err := GetOrganizationServiceProvider().
		OrgAthPrgSjtCtgSubCtgGrdAge(context.TODO(), op, orgIDs, athIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", organizations)
	fmt.Printf("%v\n", authors)
	fmt.Printf("%v\n", programs)
	fmt.Printf("%v\n", subjects)
	fmt.Printf("%v\n", categories)
	fmt.Printf("%v\n", subCategories)
	fmt.Printf("%v\n", grades)
	fmt.Printf("%v\n", ages)
}
