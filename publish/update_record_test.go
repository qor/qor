package publish_test

import "testing"

func TestUpdateStructFromDraft(t *testing.T) {
	name := "update_product_from_draft"
	newName := name + "_v2"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)

	pbdraft.Model(&product).Update("name", newName)

	if !product.PublishStatus {
		t.Errorf("Product's publish status should be DIRTY when updated from draft db")
	}

	if pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be changed in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in draft db")
	}

	if pbdraft.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	} else {
		if product.Color.PublishStatus {
			t.Errorf("Color's publish status should be PUBLISHED because it is not publishable")
		}
	}
}

func TestUpdateStructFromProduction(t *testing.T) {
	name := "update_product_from_production"
	newName := name + "_v2"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)
	pbprod.Model(&product).Update("name", newName)

	if product.PublishStatus {
		t.Errorf("Product's publish status should be PUBLISHED when updated from production db")
	}

	if pbprod.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in production db")
	}

	var productDraft Product
	if pbdraft.First(&productDraft, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in draft db")
	}

	if productDraft.PublishStatus {
		t.Errorf("Product's publish status should be PUBLISHED in draft when updated from production db")
	}

	if pbprod.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	} else {
		if product.Color.PublishStatus {
			t.Errorf("Color's publish status should be PUBLISHED because it is not publishable")
		}
	}
}
