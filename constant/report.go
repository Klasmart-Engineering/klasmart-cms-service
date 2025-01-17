package constant

const ReportTeachingLoadDays = 7

type ReportClassType string

var (
	ReportClassTypeStudy   ReportClassType = "study"
	ReportClassTypeHomeFun ReportClassType = "home_fun"
)

type LearnerReportOverviewStatus string

const (
	LearnerReportOverviewStatusNoData LearnerReportOverviewStatus = "NoData"
)
