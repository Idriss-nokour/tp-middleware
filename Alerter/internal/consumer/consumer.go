
package consumer

import (
    "context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)


var NatsConn *nats.Conn

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
	consumer, err := AlertConsumer()
	if err != nil {
		log.Printf("Erreur lors de la création du consumer NATS: %v", err)
		return
	}

	err = Consume(consumer)
	if err != nil {
		log.Printf("Erreur lors de la consommation des messages NATS: %v", err)
	}
}



// Création d'un consommateur JetStream
func AlertConsumer() (jetstream.Consumer, error) {
	js, _ := jetstream.New(NatsConn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jsc, err := NatsConn.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	// Vérifie si le flux "ALERT" existe, sinon le crée
	_, err = jsc.StreamInfo("ALERT")
	if err != nil {
		logrus.Infof("Flux 'ALERT' introuvable. Création du flux...")
		_, err = jsc.AddStream(&nats.StreamConfig{
			Name:     "ALERT",
			Subjects: []string{"alert_subject"},
		})
		if err != nil {
			return nil, err
		}
		logrus.Infof("✅ Flux 'ALERT' créé avec succès.")
	}

	// Récupération du stream existant
	stream, err := js.Stream(ctx, "ALERT")
	if err != nil {
		return nil, err
	}

	// Création du consommateur durable
	consumerName := "consumer-alerts"
	consumer, err := stream.Consumer(ctx, consumerName)
	if err != nil {
		consumer, err = stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
			Durable:     consumerName,
			Name:        consumerName,
			Description: "Consommateur des alertes Timetable",
		})
		if err != nil {
			return nil, err
		}
		logrus.Infof("✅ Consumer '%s' créé.", consumerName)
	} else {
		logrus.Infof("✅ Consumer '%s' existant trouvé.", consumerName)
	}

	return consumer, nil
}

// Consommer les messages
func Consume(consumer jetstream.Consumer) (err error) {
    logrus.Info("Subscribing")

    // consume messages from the consumer in callback
    cc, err := consumer.Consume(func(msg jetstream.Msg) {
        // Traiter le message reçu dans le callback
        logrus.Info(string(msg.Data())) // Optionnel : affiche les données du message
      
        if err := msg.Ack(); err != nil {
			logrus.Errorf("Erreur lors de l'Ack du message: %v", err)
		}
		
    })

    if err != nil {
        return err
    }

    <-cc.Closed() 
    cc.Stop()     

    return nil
}