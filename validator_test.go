package gv_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
	req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "123"})

	rr := httptest.NewRecorder()

	handler := gv.Validate(ParamsTestSchema{}, gv.Params)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*ParamsTestSchema](r, gv.Params)
		assert.Equal(t, 123, data.ID)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateParamsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})

	rr := httptest.NewRecorder()

	handler := gv.Validate(ParamsTestSchema{}, gv.Params)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateQueryOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?id=123&profile.name=John&profile.email=john@example.com", nil)

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.Query)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.Query)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateQueryError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?id=abc&profile.name=John&profile.email=john@example.com", nil)

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.Query)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateFormOK(t *testing.T) {
	form := url.Values{}
	form.Add("id", "123")
	form.Add("profile.name", "John")
	form.Add("profile.email", "john@example.com")
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.Form)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.Form)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateFormError(t *testing.T) {
	form := url.Values{}
	form.Add("id", "abc")
	form.Add("profile.name", "John")
	form.Add("profile.email", "john@example.com")
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.Form)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateFormErrorInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.Form)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateFormErrorInvalid2(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?;&", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.Form)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

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
	body := `{"id":123,"profile":{"name":"John","email":"john@example.com"},"verified":true,"follow_ids":[456,789]}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.JSON)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.JSON)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		assert.Equal(t, true, data.Verified)
		assert.Equal(t, []int{456, 789}, data.FollowIDs)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Logf("Response body: %s", rr.Body.String())
	}
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateJSONError(t *testing.T) {
	body := `{"id":"abc","profile":{"name":"John","email":"john@example.com"}}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.JSON)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateJSONErrorInvalid(t *testing.T) {
	body := `{"id":`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.JSON)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestValidateXMLOK(t *testing.T) {
	// Log validation errors
	gv.ErrorHandler(func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.Logf("Validation error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	body := `<TestSchema><id>123</id><profile><name>John</name><email>john@example.com</email></profile><follow_ids>456</follow_ids><follow_ids>789</follow_ids></TestSchema>`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/xml")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.XML)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := gv.Validated[*TestSchema](r, gv.XML)
		assert.Equal(t, 123, data.ID)
		assert.Equal(t, "John", data.Profile.Name)
		assert.Equal(t, "john@example.com", data.Profile.Email)
		assert.Equal(t, false, data.Verified)
		assert.Equal(t, []int{456, 789}, data.FollowIDs)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Logf("Response body: %s", rr.Body.String())
	}
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestValidateXMLError(t *testing.T) {
	body := `<TestSchema><id>abc</id><profile><name>John</name><email>john@example.com</email></profile></TestSchema>`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/xml")

	rr := httptest.NewRecorder()

	handler := gv.Validate(TestSchema{}, gv.XML)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestErrorHandler(t *testing.T) {
	var called bool

	gv.ErrorHandler(func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			called = true
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})

	rr := httptest.NewRecorder()

	handler := gv.Validate(ParamsTestSchema{}, gv.Params)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.True(t, called)
}

func TestWrongSourcePanics(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})

	rr := httptest.NewRecorder()

	handler := gv.Validate(ParamsTestSchema{}, gv.Source("UNEXISTING_SOURCE"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not reach here")
	}))

	assert.Panics(t, func() {
		handler.ServeHTTP(rr, req)
	})
}
