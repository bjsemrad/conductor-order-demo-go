package main

import (
	ordermodel "conductor-demo/pkg/model"
	"conductor-demo/pkg/utils"
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"

	"github.com/conductor-sdk/conductor-go/sdk/client"
	"github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/conductor-sdk/conductor-go/sdk/settings"
	"github.com/conductor-sdk/conductor-go/sdk/workflow"
	"github.com/conductor-sdk/conductor-go/sdk/workflow/executor"
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
	workflowName := "OrderDemoWrkFlow"
	workflowVersion := 1

	metaApi := client.NewMetadataClient(apiClient)
	err := setupWebhook(apiClient, workflowName, int32(workflowVersion))

	executor := executor.NewWorkflowExecutor(apiClient)

	conductorWorkflow := workflow.NewConductorWorkflow(executor).
		Name(workflowName).
		Version(int32(workflowVersion)).
		OwnerEmail("bjsemrad@gmail.com").
		Description("Test workflow for my own understanding")

	intake := workflow.NewSimpleTask("IntakeOrder", "IntakeOrder").Input("order", "${workflow.input.request}")
	intakeDef := intake.ToTaskDef()
	intakeDef.Description = "Task to Intake Orders and begin processing setting up order number etc"
	intakeDef.RetryCount = 6
	intakeDef.RetryLogic = "EXPONENTIAL_BACKOFF"

	ensurePriced := workflow.NewSimpleTask("PriceOrder", "PriceOrder").Input("order", "${IntakeOrder.output.order}")
	fraud := workflow.NewSimpleTask("FraudCheck", "FraudCheck").Input("order", "${PriceOrder.output.order}")

	creditCheck := workflow.NewSimpleTask("CreditReview", "CreditReview").Input("order", "${FraudCheck.output.order}")
	creditCheckDecision := workflow.NewSwitchTask("CreditReviewDecision", "${FraudCheck.output.order.payment.type}").
		SwitchCase(string(ordermodel.OnAccount), creditCheck)

	confirmedSignal := workflow.NewWaitTask("ConfirmedSignal").Input("order", "${IntakeOrder.output.order}")

	registerTaskDefinitions(metaApi, intakeDef, ensurePriced, fraud, creditCheck, creditCheckDecision, confirmedSignal)

	conductorWorkflow.Add(intake).
		Add(ensurePriced).
		Add(fraud).
		Add(creditCheckDecision).
		Add(confirmedSignal)

	err = conductorWorkflow.Register(true)
	if err != nil {
		println(err.Error())
	}

	kickoffSampleWorkflowInputs(executor, conductorWorkflow)
}

func kickoffSampleWorkflowInputs(executor *executor.WorkflowExecutor, conductorWorkflow *workflow.ConductorWorkflow) {
	for i := range 5 {

		request := ordermodel.OrderRequest{
			OrderedBy:   "Brian Semrad " + strconv.Itoa(i),
			Total:       utils.RoundFloat(rand.Float64()*(500-10), 2),
			DeliveryZip: "60606",
		}

		if i == 1 {
			request.Total = 0
		}

		if i%2 == 0 {
			request.Payment = &ordermodel.Payment{
				Type:       ordermodel.CreditCard,
				CreditCard: "ouroieqwuoiruqwerqwerqw",
			}
		} else {
			request.Payment = &ordermodel.Payment{
				Type:          ordermodel.OnAccount,
				AccountNumber: "1287879798798",
			}
		}

		workflowInput := ordermodel.OrderProcessingWorkflow{
			Request: request,
		}

		workflowId, err := executor.StartWorkflow(model.NewStartWorkflowRequest(conductorWorkflow.GetName(), conductorWorkflow.GetVersion(), "", workflowInput))

		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		println("WFID: " + workflowId)
	}
}

/**
* Setup the web hook,
 */
func setupWebhook(apiClient *client.APIClient, workflowName string, workflowVersion int32) error {
	whClient := client.NewWebhooksConfigClient(apiClient)
	confirmedWebhookConfig := model.WebhookConfig{
		Id:                              "OrderConfirmedHook",
		Headers:                         map[string]string{"order": "${IntakeOrder.output.order.number}"},
		ReceiverWorkflowNamesToVersions: map[string]int32{workflowName: workflowVersion},
		Verifier:                        "HEADER_BASED",
		SourcePlatform:                  "Custom",
		Name:                            "OrderConfirmed",
	}
	_, _, err := whClient.CreateWebhook(context.Background(), confirmedWebhookConfig)
	if err != nil { //I know this is not good but its a prototype, but also the API's don't prevent creating webhooks without proper setup
		_, _, err = whClient.UpdateWebhook(context.Background(), confirmedWebhookConfig, "OrderConfirmedHook")
		if err != nil {
			println(err.Error())
		}
	}
	return err
}

func registerTaskDefinitions(metaApi client.MetadataClient, intakeDef *model.TaskDef, ensurePriced *workflow.SimpleTask, fraud *workflow.SimpleTask, creditCheck *workflow.SimpleTask, creditCheckDecision *workflow.SwitchTask, confirmedSignal *workflow.WaitTask) {
	_, err := metaApi.RegisterTaskDef(context.Background(), []model.TaskDef{*intakeDef, *ensurePriced.ToTaskDef(), *fraud.ToTaskDef(), *creditCheck.ToTaskDef(),
		*creditCheckDecision.ToTaskDef(), *confirmedSignal.ToTaskDef()})
	if err != nil {
		println(err.Error())
	}

	metaApi.UpdateTaskDef(context.Background(), *intakeDef)
	metaApi.UpdateTaskDef(context.Background(), *ensurePriced.ToTaskDef())
	metaApi.UpdateTaskDef(context.Background(), *fraud.ToTaskDef())
	metaApi.UpdateTaskDef(context.Background(), *creditCheck.ToTaskDef())
	metaApi.UpdateTaskDef(context.Background(), *creditCheckDecision.ToTaskDef())
	metaApi.UpdateTaskDef(context.Background(), *confirmedSignal.ToTaskDef())
}
