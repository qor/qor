package publish

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

type resolver struct {
	Records      []interface{}
	Dependencies map[string]*dependency
	DB           *Publish
}

type dependency struct {
	Type                reflect.Type
	ManyToManyRelations []*gorm.Relationship
	PrimaryValues       [][]interface{}
}

func includeValue(value []interface{}, values [][]interface{}) bool {
	for _, v := range values {
		if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", value) {
			return true
		}
	}
	return false
}

func (resolver *resolver) AddDependency(dep *dependency) {
	name := dep.Type.String()
	var newPrimaryKeys [][]interface{}

	// append primary keys to dependency
	if d, ok := resolver.Dependencies[name]; ok {
		for _, primaryKey := range dep.PrimaryValues {
			if !includeValue(primaryKey, d.PrimaryValues) {
				newPrimaryKeys = append(newPrimaryKeys, primaryKey)
				dep.PrimaryValues = append(d.PrimaryValues, primaryKey)
			}
		}
	} else {
		resolver.Dependencies[name] = dep
		newPrimaryKeys = dep.PrimaryValues
	}

	if len(newPrimaryKeys) > 0 {
		resolver.GetDependencies(dep, newPrimaryKeys)
	}
}

func (resolver *resolver) GetDependencies(dep *dependency, primaryKeys [][]interface{}) {
	value := reflect.New(dep.Type)
	fromScope := resolver.DB.DB.NewScope(value.Interface())

	draftDB := resolver.DB.DraftDB().Unscoped()
	for _, field := range fromScope.Fields() {
		if relationship := field.Relationship; relationship != nil {
			if isPublishableModel(field.Field.Interface()) {
				toType := modelType(field.Field.Interface())
				toScope := draftDB.NewScope(reflect.New(toType).Interface())
				draftTable := draftTableName(toScope.TableName())
				var dependencyKeys [][]interface{}
				var rows *sql.Rows
				var err error

				if relationship.Kind == "belongs_to" || relationship.Kind == "has_many" {
					sql := fmt.Sprintf("%v IN (?) and publish_status = ?", relationship.ForeignDBName)
					rows, err = draftDB.Table(draftTable).Select(toScope.PrimaryKey()).Where(sql, primaryKeys, DIRTY).Rows()
				} else if relationship.Kind == "has_one" {
					fromTable := fromScope.TableName()
					fromPrimaryKey := fromScope.PrimaryKey()
					toTable := toScope.TableName()
					toPrimaryKey := toScope.PrimaryKey()

					sql := fmt.Sprintf("%v.%v IN (select %v.%v from %v where %v.%v IN (?)) and %v.publish_status = ?",
						toTable, toPrimaryKey, fromTable, relationship.ForeignDBName, fromTable, fromTable, fromPrimaryKey, toTable)

					rows, err = draftDB.Table(draftTable).Select(toTable+"."+toPrimaryKey).Where(sql, primaryKeys, DIRTY).Rows()
				}

				if rows != nil && err == nil {
					for rows.Next() {
						var primaryValues = make([]interface{}, len(toScope.PrimaryFields()))
						rows.Scan(primaryValues...)
						dependencyKeys = append(dependencyKeys, primaryValues)
					}

					resolver.AddDependency(&dependency{Type: toType, PrimaryValues: dependencyKeys})
				}
			}

			if relationship.Kind == "many_to_many" {
				dep.ManyToManyRelations = append(dep.ManyToManyRelations, relationship)
			}
		}
	}
}

func (resolver *resolver) GenerateDependencies() {
	var addToDependencies = func(data interface{}) {
		if isPublishableModel(data) {
			scope := resolver.DB.DB.NewScope(data)
			var primaryValues []interface{}
			for _, field := range scope.PrimaryFields() {
				primaryValues = append(primaryValues, field.Field.Interface())
			}
			resolver.AddDependency(&dependency{Type: modelType(data), PrimaryValues: [][]interface{}{primaryValues}})
		}
	}

	for _, record := range resolver.Records {
		reflectValue := reflect.Indirect(reflect.ValueOf(record))
		if reflectValue.Kind() == reflect.Slice {
			for i := 0; i < reflectValue.Len(); i++ {
				addToDependencies(reflectValue.Index(i).Interface())
			}
		} else {
			addToDependencies(record)
		}
	}
}

func (resolver *resolver) Publish() error {
	resolver.GenerateDependencies()
	tx := resolver.DB.DB.Begin()

	for _, dep := range resolver.Dependencies {
		value := reflect.New(dep.Type).Elem()
		productionScope := resolver.DB.ProductionDB().NewScope(value.Addr().Interface())
		productionTable := productionScope.TableName()
		draftTable := draftTableName(productionTable)
		productionPrimaryKey := scopePrimaryKeys(productionScope, productionTable)
		draftPrimaryKey := scopePrimaryKeys(productionScope, draftTable)

		var columns []string
		for _, field := range productionScope.Fields() {
			if field.IsNormal {
				columns = append(columns, field.DBName)
			}
		}

		var productionColumns, draftColumns []string
		for _, column := range columns {
			productionColumns = append(productionColumns, fmt.Sprintf("%v.%v", productionTable, column))
			draftColumns = append(draftColumns, fmt.Sprintf("%v.%v", draftTable, column))
		}

		if len(dep.PrimaryValues) > 0 {
			// delete old records
			deleteSql := fmt.Sprintf("DELETE FROM %v WHERE %v IN (%v)", productionTable, productionPrimaryKey, toQueryMarks(dep.PrimaryValues))
			tx.Exec(deleteSql, toQueryValues(dep.PrimaryValues)...)

			// insert new records
			publishSql := fmt.Sprintf("INSERT INTO %v (%v) SELECT %v FROM %v WHERE %v IN (%v)",
				productionTable, strings.Join(productionColumns, " ,"), strings.Join(draftColumns, " ,"),
				draftTable, draftPrimaryKey, toQueryMarks(dep.PrimaryValues))
			tx.Exec(publishSql, toQueryValues(dep.PrimaryValues)...)

			// publish join table data
			for _, relationship := range dep.ManyToManyRelations {
				productionTable := relationship.JoinTableHandler.Table(tx.Set("publish:draft_mode", false))
				draftTable := relationship.JoinTableHandler.Table(tx.Set("publish:draft_mode", true))
				var productionJoinKeys, draftJoinKeys []string
				var productionCondition, draftCondition string
				for _, foreignKey := range relationship.JoinTableHandler.SourceForeignKeys() {
					productionJoinKeys = append(productionJoinKeys, fmt.Sprintf("%v.%v", productionTable, productionScope.Quote(foreignKey.DBName)))
					draftJoinKeys = append(draftJoinKeys, fmt.Sprintf("%v.%v", draftTable, productionScope.Quote(foreignKey.DBName)))
				}

				if len(productionJoinKeys) > 1 {
					productionCondition = fmt.Sprintf("(%v)", strings.Join(productionJoinKeys, ","))
					draftCondition = fmt.Sprintf("(%v)", strings.Join(draftJoinKeys, ","))
				} else {
					productionCondition = strings.Join(productionJoinKeys, ",")
					draftCondition = strings.Join(draftJoinKeys, ",")
				}

				rows, _ := tx.Raw(fmt.Sprintf("select * from %v", draftTable)).Rows()
				joinColumns, _ := rows.Columns()
				rows.Close()
				var productionJoinTableColumns, draftJoinTableColumns []string
				for _, column := range joinColumns {
					productionJoinTableColumns = append(productionJoinTableColumns, fmt.Sprintf("%v.%v", productionTable, column))
					draftJoinTableColumns = append(draftJoinTableColumns, fmt.Sprintf("%v.%v", draftTable, column))
				}

				sql := fmt.Sprintf("DELETE FROM %v WHERE %v IN (%v)", productionTable, productionCondition, toQueryMarks(dep.PrimaryValues))
				tx.Exec(sql, toQueryValues(dep.PrimaryValues)...)

				publishSql := fmt.Sprintf("INSERT INTO %v (%v) SELECT %v FROM %v WHERE %v IN (%v)",
					productionTable, strings.Join(productionJoinTableColumns, " ,"), strings.Join(draftJoinTableColumns, " ,"),
					draftTable, draftCondition, toQueryMarks(dep.PrimaryValues))
				tx.Exec(publishSql, toQueryValues(dep.PrimaryValues)...)
			}

			// set status to published
			updateStateSql := fmt.Sprintf("UPDATE %v SET publish_status = ? WHERE %v IN (%v)", draftTable, draftPrimaryKey, toQueryMarks(dep.PrimaryValues))

			var params = []interface{}{bool(PUBLISHED)}
			params = append(params, toQueryValues(dep.PrimaryValues)...)
			tx.Exec(updateStateSql, params...)
		}
	}

	if err := tx.Error; err == nil {
		return tx.Commit().Error
	} else {
		tx.Rollback()
		return err
	}
}

func (resolver *resolver) Discard() error {
	resolver.GenerateDependencies()
	tx := resolver.DB.DB.Begin()

	for _, dep := range resolver.Dependencies {
		value := reflect.New(dep.Type).Elem()
		productionScope := resolver.DB.ProductionDB().NewScope(value.Addr().Interface())
		productionTable := productionScope.TableName()
		draftTable := draftTableName(productionTable)

		productionPrimaryKey := scopePrimaryKeys(productionScope, productionTable)
		draftPrimaryKey := scopePrimaryKeys(productionScope, draftTable)

		var columns []string
		for _, field := range productionScope.Fields() {
			if field.IsNormal {
				columns = append(columns, field.DBName)
			}
		}

		var productionColumns, draftColumns []string
		for _, column := range columns {
			productionColumns = append(productionColumns, fmt.Sprintf("%v.%v", productionTable, column))
			draftColumns = append(draftColumns, fmt.Sprintf("%v.%v", draftTable, column))
		}

		// delete data from draft db
		deleteSql := fmt.Sprintf("DELETE FROM %v WHERE %v IN (%v)", draftTable, draftPrimaryKey, toQueryMarks(dep.PrimaryValues))
		tx.Exec(deleteSql, toQueryValues(dep.PrimaryValues)...)

		// delete join table
		for _, relationship := range dep.ManyToManyRelations {
			productionTable := relationship.JoinTableHandler.Table(tx.Set("publish:draft_mode", false))
			draftTable := relationship.JoinTableHandler.Table(tx.Set("publish:draft_mode", true))
			var productionJoinKeys, draftJoinKeys []string
			var productionCondition, draftCondition string
			for _, foreignKey := range relationship.JoinTableHandler.SourceForeignKeys() {
				productionJoinKeys = append(productionJoinKeys, fmt.Sprintf("%v.%v", productionTable, productionScope.Quote(foreignKey.DBName)))
				draftJoinKeys = append(draftJoinKeys, fmt.Sprintf("%v.%v", draftTable, productionScope.Quote(foreignKey.DBName)))
			}

			if len(productionJoinKeys) > 1 {
				productionCondition = fmt.Sprintf("(%v)", strings.Join(productionJoinKeys, ","))
				draftCondition = fmt.Sprintf("(%v)", strings.Join(draftJoinKeys, ","))
			} else {
				productionCondition = strings.Join(productionJoinKeys, ",")
				draftCondition = strings.Join(draftJoinKeys, ",")
			}

			rows, _ := tx.Raw(fmt.Sprintf("select * from %v", draftTable)).Rows()
			joinColumns, _ := rows.Columns()
			rows.Close()
			var productionJoinTableColumns, draftJoinTableColumns []string
			for _, column := range joinColumns {
				productionJoinTableColumns = append(productionJoinTableColumns, fmt.Sprintf("%v.%v", productionTable, column))
				draftJoinTableColumns = append(draftJoinTableColumns, fmt.Sprintf("%v.%v", draftTable, column))
			}

			sql := fmt.Sprintf("DELETE FROM %v WHERE %v IN (%v)", draftTable, draftCondition, toQueryMarks(dep.PrimaryValues))
			tx.Exec(sql, toQueryValues(dep.PrimaryValues)...)

			publishSql := fmt.Sprintf("INSERT INTO %v (%v) SELECT %v FROM %v WHERE %v IN (%v)",
				draftTable, strings.Join(draftJoinTableColumns, " ,"), strings.Join(productionJoinTableColumns, " ,"),
				productionTable, productionCondition, toQueryMarks(dep.PrimaryValues))
			tx.Exec(publishSql, toQueryValues(dep.PrimaryValues)...)
		}

		// copy data from production to draft
		discardSql := fmt.Sprintf("INSERT INTO %v (%v) SELECT %v FROM %v WHERE %v IN (%v)",
			draftTable, strings.Join(draftColumns, " ,"),
			strings.Join(productionColumns, " ,"), productionTable,
			productionPrimaryKey, toQueryMarks(dep.PrimaryValues))
		tx.Exec(discardSql, toQueryValues(dep.PrimaryValues)...)
	}

	if err := tx.Error; err == nil {
		return tx.Commit().Error
	} else {
		tx.Rollback()
		return err
	}
}

func scopePrimaryKeys(scope *gorm.Scope, tableName string) string {
	var primaryKeys []string
	for _, field := range scope.PrimaryFields() {
		key := fmt.Sprintf("%v.%v", scope.Quote(tableName), scope.Quote(field.DBName))
		primaryKeys = append(primaryKeys, key)
	}
	if len(primaryKeys) > 1 {
		return fmt.Sprintf("(%v)", strings.Join(primaryKeys, ","))
	}
	return strings.Join(primaryKeys, "")
}

func toQueryMarks(primaryValues [][]interface{}) string {
	var results []string

	for _, primaryValue := range primaryValues {
		var marks []string
		for range primaryValue {
			marks = append(marks, "?")
		}

		if len(marks) > 1 {
			results = append(results, fmt.Sprintf("(%v)", strings.Join(marks, ",")))
		} else {
			results = append(results, strings.Join(marks, ""))
		}
	}
	return strings.Join(results, ",")
}

func toQueryValues(primaryValues [][]interface{}) (values []interface{}) {
	for _, primaryValue := range primaryValues {
		for _, value := range primaryValue {
			values = append(values, value)
		}
	}
	return values
}
