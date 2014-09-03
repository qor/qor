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
	Type   reflect.Type
	Values []interface{}
}

func (resolver *Resolver) AddDependency(dependency *Dependency) {
	Dependencies
}

func (resolver *Resolver) GetDependencies(dependency *Dependency) {
}

func (resolver *Resolver) Publish() {
	for _, record := range resolver.Records {
		reflectValue := reflect.ValueOf(record)
		scope := &gorm.Scope{Value: record}
		dependency = Dependency{Type: reflectValue.Type(), Value: []interface{}{scope.PrimaryKey()}}
		resolver.AddDependency(dependency)
	}

	// delete from products where products.id in (?)
	// insert into products (columns) select columns from products_draft where products_draft.id in (?);
}
