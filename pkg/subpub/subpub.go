package subpub

import (
	"context"
	"log"
	"slices"
	"sync"
)

// MessageHandler is a callback function that processes messages delivered to subscriber
type MessageHandler func(msg any)

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for.
	Unsubscribe()
}

type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given subject.
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg any) error

	// Close will shutdown sub-pub system.
	// May be blocked by data delivery until the context is canceled.
	Close(ctx context.Context) error
}

const CHANNEL_SIZE = 100

type SubscriptionInstance struct {
	Subject       string
	Handler       MessageHandler
	MsgChan       chan any
	SubPubBus     *SubPubImpl
	CancelContext context.CancelFunc
	wg            *sync.WaitGroup
}

type SubPubImpl struct {
	mu          sync.RWMutex
	Subscribers map[string][]*SubscriptionInstance
	Closed      bool
	wg          sync.WaitGroup
}

func NewSubPubImpl() *SubPubImpl {
	return &SubPubImpl{
		mu:          sync.RWMutex{},
		Subscribers: make(map[string][]*SubscriptionInstance),
		Closed:      false,
		wg:          sync.WaitGroup{},
	}
}

func (s *SubPubImpl) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Closed {
		return nil, ErrorSubPubIsClosed
	}

	ctx, cancel := context.WithCancel(context.Background())
	subsctiption := &SubscriptionInstance{
		Subject:       subject,
		Handler:       cb,
		MsgChan:       make(chan any, CHANNEL_SIZE),
		SubPubBus:     s,
		CancelContext: cancel,
		wg:            &s.wg,
	}

	s.Subscribers[subject] = append(s.Subscribers[subject], subsctiption)

	s.wg.Add(1)
	go s.messageProcessing(ctx, subsctiption)

	return subsctiption, nil
}

func (s *SubPubImpl) messageProcessing(ctx context.Context, sub *SubscriptionInstance) {
	defer s.wg.Done()
	for {
		select {
		case msg, ok := <-sub.MsgChan:
			if !ok {
				return
			}
			sub.Handler(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (s *SubPubImpl) Publish(subject string, msg any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Closed {
		return ErrorSubPubIsClosed
	}

	subs, ok := s.Subscribers[subject]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		// если канал для сообщений переполнен сообщение будет утеряно
		select {
		case sub.MsgChan <- msg:
		default:
			log.Println(ErrorChannelOverflow)
		}
	}

	return nil
}

func (s *SubPubImpl) Close(ctx context.Context) error {
	s.mu.Lock()
	if s.Closed {
		s.mu.Unlock()
		return nil
	}

	for _, subs := range s.Subscribers {
		for _, sub := range subs {
			close(sub.MsgChan)
		}
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (sub *SubscriptionInstance) Unsubscribe() {
	bus := sub.SubPubBus
	bus.mu.Lock()
	defer bus.mu.Unlock()

	subs, ok := bus.Subscribers[sub.Subject]
	if !ok {
		return
	}

	for i, inst := range subs {
		if inst == sub {
			bus.Subscribers[sub.Subject] = slices.Delete(subs, i, i)
			break
		}
	}

	sub.CancelContext()
	close(sub.MsgChan)
}
