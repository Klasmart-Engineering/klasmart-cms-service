package da

import (
	"context"
	"testing"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

func initContentFolderQueryCondition() (*CombineConditions, *FolderCondition) {
	return &CombineConditions{
			SourceCondition: &ContentCondition{
				ContentType: []int{1, 2, 10},
				VisibilitySettings: []string{
					"69d4d65c-3568-4f84-9420-a0b62e6f2c31",
					"da71af23-8e40-4677-888d-63fa35afea58",
					"aaab307d-bbe2-456e-8266-a420d03ecdad",
					"00fd1f6f-4f54-4179-a39d-93cbd8f99f78",
					"aaa18ba4-6d00-4d7e-8c5e-f24770b138c4",
					"def60b38-9518-4741-942e-0518b55d06b1",
					"94aa0387-c1a4-46b0-b2d2-2aeb4906c771",
					"55d1a188-9e40-44e3-b976-a260706e4e40",
					"7899a40f-c48e-47a4-a844-babebd1e703a",
					"4f37ac93-5718-4bd4-a460-5bfff5530fe7",
					"60ee7b42-e36e-4821-a5b2-4038a7b3cd6f",
					"bbab3288-9e5c-4501-b967-e0d5c2b92c88",
					"d0d86fdd-a895-4937-9792-6d820661e239",
					"975c7051-503e-43d3-bf9b-65f50d56e6e9",
					"fc22837e-b057-4c0d-9fb6-baede6ac099a",
					"96837ac1-1829-4faf-a09b-bddc89a1305f",
					"201591ab-a859-4ed9-97e4-6038d1bdab98",
					"b4194788-addb-4232-9870-b07ff4bf59ab",
					"338f9286-cbf6-45f8-ac02-004866e70c23",
					"30f464ab-b703-4138-8a2f-519b974cc938",
					"418de7ce-fa5c-4f9a-9ebb-f2888a57800d",
					"1de3b1cb-0ab2-47dd-8441-c537cd08b93b",
					"c4d48780-4926-4589-b671-6ea3f3f7cf55",
					"edadd7b2-06a3-486c-ace2-f246fce23e00",
					"6c388a72-c56b-4e94-93a4-4cf359e01763",
					"b6e16065-82b8-4505-a601-19c8c2283fe9",
					"c40338a1-cc5f-41ea-9815-a76110dd3409",
					"e5be1ab4-7366-42d0-9fac-2eafb8ab8a1d",
					"d4b032ac-acab-4a60-9ef9-18576786dfda",
					"5fb25a43-7fc2-41bc-9158-20f5fe41f420",
					"3a978e71-7996-4d75-a6c3-98b4276f6f66",
					"0a24ebd3-cfbe-49ce-af6e-71664ae2d905",
					"ac7fe609-92c2-4039-883d-c1007c4ae47a",
					"3be2b719-6ff8-4415-931b-0e1c4b0133d5",
					"40411300-452a-4cac-803f-a24a1555d8ab",
					"f529a34a-01c1-4656-afbf-87ccd4cf2324",
					"4f4c8050-9099-4c22-a858-f1524ca6fcfc",
					"dca9e87f-07ab-4126-82f4-dc34e784e0e7",
					"f195d9a2-3361-4e28-9cb6-956041b12163",
					"a92c81d0-73ab-4f93-9590-469b83bff6c7",
					"209bd5b2-6691-48a2-a9cc-555cb44f8d00",
					"9d5bf4ae-17e9-4b58-9d9b-1c854becab0d",
					"b534c090-0ab8-4f58-b991-cbd585109661",
					"6382e3b9-a485-4eb8-807e-19a42a9e860c",
					"c65d9548-6798-4966-ba95-32ba94eb6a91",
					"dad9f50e-19c2-4402-a4c6-72ad688861dc",
					"4a7fd784-1dfb-471a-b034-7057cea6fa80",
					"69121e0f-c725-4d06-8176-bfee84460f71",
					"30b689eb-7f26-41e2-998d-ee64b6462ef3",
					"76f30e1e-3c62-483d-a4b3-f866c0de1443",
				},
				PublishStatus: []string{"published"},
				Author:        "d4268066-9f35-4588-abfc-6f96235ebb56",
				DirPath:       entity.NullStrings{Strings: []string{"/"}, Valid: true},
				OrderBy:       ContentOrderByUpdatedAtDesc,
				Pager:         utils.Pager{PageIndex: 1, PageSize: 100},
			},
			TargetCondition: &ContentCondition{
				ContentType: []int{1, 2, 10},
				VisibilitySettings: []string{
					"4f4c8050-9099-4c22-a858-f1524ca6fcfc",
					"dca9e87f-07ab-4126-82f4-dc34e784e0e7",
					"69121e0f-c725-4d06-8176-bfee84460f71",
					"60ee7b42-e36e-4821-a5b2-4038a7b3cd6f",
					"d0d86fdd-a895-4937-9792-6d820661e239",
					"3a978e71-7996-4d75-a6c3-98b4276f6f66",
					"4f37ac93-5718-4bd4-a460-5bfff5530fe7",
					"f195d9a2-3361-4e28-9cb6-956041b12163",
					"209bd5b2-6691-48a2-a9cc-555cb44f8d00",
					"201591ab-a859-4ed9-97e4-6038d1bdab98",
					"418de7ce-fa5c-4f9a-9ebb-f2888a57800d",
					"aaab307d-bbe2-456e-8266-a420d03ecdad",
					"def60b38-9518-4741-942e-0518b55d06b1",
					"96837ac1-1829-4faf-a09b-bddc89a1305f",
					"b6e16065-82b8-4505-a601-19c8c2283fe9",
					"e5be1ab4-7366-42d0-9fac-2eafb8ab8a1d",
					"d4b032ac-acab-4a60-9ef9-18576786dfda",
					"c65d9548-6798-4966-ba95-32ba94eb6a91",
					"975c7051-503e-43d3-bf9b-65f50d56e6e9",
					"30f464ab-b703-4138-8a2f-519b974cc938",
					"1de3b1cb-0ab2-47dd-8441-c537cd08b93b",
					"fc22837e-b057-4c0d-9fb6-baede6ac099a",
					"c4d48780-4926-4589-b671-6ea3f3f7cf55",
					"edadd7b2-06a3-486c-ace2-f246fce23e00",
					"0a24ebd3-cfbe-49ce-af6e-71664ae2d905",
					"9d5bf4ae-17e9-4b58-9d9b-1c854becab0d",
					"da71af23-8e40-4677-888d-63fa35afea58",
					"aaa18ba4-6d00-4d7e-8c5e-f24770b138c4",
					"55d1a188-9e40-44e3-b976-a260706e4e40",
					"30b689eb-7f26-41e2-998d-ee64b6462ef3",
					"f529a34a-01c1-4656-afbf-87ccd4cf2324",
					"b534c090-0ab8-4f58-b991-cbd585109661",
					"4a7fd784-1dfb-471a-b034-7057cea6fa80",
					"76f30e1e-3c62-483d-a4b3-f866c0de1443",
					"7899a40f-c48e-47a4-a844-babebd1e703a",
					"6c388a72-c56b-4e94-93a4-4cf359e01763",
					"40411300-452a-4cac-803f-a24a1555d8ab",
					"c40338a1-cc5f-41ea-9815-a76110dd3409",
					"ac7fe609-92c2-4039-883d-c1007c4ae47a",
					"6382e3b9-a485-4eb8-807e-19a42a9e860c",
					"69d4d65c-3568-4f84-9420-a0b62e6f2c31",
					"b4194788-addb-4232-9870-b07ff4bf59ab",
					"338f9286-cbf6-45f8-ac02-004866e70c23",
					"5fb25a43-7fc2-41bc-9158-20f5fe41f420",
					"3be2b719-6ff8-4415-931b-0e1c4b0133d5",
					"a92c81d0-73ab-4f93-9590-469b83bff6c7",
					"dad9f50e-19c2-4402-a4c6-72ad688861dc",
					"00fd1f6f-4f54-4179-a39d-93cbd8f99f78",
					"94aa0387-c1a4-46b0-b2d2-2aeb4906c771",
					"bbab3288-9e5c-4501-b967-e0d5c2b92c88",
				},
				PublishStatus: []string{"published"},
				Author:        "d4268066-9f35-4588-abfc-6f96235ebb56",
				DirPath:       entity.NullStrings{Strings: []string{"/"}, Valid: true},
				OrderBy:       ContentOrderByUpdatedAtDesc,
				Pager:         utils.Pager{PageIndex: 1, PageSize: 100},
			},
		},
		&FolderCondition{
			OwnerType:       1,
			ItemType:        1,
			Owner:           "69d4d65c-3568-4f84-9420-a0b62e6f2c31",
			Partition:       entity.FolderPartitionMaterialAndPlans,
			ExactDirPath:    "/",
			ShowEmptyFolder: entity.NullBool{Bool: true, Valid: true},
			OrderBy:         FolderOrderByCreatedAtDesc,
			Pager:           utils.Pager{PageIndex: 0, PageSize: 0},
		}
}
func TestContentDA_SearchFolderContentUnsafe(t *testing.T) {
	ctx := context.Background()
	contentDA := GetContentDA()

	cc, fc := initContentFolderQueryCondition()

	tx := dbo.MustGetDB(ctx)
	total, records, err := contentDA.SearchFolderContentUnsafe(ctx, tx, cc, fc)
	if err != nil {
		t.Errorf("search content and folder failed due to %v", err)
	}

	_ = total
	_ = records
}

func BenchmarkContentDA_SearchFolderContentUnsafe(b *testing.B) {
	ctx := context.Background()
	contentDA := GetContentDA()

	cc, fc := initContentFolderQueryCondition()

	tx := dbo.MustGetDB(ctx)
	for n := 0; n < b.N; n++ {
		_, _, err := contentDA.SearchFolderContentUnsafe(ctx, tx, cc, fc)
		if err != nil {
			b.Errorf("search content and folder failed due to %v", err)
		}
	}
}

func BenchmarkContentMySQLDA_SearchFolderContentUnsafe(b *testing.B) {
	ctx := context.Background()
	contentDA := new(ContentMySQLDA)

	cc, fc := initContentFolderQueryCondition()

	tx := dbo.MustGetDB(ctx)
	for n := 0; n < b.N; n++ {
		_, _, err := contentDA.SearchFolderContentUnsafe(ctx, tx, cc, fc)
		if err != nil {
			b.Errorf("search content and folder failed due to %v", err)
		}
	}
}
