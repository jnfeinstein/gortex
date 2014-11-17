# GORTEX

Text search for postgres w/ ORGM using Golang

## Overview

* Based on [Textacular](https://github.com/textacular/textacular)
* Utilizes [gorm](https://github.com/jinzhu/gorm)
* Creates scopes that, in conjunction with gorm, find records and scan them into Go structs
* Offers equivalent of Textacular advanced search
* Offers equivalent of Textacular fuzzy search
* You can even make your own types of search
* Generated scopes have no state so they can be used repeatedly

## Usage
```go
gortex.NewSearchScope(language string, exclusive bool, clause {}interface)
```
#### Params
```language string``` Specify [language](http://blog.lostpropertyhq.com/postgres-full-text-search-is-good-enough/#languagesupport) for search

```exclusive bool``` Controls whether search fields should be AND'd or OR'd

```clause {}interface``` May be a ```struct``` or ```map[string]{}interface``` which specifies the fields and values to search

Note: If ```clause``` is a ```struct```, it will automatically set the table to the struct's type

## Example setup
```go
import (
  "github.com/jinzhu/gorm"
  "github.com/jnfeinstein/gortex"
  _ "github.com/lib/pq"
)

type Note struct {
  Id       int64
  Contents string
  Author   string
}

DB, _ := gorm.Open("postgres", "dbname=gortex sslmode=disable")
DB.AutoMigrate(Note{})

var notes []Note
```

## Normal search
```go
searchScope1 := gortex.NewSearchScope("english", true, Note{Contents: "brown"})
DB.Scopes(searchScope1).Select("*").Find(&notes)
//// Finds all records with 'brown' in the contents field

searchScope2 := gortex.NewSearchScope("english", false, map[]{}interface{"contents": "brown", "author": "dog"})
DB.Scopes(searchScope2).Select("*").Find(&notes)
//// Finds all records with 'brown' in the contents field OR 'dog' in the author field
```
Note: You must include .Select("*"), which instructs gorm to select all fields in addition to the search rankings

## Fuzzy search
```go
gortex.InitFuzzySearch(&DB) //// Creats pg_trgm extension
gortex.SetFuzzySearchLimit(&DB, limit) //// Optional, sets fuzzy search limit (default is 0.1)
//// Only need to be run once

searchScope3 := gortex.NewFuzzySearchScope(Note{Contents: "or"})
DB.Scopes(searchScope3).Select("*").Find(&notes)
//// Finds all records with something like 'or' in the contents, maybe 'organic' or 'boring'
```

## Custom search (not for timid souls)
All gortex searches use a series of condition queries to determine qualification, and a series of select queries to determine rank.
These queries are defined by the following interface:
```go
type SearchSqlFormat interface {
  Rank(field string, opts map[string]interface{}) string
  Condition(field string, opts map[string]interface{}) string
}
```

#### Params
```field string```: Name of the field

```opts map[string]interface{}``` The very same map passed into ```gortex.NewCustomSearchScope```

#### Example
```go
type customSearchFormat struct{}

func (c customSearchFormat) Rank(field string, opts map[string]interface{}) string {
  return fmt.Sprintf("LENGTH(%s)", field)
}

func (c customSearchFormat) Condition(field string, opts map[string]interface{}) string {
  return fmt.Sprintf("%s LIKE ?", field)
}

searchScope4 := gortex.NewCustomSearchScope(customSearchFormat{}, Note{Author: "L%"})
DB.Scopes(searchScope4).Select("*").Find(&notes)
//// Finds all records with an author beginning with 'L' and ranks them by length of author
```

