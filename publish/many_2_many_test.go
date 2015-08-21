package publish_test

import "testing"

func TestPublishManyToManyFromProduction(t *testing.T) {
	name := "create_product_with_multi_categories_from_production"
	pbprod.Create(&Product{
		Name:       name,
		Categories: []Category{{Name: "category1"}, {Name: "category2"}},
	})

	var product Product
	pbprod.First(&product, "name = ?", name)

	if pbprod.Model(&product).Association("Categories").Count() != 2 {
		t.Errorf("categories count should be 2 in production db")
	}

	if pbdraft.Model(&product).Association("Categories").Count() != 2 {
		t.Errorf("categories count should be 2 in draft db")
	}
}

func TestPublishManyToManyFromDraft(t *testing.T) {
	name := "create_product_with_multi_categories_from_draft"
	pbdraft.Create(&Product{
		Name:       name,
		Categories: []Category{{Name: "category1"}, {Name: "category2"}},
	})

	var product Product
	pbdraft.First(&product, "name = ?", name)

	if pbprod.Model(&product).Association("Categories").Count() != 0 {
		t.Errorf("categories count should be 0 in production db")
	}

	if pbdraft.Model(&product).Association("Categories").Count() != 2 {
		t.Errorf("categories count should be 2 in draft db")
	}

	pb.Publish(&product)
	var categories []Category
	pbdraft.Find(&categories)
	pb.Publish(&categories)

	if pbprod.Model(&product).Association("Categories").Count() != 2 {
		t.Errorf("categories count should be 2 in production db after publish")
	}

	if pbdraft.Model(&product).Association("Categories").Count() != 2 {
		t.Errorf("categories count should be 2 in draft db after publish")
	}
}

func TestDiscardManyToManyFromDraft(t *testing.T) {
	name := "discard_product_with_multi_categories_from_draft"
	pbdraft.Create(&Product{
		Name:       name,
		Categories: []Category{{Name: "category1"}, {Name: "category2"}},
	})

	var product Product
	pbdraft.First(&product, "name = ?", name)

	if pbprod.Model(&product).Association("Categories").Count() != 0 {
		t.Errorf("categories count should be 0 in production db")
	}

	if pbdraft.Model(&product).Association("Categories").Count() != 2 {
		t.Errorf("categories count should be 2 in draft db")
	}

	pb.Discard(&product)

	if pbprod.Model(&product).Association("Categories").Count() != 0 {
		t.Errorf("categories count should be 0 in production db after discard")
	}

	if pbdraft.Model(&product).Association("Categories").Count() != 0 {
		t.Errorf("categories count should be 0 in draft db after discard")
	}
}
