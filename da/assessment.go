package da

import (
	"context"
	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

type IAssessmentDA interface {
	Get(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Assessment, error)
	List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) ([]*entity.Assessment, error)
	Add(ctx context.Context, tx *dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error)
	UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error
}

var (
	assessmentDAInstance     IAssessmentDA
	assessmentDAInstanceOnce = sync.Once{}
)

func GetAssessmentDA() IAssessmentDA {
	assessmentDAInstanceOnce.Do(func() {
		assessmentDAInstance = &assessmentDA{}
	})
	return assessmentDAInstance
}

type assessmentDA struct{}

func (a *assessmentDA) Get(ctx context.Context, tx *dbo.DBContext, id string) (*entity.Assessment, error) {
	item := entity.Assessment{}
	if err := tx.Model(entity.Assessment{}).
		Where(a.filterDeletedAtTemplate()).
		Where("id = ?").First(&item).Error; err != nil {
		log.Error(ctx, "get assessment: get from db failed",
			log.Err(err),
			log.String("id", id),
		)
		return nil, err
	}
	return &item, nil
}

func (a *assessmentDA) List(ctx context.Context, tx *dbo.DBContext, cmd entity.ListAssessmentsCommand) ([]*entity.Assessment, error) {
	db := tx.Model(entity.Assessment{}).Where(a.filterDeletedAtTemplate())
	if cmd.Status != nil {
		db = db.Where("status = ?", *cmd.Status)
	}
	if len(cmd.TeacherIDs) > 0 {
		db = db.Where("teacher_id in ?", cmd.TeacherIDs)
	}
	if cmd.OrderBy != nil {
		switch *cmd.OrderBy {
		case entity.ListAssessmentsOrderByClassEndTime:
			db = db.Order("class_end_time")
		case entity.ListAssessmentsOrderByClassEndTimeDesc:
			db = db.Order("class_end_time desc")
		case entity.ListAssessmentsOrderByCompleteTime:
			db = db.Order("complete_time")
		case entity.ListAssessmentsOrderByCompleteTimeDesc:
			db = db.Order("complete_time desc")
		}
	}
	db = a.paging(db, cmd.Page, cmd.PageSize)
	var items []*entity.Assessment
	if err := db.Find(&items).Error; err != nil {
		log.Error(ctx, "list assessments: find from db failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		return nil, err
	}
	return items, nil
}

func (a *assessmentDA) Add(ctx context.Context, tx *dbo.DBContext, cmd entity.AddAssessmentCommand) (string, error) {
	var (
		newID   = utils.NewID()
		nowUnix = time.Now().Unix()
	)
	item := entity.Assessment{
		ID:           newID,
		ScheduleID:   cmd.ScheduleID,
		Title:        cmd.Title(),
		ProgramID:    cmd.ProgramID,
		SubjectID:    cmd.SubjectID,
		TeacherID:    cmd.TeacherID,
		ClassLength:  cmd.ClassLength,
		ClassEndTime: cmd.ClassEndTime,
		CompleteTime: cmd.CompleteTime,
		Status:       cmd.Status,
		CreatedAt:    nowUnix,
		UpdatedAt:    nowUnix,
	}
	if err := tx.Create(&item).Error; err != nil {
		log.Error(ctx, "add assessment: create from db failed",
			log.Err(err),
			log.Any("cmd", cmd),
		)
		return "", err
	}
	return newID, nil
}

func (a *assessmentDA) UpdateStatus(ctx context.Context, tx *dbo.DBContext, id string, status entity.AssessmentStatus) error {
	if err := tx.Model(entity.Assessment{}).
		Where(a.filterDeletedAtTemplate()).
		Update("status", status).Error; err != nil {
		log.Error(ctx, "update assessment status: update failed in db",
			log.Err(err),
			log.String("id", id),
			log.String("status", string(status)),
		)
		return err
	}
	return nil
}

func (a *assessmentDA) filterDeletedAtTemplate() string {
	return "deleted_at != 0"
}

func (a *assessmentDA) paging(db *gorm.DB, pagePtr, sizePtr *int) *gorm.DB {
	page, size := 0, 0
	if pagePtr != nil {
		page = *pagePtr
	}
	if page == 0 {
		page = 1
	}
	if sizePtr != nil {
		size = *sizePtr
	}
	if size == 0 {
		size = 10
	}
	if page < 0 || size < 0 {
		return db
	}
	return db.Offset((page - 1) * size).Limit(size)
}
