package route

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OnUnavailableInterceptor struct {
	onUnavailable func(context.Context) error
}

func NewOnUnavailableInterceptor() *OnUnavailableInterceptor {
	return &OnUnavailableInterceptor{
		onUnavailable: func(context.Context) error {
			return errors.New("unimplemented")
		},
	}
}

func (i *OnUnavailableInterceptor) SetOnUnavailable(fn func(context.Context) error) {
	i.onUnavailable = fn
}

func (i *OnUnavailableInterceptor) UnaryIntercept(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {

	err := invoker(ctx, method, req, reply, cc, opts...)
	if err == nil {
		return nil
	}

	if status.Code(err) != codes.Unavailable {
		return err
	}

	if onUnavailErr := i.onUnavailable(ctx); onUnavailErr != nil {
		slog.ErrorContext(ctx,
			"on unavailable handler failed",
			"method", method,
			"error", onUnavailErr,
		)
	}

	return err
}
