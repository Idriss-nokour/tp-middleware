package ressources

import (
	"encoding/json"
	"net/http"
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	"middleware/example/internal/services/ressources"
)

func GetRessource(w http.ResponseWriter, r *http.Request) {
	ressourceId := r.Context().Value("ressourceId").(uuid.UUID)

	ressource, err := ressources.GetRessourceByID(ressourceId)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ressource)
}
