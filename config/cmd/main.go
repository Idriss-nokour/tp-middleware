package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"middleware/example/internal/controllers/collections"
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
		r.Get("/", alerts.GetAlerts)  // GET /alerts pour obtenir toutes les alertes
		//r.Post("/", alerts.CreateAlert) // POST /alerts pour créer une nouvelle alerte

		r.Route("/{id}", func(r chi.Router) {
			r.Use(alerts.Ctx) // Middleware pour les alertes
			r.Get("/", alerts.GetAlert)   // GET /alerts/{id}
			//r.Put("/", alerts.UpdateAlert) // PUT /alerts/{id}
			//r.Delete("/", alerts.DeleteAlert) // DELETE /alerts/{id}
		})
	})

	logrus.Info("[INFO] Web server started. Now listening on *:8080")
	logrus.Fatalln(http.ListenAndServe(":9000", r))
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
			id TEXT PRIMARY KEY NOT NULL UNIQUE,   
			email TEXT NOT NULL,                   
			is_all BOOLEAN NOT NULL                  
		);`,

		`CREATE TABLE IF NOT EXISTS ressources (
			id TEXT PRIMARY KEY NOT NULL UNIQUE,  
			uca_id INTEGER NOT NULL,             
			name TEXT NOT NULL                   
		);`,
		
		


	}
	for _, scheme := range schemes {
		if _, err := db.Exec(scheme); err != nil {
			logrus.Fatalln("Could not generate table ! Error was : " + err.Error())
		}
	}
	helpers.CloseDB(db)
}
