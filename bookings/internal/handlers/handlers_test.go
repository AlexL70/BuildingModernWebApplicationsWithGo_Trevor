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
		ctx = addIdToChiContext(ctx, e.roomId)
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

func addIdToChiContext(parentCtx context.Context, id string) context.Context {
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", id)
	return context.WithValue(parentCtx, chi.RouteCtxKey, chiCtx)
}
