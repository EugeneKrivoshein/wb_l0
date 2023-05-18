package streaming

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/EugeneKrivoshein/wb_l0/internal/db"

	stan "github.com/nats-io/stan.go"
)

type Sub struct {
	sub      stan.Subscription
	dbObject *db.DB
	sc       *stan.Conn
	name     string
}

func NewSub(db *db.DB, conn *stan.Conn) *Sub {
	return &Sub{
		name:     "Sub",
		dbObject: db,
		sc:       conn,
	}
}

func (s *Sub) Subscribe() {
	var err error

	ackWait, err := strconv.Atoi(os.Getenv("NATS_ACK_WAIT_SECONDS"))
	if err != nil {
		log.Printf("%s: message arrived!\n", s.name)
		return
	}

	s.sub, err = (*s.sc).Subscribe(
		os.Getenv("NATS_SUBJECT"),
		func(m *stan.Msg) {
			log.Printf("%s: received a message!\n", s.name)
			if s.messageHandler(m.Data) {
				err := m.Ack()
				if err != nil {
					log.Printf("%s ack() err: %s", s.name, err)
				}
			}
		},
		stan.AckWait(time.Duration(ackWait)*time.Second),
		stan.DurableName(os.Getenv("NATS_DURABLE_NAME")),
		stan.SetManualAckMode(),
		stan.MaxInflight(10))
	if err != nil {
		log.Printf("%s: error: %v\n", s.name, err)
	}
	log.Printf("%s: subscribed to subject %s\n", s.name, os.Getenv("NATS_SUBJECT"))
}

func (s *Sub) messageHandler(data []byte) bool {
	recievedOrder := db.Order{}
	err := json.Unmarshal(data, &recievedOrder)
	if err != nil {
		log.Printf("%s:message handler error, %v\n", s.name, err)
		return true
	}
	log.Printf("%s: unmarshal Order to struct: %v\n", s.name, recievedOrder)

	_, err = s.dbObject.AddOrder(recievedOrder)
	if err != nil {
		log.Printf("%s: err add order: %v\n", s.name, err)
		return false
	}
	return true
}

func (s *Sub) Unsubscribe() {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
}
