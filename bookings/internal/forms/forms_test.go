package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func newTestPostRequest(url string) *http.Request {
	r := httptest.NewRequest("POST", url, nil)
	r.ParseForm()
	return r
}

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POSt", "/whatever", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POSt", "/whatever", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r = httptest.NewRequest("POSt", "/whatever", nil)
	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form shows not having required fields when it does")
	}
}

func TestForm_Has(t *testing.T) {
	postedData := url.Values{"myField": []string{"myValue"}, "emptyField": []string{}}
	form := New(postedData)
	tests := []struct {
		name   string
		field  string
		exists bool
	}{
		{"one value", "myField", true},
		{"empty", "emptyField", false},
		{"non-existent", "someFakeField", false},
	}

	for _, test := range tests {
		has := form.Has(test.field)
		if has && !test.exists {
			t.Errorf("%s: non-existent value found; field: %q", test.name, test.field)
		}
		if !has && test.exists {
			t.Errorf("%s: existent value not found; field: %q", test.name, test.field)
		}
	}
}

func TestForm_MinLenght(t *testing.T) {
	tests := []struct {
		name     string
		values   url.Values
		field    string
		length   int
		expected bool
	}{
		{"success", url.Values{"myField": []string{"myValue"}}, "myField", 3, true},
		{"short success", url.Values{"myField": []string{"my"}}, "myField", 1, true},
		{"too short", url.Values{"myField": []string{"my"}}, "myField", 3, false},
		{"not found", url.Values{"myField": []string{"myValue"}}, "fakeField", 3, false},
	}

	for _, e := range tests {
		form := New(e.values)
		result := form.MinLength(e.field, e.length)
		errValue := form.Errors.Get(e.field)
		if e.expected && (!result || !form.Valid() || errValue != "") {
			t.Errorf("%s: expected success, but failed; field: %q, values: %v, length: %d", e.name, e.field, e.values, e.length)
		}
		if !e.expected && (result || form.Valid() || errValue == "") {
			t.Errorf("%s: expected fail, but succeeded; field: %q, values: %v, length: %d", e.name, e.field, e.values, e.length)
		}
	}
}

func TestForm_IsEmail(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		field    string
		expected bool
	}{
		{"success", "/someurl/?myField=alex@test.com", "myField", true},
		{"two ampersands", "/someurl/?myField=my@my@test.com", "myField", false},
		{"no ampersand", "/someurl/?myField=my.my.test.com", "myField", false},
		{"prohibited character", "/someurl/?myField=my;uncle@test.com", "myField", false},
		{"not found", "/someurl/?myField=alex@test.com", "fakeField", false},
	}

	for _, e := range tests {
		r := newTestPostRequest(e.url)
		form := New(r.Form)
		result := form.IsEmail(e.field)
		if e.expected && (!result || !form.Valid()) {
			t.Errorf("%s: expected success, but failed; field: %q, url: %q", e.name, e.field, e.url)
		}
		if !e.expected && (result || form.Valid()) {
			t.Errorf("%s: expected fail, but succeeded; field: %q, url: %q", e.name, e.field, e.url)
		}
	}
}
