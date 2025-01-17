package da

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/go-sql-driver/mysql"
)

type DataAccessor interface {
	dbo.DataAccesser
	BatchInsert(ctx context.Context, models ...entity.BatchInsertModeler) (err error)
	BatchInsertTx(ctx context.Context, tx *dbo.DBContext, models ...entity.BatchInsertModeler) (err error)
}
type BaseDA struct {
	dbo.BaseDA
}

func (BaseDA) getPlaceHolder(length int) string {
	return strings.TrimRight(strings.Repeat("?,", length), ",")
}

func (s BaseDA) BatchInsert(ctx context.Context, models ...entity.BatchInsertModeler) (err error) {
	db, err := dbo.GetDB(ctx)
	if err != nil {
		return err
	}

	return s.BatchInsertTx(ctx, db, models...)

}
func (s BaseDA) BatchInsertTx(ctx context.Context, tx *dbo.DBContext, models ...entity.BatchInsertModeler) (err error) {
	if len(models) < 1 {
		return
	}
	value := models[0]
	tbName := tx.GetTableName(value)

	var insertCols []string
	var insertValues []interface{}
	rowCount := 0

	for _, v := range models {
		cols, values := v.GetBatchInsertColsAndValues()
		insertCols = cols
		insertValues = append(insertValues, values...)
		rowCount++
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("insert into %s(%s) values ", tbName, strings.Join(insertCols, ",")))
	rowPlaceHolder := "(" + strings.TrimRight(strings.Repeat("?,", len(insertCols)), ",") + "),"
	placeHolder := strings.TrimRight(strings.Repeat(rowPlaceHolder, rowCount), ",")
	sb.WriteString(placeHolder)

	start := time.Now()
	sql := sb.String()

	err = tx.Exec(sql, insertValues...).Error
	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if ok && me.Number == 1062 {
			log.Error(ctx, "insert duplicate record",
				log.Err(me),
				log.String("tableName", tbName),
				log.Any("value", value),
				log.Duration("duration", time.Since(start)))
			err = constant.ErrDuplicateRecord
			return
		}

		log.Error(ctx, "BatchInsertTx failed",
			log.Err(err),
			log.String("tableName", tbName),
			log.Any("models", models),
			log.Duration("duration", time.Since(start)))
		return
	}

	log.Debug(ctx, "BatchInsertTx success",
		log.String("tableName", tbName),
		log.Any("models", models),
		log.Duration("duration", time.Since(start)))

	return
}
