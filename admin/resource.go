package admin

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/qor/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/utils"
)

type Resource struct {
	resource.Resource
	admin          *Admin
	Config         *Config
	Metas          []*Meta
	actions        []*Action
	scopes         []*Scope
	filters        map[string]*Filter
	searchAttrs    *[]string
	sortableAttrs  *[]string
	indexAttrs     []string
	newAttrs       []*Section
	editAttrs      []*Section
	IsSetShowAttrs bool
	showAttrs      []string
	cachedMetas    *map[string][]*Meta
	SearchHandler  func(keyword string, context *qor.Context) *gorm.DB
}

func (res *Resource) Meta(meta *Meta) {
	if res.GetMeta(meta.Name) != nil {
		utils.ExitWithMsg("Duplicated meta %v defined for resource %v", meta.Name, res.Name)
	}

	meta.base = res
	meta.updateMeta()
	res.Metas = append(res.Metas, meta)
}

func (res Resource) GetAdmin() *Admin {
	return res.admin
}

func (res Resource) ToParam() string {
	if res.Config.Singleton == true {
		return utils.ToParamString(res.Name)
	}
	return utils.ToParamString(inflection.Plural(res.Name))
}

func (res Resource) UseTheme(theme string) []string {
	if res.Config != nil {
		res.Config.Themes = append(res.Config.Themes, theme)
		return res.Config.Themes
	}
	return []string{}
}

func (res *Resource) convertObjectToMap(context *Context, value interface{}, kind string) interface{} {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	switch reflectValue.Kind() {
	case reflect.Slice:
		values := []interface{}{}
		for i := 0; i < reflectValue.Len(); i++ {
			values = append(values, res.convertObjectToMap(context, reflectValue.Index(i).Interface(), kind))
		}
		return values
	case reflect.Struct:
		var metas []*Meta
		if kind == "index" {
			metas = res.indexMetas()
		} else if kind == "show" {
			metas = res.showMetas()
		}

		values := map[string]interface{}{}
		for _, meta := range metas {
			if meta.HasPermission(roles.Read, context.Context) {
				if valuer := meta.GetValuer(); valuer != nil {
					value := valuer(value, context.Context)
					if meta.Resource != nil {
						value = meta.Resource.(*Resource).convertObjectToMap(context, value, kind)
					}
					values[meta.GetName()] = value
				}
			}
		}
		return values
	default:
		panic(fmt.Sprintf("Can't convert %v (%v) to map", reflectValue, reflectValue.Kind()))
	}
}

func (res *Resource) Decode(context *qor.Context, value interface{}) error {
	return resource.Decode(context, value, res)
}

func (res *Resource) allAttrs() []string {
	var attrs []string
	scope := &gorm.Scope{Value: res.Value}

Fields:
	for _, field := range scope.GetModelStruct().StructFields {
		for _, meta := range res.Metas {
			if field.Name == meta.FieldName {
				attrs = append(attrs, meta.Name)
				continue Fields
			}
		}

		if field.IsForeignKey {
			continue
		}

		for _, value := range []string{"CreatedAt", "UpdatedAt", "DeletedAt"} {
			if value == field.Name {
				continue Fields
			}
		}

		if (field.IsNormal || field.Relationship != nil) && !field.IsIgnored {
			attrs = append(attrs, field.Name)
		}
	}

MetaIncluded:
	for _, meta := range res.Metas {
		for _, attr := range attrs {
			if attr == meta.FieldName || attr == meta.Name {
				continue MetaIncluded
			}
		}
		attrs = append(attrs, meta.Name)
	}

	return attrs
}

func (res *Resource) getAttrs(attrs []string) []string {
	if len(attrs) == 0 {
		return res.allAttrs()
	} else {
		var onlyExcludeAttrs = true
		for _, attr := range attrs {
			if !strings.HasPrefix(attr, "-") {
				onlyExcludeAttrs = false
				break
			}
		}
		if onlyExcludeAttrs {
			return append(res.allAttrs(), attrs...)
		}
		return attrs
	}
}

func (res *Resource) getAttrs1(attrs []*Section) []*Section {
	if len(attrs) == 0 {
		var sections []*Section
		for _, attr := range res.allAttrs() {
			sections = append(sections, &Section{Columns: [][]string{{attr}}})
		}
		return sections
	} else {
		var onlyExcludeAttrs = true
		for _, attr := range attrs {
			attrName := attr.Columns[0][0]
			if !strings.HasPrefix(attrName, "-") {
				onlyExcludeAttrs = false
				break
			}
		}
		if onlyExcludeAttrs {
			var sections []*Section
			for _, attr := range res.allAttrs() {
				sections = append(sections, &Section{Columns: [][]string{{attr}}})
			}
			return append(sections, attrs...)
		}
		return attrs
	}
}

func (res *Resource) IndexAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.indexAttrs = columns
	}
	return res.getAttrs(res.indexAttrs)
}

func (res *Resource) NewAttrs(values ...interface{}) []*Section {
	res.setSections(&res.newAttrs, values...)
	return res.newAttrs
}

func (res *Resource) EditAttrs(values ...interface{}) []*Section {
	res.setSections(&res.editAttrs, values...)
	return res.editAttrs
}

func (res *Resource) ShowAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.IsSetShowAttrs = true
		res.showAttrs = columns
	}
	return res.getAttrs(res.showAttrs)
}

func (res *Resource) SortableAttrs(columns ...string) []string {
	if len(columns) != 0 || res.sortableAttrs == nil {
		if len(columns) == 0 {
			columns = res.indexAttrs
		}
		res.sortableAttrs = &[]string{}
		scope := res.GetAdmin().Config.DB.NewScope(res.Value)
		for _, column := range columns {
			if field, ok := scope.FieldByName(column); ok && field.DBName != "" {
				attrs := append(*res.sortableAttrs, column)
				res.sortableAttrs = &attrs
			}
		}
	}
	return *res.sortableAttrs
}

func (res *Resource) SearchAttrs(columns ...string) []string {
	if len(columns) != 0 || res.searchAttrs == nil {
		if len(columns) == 0 {
			columns = res.IndexAttrs()
		}

		if len(columns) > 0 {
			res.searchAttrs = &columns
			res.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
				db := context.GetDB()
				var joinConditionsMap = map[string][]string{}
				var conditions []string
				var keywords []interface{}
				scope := db.NewScope(res.Value)

				for _, column := range columns {
					currentScope, nextScope := scope, scope

					if strings.Contains(column, ".") {
						for _, field := range strings.Split(column, ".") {
							column = field
							currentScope = nextScope
							if field, ok := scope.FieldByName(field); ok {
								if relationship := field.Relationship; relationship != nil {
									nextScope = currentScope.New(reflect.New(field.Field.Type()).Interface())
									key := fmt.Sprintf("LEFT JOIN %v ON", nextScope.TableName())

									for index := range relationship.ForeignDBNames {
										if relationship.Kind == "has_one" || relationship.Kind == "has_many" {
											joinConditionsMap[key] = append(joinConditionsMap[key],
												fmt.Sprintf("%v.%v = %v.%v",
													nextScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
													currentScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
												))
										} else if relationship.Kind == "belongs_to" {
											joinConditionsMap[key] = append(joinConditionsMap[key],
												fmt.Sprintf("%v.%v = %v.%v",
													currentScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
													nextScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
												))
										}
									}
								}
							}
						}
					}

					var tableName = currentScope.Quote(currentScope.TableName())
					if field, ok := currentScope.FieldByName(column); ok && field.IsNormal {
						switch field.Field.Kind() {
						case reflect.String:
							conditions = append(conditions, fmt.Sprintf("upper(%v.%v) like upper(?)", tableName, scope.Quote(field.DBName)))
							keywords = append(keywords, "%"+keyword+"%")
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							if _, err := strconv.Atoi(keyword); err == nil {
								conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
								keywords = append(keywords, keyword)
							}
						case reflect.Float32, reflect.Float64:
							if _, err := strconv.ParseFloat(keyword, 64); err == nil {
								conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
								keywords = append(keywords, keyword)
							}
						case reflect.Bool:
							if value, err := strconv.ParseBool(keyword); err == nil {
								conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
								keywords = append(keywords, value)
							}
						case reflect.Struct:
							// time ?
							if _, ok := field.Field.Interface().(time.Time); ok {
								if parsedTime, err := now.Parse(keyword); err == nil {
									conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
									keywords = append(keywords, parsedTime)
								}
							}
						case reflect.Ptr:
							// time ?
							if _, ok := field.Field.Interface().(*time.Time); ok {
								if parsedTime, err := now.Parse(keyword); err == nil {
									conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
									keywords = append(keywords, parsedTime)
								}
							}
						default:
							conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
							keywords = append(keywords, keyword)
						}
					}
				}

				// join conditions
				if len(joinConditionsMap) > 0 {
					var joinConditions []string
					for key, values := range joinConditionsMap {
						joinConditions = append(joinConditions, fmt.Sprintf("%v %v", key, strings.Join(values, " AND ")))
					}
					db = db.Joins(strings.Join(joinConditions, " "))
				}

				// search conditions
				if len(conditions) > 0 {
					return db.Where(strings.Join(conditions, " OR "), keywords...)
				} else {
					return db
				}
			}
		}
	}

	return columns
}

func (res *Resource) getCachedMetas(cacheKey string, fc func() []resource.Metaor) []*Meta {
	if res.cachedMetas == nil {
		res.cachedMetas = &map[string][]*Meta{}
	}

	if values, ok := (*res.cachedMetas)[cacheKey]; ok {
		return values
	} else {
		values := fc()
		var metas []*Meta
		for _, value := range values {
			metas = append(metas, value.(*Meta))
		}
		(*res.cachedMetas)[cacheKey] = metas
		return metas
	}
}

func (res *Resource) GetMetas(attrs []string) []resource.Metaor {
	if len(attrs) == 0 {
		attrs = res.allAttrs()
	}
	var showAttrs, ignoredAttrs []string
	for _, attr := range attrs {
		if strings.HasPrefix(attr, "-") {
			ignoredAttrs = append(ignoredAttrs, strings.TrimLeft(attr, "-"))
		} else {
			showAttrs = append(showAttrs, attr)
		}
	}

	primaryKey := res.PrimaryFieldName()

	metas := []resource.Metaor{}

Attrs:
	for _, attr := range showAttrs {
		for _, a := range ignoredAttrs {
			if attr == a {
				continue Attrs
			}
		}

		var meta *Meta
		for _, m := range res.Metas {
			if m.GetName() == attr {
				meta = m
				break
			}
		}

		if meta == nil {
			meta = &Meta{}
			meta.Name = attr
			meta.base = res
			if attr == primaryKey {
				meta.Type = "hidden"
			}
			meta.updateMeta()
		}

		metas = append(metas, meta)
	}

	return metas
}

func (res *Resource) GetMeta(name string) *Meta {
	for _, meta := range res.Metas {
		if meta.Name == name || meta.GetFieldName() == name {
			return meta
		}
	}
	return nil
}

func (res *Resource) GetMetaOrNew(name string) *Meta {
	for _, meta := range res.Metas {
		if meta.Name == name || meta.GetFieldName() == name {
			return meta
		}
	}
	for _, meta := range res.allMetas() {
		if meta.Name == name || meta.GetFieldName() == name {
			return meta
		}
	}
	return nil
}

func (res *Resource) indexMetas() []*Meta {
	return res.getCachedMetas("index_metas", func() []resource.Metaor {
		return res.GetMetas(res.IndexAttrs())
	})
}

/*func (res *Resource) newMetas() []*Meta {
	return res.getCachedMetas("new_metas", func() []resource.Metaor {
		return res.GetMetas(res.NewAttrs())
	})
}*/

/*func (res *Resource) editMetas() []*Meta {
	return res.getCachedMetas("edit_metas", func() []resource.Metaor {
		return res.GetMetas(res.EditAttrs())
	})
}*/

func (res *Resource) showMetas() []*Meta {
	return res.getCachedMetas("show_metas", func() []resource.Metaor {
		return res.GetMetas(res.ShowAttrs())
	})
}

func (res *Resource) allMetas() []*Meta {
	return res.getCachedMetas("all_metas", func() []resource.Metaor {
		return res.GetMetas([]string{})
	})
}

func (res *Resource) allowedMetas(attrs []*Meta, context *Context, roles ...roles.PermissionMode) []*Meta {
	var metas = []*Meta{}
	for _, meta := range attrs {
		for _, role := range roles {
			if meta.HasPermission(role, context.Context) {
				metas = append(metas, meta)
				break
			}
		}
	}
	return metas
}

func (res *Resource) allowedMetas1(sections []*Section, context *Context, roles ...roles.PermissionMode) []*Section {
	for _, section := range sections {
		var editableRows [][]string
		for _, row := range section.Columns {
			var editableColumns []string
			for _, column := range row {
				for _, role := range roles {
					meta := res.GetMeta(column)
					if true || (meta != nil && meta.HasPermission(role, context.Context)) {
						editableColumns = append(editableColumns, column)
						break
					}
				}
			}
			if len(editableColumns) > 0 {
				editableRows = append(editableRows, editableColumns)
			}
		}
		section.Columns = editableRows
	}
	return sections
}

// Section Related Methods
func appendSectionFromStrings(columns []string) []*Section {
	var sections []*Section
	for _, column := range columns {
		sections = append(sections, &Section{Columns: [][]string{{column}}})
	}
	return sections
}

func appendSectionFromInterfaces(values ...interface{}) []*Section {
	var sections []*Section
	var hasColumns []string
	var excludedColumns []string
	valueSize := len(values)
	for i := 0; i < len(values); i++ {
		value := values[valueSize-i-1]
		if section, ok := value.(*Section); ok {
			sections = append(sections, uniqueSection(section, hasColumns))
		} else if column, ok := value.(string); ok {
			if strings.HasPrefix(column, "-") {
				excludedColumns = append(excludedColumns, column)
			} else {
				if !isContainsColumn(excludedColumns, column) {
					sections = append(sections, &Section{Columns: [][]string{{column}}})
				}
			}
			hasColumns = append(hasColumns, column)
		} else {
			panic(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}
	return reverseSections(sections)
}

func uniqueSection(section *Section, hasColumns []string) *Section {
	newSection := Section{Title: section.Title}
	var newRows [][]string
	for _, row := range section.Columns {
		var newColumns []string
		for _, column := range row {
			if !isContainsColumn(hasColumns, column) {
				newColumns = append(newColumns, column)
				hasColumns = append(hasColumns, column)
			}
		}
		if len(newColumns) > 0 {
			newRows = append(newRows, newColumns)
		}
	}
	newSection.Columns = newRows
	return &newSection
}

func reverseSections(sections []*Section) []*Section {
	var results []*Section
	for i := 0; i < len(sections); i++ {
		results = append(results, sections[len(sections)-i-1])
	}
	return results
}

func isContainsColumn(hasColumns []string, column string) bool {
	for _, col := range hasColumns {
		if col == column || ("-"+col == column) || ("-"+column == col) {
			return true
		}
	}
	return false
}

func isSectionsAllPositive(values ...interface{}) bool {
	for _, value := range values {
		if _, ok := value.(*Section); ok {
			return true
		} else if column, ok := value.(string); ok {
			if !strings.HasPrefix(column, "-") {
				return true
			}
		} else {
			panic(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}
	return false
}

func (res *Resource) setSections(sections *[]*Section, values ...interface{}) {
	if len(*sections) > 0 && len(values) == 0 {
		return
	}
	if len(*sections) == 0 && len(values) == 0 {
		*sections = appendSectionFromStrings(res.allAttrs())
	} else {
		var flattenValues []interface{}
		for _, value := range values {
			if columns, ok := value.([]string); ok {
				for _, column := range columns {
					flattenValues = append(flattenValues, column)
				}
			} else if _sections, ok := value.([]*Section); ok {
				for _, section := range _sections {
					flattenValues = append(flattenValues, section)
				}
			} else if section, ok := value.(*Section); ok {
				flattenValues = append(flattenValues, section)
			} else if column, ok := value.(string); ok {
				flattenValues = append(flattenValues, column)
			} else {
				panic(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
			}
		}
		if isSectionsAllPositive(flattenValues...) {
			*sections = appendSectionFromInterfaces(flattenValues...)
		} else {
			var valueStrs []string
			var availbleColumns []string
			for _, value := range flattenValues {
				if column, ok := value.(string); ok {
					valueStrs = append(valueStrs, column)
				}
			}

			for _, column := range res.allAttrs() {
				if !isContainsColumn(valueStrs, column) {
					availbleColumns = append(availbleColumns, column)
				}
			}
			*sections = appendSectionFromStrings(availbleColumns)
		}
	}
}
