package subpub

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSubsctibe(t *testing.T) {
	sp := NewSubPubImpl()
	handler := func (msg any)  {}

	sub, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Errorf("Subscribe failed: %v", err)
	}
	if sub == nil {
		t.Errorf("Expected subscription instance, got nil")
	}

	if len(sp.Subscribers["test"]) != 1 {
		t.Error("Expected one subscriber for 'test' subject")
	}
}

func TestSubscribeClosed(t *testing.T) {
	sp := NewSubPubImpl()
	sp.Closed = true

	_, err := sp.Subscribe("test", func(msg any) {})
	if !errors.Is(err, ErrorSubPubIsClosed) {
		t.Errorf("Expected ErrorSubPubIsClosed, got %v", err)
	}
}

func TestPublish(t *testing.T) {
	sp := NewSubPubImpl()
	var receivedMsg any
	var wg sync.WaitGroup

	wg.Add(1)
	handler := func(msg any) {
		receivedMsg = msg
		wg.Done()
	}

	_, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Fatal(err)
	}

	testMsg := "hello world"
	err = sp.Publish("test", testMsg)
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	if receivedMsg != testMsg {
		t.Errorf("Expected message %v, got %v", testMsg, receivedMsg)
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	sp := NewSubPubImpl()
	err := sp.Publish("nonexistent", "message")
	if err != nil {
		t.Errorf("Publish with no subscribers should not return error, got %v", err)
	}
}

func TestPublishClosed(t *testing.T) {
	sp := NewSubPubImpl()
	sp.Closed = true

	err := sp.Publish("test", "message")
	if !errors.Is(err, ErrorSubPubIsClosed) {
		t.Errorf("Expected ErrorSubPubIsClosed, got %v", err)
	}
}

func TestPublishChannelOverflow(t *testing.T) {
	sp := NewSubPubImpl()
	var callCount int
	var mu sync.Mutex

	handler := func(msg any) {
		mu.Lock()
		callCount++
		mu.Unlock()
		time.Sleep(100 * time.Millisecond) // simulate slow processing
	}

	_, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Fatal(err)
	}

	// Fill the channel buffer and more
	for i := 0; i < CHANNEL_SIZE+5; i++ {
		err = sp.Publish("test", i)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Wait for all messages to be processed
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount >= CHANNEL_SIZE+5 {
		t.Error("Expected some messages to be dropped due to channel overflow")
	}
}

func TestUnsubscribe(t *testing.T) {
	sp := NewSubPubImpl()
	handler := func(msg any) {}

	sub, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Fatal(err)
	}

	if len(sp.Subscribers["test"]) != 1 {
		t.Fatal("Expected one subscriber")
	}

	sub.(*SubscriptionInstance).Unsubscribe()

	if len(sp.Subscribers["test"]) != 0 {
		t.Error("Expected no subscribers after unsubscribe")
	}
}

func TestClose(t *testing.T) {
	sp := NewSubPubImpl()
	handler := func(msg any) {}

	_, err := sp.Subscribe("test1", handler)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sp.Subscribe("test2", handler)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = sp.Close(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !sp.Closed {
		t.Error("Expected SubPub to be closed after Close()")
	}
}

func TestMultipleSubjects(t *testing.T) {
	sp := NewSubPubImpl()
	var wg sync.WaitGroup
	wg.Add(2)

	var sub1Msg, sub2Msg any

	_, err := sp.Subscribe("subject1", func(msg any) {
		sub1Msg = msg
		wg.Done()
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = sp.Subscribe("subject2", func(msg any) {
		sub2Msg = msg
		wg.Done()
	})
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Publish("subject1", "message1")
	if err != nil {
		t.Fatal(err)
	}

	err = sp.Publish("subject2", "message2")
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	if sub1Msg != "message1" {
		t.Errorf("Expected 'message1', got %v", sub1Msg)
	}
	if sub2Msg != "message2" {
		t.Errorf("Expected 'message2', got %v", sub2Msg)
	}
}