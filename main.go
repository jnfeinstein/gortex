package gortex

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"regexp"
	"strings"
)

type SearchSqlFormat interface {
	Rank(field string, opts map[string]interface{}) string
	Condition(field string, opts map[string]interface{}) string
}

type normalSearchFmt struct{}

func getLanguage(opts map[string]interface{}) (language string) {
	language = "simple"
	if l, ok := opts["language"].(string); ok {
		language = l
	}
	return
}

func getExclusive(opts map[string]interface{}) (exclusive bool) {
	exclusive = true
	if e, ok := opts["exclusive"].(bool); ok {
		exclusive = e
	}
	return
}

func getLimit(opts map[string]interface{}) (limit float64) {
	limit = -1
	if val, ok := opts["limit"]; ok {
		switch l := val.(type) {
		case float32:
			limit = float64(l)
		case float64:
			limit = l
		}
	}
	return
}

func (n normalSearchFmt) Rank(field string, opts map[string]interface{}) string {
	language := getLanguage(opts)
	return fmt.Sprintf("COALESCE(ts_rank(to_tsvector('%v', %v), to_tsquery('%v', ?)), 0)", language, field, language)
}

func (n normalSearchFmt) Condition(field string, opts map[string]interface{}) string {
	language := getLanguage(opts)
	return fmt.Sprintf("to_tsvector('%v', %v) @@ to_tsquery('%v', ?)", language, field, language)
}

type fuzzySearchFmt struct{}

func (f fuzzySearchFmt) Rank(field string, opts map[string]interface{}) string {
	return fmt.Sprintf("similarity(%v, ?)", field)
}

func (f fuzzySearchFmt) Condition(field string, opts map[string]interface{}) string {
	return fmt.Sprintf("(%v %% ?)", field)
}

func makeSearchScope(sqlFmt SearchSqlFormat, clause interface{}, o ...map[string]interface{}) func(*gorm.DB) *gorm.DB {
	var opts map[string]interface{}
	if len(o) == 0 {
		opts = map[string]interface{}{
			"language":  "simple",
			"exclusive": true,
		}
	} else {
		opts = o[0]
	}

	exclusive := getExclusive(opts)

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
			rankQuery := sqlFmt.Rank(scope.Quote(key), opts)
			rankSqls = append(rankSqls, rankQuery)
			if regexp.MustCompile("\\?").MatchString(rankQuery) {
				rankTerms = append(rankTerms, val)
			}

			whereQuery := sqlFmt.Condition(scope.Quote(key), opts)
			if regexp.MustCompile("\\?").MatchString(whereQuery) {
				if exclusive {
					scopeDB = scopeDB.Where(whereQuery, val)
				} else {
					scopeDB = scopeDB.Or(whereQuery, val)
				}
			} else {
				if exclusive {
					scopeDB = scopeDB.Where(whereQuery)
				} else {
					scopeDB = scopeDB.Or(whereQuery)
				}
			}
		}

		rankStmt := fmt.Sprintf("%s AS %s", strings.Join(rankSqls, "+"), scope.Quote("rank"))
		scopeDB = scopeDB.Select(rankStmt, rankTerms...)

		orderStmt := fmt.Sprintf("%s desc", scope.Quote("rank"))
		scopeDB = scopeDB.Order(orderStmt)

		return
	}
}

func NewSearchScope(clause interface{}, opts ...map[string]interface{}) func(*gorm.DB) *gorm.DB {
	return makeSearchScope(normalSearchFmt{}, clause, opts...)
}

func NewFuzzySearchScope(clause interface{}, opts ...map[string]interface{}) func(*gorm.DB) *gorm.DB {
	return makeSearchScope(fuzzySearchFmt{}, clause, opts...)
}

func InitFuzzySearch(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION pg_trgm").Error
}

func SetFuzzySearchLimit(db *gorm.DB, limit float64) error {
	return db.Exec("SELECT SET_LIMIT(?)", limit).Error
}

func NewCustomSearchScope(format SearchSqlFormat, clause interface{}, opts ...map[string]interface{}) func(*gorm.DB) *gorm.DB {
	return makeSearchScope(format, clause, opts...)
}
