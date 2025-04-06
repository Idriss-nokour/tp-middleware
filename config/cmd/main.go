package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"middleware/example/internal/controllers/ressources"
	"middleware/example/internal/controllers/alerts"
	_ "middleware/example/internal/models"
	"net/http"
)

func main() {
	r := chi.NewRouter()

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

	logrus.Info("[INFO] Web server started. Now listening on *:9090")
	logrus.Fatalln(http.ListenAndServe(":9090", r))
}

