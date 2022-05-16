package validator

import (
	"context"
	"fmt"
	"testing"

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
	if len(m.Name) == 0 || m.Name != "test" {
		return fmt.Errorf("name is empty")
	}
	return nil
}

func (h *Handler) Sub(ctx context.Context, req *Message) error {
	return nil
}

func TestValidator(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create server
	srv := server.NewServer(
		server.Name("helloworld"),
		server.Codec("application/json", codec.NewCodec()),
		server.WrapHandler(NewServerHandlerWrapper()),
		server.WrapSubscriber(NewServerSubscriberWrapper()),
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
		client.Wrap(NewClientWrapper()),
	)

	if err := cli.Publish(ctx, cli.NewMessage("test", &Message{Name: "test1"}, client.WithMessageContentType("application/json"))); err == nil {
		t.Fatalf("validator not works as message with bad contents")
	}

	if err := cli.Publish(ctx, cli.NewMessage("test", &Message{Name: "test"}, client.WithMessageContentType("application/json"))); err == nil {
		t.Fatal(err)
	}

	// stop server
	if err := srv.Stop(); err != nil {
		t.Fatal(err)
	}
}
