package validator

import (
	"context"
	"fmt"
	"testing"

	memory "go.unistack.org/micro/v3/broker/memory"
	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/codec"
	"go.unistack.org/micro/v3/server"
)

type Handler struct {
	t *testing.T
}

type Message struct {
	Name string
}

func (m *Message) Validate() error {
	return fmt.Errorf("SSS")
}

func (h *Handler) Sub(ctx context.Context, req *Message) error {
	return nil
}

func TestValidator(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := memory.NewBroker()

	if err := b.Init(); err != nil {
		t.Fatal(err)
	}

	if err := b.Connect(ctx); err != nil {
		t.Fatal(err)
	}

	w := NewHook()
	// create server
	srv := server.NewServer(
		server.Name("helloworld"),
		server.Codec("application/json", codec.NewCodec()),
		server.Broker(b),
		server.Hooks(
			server.HookHandler(w.ServerHandler),
			server.HookSubHandler(w.ServerSubscriber),
		),
		server.Context(ctx),
	)

	h := &Handler{t: t}

	if err := srv.Subscribe(srv.NewSubscriber("test", h.Sub)); err != nil {
		t.Fatal(err)
	}

	if err := srv.Init(); err != nil {
		t.Fatal(err)
	}

	// start server
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}

	cli := client.NewClient(
		client.ContentType("application/json"),
		client.Codec("application/json", codec.NewCodec()),
		client.Broker(b),
		client.Hooks(
			client.HookCall(w.ClientCall),
			client.HookStream(w.ClientStream),
			client.HookPublish(w.ClientPublish),
			client.HookBatchPublish(w.ClientBatchPublish),
		),
	)

	if err := cli.Init(); err != nil {
		t.Fatal(err)
	}

	if err := cli.Publish(ctx, cli.NewMessage("test", &Message{Name: "test1"}, client.WithMessageContentType("application/json"))); err == nil {
		t.Fatalf("validator not works as message with bad contents")
	}

	// stop server
	if err := srv.Stop(); err != nil {
		t.Fatal(err)
	}
}
