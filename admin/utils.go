package admin

import (
	"fmt"
	"reflect"
)

func equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func equalAsString(a interface{}, b interface{}) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}
