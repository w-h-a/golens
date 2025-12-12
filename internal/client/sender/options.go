package sender

import "context"

type Option func(*Options)

type Options struct {
	BaseURL string
	Context context.Context
}

func WithBaseURL(url string) Option {
	return func(o *Options) {
		o.BaseURL = url
	}
}

func NewOptions(opts ...Option) Options {
	options := Options{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}

type SendOption func(*SendOptions)

type SendOptions struct {
	Context context.Context
}

func NewSendOptions(opts ...SendOption) SendOptions {
	options := SendOptions{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}
