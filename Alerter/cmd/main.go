package main

import (
	"middleware/example/internal/db"
	"middleware/example/internal/consumer"
)


func main() {

	db.InitDB()
	
	consumer.InitNats()
	
	consumer.RunMyConsumer()
	
	 
}


