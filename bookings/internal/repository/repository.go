package repository

import "github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) error
}
