package ratelimit

// Bucket is a token bucket that refills by Refill tokens per tick.
type Bucket struct {
	Capacity, Refill, tokens int
}

func New(capacity, refill int) *Bucket {
	return &Bucket{Capacity: capacity, Refill: refill, tokens: capacity}
}

// Allow takes one token if one is left.
func (b *Bucket) Allow() bool {
	if b.tokens == 0 {
		return false
	}
	b.tokens--
	return true
}

// Tick adds Refill tokens, up to Capacity.
func (b *Bucket) Tick() {
	b.tokens = min(b.Capacity, b.tokens+b.Refill)
}
