package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetStudentProgressLearnOutcomeAchievement(ctx context.Context, op *entity.Operator, req *entity.LearnOutcomeAchievementRequest) (res *entity.LearnOutcomeAchievementResponse, err error) {
	res = &entity.LearnOutcomeAchievementResponse{
		Request:                               req,
		StudentAchievedCounts:                 map[string]entity.Float64Slice{},
		UnselectSubjectsStudentAchievedCounts: map[string]entity.Float64Slice{},
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
				res.UnselectSubjectsStudentAchievedCounts[c.StudentID] = append(res.UnselectSubjectsStudentAchievedCounts[c.StudentID], float64(c.AchievedCount))
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

	unselectSubjectsClassAchievedCount := entity.Float64Slice{}
	for _, s := range res.UnselectSubjectsStudentAchievedCounts {
		unselectSubjectsClassAchievedCount = append(unselectSubjectsClassAchievedCount, s.Sum())
	}
	res.UnSelectedSubjectsAverageAchieveCount = unselectSubjectsClassAchievedCount.Avg()

	return
}
