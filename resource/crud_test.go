package resource

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/qor"
	"github.com/stretchr/testify/assert"
)

type DummyStruct struct {
	gorm.Model
	CID  string `gorm:"column:cid"`
	Name string
}

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to in-memory SQLite database: %v", err)
	}

	db.LogMode(true)

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	})

	return db
}

func TestToPrimaryQueryParams(t *testing.T) {
	tests := []struct {
		name              string
		primaryValue      string
		expectedSQL       string
		expectedQueryArgs []interface{}
	}{
		{
			name:              "Single primary field 'id' with valid value",
			primaryValue:      "123",
			expectedSQL:       `"dummy_structs"."id" = ?`,
			expectedQueryArgs: []interface{}{"123"},
		},
		{
			name:              "Single primary field 'cid' with valid value",
			primaryValue:      "abc",
			expectedSQL:       `"dummy_structs"."cid" = ?`,
			expectedQueryArgs: []interface{}{"abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize GORM with in-memory SQLite for testing
			db := setupDB(t)

			// Automigrate DummyStruct to create schema
			if err := db.AutoMigrate(&DummyStruct{}).Error; err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			// Create a qor.Context with the DB
			ctx := &qor.Context{DB: db}

			// Initialize Resource directly
			res := New(&DummyStruct{})

			assert.Equal(t, 1, len(res.PrimaryFields))

			// Call the method
			sql, args := res.ToPrimaryQueryParams(tt.primaryValue, ctx)

			// Assert the results
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedQueryArgs, args)
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"00123", true},
		{"-123", false}, // Negative numbers are not handled
		{"123.45", false},
		{"", false},
		{"abc", false},
		{"123abc", false},
		{" 123 ", false},
		{"9223372036854775807", true},   // Max uint64
		{"18446744073709551615", true},  // Max uint64
		{"18446744073709551616", false}, // Overflow uint64
	}

	for _, test := range tests {
		result := isNumeric(test.input)
		if result != test.expected {
			t.Errorf("isNumeric(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}
