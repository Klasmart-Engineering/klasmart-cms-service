package da

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-playground/assert/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/ro"
)

func TestSet(t *testing.T) {
	ctx := context.TODO()
	cond := &ScheduleCacheCondition{
		UserID:     "user",
		ScheduleID: "schedule",
		SchoolID:   "school2",
	}
	data := []*entity.ScheduleListView{
		{
			ID:    "id 123",
			Title: "title 123",
		},
	}
	err := GetScheduleRedisDA().Set(ctx, "org123", cond, data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetScheduleListView(t *testing.T) {
	ctx := context.TODO()
	orgID := "org_test_1"
	cacheCondition := &ScheduleCacheCondition{
		Condition: &ScheduleCondition{
			OrgID: sql.NullString{
				String: orgID,
				Valid:  true,
			},
		},
		DataType: ScheduleListView,
	}

	data := []*entity.ScheduleListView{
		{
			ID:    "id_test_1",
			Title: "title_test_1",
		},
	}
	err := GetScheduleRedisDA().Set(ctx, orgID, cacheCondition, data)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetScheduleRedisDA().GetScheduleListView(ctx, orgID, &ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, result, data)
}

func TestGetScheduleListViewNotExist(t *testing.T) {
	ctx := context.TODO()
	orgID := "org_test_none"

	result, err := GetScheduleRedisDA().GetScheduleListView(ctx, orgID, &ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	})
	assert.Equal(t, err, ro.ErrKeyNotExist)
	assert.Equal(t, result, nil)
}

func TestGetScheduledDates(t *testing.T) {
	ctx := context.TODO()
	orgID := "org_test_1"
	cacheCondition := &ScheduleCacheCondition{
		Condition: &ScheduleCondition{
			OrgID: sql.NullString{
				String: orgID,
				Valid:  true,
			},
		},
		DataType: ScheduledDates,
	}

	data := []string{
		"test_1",
		"test_2",
		"test_3",
	}
	err := GetScheduleRedisDA().Set(ctx, orgID, cacheCondition, data)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetScheduleRedisDA().GetScheduledDates(ctx, orgID, &ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, result, data)
}

func TestGetScheduleBasic(t *testing.T) {
	ctx := context.TODO()
	orgID := "org_test_1"
	scheduleID := "schedule_test_1"
	cacheCondition := &ScheduleCacheCondition{
		ScheduleID: scheduleID,
		DataType:   ScheduleBasic,
	}

	data := &entity.ScheduleBasic{
		StudentCount: 66,
	}
	err := GetScheduleRedisDA().Set(ctx, orgID, cacheCondition, data)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetScheduleRedisDA().GetScheduleBasic(ctx, orgID, scheduleID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, result, data)
}

func TestGetScheduleDetailView(t *testing.T) {
	ctx := context.TODO()
	orgID := "org_test_1"
	userID := "user_test_1"
	scheduleID := "schedule_test_1"
	cacheCondition := &ScheduleCacheCondition{
		UserID:     userID,
		ScheduleID: scheduleID,
		DataType:   ScheduleDetailView,
	}

	data := &entity.ScheduleDetailsView{
		ID:    "id_test_1",
		Title: "title_test_1",
	}
	err := GetScheduleRedisDA().Set(ctx, orgID, cacheCondition, data)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetScheduleRedisDA().GetScheduleDetailView(ctx, orgID, userID, scheduleID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, result, data)
}

func TestGetScheduleFilterUndefinedClass(t *testing.T) {
	ctx := context.TODO()
	orgID := "org_test_1"
	permissionMap := make(map[external.PermissionName]bool)
	permissionMap[external.ScheduleViewOrgCalendar] = true
	permissionMap[external.ScheduleViewMyCalendar] = false
	cacheCondition := &ScheduleCacheCondition{
		PermissionMap: permissionMap,
		DataType:      ScheduleFilterUndefinedClass,
	}

	data := true
	err := GetScheduleRedisDA().Set(ctx, orgID, cacheCondition, data)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetScheduleRedisDA().GetScheduleFilterUndefinedClass(ctx, orgID, permissionMap)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, result, data)
}

func TestClean(t *testing.T) {
	ctx := context.TODO()
	orgID := "org_test_1"
	cacheCondition := &ScheduleCacheCondition{
		Condition: &ScheduleCondition{
			OrgID: sql.NullString{
				String: orgID,
				Valid:  true,
			},
		},
		DataType: ScheduleListView,
	}

	data := []*entity.ScheduleListView{
		{
			ID:    "id_test_1",
			Title: "title_test_1",
		},
	}
	err := GetScheduleRedisDA().Set(ctx, orgID, cacheCondition, data)
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetScheduleRedisDA().GetScheduleListView(ctx, orgID, &ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	})
	assert.Equal(t, result, data)

	err = GetScheduleRedisDA().Clean(ctx, orgID)
	if err != nil {
		t.Fatal(err)
	}

	result, err = GetScheduleRedisDA().GetScheduleListView(ctx, orgID, &ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	})

	assert.Equal(t, err, ro.ErrKeyNotExist)
	assert.Equal(t, result, nil)
}
