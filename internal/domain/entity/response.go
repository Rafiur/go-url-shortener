package entity

import "time"

type Response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	RateLimitReset time.Duration `json:"rate_limit_reset"`
}
