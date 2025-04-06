package db

import (
	"database/sql"
	_ "github.com/lib/pq" 
	"log"
	"fmt"
	
	"middleware/example/internal/models"

)

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")
	if err != nil {
		log.Fatalf("❌ Erreur de connexion à la base de données: %v", err)
	}
	log.Println("✅ Connecté à la base de données")
	return nil
}

func InsertAlert(alert models.Alert) error {
    if DB == nil {
        return fmt.Errorf("la connexion à la base de données est nulle")
    }

    query := "INSERT INTO alerts (id, message, created_at) VALUES ($1, $2, $3, $4)"
    _, err := DB.Exec(query, alert.EventID, alert.Message, alert.CreatedAt)
    if err != nil {
        return fmt.Errorf("erreur lors de l'insertion de l'alerte : %v", err)
    }
    return nil
}

