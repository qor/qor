package validations

import (
	"fmt"

	"database/sql"
	"database/sql/driver"
	"github.com/jinzhu/gorm"
	"reflect"
)

func NewError(resource interface{}, column, err string) error {
	return &Error{Resource: resource, Column: column, Message: err}
}

type Error struct {
	Resource interface{}
	Column   string
	Message  string
}

func (err Error) Label() string {
	scope := gorm.Scope{Value: err.Resource}
	return fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), err.Column)
}

func (err Error) Error() string {
	return fmt.Sprintf("%v", err.Message)
}

//********************************//
//           Validators           //
//********************************//

type resourceData struct {
	db         *gorm.DB
	scope      *gorm.Scope
	attr_name  string
	attr_value interface{}
	optionals  map[string]interface{}
}

func (rc *resourceData) GetTableName() string {
	if rc.optionals["table_name"] == nil {
		rc.optionals["table_name"] = rc.scope.GetModelStruct().TableName(rc.db)
	}
	return rc.optionals["table_name"].(string)
}

func (rc *resourceData) GetPKName() string {
	if rc.optionals["pk_name"] == nil {
		rc.optionals["pk_name"] = rc.scope.PrimaryKey()
	}
	return rc.optionals["pk_name"].(string)
}

func (rc *resourceData) GetResource() interface{} {
	return rc.scope.Value
}

type checker func(*resourceData)

func NewValidator(db *gorm.DB, resource interface{}) func(string, ...checker) {
	scope := &gorm.Scope{Value: resource}
	return func(attr_name string, checkers ...checker) {
		if fld, ok := scope.FieldByName(attr_name); ok {
			attr_value := fld.Field.Interface()
			resource_data := &resourceData{db, scope, attr_name, attr_value, make(map[string]interface{})}
			for _, chk := range checkers {
				chk(resource_data)
			}
		} else {
			panic("Attribute " + attr_name + "does not exist")
		}
	}
}

func RangeInt(mn, mx int64) checker {
	return func(rc *resourceData) {
		if val, valid := getIntVal(rc.attr_value); valid {
			if val < mn || val > mx {
				rc.db.AddError(NewError(rc.GetResource(), rc.attr_name, fmt.Sprintf("Must be between %d and %d", mn, mx)))
			}
		}
	}
}

func RangeFloat(mn, mx float64) checker {
	return func(rc *resourceData) {
		if val, valid := getFloatVal(rc.attr_value); valid {
			if val < mn || val > mx {
				rc.db.AddError(NewError(rc.GetResource(), rc.attr_name, fmt.Sprintf("Must be between %f and %f", mn, mx)))
			}
		}
	}
}

func LengthRange(mn, mx int) checker {
	return func(rc *resourceData) {
		if val, valid := getStringVal(rc.attr_value); valid {
			l := len(val)
			err := ""
			if l < mn {
				err = rc.attr_name + " must have at least " + string(mn) + "characters"
			} else if l > mx {
				err = rc.attr_name + " must have less than " + string(mx+1) + "characters"
			}

			if len(err) != 0 {
				rc.db.AddError(NewError(rc.GetResource(), rc.attr_name, err))
			}

		}
	}
}

func Unique() checker {
	return func(rc *resourceData) {
		query := make(map[string]interface{})
		query[gorm.ToDBName(rc.attr_name)] = rc.attr_value
		row := rc.db.Table(rc.GetTableName()).Where(query).Select(rc.GetPKName()).Row()
		var _pk interface{}
		if err := row.Scan(&_pk); err == nil {
			rc.db.AddError(NewError(rc.GetResource(), rc.attr_name, "Must be unique"))
		}
	}
}

func Present() checker {
	return func(rc *resourceData) {
		if val, _ := rc.attr_value.(driver.Valuer).Value(); val == nil {
			rc.db.AddError(NewError(rc.GetResource(), rc.attr_name, "Must be present"))
		}
	}
}

// Helpers

func getBoolVal(v interface{}) (bool, bool) {
	switch v.(type) {
	case bool:
		return v.(bool), true
	case sql.NullBool:
		tmp := v.(sql.NullBool)
		return tmp.Bool, tmp.Valid
	}
	panic("Boolean value expected")
}

func getIntVal(v interface{}) (int64, bool) {
	switch v.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), true
	case sql.NullInt64:
		tmp := v.(sql.NullInt64)
		return tmp.Int64, tmp.Valid
	}
	panic("Integer value expected")
}

func getFloatVal(v interface{}) (float64, bool) {
	switch v.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float(), true
	case sql.NullFloat64:
		tmp := v.(sql.NullFloat64)
		return tmp.Float64, tmp.Valid
	}
	panic("Float value expected")
}

func getStringVal(v interface{}) (string, bool) {
	switch v.(type) {
	case string:
		return v.(string), true
	case sql.NullString:
		tmp := v.(sql.NullString)
		return tmp.String, tmp.Valid
	}
	panic("String value expected")
}
