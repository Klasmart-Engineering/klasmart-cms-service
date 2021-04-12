package main

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func handleTeacherIDs(ctx context.Context, tx *dbo.DBContext) error {
	type assessment struct {
		ID         string                   `gorm:"column:id;type:varchar(64);primary_key" json:"id"`
		TeacherIDs utils.SQLJSONStringArray `gorm:"column:teacher_ids;type:json" json:"teacher_ids"`
	}
	var (
		page     = 1
		size     = 100
		inserted int
	)
	for {
		var (
			assessments              []*assessment
			assessmentIDs            []string
			assessmentAttendances    []*entity.AssessmentAttendance
			assessmentAttendancesMap map[string][]string
			columns                  = []string{"id", "assessment_id", "attendance_id", "checked", "origin", "role"}
			matrix                   [][]interface{}
		)
		if err := tx.Limit(size).Offset(size*(page-1)).Find(&assessments, "json_length(teacher_ids) != 0").Error; err != nil {
			return err
		}
		if len(assessments) == 0 {
			break
		}
		for _, a := range assessments {
			assessmentIDs = append(assessmentIDs, a.ID)
		}
		if err := tx.
			Find(&assessmentAttendances, "assessment_id in (?)", assessmentIDs).
			Error; err != nil {
			return err
		}
		assessmentAttendancesMap = make(map[string][]string, len(assessmentAttendances))
		for _, aa := range assessmentAttendances {
			assessmentAttendancesMap[aa.AssessmentID] = append(assessmentAttendancesMap[aa.AssessmentID], aa.AttendanceID)
		}
		for _, a := range assessments {
			var uniqueTeacherIDs = utils.SliceDeduplication(utils.ExcludeStrings(a.TeacherIDs, assessmentAttendancesMap[a.ID]))
			for _, tid := range uniqueTeacherIDs {
				matrix = append(matrix, []interface{}{
					utils.NewID(),
					a.ID,
					tid,
					true,
					entity.AssessmentAttendanceOriginClassRoaster,
					entity.AssessmentAttendanceRoleTeacher,
				})
				inserted++
			}
		}
		if len(matrix) > 0 {
			format, values := da.SQLBatchInsert("assessments_attendances", columns, matrix)
			if err := tx.Exec(format, values...).Error; err != nil {
				return err
			}
		}
		page++
	}

	fmt.Printf("=> Task: handle teacher ids done! inserted count: %d.\n", inserted)

	return nil
}

func handleAssessmentAttendanceDefault(ctx context.Context, tx *dbo.DBContext) error {
	changes := map[string]interface{}{
		"origin": entity.AssessmentAttendanceOriginClassRoaster,
		"role":   entity.AssessmentAttendanceRoleStudent,
	}
	return tx.Table("assessments_attendances").Where("origin = '' and role = ''").Update(changes).Error
}
