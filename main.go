package gortex

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

const rankFmt string = "COALESCE(ts_rank(to_tsvector('%v', %v), to_tsquery('%v', ?)), 0)"
const whereFmt string = "to_tsvector('%v', %v) @@ to_tsquery('%v', ?)"

func NewSearchScope(language string, exclusive bool, clause interface{}) (f func(*gorm.DB) *gorm.DB) {
	return func(db *gorm.DB) (scopeDB *gorm.DB) {
		scopeDB = db
		scope := db.NewScope("")

		var searchTerms map[string]interface{} = make(map[string]interface{})

		switch value := clause.(type) {
		case map[string]interface{}:
			searchTerms = value
		case interface{}:
			newScope := scopeDB.NewScope(value)
			scopeDB = scopeDB.Table(newScope.TableName())
			for _, field := range newScope.Fields() {
				if !field.IsBlank {
					searchTerms[field.DBName] = field.Field.Interface()
				}
			}
		}

		if len(searchTerms) == 0 {
			return scopeDB
		}

		var rankTerms []interface{}
		var rankSqls []string

		for key, val := range searchTerms {
			rankSqls = append(rankSqls, fmt.Sprintf(rankFmt, language, scope.Quote(key), language))
			rankTerms = append(rankTerms, val)

			whereQuery := fmt.Sprintf(whereFmt, language, scope.Quote(key), language)
			if exclusive {
				scopeDB = scopeDB.Where(whereQuery, val)
			} else {
				scopeDB = scopeDB.Or(whereQuery, val)
			}
		}

		rankStmt := fmt.Sprintf("%s AS %s", strings.Join(rankSqls, "+"), scope.Quote("rank"))
		scopeDB = scopeDB.Select(rankStmt, rankTerms...)

		orderStmt := fmt.Sprintf("%s desc", scope.Quote("rank"))
		scopeDB = scopeDB.Order(orderStmt)

		return
	}
}
