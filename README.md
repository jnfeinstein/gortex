# GORTEX

Text search for postgres w/ ORGM using Golang

## Overview

* Based on [Textacular](https://github.com/textacular/textacular)
* Utilizes [gorm](https://github.com/jinzhu/gorm)
* Creates scopes that, in conjunction with gorm, find records and scan them into Go structs

## Usage
```go
gortex.NewSearchScope(language string, exclusive bool, clause {}interface)
```

```language string``` Specify [language](http://blog.lostpropertyhq.com/postgres-full-text-search-is-good-enough/#languagesupport) for search

```exclusive bool``` Controls whether search fields should be AND'd or OR'd

```clause {}interface``` May be a ```struct``` or ```map[string]{}interface``` which specifies the fields and values to search

### Example setup
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

### Magic
```go
searchScope1 := gortex.NewSearchScope("english", true, Note{Contents: "brown"})
DB.Scopes(searchScope1).Find(&notes)
//// Finds all records with 'brown' in the contents field

searchScope2 := gortex.NewSearchScope("english", false, map[]{}interface{"contents": "brown", "author": "dog"})
DB.Scopes(searchScope2).Find(&notes)
//// Finds all records with 'brown' in the contents field OR 'dog' in the author field
```
