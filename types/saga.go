package types

import "context"

type Fn func(ctx context.Context, payload PayloadInput) error

type ConsumerInput interface {
	UpFn(ctx context.Context, payload PayloadInput) error
	DownFn(ctx context.Context, payload PayloadInput) error
	GetConfig() Opts
	GetEventName() string
}
