package main

import (
    "log"
    "net/http"
	"middleware/example/internal/db"
	"middleware/example/internal/api"
	"middleware/example/internal/consumer"
    "github.com/gorilla/mux"
    "github.com/sirupsen/logrus"

)

func main() {

	db.InitDB()

	consumer.InitNats()

	// Dans cmd/main.go
	consumer.RunMyConsumer()
	
	// Lancer le serveur HTTP principal
	//runMyServer()  
    // Initialisation du routeur
    router := mux.NewRouter()

    // Définition des routes
    router.HandleFunc("/events", api.GetEvents).Methods("GET")
    router.HandleFunc("/events", api.CreateEvent).Methods("POST")
    router.HandleFunc("/events/{id}", api.UpdateEvent).Methods("PUT")
    router.HandleFunc("/events/{id}", api.DeleteEvent).Methods("DELETE")

    // Lancer le serveur HTTP
    logrus.Info("[INFO] Web server started. Now listening on *:8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
