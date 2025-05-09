# gorilla-validator


[![PkgGoDev](https://pkg.go.dev/badge/github.com/iamolegga/gorilla-validator)](https://pkg.go.dev/github.com/iamolegga/gorilla-validator) [![Go Report Card](https://goreportcard.com/badge/github.com/iamolegga/gorilla-validator)](https://goreportcard.com/report/github.com/iamolegga/gorilla-validator) ![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/iamolegga/gorilla-validator/on-push-main.yml) ![Codacy coverage](https://img.shields.io/codacy/coverage/08ef22b3391240aa8806a48e6891e3d6) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/08ef22b3391240aa8806a48e6891e3d6)](https://app.codacy.com/gh/iamolegga/gorilla-validator/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)

HTTP request validation middleware for Gorilla Mux.

It simplifies the process of validating and extracting data from various HTTP request sources, including URL parameters, query strings, form data, JSON, and XML.

## Installation

```sh
go get github.com/iamolegga/gorilla-validator
```

## Features

- Easy to use middleware for Gorilla Mux
- Supports multiple data sources: URL parameters, query strings, form data, JSON, and XML
- Type-safe access to validated data
- Leverages go-playground/validator for validation rules
- Automatic HTTP 400 responses for invalid requests

## Examples

Here are examples showing how to use gorilla-validator with different data sources:

### URL Parameters

```go
// URL parameters validation example
// Define the schema for URL parameters
type Params struct {
    ID int `schema:"id" validate:"required,gt=0"`
}

router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    // Get validated URL parameters
    params := gv.Validated[*Params](r, gv.Params)
}).Methods("GET").Use(gv.Validate(Params{}, gv.Params))
```

### Query Parameters

```go
// Query parameters validation example
// Define the schema for query parameters
type Query struct {
    Page  int    `schema:"page" validate:"omitempty,gte=1"`
    Limit int    `schema:"limit" validate:"omitempty,gte=1,lte=100"`
    Sort  string `schema:"sort" validate:"omitempty,oneof=name email date"`
}

router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    // Get validated query parameters
    query := gv.Validated[*Query](r, gv.Query)
}).Methods("GET").Use(gv.Validate(Query{}, gv.Query))
```

### JSON Body

```go
// JSON body validation example
// Define the schema for JSON body
// NOTE: You must set `json` tags for all fields you want to decode from JSON.
// The standard library does NOT use `schema` tags for JSON or XML decoding.
type BodyJSON struct {
    Name     string `json:"name" schema:"name" validate:"required,min=2"`
    Email    string `json:"email" schema:"email" validate:"required,email"`
    Password string `json:"password" schema:"password" validate:"required,min=8"`
}

router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    // Get validated JSON body
    data := gv.Validated[*BodyJSON](r, gv.JSON)
}).Methods("POST").Use(gv.Validate(BodyJSON{}, gv.JSON))
```

### Form Data

```go
// Form data validation example
// Define the schema for form data
// NOTE: For form/query/params, only the `schema` tag is needed.
type BodyForm struct {
    Email    string `schema:"email" validate:"required,email"`
    Password string `schema:"password" validate:"required"`
    Remember bool   `schema:"remember"`
}

router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
    // Get validated form data
    loginData := gv.Validated[*BodyForm](r, gv.Form)
}).Methods("POST").Use(gv.Validate(BodyForm{}, gv.Form))
```

### Multiple Validators

```go
// Multiple validators example
// Define schemas for URL parameters and JSON body
// NOTE: Always specify `json` or `xml` tags for JSON/XML, and `schema` for form/query/params.
type MultiParams struct {
    ID int `schema:"id" validate:"required,gt=0"`
}

type MultiBody struct {
    Name  *string `json:"name" schema:"name" validate:"omitempty,min=2"`
    Email *string `json:"email" schema:"email" validate:"omitempty,email"`
}

router.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    // Get validated URL parameters and JSON body
    params := gv.Validated[*MultiParams](r, gv.Params)
    updateData := gv.Validated[*MultiBody](r, gv.JSON)
}).Methods("PATCH").
    Use(gv.Validate(MultiParams{}, gv.Params)).
    Use(gv.Validate(MultiBody{}, gv.JSON))
```

## Struct Tag Requirements

**Important:**

- The Go standard library decoders (`encoding/json`, `encoding/xml`) **do not** use the `schema` tag.
- You **must** specify `json` and/or `xml` tags for fields you want to decode from JSON or XML.
- The `schema` tag is only used for form, query, and params sources.
- If a tag is missing, the field name (case-sensitive) will be used by the decoder, which may not match your input.

## Validation Rules

Validation rules are defined using struct tags with the go-playground/validator syntax.
For a complete list of available validation rules, see:
https://github.com/go-playground/validator

## Sources

The library supports the following sources for validation:

- `gv.Params`: URL parameters from Gorilla Mux
- `gv.Query`: Query string parameters
- `gv.Form`: Form data from POST requests
- `gv.JSON`: JSON request body
- `gv.XML`: XML request body

## Error Handling

By default, the middleware will automatically respond with HTTP 400 (Bad Request)
when validation fails. This behavior can be customized by using the `gv.ErrorHandler` function:

```go
gv.ErrorHandler(func(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
	}
})
```
