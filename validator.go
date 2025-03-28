package gv

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
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

var schemaDecoder = schema.NewDecoder()

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
				err := schemaDecoder.Decode(schemaValue, varsFixed)
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			case Query:
				err := schemaDecoder.Decode(schemaValue, r.URL.Query())
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
				err = schemaDecoder.Decode(schemaValue, r.PostForm)
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			case JSON:
				var jsonData map[string]any
				err := json.NewDecoder(r.Body).Decode(&jsonData)
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
				data := convertToMapStringSlice(jsonData)
				err = schemaDecoder.Decode(schemaValue, data)
				if err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			case XML:
				decoder := xml.NewDecoder(r.Body)
				result := make(map[string]any)
				var currentElement string

				for {
					tok, err := decoder.Token()
					if err != nil {
						break
					}

					switch token := tok.(type) {
					case xml.StartElement:
						currentElement = token.Name.Local
					case xml.CharData:
						if currentElement != "" {
							// Handle repeated elements by converting to array as needed
							if existing, ok := result[currentElement]; ok {
								switch v := existing.(type) {
								case string:
									// If element already exists, convert to array
									result[currentElement] = []any{v, string(token)}
								case []any:
									// If it's already an array, append
									result[currentElement] = append(v, string(token))
								}
							} else {
								// First occurrence of this element
								result[currentElement] = string(token)
							}
							currentElement = ""
						}
					}
				}
				if err := schemaDecoder.Decode(schemaValue, convertToMapStringSlice(result)); err != nil {
					currentErrorHandler(err).ServeHTTP(w, r)
					return
				}
			default:
				panic("unknown source: " + src)
			}

			validate := validator.New()
			err := validate.Struct(schemaValue)
			if err != nil {
				currentErrorHandler(err).ServeHTTP(w, r)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), sourceKey(src), schemaValue))
			handler.ServeHTTP(w, r)
		})
	}
}

func Validated[T any](r *http.Request, src Source) T {
	return r.Context().Value(sourceKey(src)).(T)
}

func convertToMapStringSlice(input map[string]any) map[string][]string {
	result := make(map[string][]string)
	for key, value := range input {
		switch v := value.(type) {
		case string:
			result[key] = []string{v}
		case float64:
			result[key] = []string{fmt.Sprintf("%v", v)}
		case bool:
			result[key] = []string{fmt.Sprintf("%v", v)}
		case []any:
			var strSlice []string
			for _, elem := range v {
				strSlice = append(strSlice, fmt.Sprintf("%v", elem))
			}
			result[key] = strSlice
		default:
			result[key] = []string{fmt.Sprintf("%v", v)}
		}
	}
	return result
}

type Element struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}
