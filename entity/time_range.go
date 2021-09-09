package entity

import (
	"context"
	"strconv"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

// TimeRange start_at-end_at: eg: 1630918543-1630918555 means from 1630918543 to 1630918555
type TimeRange string

func (tr TimeRange) Value(ctx context.Context) (startAt, endAt int64, err error) {
	arr := strings.Split(string(tr), "-")
	if len(arr) < 2 {
		log.Error(ctx, "invalid param for TimeRange", log.Any("time_range", tr))
		err = constant.ErrInvalidArgs
		return
	}
	s := arr[0]
	startAt, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error(ctx, "invalid start_at for TimeRange", log.Err(err), log.Any("time_range", tr))
		err = constant.ErrInvalidArgs
		return
	}

	s = arr[1]
	endAt, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error(ctx, "invalid end_at for TimeRange", log.Err(err), log.Any("time_range", tr))
		err = constant.ErrInvalidArgs
		return
	}

	if startAt > endAt {
		log.Error(ctx, "invalid start_at and end_at for TimeRange", log.Err(err), log.Any("time_range", tr))
		err = constant.ErrInvalidArgs
		return
	}
	return
}

func (tr TimeRange) MustContain(ctx context.Context, value int64) bool {
	start, end, err := tr.Value(ctx)
	if err != nil {
		log.Panic(ctx, "MustBetween panic", log.Any("tr", tr), log.Int64("value", value))
	}
	return value >= start && value < end
}
