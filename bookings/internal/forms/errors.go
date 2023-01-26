package forms

type frmErrors map[string][]string

// Add adds an error message for the given field
func (e frmErrors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get returns the first error message for the field if any.
// If passed field has no errors, it returnd an empty string
func (e frmErrors) Get(field string) string {
	eStr := e[field]
	if len(eStr) == 0 {
		return ""
	}
	return eStr[0]
}
