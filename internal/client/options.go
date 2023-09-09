package client

type Option func(*options)

type options struct {
	Create  createClient
	Nsenter nsEnter
	Mode    mode
}

func WithMode(m mode) Option {
	return func(o *options) {
		o.Mode = m
	}
}

func WithClientCreator(c createClient) Option {
	return func(o *options) {
		o.Create = c
	}
}

func WithNsEnter(e nsEnter) Option {
	return func(o *options) {
		o.Nsenter = e
	}
}
