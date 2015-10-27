# Exchange

QOR exchange provides conversion (import/export) functionality for any Qor.Resource to CSV file


## Usage

```go
import (
  "github.com/qor/qor/exchange"
  "github.com/qor/qor/exchange/backends/csv"
)

// Define resource
product = exchange.NewResource(&Product{}, exchange.Config{PrimaryField: "Code"})
product.Meta(exchange.Meta{Name: "Code"})
product.Meta(exchange.Meta{Name: "Name"})
product.Meta(exchange.Meta{Name: "Price"})

// Define context environment
context := &qor.Context{DB: db}

// Import products.csv into database
product.Import(csv.New("products.csv"), context)

// Export products into products.csv
product.Export(csv.New("products.csv"), context)
```

Sample products.csv

```csv
Code, Name, Price
P001, Product P001, 100
P002, Product P002, 200
P003, Product P003, 300
```

## Advanced Usage

* Add Validations

```go
product.AddValidator(func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
  if f, err := strconv.ParseFloat(fmt.Sprint(metaValues.Get("Price").Value), 64); err == nil {
    if f == 0 {
      return errors.New("product's price can't be env")
    }
    return nil
  } else {
    return err
  }
})
```

* Process data before save

```go
product.AddProcessor(func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
  product := result.(*Product)
  product.Price = product.Price * 1.1 // Add 10% Tax
  return nil
})
```
