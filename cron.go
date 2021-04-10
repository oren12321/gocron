package gocron

import (
    "time"
    "context"
)

type Clock struct {
    Hour, Minutes, Seconds, Nanoseconds int
}

type Job interface {
    Run(t time.Time)
}

type Cron struct {
    ctx context.Context
    interval time.Duration

    job Job
}

func NewCron(ctx context.Context, interval time.Duration, job Job) *Cron {
    return &Cron{ctx: ctx, interval: interval, job: job}
}

func (c *Cron) StartAsync() {
    go func() {
        c.Start()
    }()
}

func (c *Cron) Start() {
    for t := range cron(c.ctx, c.interval) {
        c.runJobAsync(t)
    }
}

func (c *Cron) runJobAsync(t time.Time) {
    go func() {
        c.job.Run(t)
    }()
}

func sync(t time.Time, c Clock) time.Duration {

    req := time.Date(
        t.Year(), t.Month(), t.Day(),
        c.Hour, c.Minutes, c.Seconds, c.Nanoseconds,
        t.Location())

    if diff := req.Sub(t); diff < 0 {
        req = req.AddDate(0, 0, 1)
    }

    return req.Sub(t)
}

func cron(ctx context.Context, interval time.Duration) <-chan time.Time {

    stream := make(chan time.Time, 1)

    go func() {

        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for {
            select {
            case t := <-ticker.C:
                stream <- t
            case <-ctx.Done():
                close(stream)
                return
            }
        }

    }()

    return stream
}

