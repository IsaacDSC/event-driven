package entities

type Product struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	OrderID     string    `json:"order_id"`
	Client      string    `json:"client"`
	ClientEmail string    `json:"client_email"`
	Products    []Product `json:"products"`
	Total       float64   `json:"total"`
	Status      string    `json:"status"`
}
