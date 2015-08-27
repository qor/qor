package publish_test

import "testing"

func TestDeleteStructFromDraft(t *testing.T) {
	name := "delete_product_from_draft"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)
	pbdraft.Delete(&product)

	pbdraft.Unscoped().First(&product, product.ID)

	if !product.PublishStatus {
		t.Errorf("Product's publish status should be DIRTY when deleted from draft db")
	}

	if pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be deleted in production db")
	}

	if !pbdraft.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}

	if pbdraft.Unscoped().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}
}

func TestDeleteStructFromProduction(t *testing.T) {
	name := "delete_product_from_production"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)
	pbprod.Delete(&product)

	if !pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in production db")
	}

	if pbdraft.Unscoped().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in production db")
	}

	if !pbdraft.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}

	if pbdraft.Unscoped().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}

	pbdraft.Unscoped().First(&product, product.ID)
	if product.PublishStatus {
		t.Errorf("Product's publish status should be PUBLISHED when deleted from production db")
	}

	pbprod.Unscoped().First(&product, product.ID)
	if product.PublishStatus {
		t.Errorf("Product's publish status should be PUBLISHED when deleted from production db")
	}
}
