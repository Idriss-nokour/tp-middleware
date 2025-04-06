package db

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
    "middleware/example/internal/models"
   "time"
   "fmt" 
   "strings"

)

var db *sql.DB

func InitDB() {
    var err error
    db, err = sql.Open("sqlite3", "./events.db")
    if err != nil {
        log.Fatal(err)
    }

    createTableSQL := `DROP TABLE IF EXISTS events;
                       CREATE TABLE IF NOT EXISTS events (
                           uid TEXT PRIMARY KEY,
                           description TEXT,
                           localisation TEXT,
                           start TEXT NOT NULL,
                           end TEXT NOT NULL,
                           lastmodificated TEXT NOT NULL,
                           type TEXT
                       );`

    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Table 'events' recréée avec succès!")

    // Création de la table alerts si elle n'existe pas
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS alerts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_id TEXT NOT NULL,
        message TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`)
    if err != nil {
        log.Fatalf("Erreur lors de la création de la table alerts: %v", err)
    }
}

func GetDB() *sql.DB {
    return db
}

func GetEventByID(uid string) (*models.Event, error) {
    var event models.Event
    var startStr, endStr, lastModifiedStr string

    query := "SELECT uid, description, localisation, start, end, lastmodificated, type FROM events WHERE uid = ?"
    err := db.QueryRow(query, uid).Scan(&event.Uid, &event.Description, &event.Localisation, &startStr, &endStr, &lastModifiedStr, &event.Type)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil 
        }
        return nil, err
    }

    // Remplacer l'espace entre la date et l'heure par un "T"
    startStr = strings.Replace(startStr, " ", "T", 1)
    endStr = strings.Replace(endStr, " ", "T", 1)
    lastModifiedStr = strings.Replace(lastModifiedStr, " ", "T", 1)

    // Convertir les chaînes en time.Time (en supposant que les dates sont maintenant au format RFC3339)
    var errParse error
    event.Start, errParse = time.Parse(time.RFC3339, startStr)
    if errParse != nil {
        return nil, fmt.Errorf("erreur lors de la conversion du start: %v", errParse)
    }

    event.End, errParse = time.Parse(time.RFC3339, endStr)
    if errParse != nil {
        return nil, fmt.Errorf("erreur lors de la conversion du end: %v", errParse)
    }

    event.LastModificated, errParse = time.Parse(time.RFC3339, lastModifiedStr)
    if errParse != nil {
        return nil, fmt.Errorf("erreur lors de la conversion du lastmodificated: %v", errParse)
    }

    return &event, nil
}


func GetAllEvents() ([]models.Event, error) {
    var events []models.Event
    rows, err := db.Query("SELECT uid, description, localisation, start, end, lastmodificated, type FROM events")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var event models.Event
        var startStr, endStr, lastModifiedStr string

        if err := rows.Scan(&event.Uid, &event.Description, &event.Localisation, &startStr, &endStr, &lastModifiedStr, &event.Type); err != nil {
            return nil, err
        }

        // Parse the date strings into time.Time
        event.Start, err = time.Parse(time.RFC3339, startStr)
        if err != nil {
            log.Printf("Error parsing start time: %v", err)
            return nil, err
        }

        event.End, err = time.Parse(time.RFC3339, endStr)
        if err != nil {
            log.Printf("Error parsing end time: %v", err)
            return nil, err
        }

        event.LastModificated, err = time.Parse(time.RFC3339, lastModifiedStr)
        if err != nil {
            log.Printf("Error parsing last modified time: %v", err)
            return nil, err
        }

        events = append(events, event)
    }

    return events, nil
}


func AddEvent(event models.Event) error {
    _, err := db.Exec("INSERT INTO events (uid, description, localisation, start, end, lastmodificated, type) VALUES (?, ?, ?, ?, ?, ?, ?)",
        event.Uid, event.Description, event.Localisation, event.Start, event.End, event.LastModificated, event.Type)
    return err
}

func UpdateEvent(uid string, event models.Event) error {
    _, err := db.Exec("UPDATE events SET description = ?, localisation = ?, start = ?, end = ?, lastmodificated = ?, type = ? WHERE uid = ?",
        event.Description, event.Localisation, event.Start, event.End, event.LastModificated, event.Type, uid)
    return err
}



func DeleteEvent(uid string) error {
    db := GetDB()
    _, err := db.Exec("DELETE FROM events WHERE uid = ?", uid)
    return err
}

// Fonction pour créer une alerte dans la base de données avec sql.DB
func CreateAlert(alert models.Alert) error {
    now := time.Now()
    _, err := db.Exec("INSERT INTO alerts (event_id, message, created_at, updated_at) VALUES (?, ?, ?, ?)", 
        alert.EventID, alert.Message, now, now)
    if err != nil {
        log.Printf("Erreur lors de la création de l'alerte: %v", err)
        return err
    }

    log.Printf("Alerte créée avec succès pour l'événement: %v", alert.EventID)
    return nil
}

func DeleteAllAlerts(db *sql.DB) error {
	query := "DELETE FROM Alert"

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des alertes : %v", err)
	}

	log.Println("Toutes les alertes ont été supprimées avec succès.")
	return nil
}

// Récupérer les alertes non envoyées
func GetUnsentAlerts() ([]models.Alert, error) {
    rows, err := db.Query("SELECT event_id, message, email, event_type FROM alerts WHERE sent = 0")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var alerts []models.Alert
    for rows.Next() {
        var alert models.Alert
        if err := rows.Scan(&alert.EventID, &alert.Message); err != nil {
            return nil, err
        }
        alerts = append(alerts, alert)
    }

    return alerts, nil
}

// Mettre à jour une alerte après l'envoi
func MarkAlertAsSent(eventID string) error {
    _, err := db.Exec("UPDATE alerts SET sent = 1 WHERE event_id = ?", eventID)
    if err != nil {
        log.Printf("Erreur lors de la mise à jour de l'alerte: %v", err)
    }
    return err
}

