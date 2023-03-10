package dbrepo

import (
	"errors"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
)

// InsertReservation inserts reservation into database
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// if the room_id is 2 then fail, otherwise pass
	if res.RoomId == 2 {
		return 0, errors.New("test DB error")
	}
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	if r.RoomID == 1000 {
		return errors.New("test DB error")
	}
	return nil
}

// SearchAvailabilityByDatesAndRoomID returns true if room is available for the called period of time
// and false otherwise
func (m *testDBRepo) SearchAvailabilityByDatesAndRoomID(start, end time.Time, roomID int) (bool, error) {
	if roomID > 2 {
		return false, errors.New("test DB error")
	}
	if start == time.Date(2060, 02, 01, 0, 0, 0, 0, time.UTC) {
		return true, nil
	}
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms for a range of dates,
// if any
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var result []models.Room
	if start == time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC) {
		return result, errors.New("test DB error")
	}
	if start == time.Date(2060, 1, 5, 0, 0, 0, 0, time.UTC) {
		result = append(result, models.Room{ID: 1, RoomName: "General's Quarters"})
	}
	return result, nil
}

// GetRoomByID gets a room from DB by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("test DB error")
	}
	return room, nil
}

// GetUserById returns a user by id
func (m *testDBRepo) GetUserById(id int) (models.User, error) {
	var u models.User
	return u, nil
}

// UpdateUser updates a user in the database
func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

// Authenticate authenticates the user
func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if email == "me@here.ca" {
		return 1, "", nil
	}
	return 0, "", errors.New("invalid login")
}

// AllReservations returns a slice of all reservations
func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	if !*m.FetchError {
		return reservations, nil
	}
	return reservations, errors.New("error fetching reservations")
}

// NewReservations returns a slice of new reservations
func (m *testDBRepo) NewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	if !*m.FetchError {
		return reservations, nil
	}
	return reservations, errors.New("error fetching reservations")
}

// GetReservationByID gets reservation from the DB by ID
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var reservation models.Reservation
	if *m.FetchError {
		return reservation, errors.New("error fetching reservation")
	}
	return reservation, nil
}

// UpdateReservation updates reservation in the database
func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	if u.FirstName == "error" {
		return errors.New("error updating reservation")
	}
	return nil
}

// DeleteReservation deletes one reservation from the DB by id
func (m *testDBRepo) DeleteReservation(id int) error {
	if id == 100 {
		return errors.New("error deleting reservation")
	}
	return nil
}

// UpdateProcessedForReservation updates the processed field (status) of reservation by ID
func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	if id == 100 {
		return errors.New("error updating reservation")
	}
	return nil
}

// AllRooms returns all rooms from the database
func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms []models.Room
	if *m.FetchError {
		return rooms, errors.New("error fetching rooms")
	}
	return rooms, nil
}

// GetRestrictionsForRoomByDates returns restrictions for a room by room id and dates range
func (m *testDBRepo) GetRestrictionsForRoomByDates(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	var restrictions []models.RoomRestriction
	return restrictions, nil
}

// InsertBlockForRoom adds block to DB for the room on the date
func (m *testDBRepo) InsertBlockForRoom(roomID int, startDate time.Time) error {
	return nil
}

// DeleteBlockByID removes block (room restriction) from the DB by ID
func (m *testDBRepo) DeleteBlockByID(restrictionID int) error {
	return nil
}
