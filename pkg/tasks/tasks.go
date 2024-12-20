package ordertasks

import (
	ordermodel "conductor-demo/pkg/model"
	"conductor-demo/pkg/utils"
	"encoding/json"
	"fmt"
	"math/rand/v2"

	"github.com/conductor-sdk/conductor-go/sdk/model"
)

func AcceptOrderFrom(task *model.Task) (interface{}, error) {
	output, _ := json.Marshal(task.InputData)
	fmt.Println("Recieved Order from: " + string(output))
	orderRequest, err := ordermodel.BuildRequestFromTaskInput(task.InputData["order"].(map[string]interface{}))
	//if there is an error for now just return will come back and validate retry behavior later to ensure its not infinite.
	if err != nil {
		return buildErrorResult(task, err), nil
	}
	order := ordermodel.FromRequest(*orderRequest)
	taskResult, err := buildOrderTaskCompletedResult(task, order)
	if err != nil {
		return buildErrorResult(task, err), nil
	}

	return taskResult, nil
}

func buildOrderTaskCompletedResult(task *model.Task, order *ordermodel.Order) (*model.TaskResult, error) {
	taskResult := model.NewTaskResultFromTask(task)
	taskResult.Status = model.CompletedTask
	outputData, err := ordermodel.BuildOutputMapFromOrder(order)
	taskResult.OutputData = *outputData
	return taskResult, err
}

func buildErrorResult(task *model.Task, err error) model.TaskResult {
	taskResult := model.NewTaskResultFromTask(task)
	taskResult.Status = model.FailedTask
	taskResult.OutputData["error"] = err.Error()
	return *taskResult
}

func FraudCheck(task *model.Task) (interface{}, error) {
	output, _ := json.Marshal(task.InputData)
	fmt.Println("Starting fraud for: " + string(output))

	order, err := ordermodel.BuildOrderFromTaskInput(task.InputData["order"].(map[string]interface{}))
	if err != nil {
		return buildErrorResult(task, err), err
	}

	fmt.Println("Performing Fraud Check for: " + order.Number)
	if order.Total > 400 {
		order.Metadata.Fraud = &ordermodel.FraudCheck{
			Fraudulent: true,
			Reason:     "Too much money",
		}
	} else {
		order.Metadata.Fraud = &ordermodel.FraudCheck{
			Fraudulent: false,
		}
	}

	taskResult, err := buildOrderTaskCompletedResult(task, order)
	if err != nil {
		return buildErrorResult(task, err), nil
	}
	if order.Metadata.Fraud.Fraudulent {
		taskResult.Status = model.FailedWithTerminalErrorTask
	} else {
		taskResult.Status = model.CompletedTask
	}

	return taskResult, nil
}

func PriceOrder(task *model.Task) (interface{}, error) {
	output, _ := json.Marshal(task.InputData)
	fmt.Println("Starting fraud for: " + string(output))

	order, err := ordermodel.BuildOrderFromTaskInput(task.InputData["order"].(map[string]interface{}))
	if err != nil {
		return buildErrorResult(task, err), nil
	}

	if order.Total == 0 {
		order.Total = utils.RoundFloat(rand.Float64()*(400-10), 2)
	}
	taskResult, err := buildOrderTaskCompletedResult(task, order)
	if err != nil {
		return buildErrorResult(task, err), nil
	}

	return taskResult, nil
}

func CreditReview(task *model.Task) (interface{}, error) {
	output, _ := json.Marshal(task.InputData)
	fmt.Println("Starting Credit Review for: " + string(output))

	order, err := ordermodel.BuildOrderFromTaskInput(task.InputData["order"].(map[string]interface{}))
	if err != nil {
		return buildErrorResult(task, err), err
	}

	fmt.Println("Performing Credit Check for: " + order.Number)
	if order.Total > 300 {
		order.Metadata.CreditReview = &ordermodel.CreditReview{
			CreditExtended: false,
			Reason:         "Collateral not on hand",
		}
	} else {
		order.Metadata.CreditReview = &ordermodel.CreditReview{
			CreditExtended: true,
			Reason:         "Good Customer",
			Amount:         150.00,
		}
	}

	taskResult, err := buildOrderTaskCompletedResult(task, order)
	if err != nil {
		return buildErrorResult(task, err), nil
	}
	if !order.Metadata.CreditReview.CreditExtended {
		taskResult.Status = model.FailedWithTerminalErrorTask
	} else {
		taskResult.Status = model.CompletedTask
	}

	return taskResult, nil
}
