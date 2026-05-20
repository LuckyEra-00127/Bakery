package messaging

import (
	"log"

	"github.com/nats-io/nats.go"
)

type Listener struct {
	conn *nats.Conn
}

func NewListener(url string) *Listener {
	conn, err := nats.Connect(url)
	if err != nil {
		log.Printf("NATS listener disabled: %v", err)
		return &Listener{}
	}
	return &Listener{conn: conn}
}

func (l *Listener) Start() {
	if l.conn == nil {
		return
	}
	_, err := l.conn.Subscribe("order.created", func(msg *nats.Msg) {
		log.Printf("received NATS event order.created: %s", string(msg.Data))
		// Next implementation step: update statistics_daily and create email notification.
	})
	if err != nil {
		log.Printf("subscribe order.created error: %v", err)
	}

	_, err = l.conn.Subscribe("return.report.created", func(msg *nats.Msg) {
		log.Printf("received NATS event return.report.created: %s", string(msg.Data))
		// Next implementation step: update daily/weekly/monthly return statistics.
	})
	if err != nil {
		log.Printf("subscribe return.report.created error: %v", err)
	}
}

func (l *Listener) Close() {
	if l.conn != nil {
		l.conn.Close()
	}
}
