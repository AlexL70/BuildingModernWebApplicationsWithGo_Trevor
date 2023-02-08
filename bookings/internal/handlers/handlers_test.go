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
	reqBody := "first_name=John"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john.smith@email.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1111-222-333")
	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "post-reservation", http.StatusSeeOther, rr.Code)
	}

	// missing reservation in session case
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	_ = session.Pop(ctx, "reservation")
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "post-reservation-no-reservation", http.StatusTemporaryRedirect, rr.Code)
	}

	// missing post body case
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "post-reservation-no-body", http.StatusTemporaryRedirect, rr.Code)
	}

	// invalid form data case
	reqBody = "first_name=Jo" // less thand 3 characters long
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john.smith@email.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1111-222-333")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "post-reservation-invalid-form", http.StatusBadRequest, rr.Code)
	}

	// error inserting reservation to DB case
	reqBody = "first_name=John"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john.smith@email.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=1111-222-333")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	reservation.RoomId = 2
	session.Put(ctx, "reservation", reservation)
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "post-reservation-error-adding-reservation", http.StatusTemporaryRedirect, rr.Code)
	}

	// error inserting restriction to DB case
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	reservation.RoomId = 1000
	session.Put(ctx, "reservation", reservation)
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("%s: bad status code. Expected %d, but got %d", "post-reservation-error-adding-restriction", http.StatusTemporaryRedirect, rr.Code)
	}

}

func TestRepository_AvaliabilityJSON(t *testing.T) {
	// no available rooms
	reqBody := "start=2060-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2060-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
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
	if j.Message != "" {
		t.Errorf("not-available: got error when did not expected one: %q", j.Message)
	}
	if j.OK {
		t.Errorf("not-available: got available room when does not expected")
	}
}

func getCtx(req *http.Request) context.Context {
	sHeader := req.Header.Get("X-Session")
	ctx, err := session.Load(req.Context(), sHeader)
	if err != nil {
		log.Println(err)
	}
	return ctx
}
