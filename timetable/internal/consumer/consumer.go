package consumer

import (
    "context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"github.com/google/uuid"
	"log"
	"time"
	"middleware/example/internal/models"
	"middleware/example/internal/db"
    "fmt"
    "bytes"
	"embed"
	"html/template"
	"net/http"
	"strings"
)

var NatsConn *nats.Conn

func InitNats() {
    var err error
    NatsConn, err = nats.Connect(nats.DefaultURL)  // Remplace par l'URL de ton serveur NATS
    if err != nil {
        log.Fatalf("Erreur de connexion à NATS: %v", err)
    }
    log.Println("Connecté à NATS")
}

// Dans consumer/consumer.go
func RunMyConsumer() {
	consumer, err := EventConsumer()
	if err != nil {
		log.Printf("Erreur lors de la création du consumer NATS: %v", err)
		return
	}

	err = Consume(*consumer)
	if err != nil {
		log.Printf("Erreur lors de la consommation des messages NATS: %v", err)
	}
}


// Création d'un consommateur Jetstream
func EventConsumer() (*jetstream.Consumer, error) {
    
    js, _ := jetstream.New(NatsConn)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // getting Jetstream context
   jsc, err := NatsConn.JetStream()
   if err != nil {
    log.Fatal(err)
   }

    // Vérifie si le flux "USERS" existe, sinon le crée
    _, err = jsc.StreamInfo("USERS")
    if err != nil {
        // Si le flux n'existe pas, on le crée
        logrus.Infof("Flux 'USERS' introuvable. Création du flux...")
        _, err = jsc.AddStream(&nats.StreamConfig{
            Name:     "USERS",
            Subjects: []string{"event_subject"}, // Associer le sujet au flux
        })
        if err != nil {
            return nil, err
        }
        logrus.Infof("Flux 'USERS' créé avec succès.")
    }

    // get existing stream handle
    stream, err := js.Stream(ctx, "USERS")
    if err != nil {
        return nil, err
    }

    // Créer un consommateur durable
	consumerName := "consumer-" + uuid.New().String()
    
    // getting durable consumer
    consumer, err := stream.Consumer(ctx, consumerName)
    if err != nil {
        // if doesn't exist, create durable consumer
        consumer, err = stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
            Durable:     consumerName,
            Name:        consumerName,
            Description: "consumer_name_here consumer",
        })
        if err != nil {
            return nil, err
        }
        logrus.Infof("Created consumer")
    } else {
        logrus.Infof("Got existing consumer")
    }

    logrus.Infof("Created or got existing consumer.")

    return &consumer, nil
}


// Consommer les messages
func Consume(consumer jetstream.Consumer) (err error) {
    logrus.Info("Subscribing")

    // consume messages from the consumer in callback
    cc, err := consumer.Consume(func(msg jetstream.Msg) {
        // Traiter le message reçu dans le callback
        logrus.Info(string(msg.Data())) // Optionnel : affiche les données du message
        err := processEvent(msg.Data()) // Traiter l'événement
        if err != nil {
            logrus.Errorf("Erreur lors du traitement de l'événement: %v", err)
        }
        // Accuser la réception du message
        _ = msg.Ack()
    })

    // Important pour que le programme reste actif tant que la consommation n'est pas terminée
    if err != nil {
        return err
    }

    <-cc.Closed() // Bloque jusqu'à ce que la connexion du consommateur soit fermée
    cc.Stop()     // Arrête la consommation

    return nil
}

// Modèle de l'événement
/*type Event struct {
    Uid             string    `json:"uid"`
    Description     string    `json:"description"`
    Localisation    string    `json:"localisation"`
    Start           time.Time `json:"start"`
    End             time.Time `json:"end"`
    LastModificated time.Time `json:"lastmodificated"`
    Type            string    `json:"type"`
}*/

// Processer les événements reçus et les enregistrer
func processEvent(data []byte) error {
    // Parser le message JSON reçu en un événement structuré
    var event models.Event
    err := json.Unmarshal(data, &event)
    if err != nil {
        return err
    }

    // Vérifier si l'événement existe déjà dans la base de données
    existingEvent, err := db.GetEventByID(event.Uid)
    if err != nil {
        logrus.Errorf("Erreur lors de la récupération de l'événement: %v", err)
        return err
    }

    // Si l'événement existe déjà, vérifier s'il y a eu une modification
    if existingEvent != nil {
        isModified := false
        modificationDetails := ""

        // Vérifier les modifications
        if !existingEvent.Start.Equal(event.Start) {
            modificationDetails += fmt.Sprintf("Horaire modifié: %s -> %s. ", existingEvent.Start, event.Start)
            isModified = true
        }
        if !existingEvent.End.Equal(event.End) {
            modificationDetails += fmt.Sprintf("Horaire de fin modifié: %s -> %s. ", existingEvent.End, event.End)
            isModified = true
        }
        if existingEvent.Localisation != event.Localisation {
            modificationDetails += fmt.Sprintf("Salle modifiée: %s -> %s. ", existingEvent.Localisation, event.Localisation)
            isModified = true
        }
        if existingEvent.Description != event.Description {
            modificationDetails += fmt.Sprintf("Description modifiée: %s -> %s. ", existingEvent.Description, event.Description)
            isModified = true
        }

        // Si une modification a été détectée, créer une alerte
        if isModified {
            // Créer une alerte pour cette modification
            var err error

            alert := models.Alert{
                EventID:   event.Uid,
                Message:   fmt.Sprintf("L'événement %s a été modifié. Détails des modifications: %s", event.Description, modificationDetails),
                Email:     "Abdou_Latif.KANE@etu.uca.fr", // Adapte cette adresse email en fonction des destinataires
                EventType: event.Type,
            }

            // Enregistrer l'alerte dans la base de données
            err = db.CreateAlert(alert)
            if err != nil {
                logrus.Errorf("Erreur lors de la création de l'alerte: %v", err)
                return err
            }

            // Envoyer la notification via NATS ou un autre service
            err = sendAlertNotification(alert)
            if err != nil {
                logrus.Errorf("Erreur lors de l'envoi de la notification: %v", err)
                return err
            }

            logrus.Infof("Alerte envoyée pour l'événement modifié: %v", event.Uid)
        }
        return nil
    }

    // Si l'événement n'existe pas, l'ajouter
    err = db.AddEvent(event)
    if err != nil {
        logrus.Errorf("Erreur lors de l'insertion de l'événement: %v", err)
        return err
    }
    logrus.Infof("Nouvel événement inséré: %v", event.Uid)

    return nil
}


var embeddedTemplates embed.FS

// Envoie la notification d'alerte par email
func sendAlertNotification(alert models.Alert) error {
	// Charger et analyser le template HTML
	templatePath := "templates/alert_template.html" // Assurez-vous que ce chemin est correct
	var tpl *template.Template
	var err error
	tpl, err = template.ParseFS(embeddedTemplates, templatePath)
	if err != nil {
		return fmt.Errorf("Erreur lors du parsing du template: %v", err)
	}

	// Préparer le contenu du message
	var tplBuffer bytes.Buffer
	err = tpl.Execute(&tplBuffer, alert)
	if err != nil {
		return fmt.Errorf("Erreur lors de l'exécution du template: %v", err)
	}

	// Créer le corps de l'email (HTML)
	htmlContent := tplBuffer.String()

	// Ici, nous utilisons l'API mail.edu.forestier.re pour l'envoi d'emails
	// Assurez-vous d'avoir un token d'authentification et un endpoint correct

	// Exemple d'URL de l'API pour l'envoi d'un email
	apiURL := "https://mail.edu.forestier.re/api/v1/sendmail"
	reqBody := fmt.Sprintf(`{
		"to": "%s",
		"subject": "Alerte: %s",
		"html": "%s"
	}`, alert.Email, alert.EventType, htmlContent)

	// Effectuer la requête HTTP POST vers l'API de l'envoi de mail
	resp, err := http.Post(apiURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		log.Printf("Erreur lors de l'envoi de l'alerte: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Erreur lors de l'envoi de l'alerte, code HTTP: %d", resp.StatusCode)
		return fmt.Errorf("erreur HTTP %d lors de l'envoi de l'alerte", resp.StatusCode)
	}

	log.Printf("Alerte envoyée avec succès à %s", alert.Email)
	return nil
}