package model

import (
	"context"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetStudentUsageMaterialViewCount(ctx context.Context, op *entity.Operator, req *entity.StudentUsageMaterialViewCountReportRequest) (res *entity.StudentUsageMaterialViewCountReportResponse, err error) {
	res = &entity.StudentUsageMaterialViewCountReportResponse{
		Request: req,
	}
	usages, err := da.GetStudentUsageDA().GetMaterialUsages(ctx, req)
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

	return
}

func (m *reportModel) GetStudentUsageMaterial(ctx context.Context, op *entity.Operator, req *entity.StudentUsageMaterialReportRequest) (res *entity.StudentUsageMaterialReportResponse, err error) {

	return
}

func (m *reportModel) AddStudentUsageRecordTx(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, record *entity.StudentUsageRecord) (err error) {
	sche, err := GetScheduleModel().GetPlainByID(ctx, record.RoomID)
	if err != nil {
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
	material, _ := materials.FindByUrl(record.LessonMaterialUrl)
	// TODO record.LessonMaterialID = material.ID
	record.LessonMaterialID = material.Name

	var models []entity.BatchInsertModeler
	for _, student := range record.Students {
		usageRecord := *record
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

func (lms LiveMaterialSlice) FindByUrl(url string) (material *entity.LiveMaterial, found bool) {
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
	// TODO
	//return GetLiveTokenModel().GetMaterials(ctx, op, input)
	return
}
