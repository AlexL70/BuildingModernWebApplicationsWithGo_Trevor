package dbrepo

import (
	"errors"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

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
