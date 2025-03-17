package main

import (
   // "log"
    //"net/http"
	"middleware/example/internal/db"
	//"middleware/example/internal/api"
	"middleware/example/internal/consumer"
    //"middleware/example/internal/models"
    //"github.com/gorilla/mux"
    //"github.com/sirupsen/logrus"
	//"encoding/json"
	//"fmt"
	//"github.com/nats-io/nats.go"
    //"context"
    //"github.com/zhashkevych/scheduler"
    //"time"
    /*"os"
    "io"
    "strconv"
    "bufio"
	"bytes"
	"strings"
    "os/signal"*/

)




func main() {

	db.InitDB()

	//consumer.InitNats()
       // Initialisation de la connexion à NATS et du stream
	consumer.InitStream()
	// Dans cmd/main.go
	consumer.RunMyConsumer()
	
	// Lancer le serveur HTTP principal
	//runMyServer()  

 
    consumer.AlertScheduler()
	// Appel de la fonction pour souscrire à un sujet (en écoute)
	consumer.SubscribeToTopic()
	
}



