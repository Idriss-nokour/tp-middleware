package main

import (
	"context"
	"os"
	"os/signal"
	"time"
	"github.com/zhashkevych/scheduler"
	"middleware/example/internal/schedule"


)


func main() {
	schedule.InitStream()

	schedule.SubscribeToTopic()
	
	ctx := context.Background()
	sc := scheduler.NewScheduler()
	sc.Add(ctx, schedule.Scheduled, time.Second*5)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	sc.Stop()
}

