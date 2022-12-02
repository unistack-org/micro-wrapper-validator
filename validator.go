package validator

import (
	"context"

	"github.com/unistack-org/micro/v3/client"
	"github.com/unistack-org/micro/v3/errors"
	"github.com/unistack-org/micro/v3/server"
)

type validator interface {
	Validate() error
}

type wrapper struct {
	client.Client
}

func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		handler := &wrapper{
			Client: c,
		}
		return handler
	}
}

func NewClientCallWrapper() client.CallWrapper {
	return func(fn client.CallFunc) client.CallFunc {
		return func(ctx context.Context, addr string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if v, ok := req.Body().(validator); ok {
				if verr := v.Validate(); verr != nil {
					return errors.BadRequest(req.Service(), "%v", verr)
				}
			}
			err := fn(ctx, addr, req, rsp, opts)
			if v, ok := rsp.(validator); ok {
				if verr := v.Validate(); verr != nil {
					return errors.BadGateway(req.Service(), "%v", verr)
				}
			}
			return err
		}
	}
}

func (w *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	if v, ok := req.Body().(validator); ok {
		if verr := v.Validate(); verr != nil {
			return errors.BadRequest(req.Service(), "%v", verr)
		}
	}
	err := w.Client.Call(ctx, req, rsp, opts...)
	if v, ok := rsp.(validator); ok {
		if verr := v.Validate(); verr != nil {
			return errors.BadGateway(req.Service(), "%v", verr)
		}
	}
	return err
}

func (w *wrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	if v, ok := req.Body().(validator); ok {
		if verr := v.Validate(); verr != nil {
			return nil, errors.BadRequest(req.Service(), "%v", verr)
		}
	}
	return w.Client.Stream(ctx, req, opts...)
}

func (w *wrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	if v, ok := msg.Payload().(validator); ok {
		if err := v.Validate(); err != nil {
			return errors.BadRequest(msg.Topic(), "%v", err)
		}
	}
	return w.Client.Publish(ctx, msg, opts...)
}

func NewServerHandlerWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if v, ok := req.Body().(validator); ok {
				if verr := v.Validate(); verr != nil {
					return errors.BadRequest(req.Service(), "%v", verr)
				}
			}
			err := fn(ctx, req, rsp)
			if v, ok := rsp.(validator); ok {
				if verr := v.Validate(); verr != nil {
					return errors.BadGateway(req.Service(), "%v", verr)
				}
			}
			return err
		}
	}
}

func NewServerSubscriberWrapper() server.SubscriberWrapper {
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			if v, ok := msg.Payload().(validator); ok {
				if err := v.Validate(); err != nil {
					return errors.BadRequest(msg.Topic(), "%v", err)
				}
			}
			return fn(ctx, msg)
		}
	}
}
