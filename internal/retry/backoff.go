package retry

import (
	"context"
	"time"
)


type Config struct {
	MaxRetries int
	InitialInterval time.Duration
	MaxInterval time.Duration
	Multiplier float64
}


func DefaultConfig() Config {
	return Config{
		MaxRetries:      10,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      1.5,
	}
}


func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error
	interval := cfg.InitialInterval

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == cfg.MaxRetries {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		interval = time.Duration(float64(interval) * cfg.Multiplier)
		if interval > cfg.MaxInterval {
			interval = cfg.MaxInterval
		}
	}

	return lastErr
}


func Poll(ctx context.Context, interval time.Duration, timeout time.Duration, fn func() (bool, error)) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	done, err := fn()
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := fn()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}
