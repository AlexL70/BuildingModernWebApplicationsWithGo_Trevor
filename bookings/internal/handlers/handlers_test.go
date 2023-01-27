package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"gq", "/generals-quoters", "GET", []postData{}, http.StatusOK},
	{"ms", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"mkres", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"post-sa", "/search-availability", "POST", []postData{
		{key: "start", value: "2024-01-01"},
		{key: "end", value: "2024-01-05"},
	}, http.StatusOK},
	{"post-sa-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "2024-01-01"},
		{key: "end", value: "2024-01-05"},
	}, http.StatusOK},
	{"post-mr", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "John"},
		{key: "last_name", value: "Smith"},
		{key: "email", value: "john.smith@example.com"},
		{key: "phone", value: "1111-222-333"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	var routes = getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method != "POST" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Errorf("%s: error running request: %q", e.name, err)
			}
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("%s: bad status code. Expected %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		} else {
			values := url.Values{}
			for _, param := range e.params {
				values.Add(param.key, param.value)
			}
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Errorf("%s: error running request: %q", e.name, err)
			}
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("%s: bad status code. Expected %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
