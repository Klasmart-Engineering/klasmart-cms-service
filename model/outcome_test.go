package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

func initOperator() *entity.Operator {
	return &entity.Operator{
		OrgID:  "8a31ebab-b879-4790-af99-ee4941a778b3",
		UserID: "2013e53e-52dd-5e1c-af0b-b503e31c8a59",
	}
}

func TestOutcomeModel_GetLearningOutcomeByID(t *testing.T) {
	ctx := context.TODO()
	outcome, err := da.GetOutcomeDA().GetOutcomeByID(ctx, dbo.MustGetDB(ctx), "60616800e7c9026bb00d1d6c")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", outcome)
}

func TestGenerateShortcode(t *testing.T) {
	op := initOperator()
	ctx := context.TODO()
	shortcode, err := GetOutcomeModel().GenerateShortcode(ctx, op)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(shortcode)
}

func TestSearchSetsByOutcome(t *testing.T) {
	ctx := context.TODO()
	outcomesSets, err := da.GetOutcomeSetDA().SearchSetsByOutcome(ctx, dbo.MustGetDB(ctx), []string{"60616800e7c9026bb00d1d6c"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", outcomesSets)
}

func TestOutcomeSetModel_CreateOutcomeSet(t *testing.T) {
	ctx := context.TODO()
	set, err := GetOutcomeSetModel().CreateOutcomeSet(ctx, &entity.Operator{OrgID: "org-1"}, "math2")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", set)
}

func TestSearch(t *testing.T) {
	ctx := context.TODO()
	response, err := GetOutcomeModel().Search(ctx, &entity.Operator{OrgID: "92db7ddd-1f23-4f64-bd47-94f6d34a50c0"}, &entity.OutcomeCondition{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response.Total)
	for _, v := range response.List {
		t.Log(v)
	}
}

func TestSearchWithoutRelation(t *testing.T) {
	ctx := context.TODO()
	count, outcomes, err := GetOutcomeModel().SearchWithoutRelation(ctx, &entity.Operator{}, &entity.OutcomeCondition{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(count)
	for _, v := range outcomes {
		t.Log(v)
	}
}

func TestSearchPublished(t *testing.T) {
	ctx := context.TODO()
	result, err := GetOutcomeModel().SearchPublished(ctx, &entity.Operator{}, &entity.OutcomeCondition{
		ProgramIDs:     []string{"program_test_1"},
		SubjectIDs:     []string{"subject_test_1"},
		CategoryIDs:    []string{"category_test_1"},
		SubCategoryIDs: []string{"subcategory_test_1"},
		AgeIDs:         []string{"age_test_1"},
		GradeIDs:       []string{"grade_test_1"},
		Page:           1,
		PageSize:       5,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result.Total)
	for _, v := range result.List {
		t.Log(v)
	}
}

func TestSearchPublishedOutcome(t *testing.T) {
	ctx := context.TODO()
	result, err := GetOutcomeModel().SearchPublished(ctx, &entity.Operator{}, &entity.OutcomeCondition{
		ProgramIDs:     []string{"program_test_1"},
		SubjectIDs:     []string{"subject_test_1"},
		CategoryIDs:    []string{"category_test_1"},
		SubCategoryIDs: []string{"subcategory_test_1"},
		AgeIDs:         []string{"age_test_1"},
		GradeIDs:       []string{"grade_test_1"},
		Page:           1,
		PageSize:       5,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result.Total)
	for _, v := range result.List {
		t.Log(v)
	}
}

func TestExportOutcomes(t *testing.T) {
	ctx := context.TODO()
	isLocked := false
	cond := &entity.OutcomeCondition{
		OrganizationID: "24aee19b-ad5b-41bb-adf9-28dad0d3264a",
		Assumed:        -1,
		IsLocked:       &isLocked,
		Page:           1,
		PageSize:       50,
	}
	result, err := GetOutcomeModel().Export(ctx, &entity.Operator{OrgID: "6300b3c5-8936-497e-ba1f-d67164b59c65"}, cond)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result.TotalCount)
}

func TestVerifyImportData(t *testing.T) {
	ctx := context.TODO()
	importData := &entity.VerifyImportOutcomeRequest{
		Data: []*entity.ImportOutcomeView{
			{
				RowNumber:      12,
				OutcomeName:    "test",
				Shortcode:      "test",
				Assumed:        true,
				Description:    "test",
				Keywords:       []string{"test"},
				Program:        []string{"test"},
				Subject:        []string{"test"},
				Category:       []string{"test"},
				Subcategory:    []string{"test"},
				Age:            []string{"test"},
				Grade:          []string{"test"},
				Sets:           []string{"test"},
				ScoreThreshold: 0,
			},
			{
				RowNumber:      3,
				OutcomeName:    "test",
				Shortcode:      "0001F",
				Assumed:        true,
				Description:    "test",
				Keywords:       []string{"test"},
				Program:        []string{"test"},
				Subject:        []string{"test"},
				Category:       []string{"test"},
				Subcategory:    []string{"test"},
				Age:            []string{"test"},
				Grade:          []string{"test"},
				Sets:           []string{"test"},
				ScoreThreshold: 0,
			},
		},
	}
	result, err := GetOutcomeModel().VerifyImportData(ctx, &entity.Operator{OrgID: "6300b3c5-8936-497e-ba1f-d67164b59c65"}, importData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result.ExistError)
	for _, v := range result.CreateData {
		t.Log(v)
	}
	t.Log("update data")
	for _, v := range result.UpdateData {
		t.Log(v)
	}
}

func TestImportData(t *testing.T) {
	ctx := context.TODO()
	importData := &entity.ImportOutcomeRequest{
		CreateData: []*entity.ImportOutcomeView{
			{
				RowNumber:      3,
				OutcomeName:    "test",
				Shortcode:      "ZZ123",
				Assumed:        true,
				Description:    "test",
				Keywords:       []string{"test"},
				Program:        []string{"test"},
				Subject:        []string{"test"},
				Category:       []string{"test"},
				Subcategory:    []string{"test"},
				Age:            []string{"test"},
				Grade:          []string{"test"},
				Sets:           []string{"test"},
				ScoreThreshold: 0,
			},
		},
		UpdateData: []*entity.ImportOutcomeView{
			{
				OutcomeName:    "test",
				Shortcode:      "00002",
				Assumed:        true,
				Description:    "test",
				Keywords:       []string{"test"},
				Program:        []string{"test"},
				Subject:        []string{"test"},
				Category:       []string{"test"},
				Subcategory:    []string{"test"},
				Age:            []string{"test"},
				Grade:          []string{"test"},
				Sets:           []string{"test"},
				ScoreThreshold: 0,
			},
		},
	}
	result, err := GetOutcomeModel().Import(ctx, &entity.Operator{OrgID: "6300b3c5-8936-497e-ba1f-d67164b59c65"}, importData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result.ExistError)
	for _, v := range result.CreateData {
		t.Log(v)
	}
	t.Log("update data")
	for _, v := range result.UpdateData {
		t.Log(v)
	}
}
