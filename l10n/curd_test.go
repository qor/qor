package l10n_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/l10n"

	_ "github.com/go-sql-driver/mysql"
)

func checkHasErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func checkHasProductInLocale(db *gorm.DB, locale string, t *testing.T) {
	var count int
	if db.Where("language_code = ?", locale).Count(&count); count != 1 {
		t.Errorf("should has only one product for locale %v, but found %v", locale, count)
	}
}

func checkHasProductInAllLocales(db *gorm.DB, t *testing.T) {
	checkHasProductInLocale(db, l10n.Global, t)
	checkHasProductInLocale(db, "zh", t)
	checkHasProductInLocale(db, "en", t)
}

func TestCreateWithCreate(t *testing.T) {
	product := Product{Code: "CreateWithCreate"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasErr(t, dbEN.Create(&product).Error)

	checkHasProductInAllLocales(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "CreateWithCreate"), t)
}

func TestCreateWithSave(t *testing.T) {
	product := Product{Code: "CreateWithSave"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasErr(t, dbEN.Create(&product).Error)

	checkHasProductInAllLocales(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "CreateWithSave"), t)
}

func TestUpdate(t *testing.T) {
	product := Product{Code: "Update", Name: "global"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	sharedDB := dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "Update")

	product.Name = "中文名"
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasProductInLocale(sharedDB.Where("name = ?", "中文名"), "zh", t)

	product.Name = "English Name"
	checkHasErr(t, dbEN.Create(&product).Error)
	checkHasProductInLocale(sharedDB.Where("name = ?", "English Name"), "en", t)

	product.Name = "新的中文名"
	product.Code = "NewCode // should be ignored when update"
	dbCN.Save(&product)
	checkHasProductInLocale(sharedDB.Where("name = ?", "新的中文名"), "zh", t)

	product.Name = "New English Name"
	product.Code = "NewCode // should be ignored when update"
	dbEN.Save(&product)
	checkHasProductInLocale(sharedDB.Where("name = ?", "New English Name"), "en", t)
}

func TestQuery(t *testing.T) {
	product := Product{Code: "Query", Name: "global"}
	dbGlobal.Create(&product)
	dbCN.Create(&product)

	var productCN Product
	dbCN.First(&productCN, product.ID)
	if productCN.LanguageCode != "zh" {
		t.Error("Should find localized zh product with mixed mode")
	}

	if dbCN.Set("l10n:mode", "locale").First(&productCN, product.ID).RecordNotFound() {
		t.Error("Should find localized zh product with locale mode")
	}

	if dbCN.Set("l10n:mode", "global").First(&productCN); productCN.LanguageCode != l10n.Global {
		t.Error("Should find global product with global mode")
	}

	var productEN Product
	dbEN.First(&productEN, product.ID)
	if productEN.LanguageCode != l10n.Global {
		t.Error("Should find global product for en with mixed mode")
	}

	if !dbEN.Set("l10n:mode", "locale").First(&productEN, product.ID).RecordNotFound() {
		t.Error("Should find no record with locale mode")
	}

	if dbEN.Set("l10n:mode", "global").First(&productEN); productEN.LanguageCode != l10n.Global {
		t.Error("Should find global product with global mode")
	}
}

func TestDelete(t *testing.T) {
	product := Product{Code: "Delete", Name: "global"}
	dbGlobal.Create(&product)
	dbCN.Create(&product)

	if dbCN.Delete(&product).RowsAffected != 1 {
		t.Errorf("Should delete localized record")
	}

	if dbEN.Delete(&product).RowsAffected != 0 {
		t.Errorf("Should delete none record in unlocalized locale")
	}
}

func TestResetLanguageCodeWithGlobalDB(t *testing.T) {
	product := Product{Code: "Query", Name: "global"}
	product.LanguageCode = "test"
	dbGlobal.Save(&product)
	if product.LanguageCode != l10n.Global {
		t.Error("Should reset language code in global mode")
	}
}

func TestManyToManyRelations(t *testing.T) {
	product := Product{Code: "Delete", Name: "global", Tags: []Tag{{Name: "tag1"}, {Name: "tag2"}}}
	dbGlobal.Debug().Save(&product)
}
