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

type unmigratableModel struct {
	ID int64
	Ch chan int
}

func TestNewDB_AutoMigrateFailureFailsTheTest(t *testing.T) {
	spy := &fatalSpyTB{TB: t}

	goninjatest.NewDB(spy, &unmigratableModel{})

	if !spy.fataled {
		t.Fatal("NewDB did not fail the test when AutoMigrate returned an error")
	}
}

// fatalSpyTB wraps a *testing.T so a nested testing.TB argument's Fatalf
// can be observed without actually aborting the outer test.
type fatalSpyTB struct {
	testing.TB
	fataled bool
}

func (s *fatalSpyTB) Fatalf(format string, args ...any) {
	s.fataled = true
	s.Logf(format, args...)
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
