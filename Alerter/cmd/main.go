package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"middleware/example/internal/controllers/collections"
	"middleware/example/internal/helpers"
	_ "middleware/example/internal/models"
    "github.com/nats-io/nats.go"

	"net/http"
)

func main() {
	r := chi.NewRouter()

}

func init() {
	
}

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
	consumer, err := AlertConsumer()
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
func AlertConsumer() (*jetstream.Consumer, error) {
    
    js, _ := jetstream.New(NatsConn)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // getting Jetstream context
   jsc, err := NatsConn.JetStream()
   if err != nil {
    log.Fatal(err)
   }

    // Vérifie si le flux "USERS" existe, sinon le crée
    _, err = jsc.StreamInfo("ALERT")
    if err != nil {
        // Si le flux n'existe pas, on le crée
        logrus.Infof("Flux 'ALERT' introuvable. Création du flux...")
        _, err = jsc.AddStream(&nats.StreamConfig{
            Name:     "ALERT",
            Subjects: []string{"alert_subject"}, // Associer le sujet au flux
        })
        if err != nil {
            return nil, err
        }
        logrus.Infof("Flux 'ALERT' créé avec succès.")
    }

    // get existing stream handle
    stream, err := js.Stream(ctx, "ALERT")
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
