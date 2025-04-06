package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"bufio"
	"bytes"
	"strings"
	"github.com/nats-io/nats.go"
	"middleware/example/internal/models"

)

var jsc nats.JetStreamContext
var nc   *nats.Conn


// Fonction de récupération et de publication des événements
func Scheduled(_ context.Context) {
	// Récupérer les ressources
	var err error
	resp, err := http.Get("http://localhost:9090/ressources")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var tables []models.Ressources
	if err := json.Unmarshal(body, &tables); err != nil {
		log.Fatal(err)
	}

	tablesToFetch := ""
	for _, table := range tables {
		if tablesToFetch != "" {
			tablesToFetch += ","
		}
		tablesToFetch += strconv.Itoa(table.UcaId)
	}

	// Récupérer les événements depuis l'EDT
	resp, err = http.Get(fmt.Sprintf("https://edt.uca.fr/jsp/custom/modules/plannings/anonymous_cal.jsp?resources=%s&projectId=2&calType=ical&nbWeeks=8&displayConfigId=128", tablesToFetch))
	if err != nil {
		log.Fatal(err)
	}

	// Lire les données
	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Créer un scanner pour lire les lignes
	scanner := bufio.NewScanner(bytes.NewReader(rawData))

	var eventArray []map[string]string
	currentEvent := map[string]string{}
	currentKey := ""
	currentValue := ""
	inEvent := false

	// Analyser chaque ligne de l'ICalendar
	for scanner.Scan() {
		line := scanner.Text()

		// Ignorer les lignes de calendrier
		if !inEvent && line != "BEGIN:VEVENT" {
			continue
		}

		// Détecter le début d'un événement
		if line == "BEGIN:VEVENT" {
			inEvent = true
			currentEvent = map[string]string{}
			continue
		}

		// Fin d'un événement
		if line == "END:VEVENT" {
			eventArray = append(eventArray, currentEvent)
			inEvent = false
			continue
		}

		// Gérer les données multi-lignes
		if strings.HasPrefix(line, " ") {
			currentEvent[currentKey] += line
		} else {
			// Diviser la ligne en clé et valeur
			splitted := strings.SplitN(line, ":", 2000)
			if len(splitted) < 2 {
				continue
			}
			currentKey = splitted[0]
			currentValue = splitted[1]
			// Stocker l'attribut de l'événement
			currentEvent[currentKey] = currentValue
		}
	}

	// Convertir les événements en objets personnalisés
	var structureEvents []models.Event
	for _, event := range eventArray {
		startTime, _ := time.Parse("20060102T150405Z", event["DTSTART"])
		endTime, _ := time.Parse("20060102T150405Z", event["DTEND"])
		lastModificated, _ := time.Parse("20060102T150405Z", event["LAST-MODIFIED"])

		structureEvents = append(structureEvents, models.Event{
			Uid : event["UID"],
			Description:     event["DESCRIPTION"],
			Localisation:    event["LOCATION"],
			Start:           startTime,
			End:             endTime,
			LastModificated: lastModificated,
		})
	}

	// Convertir les événements en JSON
	eventJSON, err := json.MarshalIndent(structureEvents, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Publier les événements dans NATS
	for _, event := range structureEvents {
		if err := publishEvent("USERS.create", event); err != nil {
			log.Fatal("Erreur lors de la publication de l'événement:", err)
		}
	}

	// Afficher les événements en JSON
	fmt.Println(string(eventJSON))
}

// Fonction d'initialisation du stream NATS
func InitStream() {
	var err error

	// Connect to a server
	// create a nats connection
	nc, _ := nats.Connect(nats.DefaultURL)
	// getting Jetstream context
	jsc, err = nc.JetStream()
	if err != nil {
	 log.Fatal(err)
	}
 
	//Init stream
	_, err = jsc.AddStream(&nats.StreamConfig{
		 Name:     "USERS",             // nom du stream
		 Subjects: []string{"USERS.>"}, // tous les sujets sont sous le format "USERS.*"
	})
	if err != nil {
	 log.Fatal(err)
	}
}


// Fonction pour publier un événement dans NATS
func publishEvent(subject string, event models.Event) error {
	// Sérialiser l'événement en JSON
	messageBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("Erreur de sérialisation de l'événement: %v", err)
	}

	// Publier l'événement de manière asynchrone
	pubAckFuture, err := jsc.PublishAsync(subject, messageBytes)
	if err != nil {
		return fmt.Errorf("Erreur lors de la publication de l'événement: %v", err)
	}

	// Attendre l'accusé de réception de la publication
	select {
	case <-pubAckFuture.Ok():
		log.Println("Événement publié avec succès.")
		return nil
	case <-pubAckFuture.Err():
		return fmt.Errorf("Erreur de publication: %v", pubAckFuture.Err())
	}
}


// Fonction pour souscrire à un sujet
func SubscribeToTopic() {
     // Simple Async Subscriber
   nc.Subscribe("USERS.>", func(m *nats.Msg) {
	fmt.Printf("Received a message: %s\n", m.Subject)
	if m.Subject == "USERS.create" {
	   fmt.Println("received on USERS.create")
	   fmt.Println(string(m.Data))
	}
 })
}