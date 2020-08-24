package resource_test

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/publish2"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	testutils "github.com/qor/qor/test/utils"
	"github.com/qor/qor/utils"
	"github.com/qor/sorting"
)

func format(value interface{}) string {
	return fmt.Sprint(utils.Indirect(reflect.ValueOf(value)).Interface())
}

func checkMeta(record interface{}, meta *resource.Meta, value interface{}, t *testing.T, expectedValues ...string) {
	var (
		context       = &qor.Context{DB: testutils.TestDB()}
		metaValue     = &resource.MetaValue{Name: meta.Name, Value: value}
		expectedValue = fmt.Sprint(value)
	)

	for _, v := range expectedValues {
		expectedValue = v
	}

	meta.PreInitialize()
	meta.Initialize()

	if meta.Setter != nil {
		meta.Setter(record, metaValue, context)
		if context.HasError() {
			t.Errorf("No error should happen, but got %v", context.Errors)
		}

		result := meta.Valuer(record, context)
		if resultValuer, ok := result.(driver.Valuer); ok {
			if v, err := resultValuer.Value(); err == nil {
				result = v
			}
		}

		if format(result) != expectedValue {
			t.Errorf("Wrong value, should be %v, but got %v", expectedValue, format(result))
		}
	} else {
		t.Errorf("No setter generated for meta %v", meta.Name)
	}
}

func TestStringMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Name  string
		Name2 *string
	}{}

	res := resource.New(user)

	meta := &resource.Meta{
		Name:         "Name",
		BaseResource: res,
	}

	checkMeta(user, meta, "hello world", t)

	meta2 := &resource.Meta{
		Name:         "Name2",
		BaseResource: res,
	}

	checkMeta(user, meta2, "hello world2", t)
}

func TestIntMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Age  int
		Age2 uint
		Age3 *int8
		Age4 *uint8
	}{}

	res := resource.New(user)

	meta := &resource.Meta{
		Name:         "Age",
		BaseResource: res,
	}

	checkMeta(user, meta, 18, t)

	meta2 := &resource.Meta{
		Name:         "Age2",
		BaseResource: res,
	}

	checkMeta(user, meta2, "28", t)

	meta3 := &resource.Meta{
		Name:         "Age3",
		BaseResource: res,
	}

	checkMeta(user, meta3, 38, t)

	meta4 := &resource.Meta{
		Name:         "Age4",
		BaseResource: res,
	}

	checkMeta(user, meta4, "48", t)
}

func TestFloatMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Age  float64
		Age2 *float64
	}{}

	res := resource.New(user)

	meta := &resource.Meta{
		Name:         "Age",
		BaseResource: res,
	}

	checkMeta(user, meta, 18.5, t)

	meta2 := &resource.Meta{
		Name:         "Age2",
		BaseResource: res,
	}

	checkMeta(user, meta2, "28.5", t)
}

func TestBoolMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Actived  bool
		Actived2 *bool
	}{}

	res := resource.New(user)

	meta := &resource.Meta{
		Name:         "Actived",
		BaseResource: res,
	}

	checkMeta(user, meta, "true", t)

	meta2 := &resource.Meta{
		Name:         "Actived2",
		BaseResource: res,
	}

	checkMeta(user, meta2, "true", t)

	meta3 := &resource.Meta{
		Name:         "Actived",
		BaseResource: res,
	}

	checkMeta(user, meta3, "", t, "false")

	meta4 := &resource.Meta{
		Name:         "Actived2",
		BaseResource: res,
	}

	checkMeta(user, meta4, "f", t, "false")
}

type scanner struct {
	Body string
}

func (s *scanner) Scan(value interface{}) error {
	s.Body = fmt.Sprint(value)
	return nil
}

func (s scanner) Value() (driver.Value, error) {
	return s.Body, nil
}

func TestScannerMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Scanner scanner
	}{}

	res := resource.New(user)

	meta := &resource.Meta{
		Name:         "Scanner",
		BaseResource: res,
	}

	checkMeta(user, meta, "scanner", t)
}

func TestSliceMetaValuerAndSetter(t *testing.T) {
	t.Skip()

	user := &struct {
		Names  []string
		Names2 []*string
		Names3 *[]string
		Names4 []*string
	}{}

	res := resource.New(user)

	meta := &resource.Meta{
		Name:         "Names",
		BaseResource: res,
	}

	checkMeta(user, meta, []string{"name1", "name2"}, t)

	meta2 := &resource.Meta{
		Name:         "Names2",
		BaseResource: res,
	}

	checkMeta(user, meta2, []string{"name1", "name2"}, t)

	meta3 := &resource.Meta{
		Name:         "Names3",
		BaseResource: res,
	}

	checkMeta(user, meta3, []string{"name1", "name2"}, t)

	meta4 := &resource.Meta{
		Name:         "Names4",
		BaseResource: res,
	}

	checkMeta(user, meta4, []string{"name1", "name2"}, t)
}

type Collection struct {
	gorm.Model

	Name string

	Products       []Product `gorm:"many2many:collection_products;association_autoupdate:false"`
	ProductsSorter sorting.SortableCollection
}

type CollectionWithVersion struct {
	gorm.Model

	publish2.Version
	publish2.Schedule

	Name string

	Products       []ProductWithVersion `gorm:"many2many:collection_with_version_product_with_versions;association_autoupdate:false"`
	ProductsSorter sorting.SortableCollection
}

type ProductWithVersion struct {
	gorm.Model

	publish2.Schedule
	publish2.Version

	Name string
}

type Product struct {
	gorm.Model

	Name string
}

func WithoutVersion(db *gorm.DB) *gorm.DB {
	return db.Set(admin.DisableCompositePrimaryKeyMode, "on").Set(publish2.VersionMode, publish2.VersionMultipleMode).Set(publish2.ScheduleMode, publish2.ModeOff)
}

func updateVersionPriority() func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		if field, ok := scope.FieldByName("VersionPriority"); ok {
			createdAtField, _ := scope.FieldByName("CreatedAt")
			createdAt := createdAtField.Field.Interface().(time.Time)

			versionNameField, _ := scope.FieldByName("VersionName")
			versionName := versionNameField.Field.Interface().(string)

			versionPriority := fmt.Sprintf("%v_%v", createdAt.UTC().Format(time.RFC3339), versionName)
			field.Set(versionPriority)
		}
	}
}
func updateCallback(scope *gorm.Scope) {
	return
}
func TestMany2ManyRelation(t *testing.T) {
	db := testutils.TestDB()
	testutils.ResetDBTables(db, &Collection{}, &Product{}, "collection_products")

	adm := admin.New(&qor.Config{DB: db.Set(publish2.ScheduleMode, publish2.ModeOff)})
	c := adm.AddResource(&Collection{})

	productsMeta := resource.Meta{
		Name:         "Products",
		FieldName:    "Products",
		BaseResource: c,
		Config: &admin.SelectManyConfig{
			Collection: func(value interface{}, ctx *qor.Context) (results [][]string) {
				if c, ok := value.(*Collection); ok {
					var products []Product
					ctx.GetDB().Model(c).Related(&products, "Products")

					for _, product := range products {
						results = append(results, []string{fmt.Sprintf("%v", product.ID), product.Name})
					}
				}
				return
			},
		},
	}

	var scope = &gorm.Scope{Value: c.Value}
	var getField = func(fields []*gorm.StructField, name string) *gorm.StructField {
		for _, field := range fields {
			if field.Name == name || field.DBName == name {
				return field
			}
		}
		return nil
	}

	productsMeta.FieldStruct = getField(scope.GetStructFields(), productsMeta.FieldName)

	if err := productsMeta.Initialize(); err != nil {
		t.Fatal(err)
	}

	p1 := Product{Name: "p1"}
	p2 := Product{Name: "p2"}
	testutils.AssertNoErr(t, db.Save(&p1).Error)
	testutils.AssertNoErr(t, db.Save(&p2).Error)

	record := Collection{Name: "test"}
	testutils.AssertNoErr(t, db.Save(&record).Error)
	ctx := &qor.Context{DB: db}
	metaValue := &resource.MetaValue{Name: productsMeta.Name, Value: []string{fmt.Sprintf("%d", p1.ID), fmt.Sprintf("%d", p2.ID)}}

	productsMeta.Setter(&record, metaValue, ctx)

	testutils.AssertNoErr(t, db.Preload("Products").Find(&record).Error)
	if len(record.Products) != 2 {
		t.Error("products not set to collection")
	}
}

func TestManyToManyRelation_WithVersion(t *testing.T) {
	db := testutils.TestDB()
	registerVersionNameCallback(db)
	publish2.RegisterCallbacks(db)
	testutils.ResetDBTables(db, &CollectionWithVersion{}, &ProductWithVersion{}, "collection_with_versions_product_with_versions")

	adm := admin.New(&qor.Config{DB: db.Set(publish2.ScheduleMode, publish2.ModeOff)})
	c := adm.AddResource(&CollectionWithVersion{})

	productsMeta := resource.Meta{
		Name:         "Products",
		FieldName:    "Products",
		BaseResource: c,
		Config: &admin.SelectManyConfig{
			Collection: func(value interface{}, ctx *qor.Context) (results [][]string) {
				if c, ok := value.(*CollectionWithVersion); ok {
					var products []ProductWithVersion
					ctx.GetDB().Model(c).Related(&products, "Products")

					for _, product := range products {
						results = append(results, []string{fmt.Sprintf("%v", product.ID), product.Name})
					}
				}
				return
			},
		},
	}

	var scope = &gorm.Scope{Value: c.Value}
	var getField = func(fields []*gorm.StructField, name string) *gorm.StructField {
		for _, field := range fields {
			if field.Name == name || field.DBName == name {
				return field
			}
		}
		return nil
	}

	productsMeta.FieldStruct = getField(scope.GetStructFields(), productsMeta.FieldName)

	if err := productsMeta.Initialize(); err != nil {
		t.Fatal(err)
	}

	p1 := ProductWithVersion{Name: "p1"}
	p2_v1 := ProductWithVersion{Name: "p2"}
	testutils.AssertNoErr(t, db.Save(&p1).Error)
	testutils.AssertNoErr(t, db.Save(&p2_v1).Error)
	p2_v2 := ProductWithVersion{Name: "p2"}
	p2_v2.ID = p2_v1.ID
	testutils.AssertNoErr(t, db.Save(&p2_v2).Error)

	record := CollectionWithVersion{Name: "test"}
	testutils.AssertNoErr(t, db.Save(&record).Error)
	ctx := &qor.Context{DB: db}
	metaValue := &resource.MetaValue{Name: productsMeta.Name, Value: []map[string]string{
		{"id": fmt.Sprintf("%d", p1.ID), "version_name": p1.GetVersionName()},
		{"id": fmt.Sprintf("%d", p2_v2.ID), "version_name": p2_v2.GetVersionName()},
	}}

	productsMeta.Setter(&record, metaValue, ctx)

	testutils.AssertNoErr(t, db.Preload("Products").Find(&record).Error)
	if len(record.Products) != 2 {
		t.Error("products not set to collection")
	}

	hasCorrectVersion := false
	for _, p := range record.Products {
		if p.ID == p2_v2.ID && p.GetVersionName() == p2_v2.VersionName {
			hasCorrectVersion = true
		}
	}

	if !hasCorrectVersion {
		t.Error("p2 is not associated with collection with correct version")
	}
}

func registerVersionNameCallback(db *gorm.DB) {
	db.Callback().Create().Before("gorm:begin_transaction").Register("publish2:versions", func(scope *gorm.Scope) {
		if field, ok := scope.FieldByName("VersionName"); ok {
			if !field.IsBlank {
				return
			}

			name := time.Now().Format("2006-01-02")

			idField, _ := scope.FieldByName("ID")
			id := idField.Field.Interface().(uint)

			var count int
			scope.DB().Table(scope.TableName()).Unscoped().Scopes(WithoutVersion).Where("id = ? AND version_name like ?", id, name+"%").Count(&count)

			versionName := fmt.Sprintf("%s-v%v", name, count+1)
			field.Set(versionName)
		}
	})

	db.Callback().Create().After("gorm:begin_transaction").Register("publish2:version_priority", updateVersionPriority())
}
