
package consumer

import (
    "context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"time"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"github.com/adrg/frontmatter"
)


var NatsConn *nats.Conn

var embeddedTemplates embed.FS




func InitNats() {
    var err error
    NatsConn, err = nats.Connect(nats.DefaultURL)  
    if err != nil {
        log.Fatalf("Erreur de connexion à NATS: %v", err)
    }
    log.Println("Connecté à NATS")
}

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



func AlertConsumer() (jetstream.Consumer, error) {
	js, _ := jetstream.New(NatsConn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jsc, err := NatsConn.JetStream()
	if err != nil {
		log.Fatal(err)
	}

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

	stream, err := js.Stream(ctx, "ALERT")
	if err != nil {
		return nil, err
	}

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

type MailMatter struct {
    Subject string `yaml:"subject"`
}


func Consume(consumer jetstream.Consumer) error {
	logrus.Info("Subscribing to alerts...")

	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		var data map[string]interface{}
		if err := json.Unmarshal(msg.Data(), &data); err != nil {
			logrus.Errorf("Erreur de parsing du message : %v", err)
			_ = msg.Nak()
			return
		}

		// Générer le contenu HTML depuis le template
		htmlContent, mailMatter, err := GetStringFromEmbeddedTemplate("templates/alert.html", data)
		if err != nil {
			logrus.Errorf("Erreur de génération du template : %v", err)
			_ = msg.Nak()
			return
		}

		// TODO : à adapter selon ton modèle : récupérer l'adresse email à partir des données
		email, ok := data["email"].(string)
		if !ok || email == "" {
			logrus.Error("Aucune adresse email valide trouvée dans les données")
			_ = msg.Nak()
			return
		}

		token := "knrujlZnwerWJXNWYuVTdgfUqomTFDvhMWuvjCsu" 
		err = SendMail(email, mailMatter.Subject, htmlContent, token)
		if err != nil {
			logrus.Errorf("Erreur lors de l'envoi de l'email : %v", err)
			_ = msg.Nak()
			return
		}

		if err := msg.Ack(); err != nil {
			logrus.Errorf("Erreur lors de l'Ack du message: %v", err)
		} else {
			logrus.Infof("✅ Message traité et mail envoyé à %s", email)
		}
	})

	if err != nil {
		return err
	}

	<-cc.Closed()
	cc.Stop()
	return nil
}


func GetStringFromEmbeddedTemplate(templatePath string, body interface{}) (content string, mailMatter MailMatter, err error) {
    temp, err := template.ParseFS(embeddedTemplates, templatePath)
    if err != nil {
        return
    }

    var tpl bytes.Buffer
    if err = temp.Execute(&tpl, body); err != nil {
        return
    }

    var mailContent []byte
    mailContent, err = frontmatter.Parse(strings.NewReader(tpl.String()), &mailMatter)
    if err == nil {
        content = string(mailContent)
    }

    return
}

func SendMail(to string, subject string, content string, token string) error {
    payload := map[string]interface{}{
        "to":      []string{to},
        "subject": subject,
        "content": content,
    }

    jsonBody, _ := json.Marshal(payload)
    req, err := http.NewRequest("POST", "https://mail.edu.forestier.re/api/mail", bytes.NewBuffer(jsonBody))
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("erreur envoi mail : %s", string(body))
    }

    return nil
}
