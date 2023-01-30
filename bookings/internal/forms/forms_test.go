package forms

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestPostRequest(url string) *http.Request {
	r := httptest.NewRequest("POST", url, nil)
	r.ParseForm()
	return r
}
func TestForm_Has(t *testing.T) {
	r := newTestPostRequest("/someurl/?myField=myValue&emptyField=")
	form := New(r.Form)
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
		has := form.Has(test.field, r)
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
		url      string
		field    string
		length   int
		expected bool
	}{
		{"success", "/someurl/?myField=myValue", "myField", 3, true},
		{"short success", "/someurl/?myField=my", "myField", 1, true},
		{"too short", "/someurl/?myField=my", "myField", 3, false},
		{"not found", "/someurl/?myField=myValue", "fakeField", 3, false},
	}

	for _, e := range tests {
		r := newTestPostRequest(e.url)
		form := New(r.Form)
		result := form.MinLength(e.field, e.length, r)
		if e.expected && !result {
			t.Errorf("%s: expected success, but failed; field: %q, url: %q, length: %d", e.name, e.field, e.url, e.length)
		}
		if !e.expected && result {
			t.Errorf("%s: expected fail, but succeeded; field: %q, url: %q, length: %d", e.name, e.field, e.url, e.length)
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
		if e.expected && !result {
			t.Errorf("%s: expected success, but failed; field: %q, url: %q", e.name, e.field, e.url)
		}
		if !e.expected && result {
			t.Errorf("%s: expected fail, but succeeded; field: %q, url: %q", e.name, e.field, e.url)
		}
	}
}
