package common

import "time"

type CoolDown struct {
	MinInterval time.Duration
	lastTicked  time.Time
}

func (c *CoolDown) Tick() {
	defer func() {
		c.lastTicked = time.Now()
	}()
	if c.lastTicked.IsZero() {
		return
	}
	time.Sleep(time.Now().Sub(c.lastTicked))
}
