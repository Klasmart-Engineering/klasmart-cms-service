package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
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
	results, err := da.GetReportDA().ListAssignments(ctx, op, args)
	if err != nil {
		log.Debug(ctx, "GetAssignmentCompletion: ListAssignments failed")
		return nil, err
	}

	mapResult := t.convertToTriLevelMap(results)

	res := make([]*entity.AssignmentCompletionRate, len(args.Durations))

	for i, v := range args.Durations {
		_, rate, finish, total := t.calculateStudentSubjectAverage(ctx, mapResult, string(v), args.StudentID, args.SelectedSubjectIDList)

		averageRate := &entity.AssignmentCompletionRate{
			StudentID:                 args.StudentID,
			StudentCompleteAssignment: finish,
			StudentTotalAssignment:    total,
			StudentDesignatedSubject:  rate,
			Duration:                  v,
		}

		_, averageRate.ClassDesignatedSubject = t.calculateClassDesignedSubjectAverage(ctx, mapResult, string(v), args.SelectedSubjectIDList)
		_, averageRate.StudentNonDesignatedSubject, _, _ = t.calculateStudentSubjectAverage(ctx, mapResult, string(v), args.StudentID, args.UnSelectedSubjectIDList)

		res[i] = averageRate
	}

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
			log.String("student_id", studentID),
			log.Strings("subject_ids", subjectIDs),
			log.Any("raw", raw))
		return false, 0, sumFinishCount, sumTotalCount
	}

	log.Debug(ctx, "student_average",
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
			log.String("duration", duration),
			log.Strings("subject_ids", subjectIDs),
			log.Any("raw", raw))
		return false, 0
	}

	log.Debug(ctx, "class_average",
		log.Float64("rate_sum", sum),
		log.Int("rate_count", count))
	if count <= 0 {
		return false, 0
	}
	return true, sum / float64(count)
}
