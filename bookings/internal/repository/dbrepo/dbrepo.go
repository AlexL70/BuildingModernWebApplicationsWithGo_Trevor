package dbrepo

import (
	"database/sql"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/config"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type testDBRepo struct {
	App        *config.AppConfig
	DB         *sql.DB
	FetchError *bool
}

func NewPostresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

func NewTestingRepo(a *config.AppConfig, fetchError *bool) repository.DatabaseRepo {
	return &testDBRepo{
		App:        a,
		DB:         nil,
		FetchError: fetchError,
	}
}
