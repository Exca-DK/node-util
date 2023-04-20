package v4

func DefaultConfig() Config {
	return Config{
		NeighbourAmount:  BucketSize,
		ResponseEnr:      true,
		ResponseFindNode: true,
	}
}

type Config struct {
	NeighbourAmount int

	ResponseEnr      bool
	ResponseFindNode bool
}

func (c *Config) validate() {
	// either of those cases is invalid in disc protocol
	// fallback to default behaviour
	if c.NeighbourAmount < 0 || c.NeighbourAmount > 16 {
		c.NeighbourAmount = BucketSize
	}
}
