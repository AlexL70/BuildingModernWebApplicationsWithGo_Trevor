package handlers

import (
	"net/http"

	"github.com/AlexL70/go-hello-world/pkg/config"
	"github.com/AlexL70/go-hello-world/pkg/models"
	"github.com/AlexL70/go-hello-world/pkg/render"
)

// Repo is the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository
func NewRepo(ac *config.AppConfig) *Repository {
	return &Repository{
		App: ac,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	render.RenderTemplate(w, "home.page.gohtml", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	strMap := map[string]string{}
	strMap["test"] = "Hello again!"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	strMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, "about.page.gohtml", &models.TemplateData{StringMap: strMap})
}
