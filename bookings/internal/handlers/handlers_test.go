package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
	"github.com/go-chi/chi/v5"
)

var theTests = []struct {
	name                string
	url                 string
	expectedStatusCode  int
	userIsLoggedIn      bool
	checkPreviousStatus bool
	dbFetchError        bool
}{
	{"home", "/", http.StatusOK, false, false, false},
	{"about", "/about", http.StatusOK, false, false, false},
	{"gq", "/generals-quoters", http.StatusOK, false, false, false},
	{"ms", "/majors-suite", http.StatusOK, false, false, false},
	{"sa", "/search-availability", http.StatusOK, false, false, false},
	{"contact", "/contact", http.StatusOK, false, false, false},
	{"non-existent", "/eggs/and/ham", http.StatusNotFound, false, false, false},
	{"login", "/user/login", http.StatusOK, false, false, false},
	{"logout", "/user/logout", http.StatusSeeOther, true, true, false},
	{"dashboard", "/admin/dashboard", http.StatusOK, true, false, false},
	{"dashboard-denied", "/admin/dashboard", http.StatusSeeOther, false, true, false},
	{"new-reservations-success", "/admin/reservations-new", http.StatusOK, true, false, false},
	{"new-reservations-dberror", "/admin/reservations-new", http.StatusTemporaryRedirect, true, true, true},
	{"new-reservations-denied", "/admin/reservations-new", http.StatusSeeOther, false, true, false},
	{"all-reservations-success", "/admin/reservations-all", http.StatusOK, true, false, false},
	{"all-reservations-dberror", "/admin/reservations-all", http.StatusTemporaryRedirect, true, true, true},
	{"all-reservations-denied", "/admin/reservations-all", http.StatusSeeOther, false, true, false},
	{"show-reservation-success", "/admin/reservations/new/1", http.StatusOK, true, false, false},
	{"show-reservation-denied", "/admin/reservations/new/1", http.StatusSeeOther, false, true, false},
}

func TestGetHandlers(t *testing.T) {
	var routes = getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		fetchError = true
		IsAuthenticated = e.userIsLoggedIn
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Errorf("%s: error running request: %q", e.name, err)
		}
		if !e.checkPreviousStatus && resp.StatusCode != e.expectedStatusCode {
			t.Errorf("%s: bad status code. Expected %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
		if e.checkPreviousStatus && resp.Request.Response.StatusCode != e.expectedStatusCode {
			t.Errorf("%s: bad status code. Expected %d, but got %d", e.name, e.expectedStatusCode, resp.Request.Response.StatusCode)
		}
		fetchError = false
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]string
		reservation    *models.Reservation
		expectedStatus int
		expectedError  string
	}{
		{"invalid-start-date", map[string]string{"start": "invalid", "end": "2060-01-10"},
			nil, http.StatusTemporaryRedirect, "Error parsing start date"},
		{"invalid-end-date", map[string]string{"start": "2060-01-01", "end": "invlid"},
			nil, http.StatusTemporaryRedirect, "Error parsing end date"},
		{"wrong-date-interval", map[string]string{"start": "2060-01-02", "end": "2060-01-01"},
			nil, http.StatusTemporaryRedirect, "Error: the end date cannot be before the start date"},
		{"db-error", map[string]string{"start": "2023-01-01", "end": "2023-01-02"}, nil,
			http.StatusTemporaryRedirect, "Error searching availability in DB"},
		{"no-available-rooms", map[string]string{"start": "2060-01-01", "end": "2060-01-02"}, nil,
			http.StatusSeeOther, "No available rooms for this period. Sorry!"},
		{"success", map[string]string{"start": "2060-01-05", "end": "2060-01-06"}, nil,
			http.StatusOK, ""},
	}

	for _, e := range tests {
		var reqBody string
		var req *http.Request
		if e.params != nil {
			reqBody = composeUrlParams(e.params)
			req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
		} else {
			req, _ = http.NewRequest("POST", "/search-availability", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		handler := http.HandlerFunc(Repo.PostAvailability)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expectedStatus {
			t.Errorf("%s: bad status code. Expected %d, but got %d", e.name, e.expectedStatus, rr.Code)
		}
		errStr := session.PopString(req.Context(), "error")
		if errStr != e.expectedError {
			t.Errorf("%s: unexpected error message; expected %q but got %q", e.name, e.expectedError, errStr)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	tests := []struct {
		name           string
		reservation    *models.Reservation
		expectedStatus int
	}{
		{name: "success", reservation: &models.Reservation{
			RoomId:    1,
			StartDate: time.Date(2060, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2060, 1, 10, 0, 0, 0, 0, time.UTC),
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		}, expectedStatus: http.StatusOK},
		{name: "no-reservation", reservation: nil,
			expectedStatus: http.StatusTemporaryRedirect},
		{name: "no-room", reservation: &models.Reservation{
			RoomId: 3,
			Room: models.Room{
				ID:       3,
				RoomName: "Non-existent room",
			},
		}, expectedStatus: http.StatusTemporaryRedirect},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if e.reservation != nil {
			session.Put(ctx, "reservation", *e.reservation)
		}
		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expectedStatus {
			t.Errorf("%s: bad status code. Expected %d, but got %d", "success", e.expectedStatus, rr.Code)
		}
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

func TestRepository_ReservationSummary(t *testing.T) {
	tests := []struct {
		name            string
		reservation     *models.Reservation
		expectedStatus  int
		expectedMessage string
	}{
		{name: "success", reservation: &models.Reservation{
			FirstName: "John",
			LastName:  "Smith",
			Email:     "john.smith@email.com",
			Phone:     "5555-111-222",
			StartDate: time.Date(2060, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2060, 1, 10, 0, 0, 0, 0, time.UTC),
			RoomId:    1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quaters",
			},
		}, expectedStatus: http.StatusOK,
		},
		{name: "no-reservation", reservation: nil, expectedStatus: http.StatusTemporaryRedirect,
			expectedMessage: "Cannot get reservation from the session"},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("GET", "/reservation-summary", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		if e.reservation != nil {
			session.Put(ctx, "reservation", *e.reservation)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.ReservationSummary)
		handler.ServeHTTP(rr, req)
		if e.expectedStatus != rr.Code {
			t.Errorf("%s: wrong status code; expected %d but got %d", e.name, e.expectedStatus, rr.Code)
		}
		actualMessage := session.PopString(req.Context(), "error")
		if e.expectedMessage != actualMessage {
			t.Errorf("%s: bad error message; expected %q but got %q", e.name, e.expectedMessage, actualMessage)
		}
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	tests := []struct {
		name            string
		reservation     *models.Reservation
		roomId          string
		expectedStatus  int
		expectedMessage string
	}{
		{name: "success", reservation: &models.Reservation{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC),
		}, roomId: "1", expectedStatus: http.StatusSeeOther, expectedMessage: ""},
		{name: "no-reservation", reservation: nil, roomId: "1", expectedStatus: http.StatusTemporaryRedirect,
			expectedMessage: "Error getting reservation from the session"},
		{name: "invalid-room-id", reservation: nil, roomId: "invalid", expectedStatus: http.StatusTemporaryRedirect,
			expectedMessage: "Invalid room id"},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/choose-room/%s", e.roomId), nil)
		ctx := getCtx(req)
		ctx = addParamsToChiContext(ctx, map[string]string{"id": e.roomId})
		req = req.WithContext(ctx)
		if e.reservation != nil {
			session.Put(ctx, "reservation", *e.reservation)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.ChooseRoom)
		handler.ServeHTTP(rr, req)

		if e.expectedStatus != rr.Code {
			t.Errorf("%s: wrong status code; expected %d but got %d", e.name, e.expectedStatus, rr.Code)
		}
		actualMessage := app.Session.PopString(ctx, "error")
		if e.expectedMessage != actualMessage {
			t.Errorf("%s: bad error message; expected %q but got %q", e.name, e.expectedMessage, actualMessage)
		}
		if e.reservation != nil {
			reservation := app.Session.Get(ctx, "reservation").(models.Reservation)
			if strconv.Itoa(reservation.RoomId) != e.roomId {
				t.Errorf("%s: room id was not written to the reservation; expected %s but got %d", e.name, e.roomId, reservation.RoomId)
			}
		}
	}
}

func TestRepository_BookRoom(t *testing.T) {
	tests := []struct {
		name            string
		params          map[string]string
		expectedStatus  int
		expectedMessage string
	}{
		{name: "success", params: map[string]string{
			"id": "1", "start": "2060-01-01", "end": "2060-01-10",
		}, expectedStatus: http.StatusSeeOther, expectedMessage: ""},
		{name: "invalid-room-id", params: map[string]string{
			"id": "invalid", "start": "2060-01-01", "end": "2060-01-10",
		}, expectedStatus: http.StatusTemporaryRedirect, expectedMessage: "Error parsing room id"},
		{name: "invalid-start-date", params: map[string]string{
			"id": "1", "start": "invalid", "end": "2060-01-10",
		}, expectedStatus: http.StatusTemporaryRedirect, expectedMessage: "Error parsing start date"},
		{name: "invalid-room-id", params: map[string]string{
			"id": "1", "start": "2060-01-01", "end": "invalid",
		}, expectedStatus: http.StatusTemporaryRedirect, expectedMessage: "Error parsing end date"},
		{name: "room-does-not-exist", params: map[string]string{
			"id": "3", "start": "2060-01-01", "end": "2060-01-10",
		}, expectedStatus: http.StatusTemporaryRedirect, expectedMessage: "Error getting room from DB"},
	}
	for _, e := range tests {
		params := composeUrlParams(e.params)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/book-room?%s", params), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.BookRoom)
		handler.ServeHTTP(rr, req)
		if e.expectedStatus != rr.Code {
			t.Errorf("%s: wrong status code; expected %d but got %d", e.name, e.expectedStatus, rr.Code)
		}
		actualMessage := session.GetString(ctx, "error")
		if e.expectedMessage != actualMessage {
			t.Errorf("%s: bad error message; expected %q but gor %q", e.name, e.expectedMessage, actualMessage)
		}
		if e.expectedMessage == "" {
			expectedRoomId, _ := strconv.Atoi(e.params["id"])
			reservation := session.Get(ctx, "reservation").(models.Reservation)
			if reservation.RoomId != expectedRoomId {
				t.Errorf("%s: wrong room id in reservation; expected %d but got %d", e.name, expectedRoomId, reservation.RoomId)
			}
		}
	}
}

func TestRepository_Login(t *testing.T) {
	tests := []struct {
		name               string
		email              string
		expectedStatusCode int
		expectedHtml       string
		expectedLocation   string
	}{
		{"valid-creds", "me@here.ca", http.StatusSeeOther, "", "/"},
		{"invalid-creds", "jack@nimble.com", http.StatusSeeOther, "", "/user/login"},
		{"validation-error", "this.is.not.an.email", http.StatusOK, `action="/user/login`, ""},
	}

	for _, e := range tests {
		postedData := url.Values{
			"email":    {e.email},
			"password": {"password"},
		}
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: bad status code; expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		if e.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != e.expectedLocation {
				t.Errorf("%s: bad location; expected %q, but got %q", e.name, e.expectedLocation, actualLocation.String())
			}
		}
		if e.expectedHtml != "" {
			actualHTML := rr.Body.String()
			if !strings.Contains(actualHTML, e.expectedHtml) {
				t.Errorf("%s: expected to find %q in result but did not; actual result is %q", e.name, e.expectedHtml, actualHTML)
			}
		}
	}
}

func TestRepository_PostShowReservation(t *testing.T) {
	myForm := map[string]string{
		"first_name": "John",
		"last_name":  "Smith",
		"email":      "john.smith@email.com",
		"year":       "2023",
		"month":      "06",
	}
	tests := []struct {
		name                 string
		url                  string
		formFields           map[string]string
		reservationID        string
		source               string
		expectedStatusCode   int
		expectedLocation     string
		expectedSessionKey   string
		expectedSessionValue string
		dbFetchError         bool
	}{
		{"bad id", "/admin/reservations/{src}/{id}", myForm, "badid", "all",
			http.StatusTemporaryRedirect, "/admin/dashboard", "error", "Invalid reservation id", false},
		{"error fetching reservation", "/admin/reservations/{src}/{id}", myForm, "10", "all",
			http.StatusTemporaryRedirect, "/admin/dashboard", "error", "Error getting reservation from DB", true},
		{"error updating reservation", "/admin/reservations/{src}/{id}", map[string]string{"first_name": "error"}, "10", "all",
			http.StatusTemporaryRedirect, "/admin/dashboard", "error", "Error updating reservation in DB", false},
		{"success all", "/admin/reservations/{src}/{id}", myForm, "10", "all",
			http.StatusSeeOther, "/admin/reservations-all", "flash", "Changes successfully saved", false},
		{"success new", "/admin/reservations/{src}/{id}", myForm, "10", "new",
			http.StatusSeeOther, "/admin/reservations-new", "flash", "Changes successfully saved", false},
		{"success cal", "/admin/reservations/{src}/{id}", myForm, "10", "cal",
			http.StatusSeeOther, "/admin/reservations-calendar?y=2023&m=06", "flash", "Changes successfully saved", false},
	}
	for _, e := range tests {
		fetchError = e.dbFetchError
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(composeUrlParams(e.formFields)))
		ctx := getCtx(req)
		ctx = addParamsToChiContext(ctx, map[string]string{"id": e.reservationID, "src": e.source})
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: bad status code; expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		if e.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != e.expectedLocation {
				t.Errorf("%s: bad location; expected %q, but got %q", e.name, e.expectedLocation, actualLocation.String())
			}
		}
		if e.expectedSessionKey != "" {
			value := app.Session.Pop(ctx, e.expectedSessionKey)
			if e.expectedSessionValue != value {
				t.Errorf("%s: got an unexpected %q value from session; expected %q but got %q", e.name, e.expectedSessionKey, e.expectedSessionValue, value)
			}
		}
		fetchError = false
	}
}

func TestRepository_ReservationCalendar(t *testing.T) {
	now := time.Now()
	nowYear := now.Format("2006")
	nowMonthName := now.Format("January")
	tests := []struct {
		name                  string
		year                  string
		month                 string
		expectedStatusCode    int
		expectedLocation      string
		expectedHtml          string
		dbFetchError          bool
		expectedSessionValues map[string]string
	}{
		{"now", "", "", http.StatusOK, "", fmt.Sprintf("%s %s", nowMonthName, nowYear), false, map[string]string{}},
		{"2024-05", "2024", "05", http.StatusOK, "", "May 2024", false, map[string]string{}},
		{"error fetching rooms", "", "", http.StatusTemporaryRedirect, "/admin/dashboard", "", true, map[string]string{"error": "Error fetching rooms from DB"}},
	}

	for _, e := range tests {
		fetchError = e.dbFetchError
		var req *http.Request
		if e.year != "" {
			req, _ = http.NewRequest("GET", fmt.Sprintf("/reservations-calendar?%s", composeUrlParams(map[string]string{"y": e.year, "m": e.month})), nil)
		} else {
			req, _ = http.NewRequest("GET", "/reservations-calendar", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminReservationsCalendar)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: bad status code; expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		if e.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != e.expectedLocation {
				t.Errorf("%s: bad location; expected %q, but got %q", e.name, e.expectedLocation, actualLocation.String())
			}
		}
		if e.expectedHtml != "" {
			actualHTML := rr.Body.String()
			if !strings.Contains(actualHTML, e.expectedHtml) {
				t.Errorf("%s: expected to find %q in result but did not; actual result is %q", e.name, e.expectedHtml, actualHTML)
			}
		}
		for k, v := range e.expectedSessionValues {
			value := app.Session.Pop(ctx, k)
			if v != value {
				t.Errorf("%s: got an unexpected %q value from session; expected %q but got %q", e.name, k, v, value)
			}
		}
		fetchError = false
	}
}

func TestRepository_ProcessReservation(t *testing.T) {
	tests := []struct {
		name                  string
		id                    string
		source                string
		queryParams           map[string]string
		expectedStatusCode    int
		expectedLocation      string
		expectedSessionValues map[string]string
	}{
		{"bad id", "badid", "all", map[string]string{}, http.StatusTemporaryRedirect, "/admin/dashboard",
			map[string]string{"error": "Invalid reservation id"}},
		{"db-error", "100", "all", map[string]string{}, http.StatusTemporaryRedirect, "/admin/dashboard",
			map[string]string{"error": "Error marking reservation as processed"}},
		{"success-all", "10", "all", map[string]string{}, http.StatusSeeOther, "/admin/reservations-all",
			map[string]string{"flash": "Successfully marked as processed"}},
		{"success-new", "11", "new", map[string]string{}, http.StatusSeeOther, "/admin/reservations-new",
			map[string]string{"flash": "Successfully marked as processed"}},
		{"success-cal", "11", "cal", map[string]string{"y": "2025", "m": "04"}, http.StatusSeeOther, "/admin/reservations-calendar?y=2025&m=04",
			map[string]string{"flash": "Successfully marked as processed"}},
	}

	for _, e := range tests {
		var req *http.Request
		req, _ = http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/{src}/{id}?%s", composeUrlParams(e.queryParams)), nil)
		ctx := getCtx(req)
		ctx = addParamsToChiContext(ctx, map[string]string{"id": e.id, "src": e.source})
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: bad status code; expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		if e.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != e.expectedLocation {
				t.Errorf("%s: bad location; expected %q, but got %q", e.name, e.expectedLocation, actualLocation.String())
			}
		}
		for k, v := range e.expectedSessionValues {
			value := app.Session.Pop(ctx, k)
			if v != value {
				t.Errorf("%s: got an unexpected %q value from session; expected %q but got %q", e.name, k, v, value)
			}
		}
	}
}

func TestRepository_DeleteReservation(t *testing.T) {
	tests := []struct {
		name                  string
		id                    string
		source                string
		queryParams           map[string]string
		expectedStatusCode    int
		expectedLocation      string
		expectedSessionValues map[string]string
	}{
		{"bad id", "badid", "all", map[string]string{}, http.StatusTemporaryRedirect, "/admin/dashboard",
			map[string]string{"error": "Invalid reservation id"}},
		{"db-error", "100", "all", map[string]string{}, http.StatusTemporaryRedirect, "/admin/dashboard",
			map[string]string{"error": "Error deleting reservation"}},
		{"success-all", "10", "all", map[string]string{}, http.StatusSeeOther, "/admin/reservations-all",
			map[string]string{"flash": "Successfully deleted reservation"}},
		{"success-new", "11", "new", map[string]string{}, http.StatusSeeOther, "/admin/reservations-new",
			map[string]string{"flash": "Successfully deleted reservation"}},
		{"success-cal", "11", "cal", map[string]string{"y": "2025", "m": "04"}, http.StatusSeeOther, "/admin/reservations-calendar?y=2025&m=04",
			map[string]string{"flash": "Successfully deleted reservation"}},
	}

	for _, e := range tests {
		var req *http.Request
		req, _ = http.NewRequest("GET", fmt.Sprintf("/admin/delete-reservation/{src}/{id}?%s", composeUrlParams(e.queryParams)), nil)
		ctx := getCtx(req)
		ctx = addParamsToChiContext(ctx, map[string]string{"id": e.id, "src": e.source})
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: bad status code; expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		if e.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != e.expectedLocation {
				t.Errorf("%s: bad location; expected %q, but got %q", e.name, e.expectedLocation, actualLocation.String())
			}
		}
		for k, v := range e.expectedSessionValues {
			value := app.Session.Pop(ctx, k)
			if v != value {
				t.Errorf("%s: got an unexpected %q value from session; expected %q but got %q", e.name, k, v, value)
			}
		}
	}
}

func TestPostReservationCalendar(t *testing.T) {
	var tests = []struct {
		name                 string
		postedData           url.Values
		expectedResponseCode int
		expectedLocation     string
		expectedHTML         string
		blocks               int
		reservations         int
	}{
		{
			name: "cal",
			postedData: url.Values{
				"year":  {time.Now().Format("2006")},
				"month": {time.Now().Format("01")},
				fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
			},
			expectedResponseCode: http.StatusSeeOther,
		},
		{
			name:                 "cal-blocks",
			postedData:           url.Values{},
			expectedResponseCode: http.StatusSeeOther,
			blocks:               1,
		},
		{
			name:                 "cal-res",
			postedData:           url.Values{},
			expectedResponseCode: http.StatusSeeOther,
			reservations:         1,
		},
	}

	for _, e := range tests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

	}
}

func composeUrlParams(params map[string]string) string {
	postedData := url.Values{}
	for k, v := range params {
		postedData.Add(k, v)
	}
	return postedData.Encode()
}

func getCtx(req *http.Request) context.Context {
	sHeader := req.Header.Get("X-Session")
	ctx, err := app.Session.Load(req.Context(), sHeader)
	if err != nil {
		log.Println(err)
	}
	return ctx
}

func addParamsToChiContext(parentCtx context.Context, params map[string]string) context.Context {
	chiCtx := chi.NewRouteContext()
	for k, v := range params {
		chiCtx.URLParams.Add(k, v)
	}
	return context.WithValue(parentCtx, chi.RouteCtxKey, chiCtx)
}
