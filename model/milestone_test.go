package model

import (
	"context"
	"fmt"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/test/utils"
)

func TestDelete(t *testing.T) {
	ctx := context.TODO()
	op := utils.InitOperator(ctx, "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjI1MjJlYWUwLTVmNzItNDVkMS05OGY2LTM1ODI3YWI4MTZhNyIsImVtYWlsIjoib3Jna2lkc2xvb3AxQHlvcG1haWwuY29tIiwiZXhwIjoxNjM3NjUxMTg4LCJpc3MiOiJraWRzbG9vcCJ9.B7yv7iij8c40qa4L1NclpNkM4-PqVGHV85vi4yj2zJ0jgN47t6wuExy_ZUFl8JyxaBIVZaM9pKOT35QKEOls57vkXANYrjvupi4_AazwESWCTzRB6OPL9PCijczBzoU_oxUlrDNtfcQq7CfwjICoBAk1YWIiOZxxj_d0Y0LD8huJqpuj_dESjZNxvaqiKNx2FlQkayllMHFey2RC4GCnYmkyL9o2_f6WpQGRRu4OEo5cakMNms_sZQ2z3spZ6lYoxD3E8tJKVB_LXeuwRwUlUSp71nGgzGP5bYbDVjcaiuwAUZ9iTe3a6kdNGDAIfuqE8O1AMa0aYiIVNP-TS3BPUg",
		"org_id")
	err := GetMilestoneModel().Delete(ctx, op,
		map[external.PermissionName]bool{
			external.DeleteUnpublishedMilestone: true,
		}, []string{"org_123"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMilestoneModel_GenerateShortcode(t *testing.T) {
	setup()
	ctx := context.TODO()
	op := initOperator()
	shortcode, err := GetMilestoneModel().GenerateShortcode(ctx, op)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(shortcode)
}

func TestMilestoneModel_Create(t *testing.T) {
	// setup()
	// ctx := context.TODO()
	// op := initOperator()
	// req := MilestoneView{
	// 	Name:           "name00",
	// 	Shortcode:      "00003",
	// 	Type:           entity.CustomMilestoneType,
	// 	Organization:   &OrganizationView{OrganizationID: op.OrgID},
	// 	ProgramIDs:     []string{"75004121-0c0d-486c-ba65-4c57deacb44b"},
	// 	SubjectIDs:     []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a"},
	// 	CategoryIDs:    []string{"b4cd42b8-a09b-4f66-a03a-b9f6b6f69895", "fa8ff09d-9062-4955-9b20-5fa20757f1d9"},
	// 	SubcategoryIDs: []string{"d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb", "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"},
	// 	GradeIDs:       []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a"},
	// 	AgeIDs:         []string{"bb7982cd-020f-4e1a-93fc-4a6874917f07"},
	// 	OutcomeAncestorIDs: []string{
	// 		"607905030e4404103a3f595d",
	// 	},
	// }
	// milestone, err := req.ToMilestone(ctx, op)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// err = GetMilestoneModel().Create(ctx, op, milestone, req.OutcomeAncestorIDs)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("ok")
}

func TestMilestoneModel_Update(t *testing.T) {
	// setup()
	// ctx := context.TODO()
	// op := initOperator()
	// req := MilestoneView{
	// 	Name:           "name06",
	// 	Shortcode:      "00005",
	// 	Type:           entity.CustomMilestoneType,
	// 	Organization:   &OrganizationView{OrganizationID: op.OrgID},
	// 	ProgramIDs:     []string{"75004121-0c0d-486c-ba65-4c57deacb44b"},
	// 	SubjectIDs:     []string{"5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a"},
	// 	CategoryIDs:    []string{"b4cd42b8-a09b-4f66-a03a-b9f6b6f69895", "fa8ff09d-9062-4955-9b20-5fa20757f1d9"},
	// 	SubcategoryIDs: []string{"d50cff7c-b0c7-43be-8ec7-877fa4c9a6fb", "49e73e4f-8ffc-47e3-9b87-0f9686d361d7"},
	// 	GradeIDs:       []string{"3ee3fd4c-6208-494f-9551-d48fabc4f42a"},
	// 	AgeIDs:         []string{"bb7982cd-020f-4e1a-93fc-4a6874917f07"},
	// 	OutcomeAncestorIDs: []string{
	// 		"607905030e4404103a3f595d",
	// 	},
	// }
	// milestone, err := req.ToMilestone(ctx, op)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// milestone.ID = "607e4f4c9a752f785251dcef"
	// perms := make(map[external.PermissionName]bool)
	// perms[external.EditUnpublishedMilestone] = true
	// perms[external.EditPublishedMilestone] = true
	// err = GetMilestoneModel().Update(ctx, op, perms, milestone, req.OutcomeAncestorIDs)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("ok")
}
