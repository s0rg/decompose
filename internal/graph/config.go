package graph

import "github.com/s0rg/set"

type Config struct {
	Builder   Builder
	Meta      Enricher
	Follow    set.Unordered[string]
	SkipEnv   []string
	Proto     NetProto
	OnlyLocal bool
	NoLoops   bool
	Deep      bool
}

func (c *Config) MatchName(v string) (yes bool) {
	return c.Follow.Len() == 0 || c.Follow.Has(v)
}

func (c *Config) MatchProto(v string) (yes bool) {
	return c.Proto == ALL || v == c.Proto.String()
}
