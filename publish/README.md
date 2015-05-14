## Introduction

Publish allow user update a resource but do not show the change in website until it is get "published"

## Usage

Use "Product" and [gorm](https://github.com/jinzhu/gorm) for example.

First set `publish.Status` as field in the resource you want to use Publish

    type Product struct {
      Name        string
      Description string
      publish.Status
    }

    DB, err = gorm.Open("sqlite3", "demo_db")
    DB.AutoMigrate(&Product{})

Then initialize Publish and set which resource(table) needs Publish support.

    publish := publish.New(&DB)
    publish.Support(&Product{}).AutoMigrate()

Now, you have two tables for product. "products" and "products_draft", all changes made on product will be saved in "products_draft"

    draft_db = publish.DraftDB() // Draft resources are saved here
    production_db = publish.ProductionDB() // Published resources saved here

To process draft products

    publish.Publish(products)
    publish.Discard(products)

## Use with [QOR admin]()

Initialize QOR admin by set DB as `Publish.DraftDB()`, then add Publish to Admin resource.

    Admin := admin.New(&qor.Config{DB: Publish.DraftDB()})
    Admin.AddResource(Publish)

Now you can see a "Publish" link appears in QOR admin website, updated product will appears in this page and not available at front end. You can view difference, publish or discard changes made on product.
