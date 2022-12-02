package validator

import (
	"context"

	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/errors"
	"go.unistack.org/micro/v3/server"
)

var (
	DefaultClientErrorFunc = func(req client.Request, rsp interface{}, err error) error {
		if rsp != nil {
			return errors.BadGateway(req.Service(), "%v", err)
		}
		return errors.BadRequest(req.Service(), "%v", err)
	}

	DefaultServerErrorFunc = func(req server.Request, rsp interface{}, err error) error {
		if rsp != nil {
			return errors.BadGateway(req.Service(), "%v", err)
		}
		return errors.BadRequest(req.Service(), "%v", err)
	}

	DefaultPublishErrorFunc = func(msg client.Message, err error) error {
		return errors.BadRequest(msg.Topic(), "%v", err)
	}

	DefaultSubscribeErrorFunc = func(msg server.Message, err error) error {
		return errors.BadRequest(msg.Topic(), "%v", err)
	}
)

type (
	ClientErrorFunc    func(client.Request, interface{}, error) error
	ServerErrorFunc    func(server.Request, interface{}, error) error
	PublishErrorFunc   func(client.Message, error) error
	SubscribeErrorFunc func(server.Message, error) error
)

// Options struct holds wrapper options
type Options struct {
	ClientErrorFn    ClientErrorFunc
	ServerErrorFn    ServerErrorFunc
	PublishErrorFn   PublishErrorFunc
	SubscribeErrorFn SubscribeErrorFunc
}

// Option func signature
type Option func(*Options)

func ClientReqErrorFn(fn ClientErrorFunc) Option {
	return func(o *Options) {
		o.ClientErrorFn = fn
	}
}

func ServerErrorFn(fn ServerErrorFunc) Option {
	return func(o *Options) {
		o.ServerErrorFn = fn
	}
}

func PublishErrorFn(fn PublishErrorFunc) Option {
	return func(o *Options) {
		o.PublishErrorFn = fn
	}
}

func SubscribeErrorFn(fn SubscribeErrorFunc) Option {
	return func(o *Options) {
		o.SubscribeErrorFn = fn
	}
}

func NewOptions(opts ...Option) Options {
	options := Options{
		ClientErrorFn:    DefaultClientErrorFunc,
		ServerErrorFn:    DefaultServerErrorFunc,
		PublishErrorFn:   DefaultPublishErrorFunc,
		SubscribeErrorFn: DefaultSubscribeErrorFunc,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

type validator interface {
	Validate() error
}

type wrapper struct {
	client.Client
	opts Options
}

func NewClientWrapper(opts ...Option) client.Wrapper {
	return func(c client.Client) client.Client {
		handler := &wrapper{
			Client: c,
			opts:   NewOptions(opts...),
		}
		return handler
	}
}

func NewClientCallWrapper(opts ...Option) client.CallWrapper {
	options := NewOptions(opts...)
	return func(fn client.CallFunc) client.CallFunc {
		return func(ctx context.Context, addr string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if v, ok := req.Body().(validator); ok {
				if verr := v.Validate(); verr != nil {
					return options.ClientErrorFn(req, nil, verr)
				}
			}
			err := fn(ctx, addr, req, rsp, opts)
			if v, ok := rsp.(validator); ok {
				if verr := v.Validate(); verr != nil {
					return options.ClientErrorFn(req, rsp, verr)
				}
			}
			return err
		}
	}
}

func (w *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	if v, ok := req.Body().(validator); ok {
		if err := v.Validate(); err != nil {
			return w.opts.ClientErrorFn(req, nil, err)
		}
	}
	err := w.Client.Call(ctx, req, rsp, opts...)
	if v, ok := rsp.(validator); ok {
		if verr := v.Validate(); verr != nil {
			return w.opts.ClientErrorFn(req, rsp, verr)
		}
	}
	return err
}

func (w *wrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	if v, ok := req.Body().(validator); ok {
		if err := v.Validate(); err != nil {
			return nil, w.opts.ClientErrorFn(req, nil, err)
		}
	}
	return w.Client.Stream(ctx, req, opts...)
}

func (w *wrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	if v, ok := msg.Payload().(validator); ok {
		if err := v.Validate(); err != nil {
			return w.opts.PublishErrorFn(msg, err)
		}
	}
	return w.Client.Publish(ctx, msg, opts...)
}

func NewServerHandlerWrapper(opts ...Option) server.HandlerWrapper {
	options := NewOptions(opts...)
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if v, ok := req.Body().(validator); ok {
				if err := v.Validate(); err != nil {
					return options.ClientErrorFn(req, nil, err)
				}
			}
			err := fn(ctx, req, rsp)
			if v, ok := rsp.(validator); ok {
				if verr := v.Validate(); verr != nil {
					return options.ClientErrorFn(req, rsp, err)
				}
			}
			return err
		}
	}
}

func NewServerSubscriberWrapper(opts ...Option) server.SubscriberWrapper {
	options := NewOptions(opts...)
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			if v, ok := msg.Body().(validator); ok {
				if err := v.Validate(); err != nil {
					return options.SubscribeErrorFn(msg, err)
				}
			}
			return fn(ctx, msg)
		}
	}
}
