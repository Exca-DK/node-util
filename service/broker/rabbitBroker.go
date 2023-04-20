package broker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitBroker(address string, username string, password string, recorder Recorder) (Broker, error) {
	broker := &rabbitBroker{}
	//TODO change plain auth into ssl/tls
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%v:%v@%v/", username, password, address))
	if err != nil {
		return nil, err
	}
	broker.conn = conn

	go func() {
		err := <-broker.conn.NotifyClose(make(chan *amqp.Error))
		broker.Stop(err)
	}()

	ch, err := broker.conn.Channel()
	if err != nil {
		return nil, err
	}

	return &rabbitBroker{ch: ch, id: username, msgFeed: map[string]*Feed[BrokerMessage]{}, recorder: recorder}, nil
}

type rabbitBroker struct {
	id string //broker id

	//rabbit prims
	conn *amqp.Connection
	ch   *amqp.Channel

	//sync
	stopped atomic.Bool
	err     atomic.Value
	done    chan struct{}

	//metrics
	recorder Recorder

	stopping atomic.Bool

	//one to many per topic behind mutex
	mu      sync.Mutex
	msgFeed map[string]*Feed[BrokerMessage]
}

func (b *rabbitBroker) PublishMessage(ctx context.Context, topic string, data []byte) error {
	//TODO marshall and unmarshall in broker proto msgs
	return b.publish(ctx, topic, BrokerMessage{Data: data})
}

func (b *rabbitBroker) SubscribeTopic(topic string) (chan BrokerMessage, Subscription, error) {
	if b.stopped.Load() {
		return nil, nil, ErrBrokerStopped
	}
	queue, err := b.ch.QueueDeclare(topic, false, true, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	chConsume, err := b.ch.Consume(queue.Name, b.id, true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	feed, ok := b.msgFeed[topic]
	if !ok {
		b.msgFeed[topic] = &Feed[BrokerMessage]{}
		feed = b.msgFeed[topic]
		go func() {
			b.consume(chConsume, topic)
			b.mu.Lock()
			delete(b.msgFeed, topic)
			b.mu.Unlock()
		}()
	}
	chMsg := make(chan BrokerMessage)
	return chMsg, feed.Subscribe(chMsg), nil
}

func (b *rabbitBroker) publish(ctx context.Context, topic string, msg BrokerMessage) error {
	err := b.ch.PublishWithContext(ctx, "", topic, false, false, amqp.Publishing{ContentType: "text/plain", Body: msg.Data, UserId: b.id})
	if err != nil {
		return err
	}
	b.recorder.RecordEgress(len(msg.Data))
	return nil
}

func (b *rabbitBroker) Stop(err error) {
	if !b.stopped.CompareAndSwap(false, true) {
		return
	}
	if err != nil {
		b.err.Store(err)
	}
	close(b.done)
}

func (b *rabbitBroker) consume(queueCh <-chan amqp.Delivery, topic string) {
	deltaCh := make(chan int)
	b.mu.Lock()
	feed := b.msgFeed[topic]
	b.mu.Unlock()
	feed.callback.Store(func(i int) { deltaCh <- i })
	var count int
	for {
		select {
		case msg, ok := <-queueCh:
			if !ok {
				return
			}
			b.recorder.RecordIngress(len(msg.Body))
			feed.Notify(BrokerMessage{Data: msg.Body})
		case <-b.done:
			go b.unsubscribe(topic)
			//1st case won't consume messages since closed done will always take priority
			for {
				_, ok := <-queueCh
				if !ok {
					return
				}
			}
		case delta := <-deltaCh:
			count += delta
			//unsubscribe and try to consume in first case
			if count == 0 {
				go b.unsubscribe(topic)
			}
		}
	}
}

func (b *rabbitBroker) unsubscribe(topic string) {
	b.mu.Lock()
	delete(b.msgFeed, topic)
	b.mu.Unlock()
	if err := b.ch.Cancel(topic, false); err != nil {
		fmt.Printf("failed unsubscribing from topic %s. error: %v\n", topic, err)
	}
}
