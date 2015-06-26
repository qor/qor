package publish_test

import "testing"

func TestUpdateStructFromDraft(t *testing.T) {
	name := "update_product_from_draft"
	newName := name + "_v2"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)

	pbdraft.Model(&product).Update("name", newName)

	pbdraft.First(&product, product.ID)
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
	}

	db.Model(&Product{}).Where("id = ?", product.ID).UpdateColumns(map[string]interface{}{"quantity": 5})
	var newProduct, newDraftProduct Product
	pbprod.Find(&newProduct, product.ID)
	pbprod.Find(&newDraftProduct, product.ID)

	if newProduct.Quantity != 5 || newDraftProduct.Quantity != 5 || newProduct.Name != newName || newDraftProduct.Name != newName {
		t.Errorf("Sync update columns during production & draft db")
	}
}
