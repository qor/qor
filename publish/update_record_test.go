package publish_test

import "testing"

func TestUpdateStructFromDraft(t *testing.T) {
	name := "update_product_from_draft"
	newName := name + "_v2"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)

	pbdraft.Model(&product).Update("name", newName)

	if pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be changed in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in draft db")
	}
	if pbdraft.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	}
}

func TestUpdateStructFromProduction(t *testing.T) {
	name := "update_product_from_production"
	newName := name + "_v2"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)
	pbprod.Model(&product).Update("name", newName)

	if pbprod.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in draft db")
	}

	if pbprod.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	}
}
