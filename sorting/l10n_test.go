package sorting_test

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/l10n"
	"github.com/qor/qor/sorting"
)

type Brand struct {
	gorm.Model
	l10n.Locale
	sorting.Sorting
	Name string
}

func prepareBrand() {
	db.Delete(&Brand{})
	globalDB := db.Set("l10n:locale", l10n.Global)
	zhDB := db.Set("l10n:locale", "zh-CN")

	for i := 1; i <= 5; i++ {
		brand := Brand{Name: fmt.Sprintf("brand%v", i)}
		globalDB.Save(&brand)
		if i > 3 {
			zhDB.Save(&brand)
		}
	}
}

func getBrand(db *gorm.DB, name string) *Brand {
	var brand Brand
	db.First(&brand, "name = ?", name)
	return &brand
}

func checkBrandPosition(db *gorm.DB, t *testing.T, description string) {
	var brands []Brand
	if err := db.Set("l10n:mode", "locale").Find(&brands).Error; err != nil {
		t.Errorf("no error should happen when find brands, but got %v", err)
	}
	for i, brand := range brands {
		if brand.Position != i+1 {
			t.Errorf("Brand %v(%v)'s position should be %v after %v, but got %v", brand.ID, brand.LanguageCode, i+1, description, brand.Position)
		}
	}
}

func TestBrandPosition(t *testing.T) {
	prepareBrand()
	globalDB := db.Set("l10n:locale", "en-US")
	zhDB := db.Set("l10n:locale", "zh-CN")

	checkBrandPosition(db, t, "initalize")
	checkBrandPosition(zhDB, t, "initalize")

	if err := globalDB.Delete(getBrand(globalDB, "brand1")).Error; err != nil {
		t.Errorf("no error should happen when delete an en-US brand, but got %v", err)
	}
	checkBrandPosition(globalDB, t, "delete an brand from global db")
	checkBrandPosition(zhDB, t, "delete an brand from global db")

	if err := zhDB.Delete(getBrand(zhDB, "brand4")).Error; err != nil {
		t.Errorf("no error should happen when delete an zh-CN brand, but got %v", err)
	}
	checkBrandPosition(globalDB, t, "delete an brand from zh db")
	checkBrandPosition(zhDB, t, "delete an brand from zh db")
}
