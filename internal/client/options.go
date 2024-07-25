package client

type Option func(*options)

type options struct {
	Create  createClient
	Nsenter nsenter
	Inodes  inodes
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

func WithNsenterFn(f nsenter) Option {
	return func(o *options) {
		o.Nsenter = f
	}
}

func WithInodesFn(f inodes) Option {
	return func(o *options) {
		o.Inodes = f
	}
}
