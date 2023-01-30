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

func TestRenderTemplate(t *testing.T) {
	oldPath := pathToTemplates
	pathToTemplates = "./../../templates"

	tc, err := CreateTemplateCache()
	if err != nil {
		t.Errorf("error creating template cache: %q", err)
	}
	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Errorf("error getting session: %q", err)
	}

	ww := &myWriter{}

	err = RenderTemplate(ww, r, "home.page.gohtml", &models.TemplateData{})
	if err != nil {
		t.Errorf("error rendering template: %q", err)
	}

	err = RenderTemplate(ww, r, "non-existent.page.gohtml", &models.TemplateData{})
	if err == nil {
		t.Error("error rendering template. Non-existent template was rendered without a error")
	}

	pathToTemplates = oldPath
}

func TestNewTemplate(t *testing.T) {
	NewTemplates(app)
}

func TestCreateTemplateCache(t *testing.T) {
	oldPath := pathToTemplates
	pathToTemplates = "./../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	pathToTemplates = oldPath
}
