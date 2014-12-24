package admin

import (
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type scopeFunc func(db *gorm.DB, context *qor.Context) *gorm.DB

type Searcher struct {
	Resource *Resource
	Admin    *Admin
	scopes   []*Scope
	filters  map[string]string
}

func (admin *Admin) NewSearcher(res *Resource) *Searcher {
	return &Searcher{Resource: res, Admin: admin}
}

func (s *Searcher) Scope(names ...string) *Searcher {
	for _, name := range names {
		if scope := s.Resource.scopes[name]; scope != nil && !scope.Default {
			s.scopes = append(s.scopes, scope)
		}
	}
	return s
}

func (s *Searcher) Filter(name, query string) *Searcher {
	if s.filters == nil {
		s.filters = map[string]string{}
	}
	s.filters[name] = query
	return s
}

var filterRegexp = regexp.MustCompile(`^filters\[(.*?)\]$`)

func (s *Searcher) ParseContext(context *qor.Context) {
	if context != nil && context.Request != nil {
		// parse scopes
		scopes := strings.Split(context.Request.Form.Get("scopes"), "|")
		s.Scope(scopes...)

		// parse filters
		for key, value := range context.Request.Form {
			if matches := filterRegexp.FindStringSubmatch(key); len(matches) > 0 {
				s.Filter(matches[1], value[0])
			}
		}
	}
}

func (s *Searcher) callScopes(context *qor.Context) *qor.Context {
	s.ParseContext(context)
	db := context.GetDB()

	// call default scopes
	for _, scope := range s.Resource.scopes {
		if scope.Default {
			db = scope.Handler(db, context)
		}
	}

	// call scopes
	for _, scope := range s.scopes {
		db = scope.Handler(db, context)
	}

	// call filters
	context.SetDB(db)
	return context
}

func (s *Searcher) getContext(contexts []interface{}) *qor.Context {
	var context *qor.Context
	if len(contexts) > 0 {
		if value, ok := contexts[0].(*qor.Context); ok {
			context = value
		} else if value, ok := contexts[0].(*Context); ok {
			context = value.Context
		}
	} else {
		context = &qor.Context{DB: s.Admin.Config.DB}
	}

	return context
}

func (s *Searcher) FindAll(contexts ...interface{}) (interface{}, error) {
	context := s.getContext(contexts)
	result := s.Resource.NewSlice()
	s.callScopes(context)
	err := s.Resource.CallSearcher(result, context)
	return result, err
}

func (s *Searcher) FindOne(contexts ...interface{}) (interface{}, error) {
	context := s.getContext(contexts)
	result := s.Resource.NewStruct()
	s.callScopes(context)
	err := s.Resource.CallFinder(result, nil, context)
	return result, err
}
