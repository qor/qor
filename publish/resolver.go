package publish

import (
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
	var newPrimaryKeys []string

	if dep, ok := resolver.Dependencies[name]; ok {
		for _, primaryKey := range dependency.PrimaryKeys {
			if !IncludeValue(primaryKey, dep.PrimaryKeys) {
				newPrimaryKeys = append(newPrimaryKeys, primaryKey)
				dep.PrimaryKeys = append(dep.PrimaryKeys, primaryKey)
			}
		}
	} else {
		resolver.Dependencies[name] = dependency
		newPrimaryKeys = dependency.PrimaryKeys
	}

	if len(newPrimaryKeys) > 0 {
		resolver.GetDependencies(dependency, newPrimaryKeys)
	}
}

func (resolver *Resolver) GetDependencies(dependency *Dependency, primaryKeys []string) {
	value := reflect.New(dependency.Type)
	fromScope := resolver.DB.DB.NewScope(value.Interface())

	draftDB := resolver.DB.DraftMode().Unscoped()
	for _, field := range fromScope.Fields() {
		if relationship := field.Relationship; relationship != nil {
			toType := modelType(field.Field.Interface())
			toScope := draftDB.NewScope(reflect.New(toType).Interface())
			var dependencyKeys []interface{}

			if relationship.Kind == "belongs_to" || relationship.Kind == "has_many" {
				sql := fmt.Sprintf("%v IN (?) and publish_status = ?", gorm.ToSnake(relationship.ForeignKey))
				draftDB.Model(toScope.Value).Select(toScope.PrimaryKey()).Where(sql, primaryKeys, DIRTY).Scan(&dependencyKeys)
			} else if relationship.Kind == "has_one" {
				fromTable := fromScope.TableName()
				fromPrimaryKey := fromScope.PrimaryKey()
				toTable := toScope.TableName()
				toPrimaryKey := toScope.PrimaryKey()

				sql := fmt.Sprintf("%v.%v IN (select %v.%v from %v where %v.%v IN (?)) and %v.publish_status = ?",
					toTable, toPrimaryKey, fromTable, relationship.ForeignKey, fromTable, fromTable, fromPrimaryKey, toTable)

				draftDB.Model(toScope.Value).Select(toTable+"."+toPrimaryKey).Where(sql, primaryKeys, DIRTY).Find(&dependencyKeys)
			} else if relationship.Kind == "many_to_many" {
			}

			if len(dependencyKeys) > 0 {
				var primaryKeys []string
				for _, dependencyKey := range dependencyKeys {
					primaryKeys = append(primaryKeys, fmt.Sprintf("%v", dependencyKey))
				}

				dependency := Dependency{Type: toType, PrimaryKeys: primaryKeys}
				resolver.AddDependency(&dependency)
			}
		}
	}
}

func (resolver *Resolver) Publish() {
	for _, record := range resolver.Records {
		var supportedModels []string
		var reflectType = modelType(record)

		if value, ok := resolver.DB.DB.Get("publish:support_models"); ok {
			supportedModels = value.([]string)
		}

		if IncludeValue(reflectType.String(), supportedModels) {
			scope := &gorm.Scope{Value: record}
			dependency := Dependency{Type: reflectType, PrimaryKeys: []string{fmt.Sprintf("%v", scope.PrimaryKeyValue())}}
			resolver.AddDependency(&dependency)
		}
	}

	for _, dependency := range resolver.Dependencies {
		value := reflect.New(dependency.Type)
		productionScope := resolver.DB.ProductionMode().NewScope(value.Interface())
		productionTable := productionScope.TableName()
		productionPrimaryKey := productionScope.PrimaryKey()
		draftTable := DraftTableName(productionTable)

		var columns []string
		for _, field := range productionScope.Fields() {
			if field.IsNormal {
				columns = append(columns, field.DBName)
			}
		}

		var insertColumns []string
		for _, column := range columns {
			insertColumns = append(insertColumns, fmt.Sprintf("%v.%v", productionTable, column))
		}

		var selectColumns []string
		for _, column := range columns {
			selectColumns = append(selectColumns, fmt.Sprintf("%v.%v", draftTable, column))
		}

		deleteSql := fmt.Sprintf("DELETE FROM %v WHERE %v.%v IN (?);",
			productionTable, productionTable, productionScope.PrimaryKey())
		resolver.DB.DB.Exec(deleteSql, dependency.PrimaryKeys)

		sql := fmt.Sprintf("INSERT INTO %v (%v) SELECT %v from %v where %v.%v in (?);",
			productionTable, strings.Join(insertColumns, " ,"), strings.Join(selectColumns, " ,"),
			draftTable, draftTable, productionPrimaryKey)
		resolver.DB.DB.Exec(sql, dependency.PrimaryKeys)
	}
}
