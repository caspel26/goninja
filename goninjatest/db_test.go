package goninjatest_test

import (
	"testing"

	"github.com/caspel26/goninja/goninjatest"
)

type dbTestRow struct {
	ID   uint
	Name string
}

func TestNewDB_AutoMigratesGivenModels(t *testing.T) {
	db := goninjatest.NewDB(t, &dbTestRow{})

	if err := db.Create(&dbTestRow{Name: "hello"}).Error; err != nil {
		t.Fatalf("Create on an automigrated table: %v", err)
	}

	var count int64
	db.Model(&dbTestRow{}).Count(&count)
	if count != 1 {
		t.Errorf("row count = %d, want 1", count)
	}
}

func TestNewDB_NoModelsStillReturnsUsableDB(t *testing.T) {
	db := goninjatest.NewDB(t)

	if err := db.AutoMigrate(&dbTestRow{}); err != nil {
		t.Errorf("NewDB() with no models is not a usable *gorm.DB: %v", err)
	}
}

func TestNewDB_TwoCallsAreIndependent(t *testing.T) {
	a := goninjatest.NewDB(t, &dbTestRow{})
	b := goninjatest.NewDB(t, &dbTestRow{})

	if err := a.Create(&dbTestRow{Name: "only in a"}).Error; err != nil {
		t.Fatalf("Create on a: %v", err)
	}

	var count int64
	b.Model(&dbTestRow{}).Count(&count)
	if count != 0 {
		t.Errorf("b sees %d rows from a, want 0 (separate in-memory databases)", count)
	}
}
