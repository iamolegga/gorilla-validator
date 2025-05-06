package gv

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

// ErrorHandlerFunc is a function type that defines how validation errors are handled
type ErrorHandlerFunc func(err error) http.HandlerFunc

// defaultErrorHandler is the default implementation of error handling
var currentErrorHandler = func(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// ErrorHandler allows setting a custom error handler function
func ErrorHandler(h ErrorHandlerFunc) {
	currentErrorHandler = h
}

// sourceKey is a type for context keys used to store validated data. It is
// used to avoid conflicts with other middleware that may use the same context
// keys
type sourceKey string

// Source represents the source of input data for validation
type Source string

const (
	Params Source = "Params"
	Query  Source = "Query"
	Form   Source = "Form"
	JSON   Source = "JSON"
	XML    Source = "XML"
)

// SchemaDecoder is an instance of the schema decoder from the gorilla/schema package, could be used for setting custom options
var SchemaDecoder = schema.NewDecoder()

var currentValidator = validator.New()

// Validator allows setting a custom validator instance
func Validator(v *validator.Validate) {
	currentValidator = v
}

// Validate is a middleware factory function that validates the input data based on the provided schema and source
func Validate(schema any, src Source) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			schemaValue := reflect.New(reflect.TypeOf(schema)).Interface()

			switch src {
			case Params:
				vars := mux.Vars(r)
				varsFixed := make(map[string][]string)
				for k, v := range vars {
					varsFixed[k] = []string{v}
				}
				err := SchemaDecoder.Decode(schemaValue, varsFixed)
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			case Query:
				err := SchemaDecoder.Decode(schemaValue, r.URL.Query())
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			case Form:
				err := r.ParseForm()
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
				err = SchemaDecoder.Decode(schemaValue, r.PostForm)
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			case JSON:
				if err := json.NewDecoder(r.Body).Decode(schemaValue); err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			case XML:
				if err := xml.NewDecoder(r.Body).Decode(schemaValue); err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			default:
				panic("unknown source: " + src)
			}

			err := currentValidator.Struct(schemaValue)
			if err != nil {
				currentErrorHandler(err).ServeHTTP(w, r)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), sourceKey(src), schemaValue))
			handler.ServeHTTP(w, r)
		})
	}
}

// Validated is a function that returns the validated data from the request context
func Validated[T any](r *http.Request, src Source) T {
	return r.Context().Value(sourceKey(src)).(T)
}
