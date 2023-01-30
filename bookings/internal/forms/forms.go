package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
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

// Required checks for required fields
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be empty.")
		}
	}
}

// Has checks if form has data for field filled in
func (f *Form) Has(field string, r *http.Request) bool {
	return !(r.Form.Get(field) == "")
}

// MinLenghth checks for minimum field length
func (f *Form) MinLength(field string, length int, r *http.Request) bool {
	if len(r.Form.Get(field)) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long.", length))
		return false
	}
	return true
}

// IsEmail checks for valid email address
func (f *Form) IsEmail(field string) bool {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
		return false
	}
	return true
}
