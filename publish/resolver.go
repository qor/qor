package publish

import (
	"reflect"

	"github.com/jinzhu/gorm"
)

type Resolver struct {
	Records      []interface{}
	Dependencies map[string]*Dependency
	DB           *DB
}

type Dependency struct {
	Type        reflect.Type
	PrimaryKeys []string
}

func IncludeValue(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func (resolver *Resolver) AddDependency(dependency *Dependency) {
	name := dependency.Type.String()
	var primaryKeys []string

	if dep, ok := resolver.Dependencies[name]; ok {
		for _, primaryKey := range dependency.PrimaryKeys {
			if !IncludeValue(value, dep.PrimaryKeys) {
				dep.PrimaryKeys = append(dep.PrimaryKeys, primaryKey)
			}
		}
	} else {
		resolver.Dependencies[name] = dependency
	}
}

func (resolver *Resolver) GetDependencies(dependency *Dependency, primaryKeys []string) {
}

func (resolver *Resolver) Publish() {
	for _, record := range resolver.Records {
		reflectValue := reflect.ValueOf(record)
		scope := &gorm.Scope{Value: record}
		dependency := Dependency{Type: reflectValue.Type(), PrimaryKeys: []string{scope.PrimaryKey()}}
		resolver.AddDependency(dependency)
	}

	// delete from products where products.id in (?)
	// insert into products (columns) select columns from products_draft where products_draft.id in (?);
}
