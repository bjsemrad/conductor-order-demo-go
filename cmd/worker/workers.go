package main

import (
	ordertasks "conductor-demo/pkg/tasks"
	"os"
	"time"

	"github.com/conductor-sdk/conductor-go/sdk/client"
	"github.com/conductor-sdk/conductor-go/sdk/settings"
	"github.com/conductor-sdk/conductor-go/sdk/worker"
)

func main() {
	apiClient := client.NewAPIClient(
		settings.NewAuthenticationSettings(
			os.Getenv("ACCESS_KEY"),
			os.Getenv("SECRET_KEY"),
		),
		settings.NewHttpSettings(
			"https://developer.orkescloud.com/api",
		))

	taskRunner := worker.NewTaskRunnerWithApiClient(apiClient)
	taskRunner.StartWorker("IntakeOrder", ordertasks.AcceptOrderFrom, 1, time.Second*5)
	taskRunner.StartWorker("FraudCheck", ordertasks.FraudCheck, 1, time.Second*5)
	taskRunner.StartWorker("PriceOrder", ordertasks.PriceOrder, 1, time.Second*5)
	taskRunner.StartWorker("CreditReview", ordertasks.CreditReview, 1, time.Second*5)

	//Add more StartWorker calls as needed

	// Block
	taskRunner.WaitWorkers()
}
