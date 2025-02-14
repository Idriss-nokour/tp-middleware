package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"middleware/example/internal/controllers/collections"
	"middleware/example/internal/controllers/ressources"
	"middleware/example/internal/controllers/alerts"
	"middleware/example/internal/helpers"
	_ "middleware/example/internal/models"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Route("/collections", func(r chi.Router) {
		r.Get("/", collections.GetCollections)
		r.Route("/{id}", func(r chi.Router) {
			r.Use(collections.Ctx)
			r.Get("/", collections.GetCollection)
		})
	})

	r.Route("/alerts", func(r chi.Router) {
		r.Get("/", alerts.GetAlerts)  
		r.Post("/", alerts.CreateAlert) 

		r.Route("/{id}", func(r chi.Router) {
			r.Use(alerts.Ctx) 
			r.Get("/", alerts.GetAlert)   
			r.Put("/", alerts.UpdateAlert) 
			r.Delete("/", alerts.DeleteAlert) 
		})
	})

	r.Route("/ressources", func(r chi.Router) {
		r.Get("/", ressources.GetRessources)  
		r.Post("/", ressources.CreateRessource) 
		r.Route("/{id}", func(r chi.Router) {
			r.Use(ressources.Ctx) 
			r.Get("/", ressources.GetRessource)   
			r.Put("/", ressources.UpdateRessource) 
			r.Delete("/", ressources.DeleteRessource) 
		})
	})

	logrus.Info("[INFO] Web server started. Now listening on *:8080")
	logrus.Fatalln(http.ListenAndServe(":9090", r))
}

func init() {
	db, err := helpers.OpenDB()
	if err != nil {
		logrus.Fatalf("error while opening database : %s", err.Error())
	}
	schemes := []string{
		`CREATE TABLE IF NOT EXISTS collections (
			id VARCHAR(255) PRIMARY KEY NOT NULL UNIQUE,
			content VARCHAR(255) NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS alerts (
			id VARCHAR(255) PRIMARY KEY NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL,
			is_all BOOLEAN NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS ressources (
			id VARCHAR(255) PRIMARY KEY NOT NULL UNIQUE,
			uca_id INTEGER NOT NULL,             
			name VARCHAR(255) NOT NULL
		);`,

		
		

		
		

	}
	for _, scheme := range schemes {
		if _, err := db.Exec(scheme); err != nil {
			logrus.Fatalln("Could not generate table ! Error was : " + err.Error())
		}
	}
	helpers.CloseDB(db)
}
