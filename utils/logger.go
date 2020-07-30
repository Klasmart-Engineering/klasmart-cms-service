package utils

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-cn/helper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Ctx create zap field with context.Context
func Ctx(ctx context.Context) zap.Field {
	badaCtx, ok := helper.GetBadaCtx(ctx)
	if !ok {
		return zap.Field{
			Type: zapcore.SkipType,
		}
	}

	return zap.Field{
		Type:      zapcore.ObjectMarshalerType,
		Key:       "trace",
		Interface: TraceContext{badaCtx},
	}
}

type TraceContext struct {
	*helper.BadaCtx
}

func (c TraceContext) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("currTid", c.CurrTid)

	if c.PrevTid != "" {
		encoder.AddString("prevTid", c.PrevTid)
	}

	if c.EntryTid != "" {
		encoder.AddString("entryTid", c.EntryTid)
	}

	return nil
}
