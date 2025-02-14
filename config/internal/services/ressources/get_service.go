package ressources

import (
	"github.com/gofrs/uuid"
	"middleware/example/internal/models"
	repository "middleware/example/internal/repositories/ressources"
	"github.com/sirupsen/logrus"
	"net/http"
)

func GetAllRessources() ([]models.Ressources, error) {
	ressources, err := repository.GetAllRessources()
	if err != nil {
		logrus.Errorf("Error retrieving ressources: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Something went wrong while retrieving ressources",
			Code:    http.StatusInternalServerError,
		}
	}
	return ressources, nil
}

func GetRessourceByID(id uuid.UUID) (*models.Ressources, error) {
	ressource, err := repository.GetRessourceByID(id)
	if err != nil {
		logrus.Errorf("Error retrieving ressource by ID: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Something went wrong while retrieving the ressource",
			Code:    http.StatusInternalServerError,
		}
	}
	if ressource == nil {
		return nil, &models.CustomError{
			Message: "Ressource not found",
			Code:    http.StatusNotFound,
		}
	}
	return ressource, nil
}



func CreateRessource(ressource models.Ressources) (*models.Ressources, error) {
	logrus.Infof("Creating ressource with name: %s", ressource.Name) 

	createdRessource, err := repository.InsertRessource(ressource)
	if err != nil {
		logrus.Errorf("Error creating ressource: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Error while creating ressource",
			Code:    http.StatusInternalServerError,
		}
	}

	logrus.Infof("Ressource created with ID: %s", createdRessource.Id) 

	return createdRessource, nil
}


func UpdateRessource(id uuid.UUID, resource models.Ressources) (*models.Ressources, error) {
	updatedResource, err := repository.UpdateRessource(id, resource)
	if err != nil {
		logrus.Errorf("Error updating resource: %s", err.Error())
		return nil, &models.CustomError{
			Message: "Error while updating resource",
			Code:    http.StatusInternalServerError,
		}
	}
	return updatedResource, nil
}


func DeleteRessource(id uuid.UUID) error {
	err := repository.DeleteRessource(id)
	if err != nil {
		logrus.Errorf("Error deleting resource: %s", err.Error())
		return &models.CustomError{
			Message: "Error while deleting resource",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}
