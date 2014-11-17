package gortex_test

import (
	"fmt"
	"github.com/jnfeinstein/gorm"
	"github.com/jnfeinstein/gortex"
	_ "github.com/lib/pq"
	"testing"
)

type Note struct {
	Id       int64
	Contents string
	Author   string
}

var DB gorm.DB

var foxNote Note = Note{Contents: "The quick brown fox jumped over the green dog.", Author: "Luke"}
var grassNote Note = Note{Contents: "The grass is always green on the other side.", Author: "Lando"}

func init() {
	var err error
	fmt.Println("testing postgres...")
	DB, err = gorm.Open("postgres", "dbname=gortex sslmode=disable")

	DB.LogMode(false)

	if err != nil {
		panic(fmt.Sprintf("No error should happen when connect database, but got %+v", err))
	}

	DB.DB().SetMaxIdleConns(10)

	setupDB()
}

func setupDB() {
	DB.DropTableIfExists(Note{})
	DB.AutoMigrate(Note{})
	DB.Delete(Note{})
	DB.Save(&foxNote).Save(&grassNote)
}

func TestSingle(t *testing.T) {
	searchScope := gortex.NewSearchScope(Note{Contents: "brown"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 1 {
		t.Errorf("Should have found one note")
	}
}

func BenchmarkSingle(b *testing.B) {
	searchScope := gortex.NewSearchScope(Note{Contents: "brown"})

	var notes []Note
	for i := 0; i < b.N; i++ {
		DB.Scopes(searchScope).Select("*").Find(&notes)
	}
}

func TestMultiple(t *testing.T) {
	searchScope := gortex.NewSearchScope(Note{Contents: "green"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 2 {
		t.Errorf("Should have found two notes")
	}
}

func TestExclusive(t *testing.T) {
	searchScope := gortex.NewSearchScope(Note{Contents: "green", Author: "Luke"},
		map[string]interface{}{"exclusive": true})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 1 {
		t.Errorf("Should have found one note")
	}
}

func TestNotExclusive(t *testing.T) {
	searchScope := gortex.NewSearchScope(Note{Contents: "green", Author: "Luke"},
		map[string]interface{}{"exclusive": false})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 2 {
		t.Errorf("Should have found two notes")
	}
}

func TestAdvancedSearch(t *testing.T) {
	searchScope := gortex.NewSearchScope(Note{Contents: "!brown"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 1 {
		t.Errorf("Should have found one note")
	}
}

func TestFuzzy(t *testing.T) {
	gortex.InitFuzzySearch(&DB)
	gortex.SetFuzzySearchLimit(&DB, 0.05)
	searchScope := gortex.NewFuzzySearchScope(Note{Contents: "gre"}, map[string]interface{}{"limit": 0.05})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 2 {
		t.Errorf("Should have found two notes")
	}
}

type customSearchFormat struct{}

func (c customSearchFormat) Rank(field string, opts map[string]interface{}) string {
	return fmt.Sprintf("LENGTH(%s)", field)
}

func (c customSearchFormat) Condition(field string, opts map[string]interface{}) string {
	return fmt.Sprintf("%s LIKE ?", field)
}

func TestCustomSearch(t *testing.T) {
	searchScope := gortex.NewCustomSearchScope(customSearchFormat{}, Note{Author: "L%"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 2 {
		t.Errorf("Should have found two notes")
	}
}
