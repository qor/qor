package publish

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

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
			if !IncludeValue(primaryKey, dep.PrimaryKeys) {
				primaryKeys = append(primaryKeys, primaryKey)
				dep.PrimaryKeys = append(dep.PrimaryKeys, primaryKey)
			}
		}
	} else {
		resolver.Dependencies[name] = dependency
		primaryKeys = dependency.PrimaryKeys
	}

	if len(primaryKeys) > 0 {
		resolver.GetDependencies(dependency, primaryKeys)
	}
}

func (resolver *Resolver) GetDependencies(dependency *Dependency, primaryKeys []string) {
	// new slice
	value := reflect.New(dependency.Type)
	fromScope := resolver.DB.NewScope(value.Addr().Interface())
	for _, field := range fromScope.Fields() {
		if relationship := field.Relationship; relationship != nil {
			toType := field.Field.Type()
			toScope := resolver.DB.NewScope(reflect.New(toType).Interface())
			var dependencyKeys []string
			var rows *sql.Rows
			var err error

			if relationship.Kind == "belongs_to" || relationship.Kind == "has_many" {
				sql := fmt.Sprintf("%v IN (?) and publish_status = ?", relationship.ForeignKey)
				rows, err = resolver.DB.Select(toScope.PrimaryKey()).Where(sql, primaryKeys, DIRTY).Rows()
			} else if relationship.Kind == "has_one" {
				fromTable := fromScope.TableName()
				fromPrimaryKey := fromScope.PrimaryKey()
				toTable := toScope.TableName()
				toPrimaryKey := toScope.PrimaryKey()

				sql := fmt.Sprintf("%v.%v IN (select %v.%v from %v where %v.%v IN (?)) and %v.publish_status = ?",
					toTable, toPrimaryKey, fromTable, relationship.ForeignKey, fromTable, fromTable, fromPrimaryKey, toTable)

				rows, err = resolver.DB.Select(toTable+"."+toPrimaryKey).Where(sql, primaryKeys, DIRTY).Rows()
			} else if relationship.Kind == "many_to_many" {
			}

			if rows != nil && err == nil {
				for rows.Next() {
					var primaryKey interface{}
					rows.Scan(primaryKey)
					dependencyKeys = append(dependencyKeys, fmt.Sprintf("%v", primaryKey))
				}
			}
			if len(dependencyKeys) > 0 {
				dependency := Dependency{Type: toType, PrimaryKeys: dependencyKeys}
				resolver.AddDependency(&dependency)
			}
		}
	}
}

func (resolver *Resolver) Publish() {
	for _, record := range resolver.Records {
		var supportedModels []string
		var reflectType = modelType(record)

		if value, ok := resolver.DB.Get("publish:support_models"); ok {
			supportedModels = value.([]string)
		}

		if IncludeValue(reflectType.String(), supportedModels) {
			scope := &gorm.Scope{Value: record}
			dependency := Dependency{Type: reflectType, PrimaryKeys: []string{fmt.Sprintf("%v", scope.PrimaryKeyValue())}}
			resolver.AddDependency(&dependency)
			break
		}
	}

	for _, dependency := range resolver.Dependencies {
		value := reflect.New(dependency.Type)
		fromScope := resolver.DB.DraftMode().NewScope(value)
		fromTable := fromScope.QuotedTableName()
		fromPrimaryKey := fromScope.PrimaryKey()
		toScope := resolver.DB.ProductionMode().NewScope(value)
		toTable := toScope.QuotedTableName()

		resolver.DB.ProductionMode().Delete(value.Interface(), dependency.PrimaryKeys)

		var columns []string
		for _, field := range toScope.Fields() {
			columns = append(columns, field.DBName)
		}

		var insertColumns []string
		for _, column := range columns {
			insertColumns = append(insertColumns, fmt.Sprintf("%v.%v", toTable, column))
		}

		var selectColumns []string
		for _, column := range columns {
			selectColumns = append(selectColumns, fmt.Sprintf("%v.%v", fromTable, column))
		}

		sql := fmt.Sprintf("INSERT INTO %v (%v) SELECT %v from %v where %v.%v in (?);",
			toTable, strings.Join(insertColumns, " ,"), strings.Join(selectColumns, " ,"),
			fromTable, fromTable, fromPrimaryKey)

		resolver.DB.Exec(sql, dependency.PrimaryKeys)
	}
}
