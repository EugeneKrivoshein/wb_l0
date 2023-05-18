package streaming

import (
	"encoding/json"
	"log"
	"os"

	"github.com/EugeneKrivoshein/wb_l0/internal/db"

	stan "github.com/nats-io/stan.go"
)

type Pub struct {
	sc   *stan.Conn
	name string
}

func NewPub(conn *stan.Conn) *Pub {
	return &Pub{
		name: "Pub",
		sc:   conn,
	}
}

func (p *Pub) Publish() {
	items := db.Items{
		Chrt_id:      9934930,
		Track_number: "WBILMTESTTRACK",
		Price:        453, Rid: "ab4219087a764ae0btest",
		Name:       "Mascaras",
		Sale:       30,
		Size:       "0",
		TotalPrice: 317,
		Nm_id:      2389212,
		Brand:      "Vivienne Sabo",
		Status:     202,
	}
	delivery := db.Delivery{
		Name:    "test testov",
		Phone:   "+9720000000",
		Zip:     "2639809",
		City:    "Kiryat Mozkin",
		Address: "Ploshad Mira 15",
		Region:  "Kraiot",
		Email:   "test@gmail.com",
	}
	payment := db.Payment{
		Transaction: "b563feb7b2b84b6test",
		Request_id:  "", Currency: "USD",
		Provider: "wbpay", Amount: 1817,
		Payment_dt:    1637907727,
		Bank:          "alpha",
		Delivery_cost: 1500,
		Goods_total:   317,
	}
	order := db.Order{
		OrderUID:          "b563feb7b2b84b6test",
		Entry:             "WBIL",
		InternalSignature: "",
		Delivery:          delivery,
		Payment:           payment,
		Items:             []db.Items{items},
		Locale:            "en",
		CustomerID:        "test",
		TrackNumber:       "WBILMTESTTRACK",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		Total:             int(items.TotalPrice),
	}

	orderData, err := json.Marshal(order)
	if err != nil {
		log.Printf("%s: json.Marshal error: %v\n", p.name, err)
	}

	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			log.Printf("%s: error publishing msg id %s: %v\n", p.name, ackedNuid, err.Error())
		} else {
			log.Printf("%s: received ack for msg id: %s\n", p.name, ackedNuid)
		}
	}

	log.Printf("%s: publishing data ...\n", p.name)
	nuid, err := (*p.sc).PublishAsync(os.Getenv("NATS_SUBJECT"), orderData, ackHandler)
	if err != nil {
		log.Printf("%s: error publishing msg %s: %v\n", p.name, nuid, err.Error())
	}
}
