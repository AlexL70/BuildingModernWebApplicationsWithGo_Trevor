package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/config"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/driver"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/forms"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/helpers"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/render"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/repository"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

// Repo is the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(ac *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: ac,
		DB:  dbrepo.NewPostresRepo(db.SQL, ac),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.gohtml", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.gohtml", &models.TemplateData{})
}

// Generals is Generals' Quoters page handler
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.gohtml", &models.TemplateData{})
}

// Majors is Majors' Suite page handler
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.gohtml", &models.TemplateData{})
}

// Availability is Search Availability page hander
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.gohtml", &models.TemplateData{})
}

// PostAvailability is Search Availability page hander
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	// dates format is 2023-01-01 -- stardard format is: "01/02 03:04:05PM '06 -0700"
	const layout = "2006-01-02"
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if endDate.Before(startDate) {
		m.App.Session.Put(r.Context(), "error", "Error: the end date cannot be before the start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	available, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error searching availability in DB")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(available) == 0 {
		m.App.Session.Put(r.Context(), "error", "No available rooms for this period. Sorry!")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}
	data := map[string]any{}
	data["rooms"] = available
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}
	m.App.Session.Put(r.Context(), "reservation", res)
	render.Template(w, r, "choose-room.page.gohtml", &models.TemplateData{Data: data})
}

type availabilityResponse struct {
	OK        bool   `json:"ok"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Message   string `json:"message"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	jsonError := func(err error, params ...string) {
		// can't parse form so return appropriate JSON
		resp := availabilityResponse{
			OK:      false,
			Message: "Internal server error",
		}
		if len(params) == 1 {
			resp.Message = params[0]
		}

		out, _ := json.MarshalIndent(resp, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}

	err := r.ParseForm()
	if err != nil {
		jsonError(err)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		jsonError(err, "Error parsing start date")
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		jsonError(err, "Error parsing end date")
		return
	}
	if endDate.Before(startDate) {
		jsonError(err, "Error: the end date cannot be before the start date")
		return
	}
	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		jsonError(err, "Error parsing room id")
		return
	}
	available, err := m.DB.SearchAvailabilityByDatesAndRoomID(startDate, endDate, roomID)
	if err != nil {
		jsonError(err, "Error searching availability")
		return
	}

	resp := availabilityResponse{
		OK:        available,
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}
	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Reservation is Make Reservation page handler
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from the session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	strMap := map[string]string{}
	strMap["start_date"] = sd
	strMap["end_date"] = ed

	var err error
	reservation.Room, err = m.DB.GetRoomByID(reservation.RoomId)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Cannot find room in DB")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)
	data["reservation"] = reservation
	render.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: strMap,
	})
}

// PostReservation handles the posting of the reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Cannot pull reservation out of the session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing form.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := map[string]any{}
		data["reservation"] = reservation
		http.Error(w, "Invalid form!", http.StatusBadRequest)
		render.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error inserting reservation to DB")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomId,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error inserting room restriction to DB")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Send an email notification to guest
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong>
		<br>
		Dear %s,
		<br><br>	
		This is to confirm your reservation from %s to %s of %s room in our fantastic Room&Breakfast hotel.
		<br><br>
		Sincerely,<br>
		Honel's administration<br>
		admin@room&breakfast.com
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"),
		reservation.Room.RoomName)
	msg := models.MailData{
		To:       reservation.Email,
		From:     "admin@room&breakfast.com",
		Subject:  "Room reservation confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}
	m.App.MailChan <- msg

	// Send email notifictaion to hotel's owner
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Confirmation</strong>
		<br><br>	
		This is to inform your that reservation was made from %s to %s of %s room by %s %s.
		<br><br>
		<strong>Contact Information</strong><br>
		Email: %s<br>
		Phone#: %s
		<br><br>
		Please do the necessary preparations,<br>
		admin@room&breakfast.com
	`, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"),
		reservation.Room.RoomName, reservation.FirstName, reservation.LastName,
		reservation.Email, reservation.Phone)
	msg = models.MailData{
		To:      "admin@room&breakfast.com",
		From:    "admin@room&breakfast.com",
		Subject: "Room reservation has been made",
		Content: htmlMessage,
	}
	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Contact is Contact page handler
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.gohtml", &models.TemplateData{})
}

// ReservationSummary displays Reservation Summary page after reservation has been made
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	const errMsg = "Cannot get reservation from the session"
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println(errMsg)
		m.App.Session.Put(r.Context(), "error", errMsg)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	m.App.Session.Remove(r.Context(), "reservation")
	var data = map[string]any{}
	data["reservation"] = reservation
	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	strMap := map[string]string{}
	strMap["start_date"] = sd
	strMap["end_date"] = ed
	render.Template(w, r, "reservation-summary.page.gohtml", &models.TemplateData{
		Data:      data,
		StringMap: strMap,
	})
}

// ChooseRoom takes "id" parameter from URL, gets Reservation from the Session,
// fill in RoomID with parameter's value, put it back to the Session and redirect
// to make-reservation page
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid room id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from the session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.RoomId = roomID
	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom reads URL parameters (id,start,end), fill in Reservation
// model, put it into Session and redirect to make-reservation page
// so that user could make reservation of certain room for certain dates
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing room id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	layout := "2006-01-02"
	sd := r.URL.Query().Get("start")
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	ed := r.URL.Query().Get("end")
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error getting room from DB")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	reservation := models.Reservation{
		RoomId:    roomID,
		StartDate: startDate,
		EndDate:   endDate,
		Room:      room,
	}
	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// ShowLogin shows login screen
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := form.Get("email")
	password := form.Get("password")
	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Successful login!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles logging the user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.gohtml", &models.TemplateData{})
}

// AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.NewReservations()
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error getting reservations from DB")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}
	data := map[string]any{}
	data["reservations"] = reservations
	render.Template(w, r, "admin-new-reservations.page.gohtml", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations shows all reservations in admin tool
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error getting reservations from DB")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}
	data := map[string]any{}
	data["reservations"] = reservations
	render.Template(w, r, "admin-all-reservations.page.gohtml", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows one reservation in admin tool
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid reservation id")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from DB")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	data := map[string]any{}
	data["reservation"] = reservation
	stringMap := map[string]string{}
	stringMap["src"] = src

	render.Template(w, r, "admin-reservation-show.page.gohtml", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		Form:      forms.New(nil),
	})
}

// AdminPostShowReservation saves changes in reservation to the database
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error parsing form")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid reservation id")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from DB")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from DB")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Changes successfully saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminReservationsCalendar shows the reservations' calendar in admin tool
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// assume there is not month or year specified
	time.LoadLocation("UTC")
	now := time.Now()
	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := map[string]any{
		"now": now,
	}

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)
	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")
	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := map[string]string{
		"next_month":      nextMonth,
		"next_month_year": nextMonthYear,
		"last_month":      lastMonth,
		"last_month_year": lastMonthYear,
		"this_month":      now.Format("01"),
		"this_month_year": now.Format("2006"),
	}

	// Count a number of days in the current month
	currYear, currMonth, _ := now.Date()
	currLocation := now.Location()
	firstOfMonth := time.Date(currYear, currMonth, 1, 0, 0, 0, 0, currLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	intMap := map[string]int{
		"days_in_month": lastOfMonth.Day(),
	}

	rooms, err := m.DB.AllRooms()
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error fetching rooms from DB")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}
	data["rooms"] = rooms

	for _, room := range rooms {
		reservationMap := map[string]int{}
		blockMap := map[string]int{}
		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			reservationMap[dateStr] = 0
			blockMap[dateStr] = 0
		}

		roomRestrictions, err := m.DB.GetRestrictionsForRoomByDates(room.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			log.Println(err)
			m.App.Session.Put(r.Context(), "error", "Error fetching room restrictions from DB")
			http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		}

		for _, rr := range roomRestrictions {
			for rd := rr.StartDate; !rd.After(rr.EndDate); rd = rd.AddDate(0, 0, 1) {
				dateStr := rd.Format("2006-01-02")
				if rr.ReservationID == 0 {
					blockMap[dateStr] = rr.ID
				} else {
					reservationMap[dateStr] = rr.ReservationID
				}
			}
		}
		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.gohtml", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		IntMap:    intMap,
	})
}

// AdminProcessReservation marks reservation as processed in database
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid reservation id")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	err = m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error marking reservation as processed")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}
	m.App.Session.Put(r.Context(), "flash", "Successfully marked as processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminDeleteReservation deletes reservation from the database
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid reservation id")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	err = m.DB.DeleteReservation(id)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error deleting reservation")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}
	m.App.Session.Put(r.Context(), "flash", "Successfully deleted reservation")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminPostReservationsCalendar handles post of reservation calendar
func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Print(err)
		m.App.Session.Put(r.Context(), "error", "Error parsing the form")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	m.App.Session.Put(r.Context(), "flash", "Changes saved!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}
