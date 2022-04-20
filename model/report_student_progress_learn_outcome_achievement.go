package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"math"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetStudentProgressLearnOutcomeAchievement(ctx context.Context, op *entity.Operator, req *entity.LearnOutcomeAchievementRequest) (res *entity.LearnOutcomeAchievementResponse, err error) {
	res = &entity.LearnOutcomeAchievementResponse{
		Request:                               req,
		StudentAchievedCounts:                 map[string]entity.Float64Slice{},
		UnselectSubjectsStudentAchievedCounts: entity.Float64Slice{},
	}
	for _, duration := range req.Durations {
		res.Items = append(res.Items, &entity.LearnOutcomeAchievementResponseItem{
			Duration:                   duration,
			StudentAchievedPercentages: map[string]entity.Float64Slice{},
		})
	}
	counts, err := da.GetReportDA().GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx, req)
	if err != nil {
		return
	}

	for _, c := range counts {
		item := res.GetItem(c.Duration)
		if c.StudentID == req.StudentID {
			if utils.ContainsString(req.SelectedSubjectIDList, c.SubjectID) {
				item.FirstAchievedCount += c.FirstAchievedCount
				item.ReAchievedCount += c.AchievedCount - c.FirstAchievedCount
				item.UnAchievedCount += c.CompletedCount - c.AchievedCount

				item.FirstAchievedPercentages = append(item.FirstAchievedPercentages, float64(c.FirstAchievedCount)/float64(c.CompletedCount))
				item.ReAchievedPercentages = append(item.ReAchievedPercentages, float64(c.AchievedCount-c.FirstAchievedCount)/float64(c.CompletedCount))
			} else {
				item.UnSelectedSubjectsAchievedPercentages = append(item.UnSelectedSubjectsAchievedPercentages, float64(c.AchievedCount)/float64(c.CompletedCount))
				res.UnselectSubjectsStudentAchievedCounts = append(res.UnselectSubjectsStudentAchievedCounts, float64(c.AchievedCount))
			}
		}

		if utils.ContainsString(req.SelectedSubjectIDList, c.SubjectID) {
			item.StudentAchievedPercentages[c.StudentID] = append(item.StudentAchievedPercentages[c.StudentID], float64(c.AchievedCount)/float64(c.CompletedCount))
			res.StudentAchievedCounts[c.StudentID] = append(res.StudentAchievedCounts[c.StudentID], float64(c.AchievedCount))
		}
	}
	for _, item := range res.Items {
		item.FirstAchievedPercentage = item.FirstAchievedPercentages.Avg()
		item.ReAchievedPercentage = item.ReAchievedPercentages.Avg()
		item.UnSelectedSubjectsAverageAchievedPercentage = item.UnSelectedSubjectsAchievedPercentages.Avg()

		classAchievedPercentages := entity.Float64Slice{}
		for _, s := range item.StudentAchievedPercentages {
			classAchievedPercentages = append(classAchievedPercentages, s.Avg())
		}
		item.ClassAverageAchievedPercentage = classAchievedPercentages.Avg()

		res.FirstAchievedCount += item.FirstAchievedCount
		res.ReAchievedCount += item.ReAchievedCount
	}

	classAchievedCount := entity.Float64Slice{}
	for _, s := range res.StudentAchievedCounts {
		classAchievedCount = append(classAchievedCount, s.Sum())
	}
	res.ClassAverageAchievedCount = classAchievedCount.Avg()
	res.UnSelectedSubjectsAverageAchieveCount = res.UnselectSubjectsStudentAchievedCounts.Avg()
	labelID, labelParams := getLearnOutcomeAchievementLabelIDAndParams(res)
	res.LabelID = labelID
	res.LabelParams = labelParams
	return
}

func getLearnOutcomeAchievementLabelIDAndParams(res *entity.LearnOutcomeAchievementResponse) (labelID string, labelParams entity.LearningOutcomeAchivementLabelParams) {
	data := res.Items
	if data[0].ClassAverageAchievedPercentage == 0 && data[0].FirstAchievedPercentage == 0 && data[0].ReAchievedPercentage == 0 &&
		data[0].UnSelectedSubjectsAverageAchievedPercentage == 0 &&
		data[1].ClassAverageAchievedPercentage == 0 && data[1].FirstAchievedPercentage == 0 && data[1].ReAchievedPercentage == 0 &&
		data[1].UnSelectedSubjectsAverageAchievedPercentage == 0 &&
		data[2].ClassAverageAchievedPercentage == 0 && data[2].FirstAchievedPercentage == 0 && data[2].ReAchievedPercentage == 0 &&
		data[2].UnSelectedSubjectsAverageAchievedPercentage == 0 {
		labelID = entity.LONew
		labelParams.AchievedLoCount = data[3].FirstAchievedCount + data[3].ReAchievedCount
		labelParams.LearntLoCount = data[3].FirstAchievedCount + data[3].ReAchievedCount + data[3].UnAchievedCount
	} else if (getSum(data[1]) > data[1].ClassAverageAchievedPercentage*100 &&
		getSum(data[2]) > data[2].ClassAverageAchievedPercentage*100 &&
		getSum(data[3]) > data[3].ClassAverageAchievedPercentage*100) ||
		(getSum(data[1]) < data[1].ClassAverageAchievedPercentage*100 &&
			getSum(data[2]) < data[2].ClassAverageAchievedPercentage*100 &&
			getSum(data[3]) < data[3].ClassAverageAchievedPercentage*100) {
		if getSum(data[1]) > data[1].ClassAverageAchievedPercentage*100 &&
			getSum(data[2]) > data[2].ClassAverageAchievedPercentage*100 &&
			getSum(data[3]) > data[3].ClassAverageAchievedPercentage*100 {
			labelID = entity.LOHighClass3w
			labelParams.LOCompareClass3week = math.Ceil(getResult(data[3], data[2], data[1]) / 3)
		} else {
			labelID = entity.LOLowClass3w
			labelParams.LOCompareClass3week = math.Ceil(getAdverseResult(data[3], data[2], data[1]) / 3)
		}

	} else if getPercentage(data[3], data[2]) >= 20 || getPercentage(data[2], data[3]) >= 20 {
		if getPercentage(data[3], data[2]) >= 20 {
			labelID = entity.LOIncreasePreviousLargeW
			labelParams.LOCompareLastWeek = math.Ceil(getPercentage(data[3], data[2]))
		} else {
			labelID = entity.LODecreasePreviousLargeW
			labelParams.LOCompareLastWeek = math.Ceil(getPercentage(data[2], data[3]))
		}
	} else if (data[3].ReAchievedPercentage*100-data[3].ClassAverageAchievedPercentage*100 >= 10) ||
		(data[3].ClassAverageAchievedPercentage*100-data[3].ReAchievedPercentage*100 >= 10) {
		if data[3].ReAchievedPercentage*100-data[3].ClassAverageAchievedPercentage*100 >= 10 {
			labelID = entity.LOHighClassReviewW
			labelParams.LOReviewCompareClass = math.Ceil((data[3].ReAchievedPercentage - data[3].ClassAverageAchievedPercentage) * 100)
		} else {
			labelID = entity.LOLowClassReviewW
			labelParams.LOReviewCompareClass = math.Ceil((data[3].ClassAverageAchievedPercentage - data[3].ReAchievedPercentage) * 100)
		}
	} else if (getPercentage(data[2], data[1]) > 0 && getPercentage(data[1], data[0]) > 0 && getPercentage(data[3], data[2]) > 0) ||
		(getPercentage(data[2], data[1]) < 0 && getPercentage(data[1], data[0]) < 0 && getPercentage(data[3], data[2]) < 0) {
		if getPercentage(data[2], data[1]) > 0 && getPercentage(data[3], data[2]) > 0 && getPercentage(data[1], data[0]) > 0 {
			labelID = entity.LOIncrease3w
			labelParams.LOCompareLast3Week = math.Ceil(getPercentage(data[3], data[0]))
		} else {
			labelID = entity.LODecrease3w
			labelParams.LOCompareLast3Week = math.Ceil(getPercentage(data[0], data[3]))
		}
	} else if getSum(data[3]) > data[3].ClassAverageAchievedPercentage*100 || getSum(data[3]) < data[3].ClassAverageAchievedPercentage*100 {
		if getSum(data[3]) > data[3].ClassAverageAchievedPercentage*100 {
			labelID = entity.LOHighClassW
			labelParams.LOCompareClass = math.Ceil(getSum(data[3]) - data[3].ClassAverageAchievedPercentage*100)
		} else {
			labelID = entity.LOLowClassW
			labelParams.LOCompareClass = math.Ceil(data[3].ClassAverageAchievedPercentage*100 - getSum(data[3]))
		}
	} else if (getPercentage(data[3], data[2]) < 20 && getPercentage(data[3], data[2]) > 0) ||
		(getPercentage(data[2], data[3]) < 20 && getPercentage(data[2], data[3]) > 0) {
		if getPercentage(data[3], data[2]) < 20 && getPercentage(data[3], data[2]) > 0 {
			labelID = entity.LOIncreasePreviousW
			labelParams.LOCompareLastWeek = math.Ceil(getPercentage(data[3], data[2]))
		} else {
			labelID = entity.LODecreasePreviousW
			labelParams.LOCompareLastWeek = math.Ceil(getPercentage(data[2], data[3]))
		}
	} else {
		labelID = entity.LODefault
		labelParams.AchievedLoCount = data[3].FirstAchievedCount + data[3].ReAchievedCount
		labelParams.LearntLoCount = data[3].FirstAchievedCount + data[3].ReAchievedCount + data[3].UnAchievedCount
	}
	return
}

func getPercentage(data1, data2 *entity.LearnOutcomeAchievementResponseItem) (result float64) {
	return (data1.FirstAchievedPercentage + data1.ReAchievedPercentage -
		(data2.FirstAchievedPercentage + data2.ReAchievedPercentage)) * 100
}

func getSum(data *entity.LearnOutcomeAchievementResponseItem) (result float64) {
	return (data.FirstAchievedPercentage + data.ReAchievedPercentage) * 100
}

func getResult(data1, data2, data3 *entity.LearnOutcomeAchievementResponseItem) (result float64) {
	return (data1.FirstAchievedPercentage + data1.ReAchievedPercentage -
		data1.ClassAverageAchievedPercentage +
		(data2.FirstAchievedPercentage +
			data2.ReAchievedPercentage -
			data2.ClassAverageAchievedPercentage) +
		(data3.FirstAchievedPercentage +
			data3.ReAchievedPercentage -
			data3.ClassAverageAchievedPercentage)) * 100
}

func getAdverseResult(data1, data2, data3 *entity.LearnOutcomeAchievementResponseItem) (result float64) {
	return (data1.ClassAverageAchievedPercentage -
		(data1.FirstAchievedPercentage + data1.ReAchievedPercentage) +
		(data2.ClassAverageAchievedPercentage -
			(data2.FirstAchievedPercentage + data2.ReAchievedPercentage)) +
		(data3.ClassAverageAchievedPercentage -
			(data3.FirstAchievedPercentage + data3.ReAchievedPercentage))) * 100
}
