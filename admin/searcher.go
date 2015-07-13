package admin

import (
	"regexp"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type scopeFunc func(db *gorm.DB, context *qor.Context) *gorm.DB

type Pagination struct {
	Total       int
	Pages       int
	CurrentPage int
	PrePage     int
}

type Searcher struct {
	*Context
	scopes         []*Scope
	filters        map[string]string
	withPagination bool
	Pagination     Pagination
}

func (s *Searcher) WithPagination() *Searcher {
	s.withPagination = true
	return s
}

func (s *Searcher) Page(num int) *Searcher {
	s.Pagination.CurrentPage = num
	return s
}

func (s *Searcher) PrePage(num int) *Searcher {
	s.Pagination.PrePage = num
	return s
}

func (s *Searcher) clone() *Searcher {
	return &Searcher{Context: s.Context, scopes: s.scopes, filters: s.filters}
}

func (s *Searcher) Scope(names ...string) *Searcher {
	newSearcher := s.clone()
	for _, name := range names {
		for _, scope := range s.Resource.scopes {
			if scope.Name == name && !scope.Default {
				newSearcher.scopes = append(newSearcher.scopes, scope)
				break
			}
		}
	}
	return newSearcher
}

func (s *Searcher) Filter(name, query string) *Searcher {
	newSearcher := s.clone()
	if newSearcher.filters == nil {
		newSearcher.filters = map[string]string{}
	}
	newSearcher.filters[name] = query
	return newSearcher
}

var filterRegexp = regexp.MustCompile(`^filters\[(.*?)\]$`)

func (s *Searcher) callScopes(context *qor.Context) *qor.Context {
	db := context.GetDB()

	// call default scopes
	for _, scope := range s.Resource.scopes {
		if scope.Default {
			db = scope.Handle(db, context)
		}
	}

	// call scopes
	for _, scope := range s.scopes {
		db = scope.Handle(db, context)
	}

	// call filters
	if s.filters != nil {
		for key, value := range s.filters {
			filter := s.Resource.filters[key]
			if filter != nil && filter.Handler != nil {
				db = filter.Handler(key, value, db, context)
			} else {
				db = DefaultHandler(key, value, db, context)
			}
		}
	}

	context.SetDB(db)

	// call search
	if keyword := context.Request.Form.Get("keyword"); keyword != "" {
		if s.Resource.SearchHandler != nil {
			context.SetDB(s.Resource.SearchHandler(keyword, context))
		}
	}
	return context
}

func (s *Searcher) cloneContext() *qor.Context {
	context := s.Context.Context
	return &qor.Context{
		Request:    context.Request,
		Writer:     context.Writer,
		ResourceID: context.ResourceID,
		Config:     context.Config,
		DB:         context.DB,
	}
}

func (s *Searcher) parseContext() *qor.Context {
	var context = s.cloneContext()
	var searcher = s.clone()

	if context != nil && context.Request != nil {
		// parse scopes
		scopes := context.Request.Form["scopes"]
		searcher = searcher.Scope(scopes...)

		// parse filters
		for key, value := range context.Request.Form {
			if matches := filterRegexp.FindStringSubmatch(key); len(matches) > 0 {
				searcher = searcher.Filter(matches[1], value[0])
			}
		}
	}

	searcher.callScopes(context)

	// pagination
	db := context.GetDB()
	paginationDB := db.Select("count(*) total").Model(s.Resource.Value).Set("qor:no_ordering", true).Set("gorm:query_destination", &s.Pagination)
	context.SetDB(paginationDB)
	context.GetDB().Find(s.Resource.Value)

	if s.Pagination.CurrentPage == 0 {
		if s.Context.Request != nil {
			if page, err := strconv.Atoi(s.Context.Request.Form.Get("page")); err == nil {
				s.Pagination.CurrentPage = page
			}
		}

		if s.Pagination.CurrentPage == 0 {
			s.Pagination.CurrentPage = 1
		}
	}

	if s.Pagination.PrePage == 0 {
		s.Pagination.PrePage = s.Resource.Config.PageCount
	}

	if s.Pagination.CurrentPage > 0 {
		s.Pagination.Pages = (s.Pagination.Total-1)/s.Pagination.PrePage + 1

		db = db.Limit(s.Pagination.PrePage).Offset((s.Pagination.CurrentPage - 1) * s.Pagination.PrePage)
	}

	context.SetDB(db)

	return context
}

func (s *Searcher) FindMany() (interface{}, error) {
	context := s.parseContext()
	result := s.Resource.NewSlice()
	err := s.Resource.CallFindMany(result, context)
	return result, err
}

func (s *Searcher) FindOne() (interface{}, error) {
	context := s.parseContext()
	result := s.Resource.NewStruct()
	err := s.Resource.CallFindOne(result, nil, context)
	return result, err
}
