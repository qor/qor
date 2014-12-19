package admin

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type scopeFunc func(db *gorm.DB, context *qor.Context) *gorm.DB

type Searcher struct {
	Resource *Resource
	scopes   []scopeFunc
}

func NewSearcher(res *Resource) *Searcher {
	return &Searcher{Resource: res}
}

func (s *Searcher) Scope(names ...string) *Searcher {
	for _, name := range names {
		s.scopes = append(s.scopes, s.Resource.scopes[name])
	}
	return s
}

func (s *Searcher) Filter(name, query string) *Searcher {
	return s
}

func (s *Searcher) ParseContext(context *qor.Context) {
	// scopes
	if context != nil {
		scopes := strings.Split(context.Request.Form.Get("scopes"), "|")
		s.Scope(scopes...)
	}
}

func (s *Searcher) callScopes(context *qor.Context) {
	s.ParseContext(context)
	db := context.GetDB()
	for _, scope := range s.scopes {
		db = scope(db, context)
	}
	context.SetDB(db)
}

func (s *Searcher) FindAll(result interface{}, context *qor.Context) error {
	return s.Resource.CallSearcher(result, context)
}

func (s *Searcher) FindOne(result interface{}, context *qor.Context) error {
	return s.Resource.CallFinder(result, nil, context)
}
