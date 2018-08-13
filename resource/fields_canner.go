package resource

import "reflect"

type FieldScanner interface {
	FieldScan(field *reflect.StructField, src interface{}) error
}
