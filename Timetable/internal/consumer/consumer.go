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
		 Name:     "ALERT",            
		 Subjects: []string{"ALERT.>"}, 
	})
	if err != nil {
	 log.Fatal(err)
	}
}

// Fonction pour souscrire à un sujet
func SubscribeToTopic() {
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

    NatsConn, err = nats.Connect(nats.DefaultURL)  
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

    _, err := jsc.StreamInfo("USERS")
    if err != nil {

        logrus.Infof("Flux 'USERS' introuvable. Création du flux...")
        _, err = jsc.AddStream(&nats.StreamConfig{
            Name:     "USERS",
            Subjects: []string{"event_subject"}, 
        })
        if err != nil {
            return nil, err
        }
        logrus.Infof("Flux 'USERS' créé avec succès.")
    }

    stream, err := js.Stream(ctx, "USERS")
    if err != nil {
        return nil, err
    }

	consumerName := "consumer-" + uuid.New().String()
    
    consumer, err := stream.Consumer(ctx, consumerName)
    if err != nil {
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

    cc, err := consumer.Consume(func(msg jetstream.Msg) {
        logrus.Info(string(msg.Data())) 
        err := processEvent(msg.Data()) 
        if err != nil {
            logrus.Errorf("Erreur lors du traitement de l'événement: %v", err)
        }
        _ = msg.Ack()
    })

    if err != nil {
        return err
    }

    <-cc.Closed() 
    cc.Stop()     

    return nil
}

// Processer les événements reçus et les enregistrer
func processEvent(data []byte) error {
    var event models.Event
    err := json.Unmarshal(data, &event)
    if err != nil {
        return err
    }

    existingEvent, err := db.GetEventByID(event.Uid)
    if err != nil {
        logrus.Errorf("Erreur lors de la récupération de l'événement: %v", err)
        return err
    }

    if existingEvent != nil {
        isModified := false
        modificationDetails := ""

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

        if isModified {
            var err error

            alert := models.Alert{
                EventID:   event.Uid,
                Message:   fmt.Sprintf("L'événement %s a été modifié. Détails des modifications: %s", event.Description, modificationDetails),
            }
            
            logrus.Infof("Alerte à enregistrer: %+v", alert)
            err = db.CreateAlert(alert)
            if err != nil {
                logrus.Errorf("Erreur lors de la création de l'alerte: %v", err)
                return err
            }

            err = publishAlert("alert_subject", alert)
            if err != nil {
                logrus.Errorf("Erreur lors de la publication de l'alerte: %v", err)
                return err
            }

        }
        return nil
    }

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
    ticker := time.NewTicker(10 * time.Second) 
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


