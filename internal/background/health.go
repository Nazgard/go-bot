package background

import (
	"context"
	"makarov.dev/bot/internal/config"
	"time"
)

type healthBackgroundJob struct {
	ctx context.Context
}

func newHealthBackgroundJob(ctx context.Context) *healthBackgroundJob {
	return &healthBackgroundJob{ctx: ctx}
}

func (h *healthBackgroundJob) Start() {
	log := config.GetLogger()
	for {
		select {
		case <-h.ctx.Done():
			log.Infof("Health background job stopped")
			return
		default:
			log.Debug("Health ok")
			time.Sleep(1 * time.Hour)
		}
	}
}
