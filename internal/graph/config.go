package graph

type Config struct {
	Builder   Builder
	Meta      Enricher
	Follow    string
	SkipEnv   []string
	Proto     NetProto
	OnlyLocal bool
	FullInfo  bool
	NoLoops   bool
}

func (c *Config) MatchName(v string) (yes bool) {
	return c.Follow == "" || v == c.Follow
}

func (c *Config) MatchProto(v string) (yes bool) {
	return c.Proto == ALL || v == c.Proto.String()
}
