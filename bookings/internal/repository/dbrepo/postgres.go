package dbrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts reservation into database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var newId int
	stmt := `
		insert into reservations(first_name, last_name, email, phone,
			start_date, end_date, room_id, created_at, updated_at)
			values($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomId,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		return 0, err
	}
	return newId, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id,
		created_at, updated_at, restriction_id) values ($1, $2, $3, $4, $5, $6, $7)`
	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)
	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesAndRoomID returns true if room is available for the called period of time
// and false otherwise
func (m *postgresDBRepo) SearchAvailabilityByDatesAndRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `
		select  count(id)
		  from  room_restrictions rr
		 where  room_id = $1 and $2 < rr.end_date and $3 > start_date
	`
	var numRows int

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	return numRows == 0, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms for a range of dates,
// if any
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `
		select  r.id, r.room_name
		  from  rooms r
		 where  r.id not in (
			select  rr.room_id
			  from  room_restrictions rr
			 where  $1 < rr.end_date and $2 > rr.start_date
		 )	
	`
	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Room
	for rows.Next() {
		var room models.Room
		err = rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return nil, err
		}
		result = append(result, room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetRoomByID gets a room from DB by id
func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room
	query := `
		select  id, room_name, created_at, updated_at
		  from  rooms
		 where  id = $1
	`
	row, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return room, err
	}
	if row.Next() {
		err = row.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)
		return room, err
	}
	return room, fmt.Errorf("room with id %d is not found in DB", id)
}

// GetUserById returns a user by id
func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select  id, first_name, last_name, email, password, access_level, created_at, updated_at
		  from  users u
		 where  u.id = $1
	`
	row := m.DB.QueryRowContext(ctx, query, id)
	var u models.User
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.AccessLevel, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return u, err
	}
	return u, nil
}

// UpdateUser updates a user in the database
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update  users
		   set  first_name = &1
		   		last_name = &2
				email = &3
				access_level = &4
				updated_at = &5
	`
	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now())
	return err
}

// Authenticate authenticates the user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return 0, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

// AllReservations returns a slice of all reservations
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation
	query := `
		select  r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date,
				r.room_id, r.created_at, r.updated_at, r.processed, rm.room_name
		  from  reservations r
		  left
		  join  rooms rm
		    on  r.room_id = rm.id
		 order  by
		        r.start_date desc
`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Reservation
		err = rows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.Email, &r.Phone, &r.StartDate, &r.EndDate,
			&r.RoomId, &r.CreatedAt, &r.UpdatedAt, &r.Processed, &r.Room.RoomName)
		if err != nil {
			return reservations, err
		}
		r.Room.ID = r.RoomId
		reservations = append(reservations, r)
	}
	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

// NewReservations returns a slice of all reservations
func (m *postgresDBRepo) NewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation
	query := `
		select  r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date,
				r.room_id, r.created_at, r.updated_at, rm.room_name
		  from  reservations r
		  left
		  join  rooms rm
		    on  r.room_id = rm.id
		 where  r.processed = 0
		 order  by
		        r.start_date desc
`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Reservation
		err = rows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.Email, &r.Phone, &r.StartDate, &r.EndDate,
			&r.RoomId, &r.CreatedAt, &r.UpdatedAt, &r.Room.RoomName)
		if err != nil {
			return reservations, err
		}
		r.Room.ID = r.RoomId
		reservations = append(reservations, r)
	}
	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}
