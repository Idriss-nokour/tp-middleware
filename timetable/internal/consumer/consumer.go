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

// Traiter les événements reçus et les enregistrer
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

    // Si l'événement existe déjà, ne rien faire
    if existingEvent != nil {
        logrus.Infof("L'événement existe déjà, aucun changement: %v", event.Uid)
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
