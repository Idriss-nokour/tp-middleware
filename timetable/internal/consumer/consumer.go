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
    /*"bytes"
	"embed"
	"html/template"
	"net/http"
	"strings"*/
)

var NatsConn *nats.Conn
var jsc nats.JetStreamContext
var nc   *nats.Conn

// Fonction d'initialisation du stream NATS
func InitStream() {
	var err error

	// Connect to a server
	// create a nats connection
	nc, err = nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    } 
	// getting Jetstream context
	jsc, err = nc.JetStream()
	if err != nil {
	 log.Fatal(err)
	}
 
	//Init stream
	_, err = jsc.AddStream(&nats.StreamConfig{
		 Name:     "ALERT",             // nom du stream
		 Subjects: []string{"ALERT.>"}, // tous les sujets sont sous le format "USERS.*"
	})
	if err != nil {
	 log.Fatal(err)
	}
}

// Fonction pour souscrire à un sujet
func SubscribeToTopic() {
    // Simple Async Subscriber
  nc.Subscribe("ALERT.>", func(m *nats.Msg) {
   fmt.Printf("Received a message: %s\n", m.Subject)
   if m.Subject == "ALERT.create" {
      fmt.Println("received on USERS.create")
      fmt.Println(string(m.Data))
   }
})
}

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
    
    js, _ := jetstream.New(nc)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // getting Jetstream context
   /*jsc, err := nc.JetStream()
   if err != nil {
    log.Fatal(err)
   }*/

    // Vérifie si le flux "USERS" existe, sinon le crée
    _, err := jsc.StreamInfo("USERS")
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
            
             // Afficher le contenu de l'alerte avant de l'enregistrer
            logrus.Infof("Alerte à enregistrer: %+v", alert)
            // Enregistrer l'alerte dans la base de données
            err = db.CreateAlert(alert)
            if err != nil {
                logrus.Errorf("Erreur lors de la création de l'alerte: %v", err)
                return err
            }

            // Publier l'alerte dans NATS
            err = publishAlert("ALERT", alert)
            if err != nil {
                logrus.Errorf("Erreur lors de la publication de l'alerte: %v", err)
                return err
            }

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


// Scheduler pour publier les alertes non envoyées
func AlertScheduler() {
    ticker := time.NewTicker(10 * time.Second) // Exécute toutes les 10 secondes
    defer ticker.Stop()

    for {
        <-ticker.C
        fmt.Println("Exécution du scheduler pour envoyer les alertes...")

        // Récupérer les alertes non envoyées
        alerts, err := db.GetUnsentAlerts()
        if err != nil {
            log.Printf("Erreur lors de la récupération des alertes: %v", err)
            continue
        }

        if len(alerts) == 0 {
            fmt.Println("Aucune alerte à envoyer.")
            continue
        }

        // Connexion à NATS
        nc, err := nats.Connect(nats.DefaultURL)
        if err != nil {
            log.Printf("Erreur de connexion à NATS: %v", err)
            continue
        }
        defer nc.Close()

        for _, alert := range alerts {
            // Convertir l'alerte en JSON
            alertJSON, err := json.Marshal(alert)
            if err != nil {
                log.Printf("Erreur lors de la conversion de l'alerte en JSON: %v", err)
                continue
            }

            // Publier sur NATS
            err = nc.Publish("alerts", alertJSON)
            if err != nil {
                log.Printf("Erreur lors de la publication de l'alerte: %v", err)
                continue
            }

            fmt.Printf("Alerte envoyée: %+v\n", alert)

            // Marquer l'alerte comme envoyée
            err = db.MarkAlertAsSent(alert.EventID)
            if err != nil {
                log.Printf("Erreur lors de la mise à jour de l'alerte: %v", err)
            }
        }
    }
}



func publishAlert(subject string, alert models.Alert) error {
    // Sérialiser l'événement en JSON
    messageBytes, err := json.Marshal(alert)
    if err != nil {
        return fmt.Errorf("Erreur de sérialisation de l'alert: %v", err)
    }

    // Publier l'événement de manière asynchrone
    pubAckFuture, err := jsc.PublishAsync(subject, messageBytes)
    if err != nil {
        return fmt.Errorf("Erreur lors de la publication de l'alert: %v", err)
    }

    // Attendre l'accusé de réception de la publication
    select {
    case <-pubAckFuture.Ok():
        log.Println("Alert publié avec succès.")
        return nil
    case <-pubAckFuture.Err():
        return fmt.Errorf("Erreur de publication: %v", pubAckFuture.Err())
    }
}


