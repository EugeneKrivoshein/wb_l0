package db

type Order struct {
	OrderUID          string   `json:"order_uid"`
	Entry             string   `json:"entry"`
	InternalSignature string   `json:"internal_signature"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Items  `json:"items"`
	Locale            string   `json:"locale"`
	CustomerID        string   `json:"customer_id"`
	TrackNumber       string   `json:"track_number"`
	DeliveryService   string   `json:"delivery_service"`
	Shardkey          string   `json:"shardkey"`
	SmID              int      `json:"sm_id"`
	Total             int      `json:"total"`
}

func (o *Order) GetTotalPrice() int {
	var total int64
	for _, item := range o.Items {
		total += item.TotalPrice
	}
	return int(total)
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction   string `json:"transaction"`
	Request_id    string `json:"request_id"`
	Currency      string `json:"currency"`
	Provider      string `json:"provider"`
	Amount        int64  `json:"amount"`
	Payment_dt    int64  `json:"payment_dt"`
	Bank          string `json:"bank"`
	Delivery_cost int64  `json:"delivery_cost"`
	Goods_total   int64  `json:"goods_total"`
	Custom_fee    int64  `json:"custom_fee"`
}
type Items struct {
	Chrt_id      int64  `json:"chrt_id"`
	Track_number string `json:"track_number"`
	Price        int64  `json:"price"`
	Rid          string `json:"rid"`
	Name         string `json:"name"`
	Sale         int64  `json:"sale"`
	Size         string `json:"size"`
	TotalPrice   int64  `json:"total_price"`
	Nm_id        int64  `json:"nm_id"`
	Brand        string `json:"brand"`
	Status       int64  `json:"status"`
}

type OrderOut struct {
	OrderUID        string `json:"order_uid"`
	Entry           string `json:"entry"`
	Total           int    `json:"total"`
	CustomerID      string `json:"customer_id"`
	TrackNumber     string `json:"track_number"`
	DeliveryService string `json:"delivery_service"`
}
