package resource

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"github.com/qor/validations"
)

// CompositePrimaryKeySeparator to separate composite primary keys like ID and version_name
const CompositePrimaryKeySeparator = "^|^"

// CompositePrimaryKey the string that represents the composite primary key
const CompositePrimaryKeyFieldName = "CompositePrimaryKeyField"

// CompositePrimaryKeyField to embed into the struct that requires composite primary key in select many
type CompositePrimaryKeyField struct {
	CompositePrimaryKey string `gorm:"-"`
}

// CompositePrimaryKeyStruct container for store id & version combination temporarily
type CompositePrimaryKeyStruct struct {
	ID          uint   `json:"id"`
	VersionName string `json:"version_name"`
}

// GenCompositePrimaryKey generates composite primary key in a specific format
func GenCompositePrimaryKey(id interface{}, versionName string) string {
	return fmt.Sprintf("%d%s%s", id, CompositePrimaryKeySeparator, versionName)
}

// Metaor interface
type Metaor interface {
	GetName() string
	GetFieldName() string
	GetSetter() func(resource interface{}, metaValue *MetaValue, context *qor.Context)
	GetFormattedValuer() func(interface{}, *qor.Context) interface{}
	GetValuer() func(interface{}, *qor.Context) interface{}
	GetResource() Resourcer
	GetMetas() []Metaor
	SetPermission(*roles.Permission)
	HasPermission(roles.PermissionMode, *qor.Context) bool
}

// ConfigureMetaBeforeInitializeInterface if a struct's field's type implemented this interface, it will be called when initializing a meta
type ConfigureMetaBeforeInitializeInterface interface {
	ConfigureQorMetaBeforeInitialize(Metaor)
}

// ConfigureMetaInterface if a struct's field's type implemented this interface, it will be called after configed
type ConfigureMetaInterface interface {
	ConfigureQorMeta(Metaor)
}

// MetaConfigInterface meta configuration interface
type MetaConfigInterface interface {
	ConfigureMetaInterface
}

// MetaConfig base meta config struct
type MetaConfig struct {
}

// ConfigureQorMeta implement the MetaConfigInterface
func (MetaConfig) ConfigureQorMeta(Metaor) {
}

// Meta meta struct definition
type Meta struct {
	Name            string
	FieldName       string
	FieldStruct     *gorm.StructField
	Setter          func(resource interface{}, metaValue *MetaValue, context *qor.Context)
	Valuer          func(interface{}, *qor.Context) interface{}
	FormattedValuer func(interface{}, *qor.Context) interface{}
	Config          MetaConfigInterface
	BaseResource    Resourcer
	Resource        Resourcer
	Permission      *roles.Permission
}

// GetBaseResource get base resource from meta
func (meta Meta) GetBaseResource() Resourcer {
	return meta.BaseResource
}

// GetName get meta's name
func (meta Meta) GetName() string {
	return meta.Name
}

// GetFieldName get meta's field name
func (meta Meta) GetFieldName() string {
	return meta.FieldName
}

// SetFieldName set meta's field name
func (meta *Meta) SetFieldName(name string) {
	meta.FieldName = name
}

// GetSetter get setter from meta
func (meta Meta) GetSetter() func(resource interface{}, metaValue *MetaValue, context *qor.Context) {
	return meta.Setter
}

// SetSetter set setter to meta
func (meta *Meta) SetSetter(fc func(resource interface{}, metaValue *MetaValue, context *qor.Context)) {
	meta.Setter = fc
}

// GetValuer get valuer from meta
func (meta Meta) GetValuer() func(interface{}, *qor.Context) interface{} {
	return meta.Valuer
}

// SetValuer set valuer for meta
func (meta *Meta) SetValuer(fc func(interface{}, *qor.Context) interface{}) {
	meta.Valuer = fc
}

// GetFormattedValuer get formatted valuer from meta
func (meta *Meta) GetFormattedValuer() func(interface{}, *qor.Context) interface{} {
	if meta.FormattedValuer != nil {
		return meta.FormattedValuer
	}
	return meta.Valuer
}

// SetFormattedValuer set formatted valuer for meta
func (meta *Meta) SetFormattedValuer(fc func(interface{}, *qor.Context) interface{}) {
	meta.FormattedValuer = fc
}

// HasPermission check has permission or not
func (meta Meta) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	var roles = []interface{}{}
	for _, role := range context.Roles {
		roles = append(roles, role)
	}
	return meta.Permission.HasPermission(mode, roles...)
}

// SetPermission set permission for meta
func (meta *Meta) SetPermission(permission *roles.Permission) {
	meta.Permission = permission
}

// PreInitialize when will be run before initialize, used to fill some basic necessary information
func (meta *Meta) PreInitialize() error {
	if meta.Name == "" {
		utils.ExitWithMsg("Meta should have name: %v", reflect.TypeOf(meta))
	} else if meta.FieldName == "" {
		meta.FieldName = meta.Name
	}

	// parseNestedField used to handle case like Profile.Name
	var parseNestedField = func(value reflect.Value, name string) (reflect.Value, string) {
		fields := strings.Split(name, ".")
		value = reflect.Indirect(value)
		for _, field := range fields[:len(fields)-1] {
			value = value.FieldByName(field)
		}

		return value, fields[len(fields)-1]
	}

	var getField = func(fields []*gorm.StructField, name string) *gorm.StructField {
		for _, field := range fields {
			if field.Name == name || field.DBName == name {
				return field
			}
		}
		return nil
	}

	var nestedField = strings.Contains(meta.FieldName, ".")
	var scope = &gorm.Scope{Value: meta.BaseResource.GetResource().Value}
	if nestedField {
		subModel, name := parseNestedField(reflect.ValueOf(meta.BaseResource.GetResource().Value), meta.FieldName)
		meta.FieldStruct = getField(scope.New(subModel.Interface()).GetStructFields(), name)
	} else {
		meta.FieldStruct = getField(scope.GetStructFields(), meta.FieldName)
	}
	return nil
}

// Initialize initialize meta, will set valuer, setter if haven't configure it
func (meta *Meta) Initialize() error {
	// Set Valuer for Meta
	if meta.Valuer == nil {
		setupValuer(meta, meta.FieldName, meta.GetBaseResource().NewStruct())
	}

	if meta.Valuer == nil {
		utils.ExitWithMsg("Meta %v is not supported for resource %v, no `Valuer` configured for it", meta.FieldName, reflect.TypeOf(meta.BaseResource.GetResource().Value))
	}

	// Set Setter for Meta
	if meta.Setter == nil {
		setupSetter(meta, meta.FieldName, meta.GetBaseResource().NewStruct())
	}
	return nil
}

// setCompositePrimaryKey if the association has CompositePrimaryKey integrated, generates value for it by our conventional format
// the PrimaryKeyOf function will return the value from CompositePrimaryKey instead of ID, so that frontend could find correct version
func setCompositePrimaryKey(f *gorm.Field) {
	for i := 0; i < f.Field.Len(); i++ {
		associatedRecord := reflect.Indirect(f.Field.Index(i))
		if v := associatedRecord.FieldByName(CompositePrimaryKeyFieldName); v.IsValid() {
			id := associatedRecord.FieldByName("ID").Uint()
			versionName := associatedRecord.FieldByName("VersionName").String()
			associatedRecord.FieldByName("CompositePrimaryKey").SetString(fmt.Sprintf("%d%s%s", id, CompositePrimaryKeySeparator, versionName))
		}
	}
}

func setupValuer(meta *Meta, fieldName string, record interface{}) {
	nestedField := strings.Contains(fieldName, ".")

	// Setup nested fields
	if nestedField {
		fieldNames := strings.Split(fieldName, ".")
		setupValuer(meta, strings.Join(fieldNames[1:], "."), getNestedModel(record, strings.Join(fieldNames[0:2], "."), nil))

		oldValuer := meta.Valuer
		meta.Valuer = func(record interface{}, context *qor.Context) interface{} {
			return oldValuer(getNestedModel(record, strings.Join(fieldNames[0:2], "."), context), context)
		}
		return
	}

	if meta.FieldStruct != nil {
		meta.Valuer = func(value interface{}, context *qor.Context) interface{} {
			// get scope of current record. like Collection, then iterate its fields
			scope := context.GetDB().NewScope(value)

			if f, ok := scope.FieldByName(fieldName); ok {
				if relationship := f.Relationship; relationship != nil && f.Field.CanAddr() && !scope.PrimaryKeyZero() {
					// Iterate each field see if it is an relationship field like
					// Factories []factory.Factory
					// If so, set the CompositePrimaryKey value for PrimaryKeyOf to read
					if (relationship.Kind == "has_many" || relationship.Kind == "many_to_many") && f.Field.Len() == 0 {
						// Retrieve the associated records from db
						context.GetDB().Set("publish:version:mode", "multiple").Model(value).Related(f.Field.Addr().Interface(), fieldName)

						setCompositePrimaryKey(f)
					} else if (relationship.Kind == "has_one" || relationship.Kind == "belongs_to") && context.GetDB().NewScope(f.Field.Interface()).PrimaryKeyZero() {
						if f.Field.Kind() == reflect.Ptr && f.Field.IsNil() {
							f.Field.Set(reflect.New(f.Field.Type().Elem()))
						}

						context.GetDB().Set("publish:version:mode", "multiple").Model(value).Related(f.Field.Addr().Interface(), fieldName)
					}
				}

				return f.Field.Interface()
			}

			return ""
		}
	}
}

// switchRecordToNewVersionIfNeeded is for switching to new version of the record when creating a new version.
// The given record must has function 'AssignVersionName' defined, with *Pointer* receiver to create associations on new version
// Otherwise, the operation would be omitted
// e.g. the user is creating a new version based on version "2021-3-3-v1". which would be "2021-3-3-v2".
//      the associations added during the creation should be associated with "2021-3-3-v2" rather than "2021-3-3-v1"
func switchRecordToNewVersionIfNeeded(context *qor.Context, record interface{}) interface{} {
	if context.Request == nil {
		return record
	}

	currentVersionName := context.Request.Form.Get("QorResource.VersionName")
	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}

	// Handle situation when the primary key is a uint64 not general uint
	var id uint64
	idUint, ok := recordValue.FieldByName("ID").Interface().(uint)
	if !ok {
		id64, ok := recordValue.FieldByName("ID").Interface().(uint64)
		if !ok {
			panic("ID filed must be uint or uint64")
		}
		id = id64
	} else {
		id = uint64(idUint)
	}

	// if currentVersionName is blank, we consider it is creating a new version
	if id != 0 && currentVersionName == "" {
		arguments := []reflect.Value{reflect.ValueOf(context.GetDB())}

		// Handle the situation when record is NOT a pointer
		if reflect.ValueOf(record).Kind() != reflect.Ptr {
			// We create a new pointer to be able to invoke the AssignVersionName method on Pointer receiver
			recordPtr := reflect.New(reflect.TypeOf(record))
			recordPtr.Elem().Set(reflect.ValueOf(record))
			fn := recordPtr.MethodByName("AssignVersionName")

			if !fn.IsValid() {
				log.Printf("Struct %v must has function 'AssignVersionName' defined, with *Pointer* receiver to create associations on new version", reflect.TypeOf(record).Name())
				return record
			}
			fn.Call(arguments)

			// Since it is a new pointer, we have to return the new record
			return recordPtr.Elem().Interface()
		}

		// When the record is a pointer
		fn := reflect.ValueOf(record).MethodByName("AssignVersionName")
		if !fn.IsValid() {
			log.Printf("Struct %v must has function 'AssignVersionName' defined, with *Pointer* receiver to create associations on new version", reflect.TypeOf(record).Name())
			return record
		}

		// AssignVersionName set the record's version name as the new version, so when execute the SQL, we can find correct object to apply the association
		fn.Call(arguments)
		return record
	}

	return record
}

func HandleBelongsTo(context *qor.Context, record reflect.Value, field reflect.Value, relationship *gorm.Relationship, primaryKeys []string) {
	// Read value from foreign key field. e.g.  TagID => 1
	oldPrimaryKeys := utils.ToArray(record.FieldByName(relationship.ForeignFieldNames[0]).Interface())
	// if not changed, return immediately
	if fmt.Sprint(primaryKeys) == fmt.Sprint(oldPrimaryKeys) {
		return
	}

	foreignKeyField := record.FieldByName(relationship.ForeignFieldNames[0])
	if len(primaryKeys) == 0 {
		// if foreign key removed
		foreignKeyField.Set(reflect.Zero(foreignKeyField.Type()))
	} else {
		// if foreign key changed. We need to make sure the field is a blank object
		// Suppose this is a Collection belongs to Tag association
		// non-blank field will perform a query like `SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND "tags"."id" = 1 AND (("tags"."id" IN ('2')))`
		// Usually this won't happen, cause the Tag field of Collection will be blank by default. it is a double assurance.
		field.FieldByName("ID").SetUint(0)
		context.GetDB().Set("publish:version:mode", "multiple").Where(primaryKeys).Find(field.Addr().Interface())
	}
}

func HandleVersioningBelongsTo(context *qor.Context, record reflect.Value, field reflect.Value, relationship *gorm.Relationship, primaryKeys []string, fieldHasVersion bool) {
	foreignKeyName := relationship.ForeignFieldNames[0]
	// Construct version name foreign key. e.g.  ManagerID -> ManagerVersionName
	foreignVersionName := strings.Replace(foreignKeyName, "ID", "VersionName", -1)

	foreignKeyField := record.FieldByName(foreignKeyName)
	foreignVersionField := record.FieldByName(foreignVersionName)

	oldPrimaryKeys := utils.ToArray(foreignKeyField.Interface())
	// If field struct has version and it defined XXVersionName foreignKey field
	// then construct ID+VersionName and compare with composite primarykey
	if fieldHasVersion && len(oldPrimaryKeys) != 0 && foreignVersionField.IsValid() {
		oldPrimaryKeys[0] = GenCompositePrimaryKey(oldPrimaryKeys[0], foreignVersionField.String())
	}

	// if not changed
	if fmt.Sprint(primaryKeys) == fmt.Sprint(oldPrimaryKeys) {
		return
	}

	// foreignkey removed
	if len(primaryKeys) == 0 {
		foreignKeyField.Set(reflect.Zero(foreignKeyField.Type()))
		// if field has version, we have to set both the id and version_name to zero value.
		if fieldHasVersion {
			foreignKeyField.Set(reflect.Zero(foreignKeyField.Type()))
			foreignVersionField.Set(reflect.Zero(foreignVersionField.Type()))
		}
		// foreignkey updated
	} else {
		// if foreign key changed. We need to make sure the field is a blank object
		// Suppose this is a Collection belongs to Tag association
		// non-blank field will perform a query like `SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND "tags"."id" = 1 AND (("tags"."id" IN ('2')))`
		// Usually this won't happen, cause the Tag field of Collection will be blank by default. it is a double assurance.
		field.FieldByName("ID").SetUint(0)

		compositePKeys := strings.Split(primaryKeys[0], CompositePrimaryKeySeparator)
		// If primaryKeys doesn't include version name, process it as an ID
		if len(compositePKeys) == 1 {
			context.GetDB().Set("publish:version:mode", "multiple").Where(primaryKeys).Find(field.Addr().Interface())
		} else {
			context.GetDB().Set("publish:version:mode", "multiple").Where("id = ? AND version_name = ?", compositePKeys[0], compositePKeys[1]).Find(field.Addr().Interface())
		}
	}
}

func CollectPrimaryKeys(metaValueForCompositePrimaryKeys []string) (compositePKeys []CompositePrimaryKeyStruct, compositePKeyConvertErr error) {
	// To convert []string{"1^|^2020-09-14-v1", "2^|^2020-09-14-v3"} to []compositePrimaryKey
	for _, rawCpk := range metaValueForCompositePrimaryKeys {
		// Skip blank string when it is not the only element
		if len(rawCpk) == 0 && len(metaValueForCompositePrimaryKeys) > 1 {
			continue
		}

		pks := strings.Split(rawCpk, CompositePrimaryKeySeparator)
		if len(pks) != 2 {
			compositePKeyConvertErr = errors.New("metaValue is not for composite primary key")
			break
		}

		id, convErr := strconv.ParseUint(pks[0], 10, 32)
		if convErr != nil {
			compositePKeyConvertErr = fmt.Errorf("composite primary key has incorrect id %s", pks[0])
			break
		}

		cpk := CompositePrimaryKeyStruct{
			ID:          uint(id),
			VersionName: pks[1],
		}

		compositePKeys = append(compositePKeys, cpk)
	}

	return
}

func HandleManyToMany(context *qor.Context, scope *gorm.Scope, meta *Meta, record interface{}, metaValue *MetaValue, field reflect.Value, fieldHasVersion bool) {
	metaValueForCompositePrimaryKeys, ok := metaValue.Value.([]string)
	compositePKeys := []CompositePrimaryKeyStruct{}
	var compositePKeyConvertErr error

	if ok {
		compositePKeys, compositePKeyConvertErr = CollectPrimaryKeys(metaValueForCompositePrimaryKeys)
	}

	// If the field is a struct with version and metaValue is present and we can collect id + version_name combination
	// It means we can make the query by specific condition
	if fieldHasVersion && metaValue.Value != nil && compositePKeyConvertErr == nil && len(compositePKeys) > 0 {
		HandleVersionedManyToMany(context, field, compositePKeys)
	} else {
		HandleNormalManyToMany(context, field, metaValue, fieldHasVersion, compositePKeyConvertErr)
	}

	if !scope.PrimaryKeyZero() {
		context.GetDB().Model(record).Association(meta.FieldName).Replace(field.Interface())
		field.Set(reflect.Zero(field.Type()))
	}
}

// HandleNormalManyToMany not only handle normal many_to_many relationship, it also handled the situation that user set the association to blank
func HandleNormalManyToMany(context *qor.Context, field reflect.Value, metaValue *MetaValue, fieldHasVersion bool, compositePKeyConvertErr error) {
	if fieldHasVersion && metaValue.Value != nil && compositePKeyConvertErr != nil {
		fmt.Println("given meta value contains no version name, this might cause the association is incorrect")
	}

	primaryKeys := utils.ToArray(metaValue.Value)
	if metaValue.Value == nil {
		primaryKeys = []string{}
	}

	// set current field value to blank. This line responsible for set field to blank value when metaValue is nil
	// which means user removed all associations
	field.Set(reflect.Zero(field.Type()))

	if len(primaryKeys) > 0 {
		// replace it with new value
		context.GetDB().Set("publish:version:mode", "multiple").Where(primaryKeys).Find(field.Addr().Interface())
	}
}

// HandleVersionedManyToMany handle id+version_name composite primary key, query and set the correct result to the "Many" field
// e.g. Collection.Products
// This doesn't handle compositePKeys is blank logic, if it is, this function should not be invoked
func HandleVersionedManyToMany(context *qor.Context, field reflect.Value, compositePKeys []CompositePrimaryKeyStruct) {
	// set current field value to blank
	field.Set(reflect.Zero(field.Type()))

	// eliminate potential version_name condition on the main object, we don't need it when querying associated records
	// it usually added by qor/publish2.
	db := context.GetDB().Set("publish:version:mode", "multiple")
	for i, compositePKey := range compositePKeys {
		if i == 0 {
			db = db.Where("id = ? AND version_name = ?", compositePKey.ID, compositePKey.VersionName)
		} else {
			db = db.Or("id = ? AND version_name = ?", compositePKey.ID, compositePKey.VersionName)
		}
	}

	db.Find(field.Addr().Interface())
}

func setupSetter(meta *Meta, fieldName string, record interface{}) {
	nestedField := strings.Contains(fieldName, ".")

	// Setup nested fields
	if nestedField {
		fieldNames := strings.Split(fieldName, ".")
		setupSetter(meta, strings.Join(fieldNames[1:], "."), getNestedModel(record, strings.Join(fieldNames[0:2], "."), nil))

		oldSetter := meta.Setter
		meta.Setter = func(record interface{}, metaValue *MetaValue, context *qor.Context) {
			oldSetter(getNestedModel(record, strings.Join(fieldNames[0:2], "."), context), metaValue, context)
		}
		return
	}

	commonSetter := func(setter func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{})) func(record interface{}, metaValue *MetaValue, context *qor.Context) {
		return func(record interface{}, metaValue *MetaValue, context *qor.Context) {
			if metaValue == nil {
				return
			}

			defer func() {
				if r := recover(); r != nil {
					fmt.Println(r)
					debug.PrintStack()
					context.AddError(validations.NewError(record, meta.Name, fmt.Sprintf("Failed to set Meta %v's value with %v, got %v", meta.Name, metaValue.Value, r)))
				}
			}()

			field := utils.Indirect(reflect.ValueOf(record)).FieldByName(fieldName)
			if field.Kind() == reflect.Ptr {
				if field.IsNil() && utils.ToString(metaValue.Value) != "" {
					field.Set(utils.NewValue(field.Type()).Elem())
				}

				if utils.ToString(metaValue.Value) == "" {
					field.Set(reflect.Zero(field.Type()))
					return
				}

				for field.Kind() == reflect.Ptr {
					field = field.Elem()
				}
			}

			if field.IsValid() && field.CanAddr() {
				setter(field, metaValue, context, record)
			}
		}
	}

	// Setup belongs_to / many_to_many Setter
	if meta.FieldStruct != nil {
		if relationship := meta.FieldStruct.Relationship; relationship != nil {
			if relationship.Kind == "belongs_to" || relationship.Kind == "many_to_many" {
				meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
					var (
						scope         = context.GetDB().NewScope(record)
						recordAsValue = reflect.Indirect(reflect.ValueOf(record))
					)
					switchRecordToNewVersionIfNeeded(context, record)

					// If the field struct has version
					fieldHasVersion := fieldIsStructAndHasVersion(field)

					if relationship.Kind == "belongs_to" {
						primaryKeys := utils.ToArray(metaValue.Value)
						if metaValue.Value == nil {
							primaryKeys = []string{}
						}

						// For normal belongs_to association
						if len(relationship.ForeignFieldNames) == 1 {
							HandleBelongsTo(context, recordAsValue, field, relationship, primaryKeys)
						}

						// For versioning association
						if len(relationship.ForeignFieldNames) == 2 {
							HandleVersioningBelongsTo(context, recordAsValue, field, relationship, primaryKeys, fieldHasVersion)
						}
					}

					if relationship.Kind == "many_to_many" {
						// The reason why we use `record` as an interface{} here rather than `recordAsValue` is
						// we need make query by record, it must be a pointer, but belongs_to make query based on field, no need to be pointer.
						HandleManyToMany(context, scope, meta, record, metaValue, field, fieldHasVersion)
					}
				})

				return
			}
		}
	}

	field := reflect.Indirect(reflect.ValueOf(record)).FieldByName(fieldName)
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(utils.NewValue(field.Type().Elem()))
		}
		field = field.Elem()
	}

	if !field.IsValid() {
		return
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
			field.SetInt(utils.ToInt(metaValue.Value))
		})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
			field.SetUint(utils.ToUint(metaValue.Value))
		})
	case reflect.Float32, reflect.Float64:
		meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
			field.SetFloat(utils.ToFloat(metaValue.Value))
		})
	case reflect.Bool:
		meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
			if utils.ToString(metaValue.Value) == "true" {
				field.SetBool(true)
			} else {
				field.SetBool(false)
			}
		})
	default:
		if _, ok := field.Addr().Interface().(sql.Scanner); ok {
			meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
				if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
					if metaValue.Value == nil && len(metaValue.MetaValues.Values) > 0 {
						decodeMetaValuesToField(meta.Resource, field, metaValue, context)
						return
					}

					if scanner.Scan(metaValue.Value) != nil {
						if err := scanner.Scan(utils.ToString(metaValue.Value)); err != nil {
							context.AddError(err)
							return
						}
					}
				}
			})
		} else if reflect.TypeOf("").ConvertibleTo(field.Type()) {
			meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
				field.Set(reflect.ValueOf(utils.ToString(metaValue.Value)).Convert(field.Type()))
			})
		} else if reflect.TypeOf([]string{}).ConvertibleTo(field.Type()) {
			meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
				field.Set(reflect.ValueOf(utils.ToArray(metaValue.Value)).Convert(field.Type()))
			})
		} else if _, ok := field.Addr().Interface().(*time.Time); ok {
			meta.Setter = commonSetter(func(field reflect.Value, metaValue *MetaValue, context *qor.Context, record interface{}) {
				if str := utils.ToString(metaValue.Value); str != "" {
					if newTime, err := utils.ParseTime(str, context); err == nil {
						field.Set(reflect.ValueOf(newTime))
					}
				} else {
					field.Set(reflect.Zero(field.Type()))
				}
			})
		}
	}
}

func getNestedModel(value interface{}, fieldName string, context *qor.Context) interface{} {
	model := reflect.Indirect(reflect.ValueOf(value))
	fields := strings.Split(fieldName, ".")
	for _, field := range fields[:len(fields)-1] {
		if model.CanAddr() {
			submodel := model.FieldByName(field)
			if context != nil && context.GetDB() != nil && context.GetDB().NewRecord(submodel.Interface()) && !context.GetDB().NewRecord(model.Addr().Interface()) {
				if submodel.CanAddr() {
					context.GetDB().Model(model.Addr().Interface()).Association(field).Find(submodel.Addr().Interface())
					model = submodel
				} else {
					break
				}
			} else {
				model = submodel
			}
		}
	}

	if model.CanAddr() {
		return model.Addr().Interface()
	}
	return nil
}

// fieldStructHasVersion determine if the given field is a struct
// if so, detect if it has publish2.Version integrated
func fieldIsStructAndHasVersion(field reflect.Value) bool {
	// If the field struct has version
	if field.Type().Kind() == reflect.Slice || field.Type().Kind() == reflect.Struct {
		underlyingType := field.Type()
		// If the field is a slice of struct, we retrive one element(struct) as a sample to determine whether it has version
		// e.g. []User -> User
		if field.Type().Kind() == reflect.Slice {
			underlyingType = underlyingType.Elem()
			if underlyingType.Kind() == reflect.Ptr {
				underlyingType = underlyingType.Elem()
			}
		}

		for i := 0; i < underlyingType.NumField(); i++ {
			if underlyingType.Field(i).Name == "Version" && underlyingType.Field(i).Type.String() == "publish2.Version" {
				return true
			}
		}
	}

	return false
}
