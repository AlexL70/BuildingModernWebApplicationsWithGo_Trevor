package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
)

var theTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{"home", "/", http.StatusOK},
	{"about", "/about", http.StatusOK},
	{"gq", "/generals-quoters", http.StatusOK},
	{"ms", "/majors-suite", http.StatusOK},
	{"sa", "/search-availability", http.StatusOK},
	{"contact", "/contact", http.StatusOK},
	//{"mkres", "/make-reservation", "GET", []postData{}, http.StatusOK},
	//{"post-sa", "/search-availability", "POST", []postData{
	//	{key: "start", value: "2024-01-01"},
	//	{key: "end", value: "2024-01-05"},
	//}, http.StatusOK},
	//{"post-sa-json", "/search-availability-json", "POST", []postData{
	//	{key: "start", value: "2024-01-01"},
	//	{key: "end", value: "2024-01-05"},
	//}, http.StatusOK},
	//{"post-mr", "/make-reservation", "POST", []postData{
	//	{key: "first_name", value: "John"},
	//	{key: "last_name", value: "Smith"},
	//	{key: "email", value: "john.smith@example.com"},
	//	{key: "phone", value: "1111-222-333"},
	//}, http.StatusOK},
}

func TestGetHandlers(t *testing.T) {
	var routes = getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Errorf("%s: error running request: %q", e.name, err)
		}
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("%s: bad status code. Expected %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomId: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quoters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "mkres", http.StatusOK, rr.Code)
	}

	// test situation when there is no reservation stored in session
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "mkres", http.StatusTemporaryRedirect, rr.Code)
	}
	// test situation when room does not exist in DB
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomId = 3
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "mkres", http.StatusTemporaryRedirect, rr.Code)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	reservation := models.Reservation{
		StartDate: time.Date(2060, 11, 11, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2060, 11, 15, 0, 0, 0, 0, time.UTC),
		RoomId:    1,
	}
	tests := []struct {
		name           string
		params         map[string]string
		reservation    *models.Reservation
		expectedStatus int
	}{
		{"success", map[string]string{
			"first_name": "John",
			"last_name":  "Smith",
			"email":      "john.smith@email.com",
			"phone":      "1111-222-333",
		}, &reservation, http.StatusSeeOther,
		},
		{"no-reservation-in-session", map[string]string{
			"first_name": "John",
			"last_name":  "Smith",
			"email":      "john.smith@email.com",
			"phone":      "1111-222-333",
		}, nil, http.StatusTemporaryRedirect,
		},
		{"no-request-body", nil, &reservation, http.StatusTemporaryRedirect},
		{"invalid-form", map[string]string{
			"first_name": "Jo", // name must be at least 3 characters long
			"last_name":  "Smith",
			"email":      "john.smith@email.com",
			"phone":      "1111-222-333",
		}, &reservation, http.StatusBadRequest,
		},
		{"error-inserting-reservation", map[string]string{
			"first_name": "John",
			"last_name":  "Smith",
			"email":      "john.smith@email.com",
			"phone":      "1111-222-333",
		}, &models.Reservation{
			StartDate: time.Date(2060, 11, 11, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2060, 11, 15, 0, 0, 0, 0, time.UTC),
			RoomId:    2, // inserting reservation in testDBRepo fails when RoomId = 2
		}, http.StatusTemporaryRedirect,
		},
		{"error-inserting-restriction", map[string]string{
			"first_name": "John",
			"last_name":  "Smith",
			"email":      "john.smith@email.com",
			"phone":      "1111-222-333",
		}, &models.Reservation{
			StartDate: time.Date(2060, 11, 11, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2060, 11, 15, 0, 0, 0, 0, time.UTC),
			RoomId:    1000, // inserting restriction in testDBRepo fails when RoomId = 1000
		}, http.StatusTemporaryRedirect,
		},
	}

	for _, e := range tests {
		var reqBody string
		var req *http.Request
		if e.params != nil {
			reqBody = composeUrlParams(e.params)
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		if e.reservation != nil {
			session.Put(ctx, "reservation", *e.reservation)
		}
		handler := http.HandlerFunc(Repo.PostReservation)
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expectedStatus {
			t.Errorf("%s: bad status code. Expected %d, but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}

func TestRepository_AvaliabilityJSON(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]string
		available bool
		message   string
	}{
		{"no-available-rooms", map[string]string{
			"start":   "2060-01-01",
			"end":     "2060-01-02",
			"room_id": "1",
		}, false, ""},
		{"invalid-start-date", map[string]string{
			"start":   "invalid",
			"end":     "2060-01-02",
			"room_id": "1",
		}, false, "Error parsing start date"},
		{"invalid-end-date", map[string]string{
			"start":   "2060-01-01",
			"end":     "invalid",
			"room_id": "1",
		}, false, "Error parsing end date"},
		{"invalid-room-id", map[string]string{
			"start":   "2060-01-01",
			"end":     "2060-01-02",
			"room_id": "invalid",
		}, false, "Error parsing room id"},
		{"wrong-date-interval", map[string]string{
			"start":   "2060-01-02",
			"end":     "2060-01-01",
			"room_id": "2",
		}, false, "Error: the end date cannot be before the start date"},
		{"db-search-error", map[string]string{
			"start":   "2060-01-01",
			"end":     "2060-01-02",
			"room_id": "3",
		}, false, "Error searching availability"},
		{"success", map[string]string{
			"start":   "2060-02-01",
			"end":     "2060-02-02",
			"room_id": "1",
		}, true, ""},
	}
	for _, e := range tests {
		reqBody := composeUrlParams(e.params)
		req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		var j availabilityResponse
		err := json.Unmarshal(rr.Body.Bytes(), &j)
		if err != nil {
			t.Error("error parsing json")
		}
		if j.Message != e.message {
			t.Errorf("%s: wrong error message; expected %q but got %q", e.name, e.message, j.Message)
		}
		if j.OK != e.available {
			t.Errorf("%s: error getting availability; expected %t but got %t", e.name, e.available, j.OK)
		}
	}
}

func composeUrlParams(params map[string]string) string {
	var result string
	for k, v := range params {
		if result == "" {
			result = fmt.Sprintf("%s=%s", k, v)
		} else {
			result = fmt.Sprintf("%s&%s=%s", result, k, v)
		}
	}
	return result
}

func getCtx(req *http.Request) context.Context {
	sHeader := req.Header.Get("X-Session")
	ctx, err := session.Load(req.Context(), sHeader)
	if err != nil {
		log.Println(err)
	}
	return ctx
}
