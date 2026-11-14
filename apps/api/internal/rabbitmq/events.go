package rabbitmq

type OrderCreatedEvent struct {
	Event     string `json:"event"`
	OrderID   int64  `json:"orderId"`
	ProductID int    `json:"productId"`
	Quantity  int    `json:"quantity"`
}
