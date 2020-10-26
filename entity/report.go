package entity

type StudentReportList struct {
	Items []StudentReportItem `json:"items"`
}

type StudentReportItem struct {
	StudentName       string `json:"student_name"`
	AllAchievedCount  int    `json:"all_achieved_count"`
	NotAchievedCount  int    `json:"not_achieved_count"`
	NotAttemptedCount int    `json:"not_attempted_count"`
}

type StudentReportDetail struct {
	Categories []StudentReportCategory `json:"categories"`
}

type StudentReportCategory struct {
	CategoryName      ReportCategory `json:"category_name"`
	AllAchievedCount  int            `json:"all_achieved_count"`
	NotAchievedCount  int            `json:"not_achieved_count"`
	NotAttemptedCount int            `json:"not_attempted_count"`
}

type ReportCategory string

const (
	ReportCategorySpeechLanguagesSkills     = "Speech & Language Skills"
	ReportCategoryFineMotorSkills           = "Fine Motor Skills"
	ReportCategoryGrossMotorSkills          = "Gross Motor Skills"
	ReportCategoryCognitiveSkills           = "Cognitive Skills"
	ReportCategoryPersonalDevelopment       = "Personal Development"
	ReportCategoryLanguageAndNumeracySkills = "Language and Numeracy Skills"
	ReportCategorySocialAndEmotional        = "Social and Emotional"
	ReportCategoryOral                      = "Oral"
	ReportCategoryLiteracy                  = "Literacy"
	ReportCategoryWholeChild                = "Whole-Child"
	ReportCategoryKnowledge                 = "Knowledge"
)

var reportCategoryOrderMap = map[ReportCategory]int{
	ReportCategorySpeechLanguagesSkills:     0,
	ReportCategoryFineMotorSkills:           1,
	ReportCategoryGrossMotorSkills:          2,
	ReportCategoryCognitiveSkills:           3,
	ReportCategoryPersonalDevelopment:       4,
	ReportCategoryLanguageAndNumeracySkills: 5,
	ReportCategorySocialAndEmotional:        6,
	ReportCategoryOral:                      7,
	ReportCategoryLiteracy:                  8,
	ReportCategoryWholeChild:                9,
	ReportCategoryKnowledge:                 10,
}

func ReportCategoryOrder(category ReportCategory) int {
	return reportCategoryOrderMap[category]
}
