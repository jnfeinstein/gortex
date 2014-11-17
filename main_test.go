package gortex_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"gortex"
	"testing"
)

type Note struct {
	Id       int64
	Contents string
	Author   string
}

var DB gorm.DB

var foxNote Note = Note{Contents: "The quick brown fox jumped over the green dog.", Author: "Luke"}
var grassNote Note = Note{Contents: "The grass is always green on the other side.", Author: "Leah"}

func init() {
	var err error
	fmt.Println("testing postgres...")
	DB, err = gorm.Open("postgres", "dbname=gortex sslmode=disable")

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
	searchScope := gortex.NewSearchScope("english", true, Note{Contents: "brown"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 1 {
		t.Errorf("Should have found one note")
	}
}

func TestMultiple(t *testing.T) {
	searchScope := gortex.NewSearchScope("english", true, Note{Contents: "green"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 2 {
		t.Errorf("Should have found two notes")
	}
}

func TestExclusive(t *testing.T) {
	searchScope := gortex.NewSearchScope("english", true, Note{Contents: "green", Author: "Luke"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 1 {
		t.Errorf("Should have found one note")
	}
}

func TestNotExclusive(t *testing.T) {
	searchScope := gortex.NewSearchScope("english", false, Note{Contents: "grass", Author: "Luke"})

	var notes []Note
	if err := DB.Scopes(searchScope).Select("*").Find(&notes).Error; err != nil {
		t.Errorf("Encountered DB error - %s", err.Error())
	} else if len(notes) != 2 {
		t.Errorf("Should have found two notes")
	}
}
