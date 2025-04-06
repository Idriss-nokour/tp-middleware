package api

import (
    "encoding/json"
    "net/http"
    "middleware/example/internal/models"
    "middleware/example/internal/db"
	
    "github.com/gorilla/mux"
)

func GetEvents(w http.ResponseWriter, r *http.Request) {
    events, err := db.GetAllEvents()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(events)
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
    var event models.Event
    err := json.NewDecoder(r.Body).Decode(&event)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = db.AddEvent(event)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var event models.Event
    err := json.NewDecoder(r.Body).Decode(&event)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = db.UpdateEvent(params["id"], event)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}

func DeleteEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    err := db.DeleteEvent(params["id"])
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
