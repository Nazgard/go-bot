package background

import (
	"context"
	"makarov.dev/bot/internal/integration/telegram"
)

type telegramBackgroundJob struct {
	ctx context.Context
}

func newTelegramBackgroundJob(ctx context.Context) *telegramBackgroundJob {
	return &telegramBackgroundJob{ctx: ctx}
}

func (t *telegramBackgroundJob) Start() {
	telegram.Start(t.ctx)
}
