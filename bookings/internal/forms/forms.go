package forms

import (
	"net/http"
	"net/url"
)

// Form is a type encapsulating HTML form fields, their values,
// and errors in form's data
type Form struct {
	url.Values
	Errors frmErrors
}

// Valid returns true if form does not have any errors
// otherwise it returns false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New initializes a form struct out of url.Values
func New(data url.Values) *Form {
	return &Form{data, frmErrors{}}
}

// Has checks if form has data for required field filled in
func (f *Form) Has(field string, r *http.Request) bool {
	if r.Form.Get(field) == "" {
		f.Errors.Add(field, "This field cannot be empty.")
		return false
	}
	return true
}
