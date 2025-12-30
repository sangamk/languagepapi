package handlers

import (
	"net/http"

	"languagepapi/components"
	"languagepapi/internal/service"
)

// lessonService is the shared lesson service instance
var lessonService = service.NewLessonService()

// HandleHome renders the journey home page
func HandleHome(w http.ResponseWriter, r *http.Request) {
	// Get journey home data
	data, err := lessonService.GetJourneyHomeData(defaultUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	components.Home(data).Render(r.Context(), w)
}
