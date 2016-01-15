package admin

import (
	"fmt"
	"github.com/qor/qor/utils"
	"strings"
)

type Section struct {
	Resource Resource
	Title    string
	Rows     [][]string
}

func (res *Resource) generateSections(values ...interface{}) []*Section {
	var sections []*Section
	var hasColumns []string
	var excludedColumns []string
	// Reverse values to make the last one as a key one
	// e.g. Name, Code, -Name (`-Name` will get first and will skip `Name`)
	for i := len(values) - 1; i >= 0; i-- {
		value := values[i]
		if section, ok := value.(*Section); ok {
			sections = append(sections, uniqueSection(section, &hasColumns))
		} else if column, ok := value.(string); ok {
			if strings.HasPrefix(column, "-") {
				excludedColumns = append(excludedColumns, column)
			} else if !isContainsColumn(excludedColumns, column) {
				sections = append(sections, &Section{Rows: [][]string{{column}}})
			}
			hasColumns = append(hasColumns, column)
		} else if row, ok := value.([]string); ok {
			for j := len(row) - 1; j >= 0; j-- {
				column = row[j]
				sections = append(sections, &Section{Rows: [][]string{{column}}})
				hasColumns = append(hasColumns, column)
			}
		} else {
			utils.ExitWithMsg(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}
	sections = reverseSections(sections)
	for _, section := range sections {
		section.Resource = *res
	}
	return sections
}

func uniqueSection(section *Section, hasColumns *[]string) *Section {
	newSection := Section{Title: section.Title}
	var newRows [][]string
	for _, row := range section.Rows {
		var newColumns []string
		for _, column := range row {
			if !isContainsColumn(*hasColumns, column) {
				newColumns = append(newColumns, column)
				*hasColumns = append(*hasColumns, column)
			}
		}
		if len(newColumns) > 0 {
			newRows = append(newRows, newColumns)
		}
	}
	newSection.Rows = newRows
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
		if strings.TrimLeft(col, "-") == strings.TrimLeft(column, "-") {
			return true
		}
	}
	return false
}

func isContainsPositiveValue(values ...interface{}) bool {
	for _, value := range values {
		if _, ok := value.(*Section); ok {
			return true
		} else if column, ok := value.(string); ok {
			if !strings.HasPrefix(column, "-") {
				return true
			}
		} else {
			utils.ExitWithMsg(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
		}
	}
	return false
}

func (res *Resource) ConvertSectionToMetas(sections []*Section) []*Meta {
	var metas []*Meta
	for _, section := range sections {
		for _, row := range section.Rows {
			for _, col := range row {
				meta := res.GetMetaOrNew(col)
				if meta != nil {
					metas = append(metas, meta)
				}
			}
		}
	}
	return metas
}

func (res *Resource) ConvertSectionToStrings(sections []*Section) []string {
	var columns []string
	for _, section := range sections {
		for _, row := range section.Rows {
			for _, col := range row {
				columns = append(columns, col)
			}
		}
	}
	return columns
}

func (res *Resource) setSections(sections *[]*Section, values ...interface{}) {
	if len(*sections) > 0 && len(values) == 0 {
		return
	}
	var excludedColumns = []string{"CreatedAt", "UpdatedAt", "DeletedAt"}
	if len(*sections) == 0 && len(values) == 0 {
		*sections = res.generateSections(res.allAttrs(excludedColumns...))
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
				utils.ExitWithMsg(fmt.Sprintf("Qor Resource: attributes should be Section or String, but it is %+v", value))
			}
		}
		if isContainsPositiveValue(flattenValues...) {
			// Contains Positive attributes will not get attribute from allAttrs
			*sections = res.generateSections(flattenValues...)
		} else {
			// All attributes are negative, will get attributes from allAttrs then minus negative attributes
			var valueStrs []string
			var availbleColumns []string
			for _, value := range flattenValues {
				if column, ok := value.(string); ok {
					valueStrs = append(valueStrs, column)
				}
			}

			for _, column := range res.allAttrs(excludedColumns...) {
				if !isContainsColumn(valueStrs, column) {
					availbleColumns = append(availbleColumns, column)
				}
			}
			*sections = res.generateSections(availbleColumns)
		}
	}
}
