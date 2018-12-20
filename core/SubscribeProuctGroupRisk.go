package core

import (
	"context"
	"io"

	pb "github.com/fengdu/risk-monitor-server/pb"

	log "github.com/sirupsen/logrus"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

type prouctGroupRiskSubscriber struct {
	ctx     context.Context
	client  pb.RiskMonitorServerClient
	session *r.Session
}

func NewprouctGroupRiskSubscriber(ctx context.Context, client pb.RiskMonitorServerClient, session *r.Session) Subscriber {
	return &prouctGroupRiskSubscriber{
		ctx:     ctx,
		client:  client,
		session: session,
	}
}

func (c *prouctGroupRiskSubscriber) Read(buffer int) <-chan interface{} {
	log.Infoln("SubscribeProuctGroupRisk started...")

	out := make(chan interface{}, buffer)

	go func() {
		defer close(out)

		stream, err := c.client.SubscribeProuctGroupRisk(c.ctx, &pb.SubscribeReq{})
		if err != nil {
			log.Errorf("SubscribeProuctGroupRisk failed : %v", err)
			return
		}
		for {
			select {
			default:
				item, err := stream.Recv()
				if err == io.EOF {
					log.Warnf("SubscribeProuctGroupRisk receive EOF : %v", err)
					return
				}
				if err != nil {
					log.Errorf("SubscribeProuctGroupRisk receive failed : %v", err)
					return
				}

				out <- item
			case <-c.ctx.Done():
				return
			}

		}
	}()

	return out
}

func (c *prouctGroupRiskSubscriber) Convert(in <-chan interface{}) <-chan *Message {
	out := make(chan *Message)
	go func() {
		defer close(out)
		for n := range in {
			item := n.(*pb.ProuctGroupRiskRtn)

			msg := &Message{
				TableName:  TableName_SubscribeProuctGroupRisk,
				ActionFlag: item.ActionFlag,
				ActionKey:  string(item.MonitorNo) + "#" + string(item.ProductGroupNo) + "#" + string(item.ContractCode),
				Msg:        *item,
			}

			out <- msg
		}

	}()
	return out
}

func (c *prouctGroupRiskSubscriber) Write(in <-chan *Message) {
	for msg := range in {

		err := msg.Replace(c.session)
		if err != nil {
			log.Errorf("Write ProuctGroupRiskRtn message failed : err: %v, message: %+v", err, *msg)
			return
		}
	}
}