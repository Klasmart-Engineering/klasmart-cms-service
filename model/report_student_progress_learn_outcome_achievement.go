package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

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
	labelID, labelParams, err := getLabelIDAndParams(res)
	if err != nil {
		return
	}
	res.LabelID = labelID
	res.LabelParams = labelParams
	return
}

func getLabelIDAndParams(res *entity.LearnOutcomeAchievementResponse) (labelID string, labelParams string, err error) {
	data := res.Items
	if data[0].ClassAverageAchievedPercentage*100 == 0 && data[0].FirstAchievedPercentage*100 == 0 && data[0].ReAchievedPercentage*100 == 0 &&
		data[0].UnSelectedSubjectsAverageAchievedPercentage*100 == 0 &&
		data[1].ClassAverageAchievedPercentage*100 == 0 && data[1].FirstAchievedPercentage*100 == 0 && data[1].ReAchievedPercentage*100 == 0 &&
		data[1].UnSelectedSubjectsAverageAchievedPercentage*100 == 0 &&
		data[2].ClassAverageAchievedPercentage*100 == 0 && data[2].FirstAchievedPercentage*100 == 0 && data[2].ReAchievedPercentage*100 == 0 &&
		data[2].UnSelectedSubjectsAverageAchievedPercentage*100 == 0 {
		labelID = constant.LO_new
		labelParams = ""
	} else if ((data[1].FirstAchievedPercentage+data[1].ReAchievedPercentage) > data[1].ClassAverageAchievedPercentage &&
		(data[2].FirstAchievedPercentage+data[2].ReAchievedPercentage) > data[2].ClassAverageAchievedPercentage &&
		(data[3].FirstAchievedPercentage+data[3].ReAchievedPercentage) > data[3].ClassAverageAchievedPercentage) ||
		((data[1].FirstAchievedPercentage+data[1].ReAchievedPercentage) < data[1].ClassAverageAchievedPercentage &&
			(data[2].FirstAchievedPercentage+data[2].ReAchievedPercentage) < data[2].ClassAverageAchievedPercentage &&
			(data[3].FirstAchievedPercentage+data[3].ReAchievedPercentage) < data[3].ClassAverageAchievedPercentage) {
		if (data[1].FirstAchievedPercentage+data[1].ReAchievedPercentage) > data[1].ClassAverageAchievedPercentage &&
			(data[2].FirstAchievedPercentage+data[2].ReAchievedPercentage) > data[2].ClassAverageAchievedPercentage &&
			(data[3].FirstAchievedPercentage+data[3].ReAchievedPercentage) > data[3].ClassAverageAchievedPercentage {
			labelID = constant.LO_high_class_3w
			labelParams = ""
		} else {
			labelID = constant.LO_low_class_3w
			labelParams = ""
		}

	} else if getPercentage(data[3], data[2]) >= 20 || getPercentage(data[2], data[3]) >= 20 {
		if getPercentage(data[3], data[2]) >= 20 {
			labelID = constant.LO_increase_previous_large_w
			labelParams = ""
		} else {
			labelID = constant.LO_decrease_previous_large_w
			labelParams = ""
		}
	}
	return
}

func getPercentage(data1, data2 *entity.LearnOutcomeAchievementResponseItem) (result float64) {
	return data1.FirstAchievedPercentage*100 + data1.ReAchievedPercentage*100 -
		(data2.FirstAchievedPercentage*100 + data2.ReAchievedPercentage*100)
}
