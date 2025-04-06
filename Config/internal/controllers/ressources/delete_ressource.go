package ressources

import (
	"net/http"
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	"middleware/example/internal/services/ressources"
)

func DeleteRessource(w http.ResponseWriter, r *http.Request) {
	ressourceId := r.Context().Value("ressourceId").(uuid.UUID)

	err := ressources.DeleteRessource(ressourceId)
	if err != nil {
		http.Error(w, err.(*models.CustomError).Message, err.(*models.CustomError).Code)
		return
	}

	w.WriteHeader(http.StatusNoContent) 
}
