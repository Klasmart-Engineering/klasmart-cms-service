package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"math"
)

// First-Level: key is duration, value is student map
// Second-Level: key is student id, value is subject map
// Third-Lever: key is subject id, value is slice of complete and total count
// example: map[duration: '0-1630016787']map[student_id: 'xxxxx']map[subject-id: 'xxxxx'][]int64{5, 10}
func (t *reportModel) convertToTriLevelMap(sas []*entity.StudentAssignmentStatus) map[string]map[string]map[string][]int64 {
	result := make(map[string]map[string]map[string][]int64)
	for _, s := range sas {
		if _, ok := result[s.Duration]; !ok {
			result[s.Duration] = make(map[string]map[string][]int64)
		}
		studentMap := result[s.Duration]

		if _, ok := studentMap[s.StudentID]; !ok {
			studentMap[s.StudentID] = make(map[string][]int64)
		}
		subjectMap := studentMap[s.StudentID]

		subjectMap[s.SubjectID] = []int64{s.Finish, s.Total}
	}
	return result
}

func (t *reportModel) GetAssignmentCompletion(ctx context.Context, op *entity.Operator, args *entity.AssignmentRequest) (entity.AssignmentResponse, error) {
	var res entity.AssignmentResponse
	results, err := da.GetReportDA().ListAssignments(ctx, op, args)
	if err != nil {
		log.Debug(ctx, "GetAssignmentCompletion: ListAssignments failed")
		return res, err
	}

	mapResult := t.convertToTriLevelMap(results)

	log.Debug(ctx, "GetAssignmentCompletion: average_raw",
		log.Any("args", args),
		log.Any("results", mapResult))

	result := make([]*entity.AssignmentCompletionRate, len(args.Durations))

	for i, v := range args.Durations {
		selected := utils.SliceDeduplication(args.SelectedSubjectIDList)
		unSelected := utils.SliceDeduplication(args.UnSelectedSubjectIDList)
		_, rate, finish, total := t.calculateStudentSubjectAverage(ctx, mapResult, string(v), args.StudentID, selected)

		averageRate := &entity.AssignmentCompletionRate{
			StudentID:                 args.StudentID,
			StudentCompleteAssignment: finish,
			StudentTotalAssignment:    total,
			StudentDesignatedSubject:  rate,
			Duration:                  v,
		}

		_, averageRate.ClassDesignatedSubject = t.calculateClassDesignedSubjectAverage(ctx, mapResult, string(v), selected)
		_, averageRate.StudentNonDesignatedSubject, _, _ = t.calculateStudentSubjectAverage(ctx, mapResult, string(v), args.StudentID, unSelected)

		result[i] = averageRate
	}
	res.Assignments = result
	labelID, labelParams := getAssignmentLabelIDAndParams(res)
	res.LabelID = labelID
	res.LabelParams = labelParams
	return res, nil
}

func (t *reportModel) calculateStudentSubjectAverage(ctx context.Context, raw map[string]map[string]map[string][]int64, duration, studentID string, subjectIDs []string) (bool, float64, int64, int64) {
	var sum float64
	var count int
	var sumFinishCount, sumTotalCount int64

	if raw != nil && raw[duration] != nil && raw[duration][studentID] != nil {
		data := raw[duration][studentID]
		for _, s := range subjectIDs {
			if data[s] != nil && len(data[s]) >= 2 {
				numerator := data[s][0]
				denominator := data[s][1]
				if denominator > 0 {
					sum += float64(numerator) / float64(denominator)
					count++

					sumFinishCount += numerator
					sumTotalCount += denominator
				}
			}

			log.Debug(ctx, "student_average",
				log.String("duration", duration),
				log.String("student_id", studentID),
				log.String("subject_id", s),
				log.Any("finish_total", data[s]))
		}
	} else {
		log.Debug(ctx, "student_average",
			log.String("duration", duration),
			log.String("student_id", studentID))
		return false, 0, sumFinishCount, sumTotalCount
	}

	log.Debug(ctx, "student_average---",
		log.String("duration", duration),
		log.String("student_id", studentID),
		log.Float64("rate_sum", sum),
		log.Int("rate_count", count),
		log.Int64("finish_sum", sumFinishCount),
		log.Int64("total_sum", sumTotalCount))

	if count <= 0 {
		return false, 0, sumFinishCount, sumTotalCount
	}
	return true, sum / float64(count), sumFinishCount, sumTotalCount
}

func (t *reportModel) calculateClassDesignedSubjectAverage(ctx context.Context, raw map[string]map[string]map[string][]int64, duration string, subjectIDs []string) (bool, float64) {
	var sum float64
	var count int

	if raw != nil && raw[duration] != nil {
		data := raw[duration]
		for k, v := range data {
			need, rate, _, _ := t.calculateStudentSubjectAverage(ctx, raw, duration, k, subjectIDs)
			if need {
				sum += rate
				count++
			}
			log.Debug(ctx, "class_average",
				log.String("duration", duration),
				log.Any("student_data", v),
				log.String("student_id", k),
				log.Float64("rate", rate),
				log.Bool("need", need))
		}
	} else {
		log.Debug(ctx, "class_average",
			log.String("duration", duration))
		return false, 0
	}

	log.Debug(ctx, "class_average---",
		log.Float64("rate_sum", sum),
		log.Int("rate_count", count))
	if count <= 0 {
		return false, 0
	}
	return true, sum / float64(count)
}

func getAssignmentLabelIDAndParams(res entity.AssignmentResponse) (labelID string, labelParams entity.AssignmentLabelParams) {
	data := res.Assignments
	if data[0].ClassDesignatedSubject == 0 && data[0].StudentDesignatedSubject == 0 &&
		data[0].StudentNonDesignatedSubject == 0 &&
		data[1].ClassDesignatedSubject == 0 &&
		data[1].StudentDesignatedSubject == 0 &&
		data[1].StudentNonDesignatedSubject == 0 &&
		data[2].ClassDesignatedSubject == 0 &&
		data[2].StudentDesignatedSubject == 0 &&
		data[2].StudentNonDesignatedSubject == 0 {
		labelID = entity.AssignNew
		labelParams.AssignmentCompleteCount = data[3].StudentCompleteAssignment
		labelParams.AssignmentCount = data[3].StudentTotalAssignment
	} else if (data[1].StudentDesignatedSubject > data[1].ClassDesignatedSubject &&
		data[2].StudentDesignatedSubject > data[2].ClassDesignatedSubject &&
		data[3].StudentDesignatedSubject > data[3].ClassDesignatedSubject) ||
		(data[1].StudentDesignatedSubject < data[1].ClassDesignatedSubject &&
			data[2].StudentDesignatedSubject < data[2].ClassDesignatedSubject &&
			data[3].StudentDesignatedSubject < data[3].ClassDesignatedSubject) {
		if data[1].StudentDesignatedSubject > data[1].ClassDesignatedSubject &&
			data[2].StudentDesignatedSubject > data[2].ClassDesignatedSubject &&
			data[3].StudentDesignatedSubject > data[3].ClassDesignatedSubject {
			labelID = entity.AssignHighClass3w
			labelParams.AssignCompareClass3week = math.Ceil(getDesignatedSub(data[3], data[2], data[1]) / 3)
		} else {
			labelID = entity.AssignLowClass3w
			labelParams.AssignCompareClass3week = math.Ceil(getAbverseDesignatedSub(data[3], data[2], data[1]) / 3)
		}
	} else if getDesignatedSubResult(data[3], data[2]) >= 20 || getDesignatedSubResult(data[2], data[3]) >= 20 {
		if getDesignatedSubResult(data[3], data[2]) >= 20 {
			labelID = entity.AssignIncreasePreviousLargeW
			labelParams.AssignCompareLastWeek = math.Ceil(getDesignatedSubResult(data[3], data[2]))
		} else {
			labelID = entity.AssignDecreasePreviousLargeW
			labelParams.AssignCompareLastWeek = math.Ceil(getDesignatedSubResult(data[2], data[3]))
		}
	} else if (getDesignatedSubResult(data[2], data[1]) > 0 &&
		getDesignatedSubResult(data[1], data[0]) > 0 &&
		getDesignatedSubResult(data[3], data[2]) > 0) ||
		(getDesignatedSubResult(data[2], data[1]) < 0 &&
			getDesignatedSubResult(data[1], data[0]) < 0 &&
			getDesignatedSubResult(data[3], data[2]) < 0) {
		if getDesignatedSubResult(data[2], data[1]) > 0 &&
			getDesignatedSubResult(data[1], data[0]) < 0 &&
			getDesignatedSubResult(data[3], data[2]) > 0 {
			labelID = entity.AssignIncrease3w
			labelParams.AssignCompare3Week = math.Ceil(getDesignatedSubResult(data[3], data[0]))
		} else {
			labelID = entity.AssignDecrease3w
			labelParams.AssignCompare3Week = math.Ceil(getDesignatedSubResult(data[0], data[3]))
		}
	} else if data[3].StudentDesignatedSubject > data[3].ClassDesignatedSubject ||
		data[3].StudentDesignatedSubject < data[3].ClassDesignatedSubject {
		if data[3].StudentDesignatedSubject > data[3].ClassDesignatedSubject {
			labelID = entity.AssignHighClassW
			labelParams.AssignCompareClass = math.Ceil((data[3].StudentDesignatedSubject - data[3].ClassDesignatedSubject) * 100)
		} else {
			labelID = entity.AssignLowClassW
			labelParams.AssignCompareClass = math.Ceil((data[3].ClassDesignatedSubject - data[3].StudentDesignatedSubject) * 100)
		}
	} else if (getDesignatedSubResult(data[3], data[2]) < 20 && getDesignatedSubResult(data[3], data[2]) > 0) ||
		(getDesignatedSubResult(data[2], data[3]) < 20 && getDesignatedSubResult(data[2], data[3]) > 0) {
		if getDesignatedSubResult(data[3], data[2]) < 20 && getDesignatedSubResult(data[3], data[2]) > 0 {
			labelID = entity.AssignIncreasePreviousW
			labelParams.AssignCompareLastWeek = math.Ceil(getDesignatedSubResult(data[3], data[2]))
		} else {
			labelID = entity.AssignDecreasePreviousW
			labelParams.AssignCompareLastWeek = math.Ceil(getDesignatedSubResult(data[2], data[3]))
		}
	} else {
		labelID = entity.AssignDefault
		labelParams.AssignCompleteCount = data[3].StudentCompleteAssignment
		labelParams.AssignmentCount = data[3].StudentTotalAssignment
	}
	return
}

func getDesignatedSub(data1, data2, data3 *entity.AssignmentCompletionRate) (result float64) {
	return (data1.StudentDesignatedSubject - data1.ClassDesignatedSubject +
		(data2.StudentDesignatedSubject - data2.ClassDesignatedSubject) +
		(data3.StudentDesignatedSubject - data3.ClassDesignatedSubject)) * 100

}

func getAbverseDesignatedSub(data1, data2, data3 *entity.AssignmentCompletionRate) (result float64) {
	return (data1.ClassDesignatedSubject -
		data1.StudentDesignatedSubject +
		(data2.ClassDesignatedSubject -
			data2.StudentDesignatedSubject) +
		(data3.ClassDesignatedSubject -
			data3.StudentDesignatedSubject)) * 100

}

func getDesignatedSubResult(data1, data2 *entity.AssignmentCompletionRate) (result float64) {
	return (data1.StudentDesignatedSubject - data2.StudentDesignatedSubject) * 100
}
