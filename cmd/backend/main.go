package main

import (
	"github.com/mnichols/go-temporal-sample/pkg/auth"
	"github.com/mnichols/go-temporal-sample/pkg/cicd"
	"github.com/mnichols/go-temporal-sample/pkg/clients/temporal"
	"github.com/mnichols/go-temporal-sample/pkg/notifications"
	"github.com/mnichols/go-temporal-sample/pkg/orchestrations"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	temporalClient := temporal.MustNewClient()

	defer temporalClient.Close()

	const taskQueue = "apps"
	w := worker.New(temporalClient, taskQueue, worker.Options{})

	authHandlers := &auth.Handlers{}
	notificationHandlers := &notifications.Handlers{}
	cicdHandlers := &cicd.Handlers{}
	w.RegisterWorkflow(orchestrations.TypeOrchestrations.OnboardApplication)
	w.RegisterActivity(authHandlers)
	w.RegisterActivity(cicdHandlers)
	w.RegisterActivity(notificationHandlers)

	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
