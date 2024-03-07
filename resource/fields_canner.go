package resource

import "reflect"

// FieldScanner interface
// Allow to scan value with reflect.StructField arg
type FieldScanner interface {
	// FieldScan scan value
	FieldScan(field *reflect.StructField, src interface{}) error
}
