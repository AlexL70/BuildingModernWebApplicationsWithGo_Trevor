package render

import (
	"net/http"
	"testing"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}
	const flashVal = "My flash"
	session.Put(r.Context(), "flash", flashVal)
	result := AddDefaultData(&td, r)
	if result == nil {
		t.Error("failed!")
	}
	if result.Flash != flashVal {
		t.Errorf("Flash does not work right. Expected %q, but got %q", flashVal, result.Flash)
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, "X-Session")
	r = r.WithContext(ctx)

	return r, nil
}
