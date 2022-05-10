package da

import (
	"bytes"
	"context"
	"regexp"
	"text/template"

	"github.com/KL-Engineering/common-log/log"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
)

type sqlBuilder struct {
	Sql          string
	PlaceHolders map[string]*sqlBuilder
	Args         []interface{}
}

func (sb *sqlBuilder) Build(ctx context.Context) (sql string, args []interface{}, err error) {
	tp, err := template.New("").Parse(sb.Sql)
	if err != nil {
		log.Error(ctx, "sqlBuilder build failed", log.Err(err), log.Any("sb", sb))
		err = constant.ErrSqlBuilderFailed
		return
	}

	data := map[string]string{}
	args = sb.Args

	reg := regexp.MustCompile(`\{\{\.([A-Za-z0-9_]+)\}\}`)
	plKeys := reg.FindAllStringSubmatch(sb.Sql, -1)
	if len(plKeys) > 0 && len(sb.Args) > 0 {
		log.Error(ctx, "plKeys and Args cannot have items in the same time",
			log.Any("plKeys", plKeys),
			log.Any("Args", sb.Args),
			log.Any("sql", sb.Sql),
		)
		err = constant.ErrSqlBuilderFailed
	}
	for _, subPlKeys := range plKeys {
		if len(subPlKeys) < 2 {
			continue
		}
		plKey := subPlKeys[1]
		sb1, ok := sb.PlaceHolders[plKey]
		if !ok {
			log.Error(ctx, "sqlBuilder build placeHolderKey not found", log.Any("plKey", plKey), log.Err(err), log.Any("sb", sb))
			err = constant.ErrSqlBuilderFailed
			return
		}
		var args1 []interface{}
		data[plKey], args1, err = sb1.Build(ctx)
		if err != nil {
			log.Error(ctx, "sqlBuilder build failed", log.Err(err), log.Any("sb", sb1), log.Any("data", data))
			return
		}
		args = append(args, args1...)
	}
	bf := new(bytes.Buffer)
	err = tp.Execute(bf, data)
	if err != nil {
		log.Error(ctx, "sqlBuilder build Execute failed", log.Err(err), log.Any("sb", sb), log.Any("data", data))
		err = constant.ErrSqlBuilderFailed
		return
	}
	sql = bf.String()
	return
}

func NewSqlBuilder(ctx context.Context, sql string, args ...interface{}) (sb *sqlBuilder) {
	sb = &sqlBuilder{
		Sql:          sql,
		Args:         args,
		PlaceHolders: map[string]*sqlBuilder{},
	}
	return
}
func (sb *sqlBuilder) Replace(ctx context.Context, placeHolderKey string, sb1 *sqlBuilder) *sqlBuilder {
	sb.PlaceHolders[placeHolderKey] = sb1
	return sb
}
