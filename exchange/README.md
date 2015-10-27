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
