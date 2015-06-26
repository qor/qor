package publish_test

import "testing"

func TestPublishManyToMany(t *testing.T) {
	name := "create_product_with_multi_categories_from_production"
	// pbdraft.Debug().Create(&Product{
	// 	Name:       name,
	// 	Categories: []Category{{Name: "category1"}, {Name: "category2"}},
	// })

	pbprod.Debug().Create(&Product{
		Name:       name,
		Categories: []Category{{Name: "category1"}, {Name: "category2"}},
	})
}
