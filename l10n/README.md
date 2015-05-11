## Introduction

L10n support translate resource in different locales.

## Usage

First set `l10n.Locale` as field in the resource you want to use L10n

    type Product struct {
      Name        string
      Description string `l10n:"sync"` // `l10n:"sync"` means this field will always follow global product's change
      l10n.Locale
    }

Then register callbacks for L10n

    DB, err = gorm.Open("sqlite3", "demo_db") // [gorm](https://github.com/jinzhu/gorm)
    DB.AutoMigrate(&Product{})

    l10n.RegisterCallbacks(&DB)

And you're done ! Now if you want to create product for zh and en

    var productCN Product
    var productEN Product

    product := Product{Name: "Global product", Description: "Global product description"}
    DB.Create(&product)
    fmt.Println(product.language_code) // global

    product.Name = "中文产品"
    DB.Set("l10n:locale", "zh").Create(&product)
    // Query zh version of product
    DB.Set("l10n:locale", "zh").First(&productCN, product.ID)
    fmt.Println(productCN.Name) // "中文产品"

    product.Name = "English product"
    DB.Set("l10n:locale", "en").Create(&product)
    // Query en version of product
    DB.Set("l10n:locale", "en").First(&productEN, product.ID)
    fmt.Println(productEN.Name) // "English product"

## Set attribute that always sync with global record

Set `l10n:"sync"` to the field you want it always sync with global record like this

    type Product struct {
      Name        string
      Description string `l10n:"sync"`
      l10n.Locale
    }

Now, localized product's Description will keep same with global product's Description. NOTE: Description can't be changed on localized product by this set.

## Query
L10n support 4 modes on query, you can specify the mode by

    mode := "global"
    db.Set("l10n:mode", mode).First(&product, product.ID)

#### global

find global record only

#### locale

find localized record only

#### unscoped

append nothing in query

#### default

find localized record first. if nothing returned, return global record

## Use with [QOR admin]()
Add two functions to admin "user" to set viewable and editable locales for different users, you can see below selector in QOR admin

    func (User) ViewableLocales() []string {
      return []string{l10n.Global, "zh-CN", "JP", "EN", "DE"}
    }

    func (user User) EditableLocales() []string {
      if user.role == "global_admin" {
        return []string{l10n.Global, "zh-CN", "EN"}
      } else {
        return []string{"zh-CN", "EN"}
      }
    }

// TODO: add QOR admin l10n selector screenshot here
