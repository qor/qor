package validations_test

import (
	"database/sql"
	"testing"

	"github.com/JosephBuchma/qor/test/utils"
	"github.com/JosephBuchma/qor/validations"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db *gorm.DB

func init() {
	db = utils.TestDB()
	validations.RegisterCallbacks(db)
	db.AutoMigrate(&IntRangeTestModel{}, &FloatRangeTestModel{}, &PresentTestModel{}, &UniqueTestModel{}, &MultipleValidationsTestModel{})
	db.Exec("truncate int_range_test_models;")
	db.Exec("truncate float_range_test_models;")
	db.Exec("truncate present_test_models;")
	db.Exec("truncate unique_test_models;")
	db.Exec("truncate multiple_validations_test_models;")
}

// RangeInt validator

type IntRangeTestModel struct {
	gorm.Model
	Normal   int64
	Nullable sql.NullInt64
}

func (rc *IntRangeTestModel) Validate(db *gorm.DB) {
	validates := validations.NewValidator(db, rc)
	validates("Normal",
		validations.RangeInt(5, 50))
	validates("Nullable",
		validations.RangeInt(5, 50))
}

func TestRangeInt(t *testing.T) {
	ti := IntRangeTestModel{Normal: 33, Nullable: sql.NullInt64{Int64: 42, Valid: true}}
	if res := db.Save(&ti); res.Error != nil {
		t.Error("Should not get error when save ti record")
	}

	valid_with_null := IntRangeTestModel{Normal: 33, Nullable: sql.NullInt64{Int64: 0, Valid: false}}
	if res := db.Save(&valid_with_null); res.Error != nil {
		t.Error("Should not get error when save ti record with null value")
	}

	invalid := IntRangeTestModel{Normal: 1, Nullable: sql.NullInt64{Int64: 452, Valid: true}}
	if res := db.Save(&invalid); res.Error == nil {
		t.Error("Should get error when save invalid record with null value")
	}
}

// RangeFloat validator

type FloatRangeTestModel struct {
	gorm.Model
	Normal   float64
	Nullable sql.NullFloat64
}

func (rc *FloatRangeTestModel) Validate(db *gorm.DB) {
	validates := validations.NewValidator(db, rc)
	validates("Normal",
		validations.RangeFloat(5.0, 50.0))
	validates("Nullable",
		validations.RangeFloat(5.0, 50.0))
}

func TestRangeFloat(t *testing.T) {
	ti := FloatRangeTestModel{Normal: 33.0, Nullable: sql.NullFloat64{Float64: 42.0, Valid: true}}
	if res := db.Save(&ti); res.Error != nil {
		t.Error("Should not get error when save ti record")
	}

	valid_with_null := FloatRangeTestModel{Normal: 33.0, Nullable: sql.NullFloat64{Float64: 0.0, Valid: false}}
	if res := db.Save(&valid_with_null); res.Error != nil {
		t.Error("Should not get error when save valid record with null value")
	}

	invalid := FloatRangeTestModel{Normal: 1.3, Nullable: sql.NullFloat64{Float64: 452.0, Valid: true}}
	if res := db.Save(&invalid); res.Error == nil {
		t.Error("Should get error when save invalid record with null value")
	}
}

// Present validator

type PresentTestModel struct {
	gorm.Model
	Int    sql.NullInt64
	String sql.NullString
	Bool   sql.NullBool
}

func (rc *PresentTestModel) Validate(db *gorm.DB) {
	validates := validations.NewValidator(db, rc)
	validates("Int",
		validations.Present())
	validates("String",
		validations.Present())
	validates("Bool",
		validations.Present())
}

func TestPresent(t *testing.T) {
	empty := &PresentTestModel{}
	if res := db.Save(empty); res.Error == nil {
		t.Error("Should get error when save unitialized struct")
	}

	ti := &PresentTestModel{}
	ti.Int.Scan(24)
	ti.String.Scan("ti string")
	ti.Bool.Scan(true)
	if res := db.Save(ti); res.Error != nil {
		t.Error("Should not get error when save ti struct")
	}
}

// Unique validator

type UniqueTestModel struct {
	gorm.Model
	UniqueInt    int
	UniqueString sql.NullString
}

func (rc *UniqueTestModel) Validate(db *gorm.DB) {
	validates := validations.NewValidator(db, rc)
	validates("UniqueInt",
		validations.Unique())
	validates("UniqueString",
		validations.Unique())
}

func TestUnique(t *testing.T) {
	ti := &UniqueTestModel{}
	ti.UniqueInt = 3
	ti.UniqueString.Scan("unique string")
	db.Save(ti)
	ti.ID = 0
	if err := db.Save(ti); err == nil {
		t.Errorf("Should get error when save non-unique record")
	}
	ti.ID = 0
	ti.UniqueInt = 4
	ti.UniqueString.Scan("Another unique string")
	if res := db.Save(ti); res.Error != nil {
		t.Errorf("Should not get error when save unique record")
	}
	ti.ID = 0
	ti.UniqueInt = 5
	ti.UniqueString = sql.NullString{String: "", Valid: false}
	if res := db.Save(ti); res.Error != nil {
		t.Errorf("Should not get error when save unique record with null value")
	}
}

// Multivalidations

type MultipleValidationsTestModel struct {
	gorm.Model
	SomeString sql.NullString
}

func (rc *MultipleValidationsTestModel) Validate(db *gorm.DB) {
	validates := validations.NewValidator(db, rc)
	validates("SomeString",
		validations.Present(),
		validations.Unique(),
		validations.LengthRange(8, 8))
}

func TestMultipleValidators(t *testing.T) {
	ti := MultipleValidationsTestModel{}
	if res := db.Save(&ti); res.Error == nil {
		t.Errorf("Should get error when SomeString of record is NULL")
	}
	ti.SomeString.Scan("12345678")
	if res := db.Save(&ti); res.Error != nil {
		t.Errorf("Should not get error when SomeString of record is unique, not NULL, has 8 characters")
	}
	ti.ID = 0
	if res := db.Save(&ti); res.Error == nil {
		t.Errorf("Should get error when SomeString of record is not NULL, has 8 characters, but not unique")
	}
	ti.ID = 0
	ti.SomeString.Scan("123")
	if res := db.Save(&ti); res.Error == nil {
		t.Errorf("Should get error when SomeString of record is unique, not NULL, but hasn't 8 characters")
	}
}
