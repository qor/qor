# Qor Admin

## Usage


HTTP Mux

```
mux := http.NewServeMux()
web := admin.New(&qor.Config{DB: &db})
web.AddToMux("/admin", mux)
http.ListenAndServe(":8080", mux)
```

Gin

```
mux := http.NewServeMux()

web := admin.New(&qor.Config{DB: &db.DB})
web.AddToMux("/admin", mux)

router := gin.Default()
// ...
mux.Handle("/", router)
http.ListenAndServe(":8080", mux)
```
