package ordermodel

import (
	"encoding/json"
	"math/rand"
	"strconv"
)

type OrderProcessingWorkflow struct {
	Request OrderRequest `json:"request,omitempty"`
}

type OrderRequest struct {
	OrderedBy   string   `json:"orderedBy,omitempty"`
	Total       float64  `json:"total"`
	DeliveryZip string   `json:"deliveryZip"`
	Payment     *Payment `json:"payment,omitempty"`
}

type Order struct {
	Number      string         `json:"number,omitempty"`
	OrderedBy   string         `json:"orderedBy,omitempty"`
	Total       float64        `json:"total,omitempty"`
	DeliveryZip string         `json:"deliveryZip,omitempty"`
	Payment     *Payment       `json:"payment,omitempty"`
	Metadata    *OrderMetadata `json:"metadata,omitempty"`
}

type OrderMetadata struct {
	Fraud        *FraudCheck   `json:"fraud,omitempty"`
	CreditReview *CreditReview `json:"creditReview,omitempty"`
}

type FraudCheck struct {
	Fraudulent bool   `json:"fradulent"`
	Reason     string `json:"reason,omitempty"`
}

type CreditReview struct {
	CreditExtended bool   `json:"creditExtended,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Amount         float64
}

type Payment struct {
	Type          PaymentType `json:"type,omitempty"`
	CreditCard    string      `json:"creditCard,omitempty"`
	AccountNumber string      `json:"accountNumber,omitempty"`
}

type PaymentType string

const (
	OnAccount  PaymentType = "ON_ACCOUNT"
	CreditCard PaymentType = "CREDIT_CARD"
)

func BuildRequestFromTaskInput(input map[string]interface{}) (*OrderRequest, error) {
	request := OrderRequest{}
	mar, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(mar, &request)
	return &request, err
}

// TODO: There is 1000% a better way to handle this and the above, come back to it this is quick and dirty for now
func BuildOrderFromTaskInput(input map[string]interface{}) (*Order, error) {
	order := Order{}
	mar, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(mar, &order)
	return &order, err
}

func maptostruct(m map[string]interface{}, result interface{}) error {
	mar, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(mar, &result)

	return err
}

func FromRequest(request OrderRequest) *Order {
	randomnumber := rand.Int63()
	return &Order{
		Number:      strconv.FormatInt(randomnumber, 10),
		OrderedBy:   request.OrderedBy,
		Total:       request.Total,
		DeliveryZip: request.DeliveryZip,
		Metadata:    &OrderMetadata{},
		Payment:     request.Payment,
	}
}

func BuildOutputMapFromOrder(order *Order) (*map[string]interface{}, error) {
	output := make(map[string]interface{})
	mar, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	order_output := make(map[string]interface{})
	_ = json.Unmarshal(mar, &order_output)
	output["order"] = order_output
	return &output, nil
}
