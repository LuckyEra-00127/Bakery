package messaging

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) *Publisher {
	conn, err := nats.Connect(url)
	if err != nil {
		log.Printf("NATS disabled: %v", err)
		return &Publisher{}
	}
	return &Publisher{conn: conn}
}

func (p *Publisher) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}

func (p *Publisher) PublishJSON(subject string, value any) {
	if p.conn == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("marshal NATS event error: %v", err)
		return
	}
	if err := p.conn.Publish(subject, data); err != nil {
		log.Printf("publish NATS event error: %v", err)
	}
}
