package intergrate_academic_profile

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

type Mapper interface {
	Program(ctx context.Context, organizationID, programID string) (string, error)
	Subject(ctx context.Context, organizationID, programID, subjectID string) (string, error)
	// developmental
	Category(ctx context.Context, organizationID, programID, categoryID string) (string, error)
	// skills
	SubCategory(ctx context.Context, organizationID, programID, categoryID, subCategoryID string) (string, error)
	Age(ctx context.Context, organizationID, programID, AgeID string) (string, error)
	Grade(ctx context.Context, organizationID, programID, gradeID string) (string, error)
}

func NewMapper(operator *entity.Operator) Mapper {
	ctx := context.TODO()

	impl := &MapperImpl{
		operator: operator,
		headquarters: map[string]bool{
			"10f38ce9-5152-4049-b4e7-6d2e2ba884e6": true, // Badanamu HQ
			"98988158-a2cd-47d1-b199-5afa74639b80": true, // International Sales Team NEW
			"740ec808-bd56-46c6-8bcb-babbe1666dc4": true, // Badanamu KR
			"f5b80992-9b4a-4f05-8f6e-27090569b75a": true, // KidsLoop Indonesia
			"af405e51-b65a-49fe-ab5f-2cd9dd435aee": true, // KidsloopVN
		},
	}

	err := impl.initProgramMapper(ctx)
	if err != nil {
		log.Panic(ctx, "init program mapping failed", log.Err(err))
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()

		err = impl.initSubjectMapper(ctx)
		if err != nil {
			log.Panic(ctx, "init subject mapping failed", log.Err(err))
		}
	}()

	go func() {
		defer wg.Done()

		err = impl.initAgeMapper(ctx)
		if err != nil {
			log.Panic(ctx, "init age mapping failed", log.Err(err))
		}
	}()

	go func() {
		defer wg.Done()

		err = impl.initGradeMapper(ctx)
		if err != nil {
			log.Panic(ctx, "init grade mapping failed", log.Err(err))
		}
	}()

	go func() {
		defer wg.Done()

		err = impl.initCategoryMapper(ctx)
		if err != nil {
			log.Panic(ctx, "init category mapping failed", log.Err(err))
		}

		err = impl.initSubCategoryMapper(ctx)
		if err != nil {
			log.Panic(ctx, "init sub category mapping failed", log.Err(err))
		}
	}()

	wg.Wait()

	return impl
}

type MapperImpl struct {
	operator *entity.Operator

	// key: org_id  value: isHQ
	headquarters map[string]bool

	// key: ams program name
	amsPrograms map[string]*external.Program
	// key: our program id
	ourPrograms map[string]*entity.Program
	HQPrograms  map[string]bool

	MapperCategory    MapperCategory
	MapperSubCategory MapperSubCategory

	subjectMutex sync.Mutex
	// key: {ams program id}:{ams subject name}
	amsSubjects map[string]*external.Subject
	// key: our subject id
	ourSubjects map[string]*entity.Subject

	gradeMutex sync.Mutex
	// key: {ams program id}:{ams grade name}
	amsGrades map[string]*external.Grade
	// key: our grade id
	ourGrades map[string]*entity.Grade

	ageMutex sync.Mutex
	// key: {ams program id}:{ams age name}
	amsAges map[string]*external.Age
	// key: our age id
	ourAges map[string]*entity.Age
}

type MapperCategory struct {
	amsCategoryMutex sync.Mutex

	// key:program id
	// val:category map(name:category)
	amsCategorys map[string]map[string]*external.Category
	// key:category id
	// val:category
	amsCategoryIDMap map[string]*external.Category
	// key: our category id
	ourCategorys map[string]*entity.Developmental

	// key: our category id
	// value: ams category id
	categoryMapping map[string]string
}

type MapperSubCategory struct {
	amsSubCategoryMutex sync.Mutex

	// key:program id
	// val:map(categoryID+subCategoryName:subCategory)
	amsSubCategorys map[string]map[string]*external.SubCategory
	// key: our subCategory id
	ourSubCategorys map[string]*entity.Skill

	// key: our subCategory id
	// value: ams subCategory id
	subCategoryMapping map[string]string
}
