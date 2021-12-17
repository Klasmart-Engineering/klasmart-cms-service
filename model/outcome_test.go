package model

import (
	"context"
	"fmt"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
