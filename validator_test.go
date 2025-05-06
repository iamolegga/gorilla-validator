package gv_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	gv "github.com/iamolegga/gorilla-validator"
	"github.com/stretchr/testify/assert"
)

type Profile struct {
	Name  string `schema:"name" json:"name" xml:"name" validate:"required"`
	Email string `schema:"email" json:"email" xml:"email" validate:"required,email"`
}

type TestSchema struct {
	ID        int     `schema:"id" json:"id" xml:"id" validate:"required"`
	Profile   Profile `schema:"profile" json:"profile" xml:"profile" validate:"required"`
	Verified  bool    `schema:"verified" json:"verified" xml:"verified"`
	FollowIDs []int   `schema:"follow_ids" json:"follow_ids" xml:"follow_ids"`
}

type ParamsTestSchema struct {
	ID int `schema:"id" validate:"required"`
}

func TestValidateParamsOK(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that will be called after validation succeeds
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*ParamsTestSchema](r, gv.Params)
		assert.Equal(t, 123, data.ID)
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(ParamsTestSchema{}, gv.Params)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test/{id}", validatedHandler).Methods(http.MethodGet)
	
	// Create a request with the URL that contains the parameter
	req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify the response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateParamsError(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(ParamsTestSchema{}, gv.Params)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test/{id}", validatedHandler).Methods(http.MethodGet)
	
	// Create a request with a non-integer ID that will fail validation
	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateQueryOK(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that will be called after validation succeeds
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.Query)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.Query)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodGet)
	
	// Create a request with query parameters
	req := httptest.NewRequest(http.MethodGet, "/test?id=123&profile.name=John&profile.email=john@example.com", nil)
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify the response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateQueryError(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.Query)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodGet)
	
	// Create a request with invalid query parameters (id=abc is not an integer)
	req := httptest.NewRequest(http.MethodGet, "/test?id=abc&profile.name=John&profile.email=john@example.com", nil)
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateFormOK(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that will be called after validation succeeds
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.Form)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.Form)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create form data
	form := url.Values{}
	form.Add("id", "123")
	form.Add("profile.name", "John")
	form.Add("profile.email", "john@example.com")
	
	// Create a request with form data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify the response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateFormError(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.Form)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create form data with invalid ID (abc is not an integer)
	form := url.Values{}
	form.Add("id", "abc")
	form.Add("profile.name", "John")
	form.Add("profile.email", "john@example.com")
	
	// Create a request with form data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateFormErrorInvalid(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.Form)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create a request with no form data (will fail validation)
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateFormErrorInvalid2(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.Form)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodGet)
	
	// Create a request with invalid query string format
	req := httptest.NewRequest(http.MethodGet, "/test?;&", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateJSONOK(t *testing.T) {
	// Log validation errors
	gv.ErrorHandler(func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.Logf("Validation error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that will be called after validation succeeds
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.JSON)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		assert.Equal(t, true, data.Verified)
		assert.Equal(t, []int{456, 789}, data.FollowIDs)
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.JSON)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create JSON request body
	body := `{"id":123,"profile":{"name":"John","email":"john@example.com"},"verified":true,"follow_ids":[456,789]}`
	
	// Create a request with JSON data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Log response body if there's an error
	if rr.Code != http.StatusOK {
		t.Logf("Response body: %s", rr.Body.String())
	}
	
	// Verify the response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateJSONError(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.JSON)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create JSON request body with invalid ID (abc is not an integer)
	body := `{"id":"abc","profile":{"name":"John","email":"john@example.com"}}`
	
	// Create a request with JSON data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateJSONErrorInvalid(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.JSON)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create invalid JSON request body
	body := `{invalid json}`
	
	// Create a request with invalid JSON data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateXMLOK(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that will be called after validation succeeds
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.XML)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		assert.Equal(t, true, data.Verified)
		assert.Equal(t, []int{456, 789}, data.FollowIDs)
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.XML)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create XML request body
	body := `<TestSchema>
	<id>123</id>
	<profile>
		<name>John</name>
		<email>john@example.com</email>
	</profile>
	<verified>true</verified>
	<follow_ids>456</follow_ids>
	<follow_ids>789</follow_ids>
</TestSchema>`
	
	// Create a request with XML data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/xml")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify the response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateXMLError(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(TestSchema{}, gv.XML)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test", validatedHandler).Methods(http.MethodPost)
	
	// Create XML request body with invalid ID (abc is not an integer)
	body := `<TestSchema>
	<id>abc</id>
	<profile>
		<name>John</name>
		<email>john@example.com</email>
	</profile>
</TestSchema>`
	
	// Create a request with XML data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/xml")
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestErrorHandler(t *testing.T) {
	// Set a flag to track if our custom error handler was called
	var called bool

	// Set up a custom error handler
	gv.ErrorHandler(func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			called = true
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will fail
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware to the handler
	validatedHandler := gv.Validate(ParamsTestSchema{}, gv.Params)(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test/{id}", validatedHandler).Methods(http.MethodGet)
	
	// Create a request with an invalid ID (abc is not an integer)
	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	rr := httptest.NewRecorder()
	
	// Let the router handle the request
	router.ServeHTTP(rr, req)
	
	// Verify that we got a bad request response due to validation failure
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	
	// Verify that our custom error handler was called
	assert.True(t, called)
}

func TestWrongSourcePanics(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	
	// Define the handler that should never be called because validation will panic
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	})
	
	// Apply the validator middleware with an invalid source
	validatedHandler := gv.Validate(ParamsTestSchema{}, gv.Source("UNEXISTING_SOURCE"))(handlerFunc)
	
	// Register the route with the validated handler
	router.Handle("/test/{id}", validatedHandler).Methods(http.MethodGet)
	
	// Create a request
	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	rr := httptest.NewRecorder()
	
	// Verify that the router panics when handling the request with an invalid source
	assert.Panics(t, func() {
		router.ServeHTTP(rr, req)
	})
}

// CustomValidationSchema defines a schema with a custom validation tag
type CustomValidationSchema struct {
	ID int `schema:"id" validate:"even"`
}

func TestCustomValidator(t *testing.T) {
	// Save the original validator to restore it after the test
	defer func() {
		// Reset the validator to default after the test
		gv.Validator(validator.New())
	}()
	
	// Create a custom validator with a custom validation rule
	customValidator := validator.New()
	// Register a custom validation rule for "even" numbers
	customValidator.RegisterValidation("even", func(fl validator.FieldLevel) bool {
		// Get the field value as int
		val, ok := fl.Field().Interface().(int)
		if !ok {
			return false
		}
		// Check if the number is even
		return val%2 == 0
	})
	
	// Set the custom validator
	gv.Validator(customValidator)
	
	// Test with an even number (should pass validation)
	t.Run("EvenNumber", func(t *testing.T) {
		// Create a new router
		router := mux.NewRouter()
		
		// Define the handler that will be called after validation succeeds
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data := gv.Validated[*CustomValidationSchema](r, gv.Params)
			assert.Equal(t, 2, data.ID) // Should be 2 (even)
			w.WriteHeader(http.StatusOK)
		})
		
		// Apply the validator middleware to the handler
		validatedHandler := gv.Validate(CustomValidationSchema{}, gv.Params)(handlerFunc)
		
		// Register the route with the validated handler
		router.Handle("/test/{id}", validatedHandler).Methods(http.MethodGet)
		
		// Create a request with an even ID (2)
		req := httptest.NewRequest(http.MethodGet, "/test/2", nil)
		rr := httptest.NewRecorder()
		
		// Let the router handle the request
		router.ServeHTTP(rr, req)
		
		// Verify the response (should be OK)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
	
	// Test with an odd number (should fail validation)
	t.Run("OddNumber", func(t *testing.T) {
		// Create a new router
		router := mux.NewRouter()
		
		// Define the handler that should never be called because validation will fail
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Fail(t, "should not reach here")
		})
		
		// Apply the validator middleware to the handler
		validatedHandler := gv.Validate(CustomValidationSchema{}, gv.Params)(handlerFunc)
		
		// Register the route with the validated handler
		router.Handle("/test/{id}", validatedHandler).Methods(http.MethodGet)
		
		// Create a request with an odd ID (3)
		req := httptest.NewRequest(http.MethodGet, "/test/3", nil)
		rr := httptest.NewRecorder()
		
		// Let the router handle the request
		router.ServeHTTP(rr, req)
		
		// Verify that we got a bad request response due to validation failure
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
