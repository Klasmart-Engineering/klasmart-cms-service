package model

import (
	"testing"
	"time"

	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/stretchr/testify/assert"
)

func TestLearningSummaryReportModel_QueryTimeFilter(t *testing.T) {
	l := learningSummaryReportModel{}
	fakeLoc := time.FixedZone("fake_loc_0800", 3600*8)
	fakeNow := time.Date(2050, 1, 1, 12, 23, 56, 0, fakeLoc)
	ret := l.getYearsWeeksData(fakeNow)

	testDataTable := []struct {
		actualData                      *entity.LearningSummaryFilterYear
		yearExpect                      int
		firstWeekExpect, lastWeekExpect entity.LearningSummaryFilterWeek
	}{
		{
			actualData: ret[0],
			yearExpect: 2049,
			firstWeekExpect: entity.LearningSummaryFilterWeek{
				WeekStart: time.Date(2048, 12, 28, 0, 0, 0, 0, fakeLoc).Unix(),
				WeekEnd:   time.Date(2049, 1, 4, 0, 0, 0, 0, fakeLoc).Unix(),
			},
			lastWeekExpect: entity.LearningSummaryFilterWeek{
				WeekStart: time.Date(2049, 12, 20, 0, 0, 0, 0, fakeLoc).Unix(),
				WeekEnd:   time.Date(2049, 12, 27, 0, 0, 0, 0, fakeLoc).Unix(),
			},
		}, {
			actualData: ret[10],
			yearExpect: 2039,
			firstWeekExpect: entity.LearningSummaryFilterWeek{
				WeekStart: time.Date(2038, 12, 27, 0, 0, 0, 0, fakeLoc).Unix(),
				WeekEnd:   time.Date(2039, 1, 3, 0, 0, 0, 0, fakeLoc).Unix(),
			},
			lastWeekExpect: entity.LearningSummaryFilterWeek{
				WeekStart: time.Date(2039, 12, 19, 0, 0, 0, 0, fakeLoc).Unix(),
				WeekEnd:   time.Date(2039, 12, 26, 0, 0, 0, 0, fakeLoc).Unix(),
			},
		}, {
			actualData: ret[29],
			yearExpect: 2020,
			firstWeekExpect: entity.LearningSummaryFilterWeek{
				WeekStart: time.Date(2019, 12, 30, 0, 0, 0, 0, fakeLoc).Unix(),
				WeekEnd:   time.Date(2020, 1, 6, 0, 0, 0, 0, fakeLoc).Unix(),
			},
			lastWeekExpect: entity.LearningSummaryFilterWeek{
				WeekStart: time.Date(2020, 12, 21, 0, 0, 0, 0, fakeLoc).Unix(),
				WeekEnd:   time.Date(2020, 12, 28, 0, 0, 0, 0, fakeLoc).Unix(),
			},
		},
	}

	for _, testData := range testDataTable {
		assertOneTimeFilter(t, testData.actualData, testData.yearExpect, testData.firstWeekExpect, testData.lastWeekExpect)
	}
}

func assertOneTimeFilter(t *testing.T, year *entity.LearningSummaryFilterYear,
	yearExpect int, firstWeekExpect, lastWeekExpect entity.LearningSummaryFilterWeek) {
	t.Logf("asserting year %d", yearExpect)
	assert.Equal(t, yearExpect, year.Year, "year assertion failed")
	weeks := year.Weeks
	lastWeek, firstWeek := weeks[0], weeks[len(weeks)-1]
	assert.Equalf(t, firstWeekExpect, firstWeek, "asserting first week failed on year %d", year)
	assert.Equalf(t, lastWeekExpect, lastWeek, "asserting last week failed on year %d", year)
}
