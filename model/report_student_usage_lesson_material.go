package model

import (
	"context"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetStudentUsageMaterialViewCount(ctx context.Context, op *entity.Operator, req *entity.StudentUsageMaterialViewCountReportRequest) (res *entity.StudentUsageMaterialViewCountReportResponse, err error) {
	res = &entity.StudentUsageMaterialViewCountReportResponse{
		Request:          req,
		ContentUsageList: make([]*entity.ContentUsage, 0),
	}
	usages, err := da.GetStudentUsageDA().GetMaterialViewCountUsages(ctx, req)
	if err != nil {
		return
	}

	for _, usage := range usages {
		contentUsage := &entity.ContentUsage{
			Type:  usage.ContentType,
			Count: usage.UsedCount,
		}
		res.ContentUsageList = append(res.ContentUsageList, contentUsage)
	}

	res.ContentUsageList = res.ContentUsageList.FillZeroItems(req.TimeRangeList, req.ContentTypeList)
	return
}

func (m *reportModel) GetStudentUsageMaterial(ctx context.Context, op *entity.Operator, req *entity.StudentUsageMaterialReportRequest) (res *entity.StudentUsageMaterialReportResponse, err error) {
	res = &entity.StudentUsageMaterialReportResponse{
		Request:        req,
		ClassUsageList: make([]*entity.ClassUsage, 0),
	}
	usages, err := da.GetStudentUsageDA().GetMaterialUsages(ctx, req)
	if err != nil {
		return
	}

	mClassUsage := map[string]*entity.ClassUsage{}
	for _, usage := range usages {
		u, ok := mClassUsage[usage.ClassID]
		if !ok {
			u = &entity.ClassUsage{
				ID: usage.ClassID,
			}
			mClassUsage[usage.ClassID] = u
		}
		u.ContentUsageList = append(u.ContentUsageList, &entity.ContentUsage{
			TimeRange: usage.TimeRange,
			Type:      usage.ContentType,
			Count:     usage.UsedCount,
		})
	}

	for _, classUsage := range mClassUsage {
		res.ClassUsageList = append(res.ClassUsageList, classUsage)
	}

	for _, classID := range req.ClassIDList {
		classUsage, found := res.ClassUsageList.Find(classID)
		if !found {
			classUsage = &entity.ClassUsage{
				ID:               classID,
				ContentUsageList: make(entity.ContentUsageSlice, 0),
			}
			res.ClassUsageList = append(res.ClassUsageList, classUsage)
		}
		classUsage.ContentUsageList = classUsage.ContentUsageList.FillZeroItems(req.TimeRangeList, req.ContentTypeList)
	}
	return
}

func (m *reportModel) AddStudentUsageRecordTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, record *entity.StudentUsageRecord) (err error) {
	sche, err := GetScheduleModel().GetPlainByID(ctx, record.RoomID)
	if err != nil {
		log.Error(ctx, "can not find schedule by id", log.Any("schedule_id", record.RoomID))
		err = constant.ErrInvalidArgs
		return
	}
	if sche.LessonPlanID == "" {
		return
	}
	record.LessonPlanID = sche.LessonPlanID
	record.ScheduleStartAt = sche.StartAt
	classID, err := GetScheduleRelationModel().GetClassRosterID(ctx, op, record.RoomID)
	if err != nil {
		return
	}
	record.ClassID = classID

	materials, err := m.getMaterials(ctx, op, &entity.MaterialInput{
		ContentID:  sche.LessonPlanID,
		ScheduleID: sche.ID,
		TokenType:  entity.LiveTokenTypeLive,
	})
	if err != nil {
		return
	}
	material, found := materials.FindByUrl(ctx, record.LessonMaterialUrl)
	if found {
		record.LessonMaterialID = material.ID
	}

	var models []entity.BatchInsertModeler
	for _, student := range record.Students {
		usageRecord := *record
		usageRecord.ID = utils.NewID()
		usageRecord.StudentUserID = student.UserID
		usageRecord.StudentName = student.Name
		usageRecord.StudentEmail = student.Email
		models = append(models, &usageRecord)
	}
	err = da.GetStudentUsageDA().BatchInsertTx(ctx, tx, models...)
	if err != nil {
		return
	}
	return
}

type LiveMaterialSlice []*entity.LiveMaterial

func (lms LiveMaterialSlice) FindByUrl(ctx context.Context, url string) (material *entity.LiveMaterial, found bool) {
	defer func() {
		log.Info(
			ctx,
			"find material by url",
			log.Any("materials", lms),
			log.Any("url", url),
			log.Any("material", material),
			log.Any("found", found),
		)
	}()
	for _, lm := range lms {
		if strings.TrimSpace(lm.URL) == strings.TrimSpace(url) {
			material = lm
			found = true
			return
		}
	}
	return
}

func (m *reportModel) getMaterials(ctx context.Context, op *entity.Operator, input *entity.MaterialInput) (materials LiveMaterialSlice, err error) {
	return GetLiveTokenModel().GetMaterials(ctx, op, input, true)
}
