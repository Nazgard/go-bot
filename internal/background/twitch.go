package background

import (
	"context"
	"makarov.dev/bot/internal/integration/twitch"
)

type twitchBackgroundJob struct {
	ctx context.Context
}

func newTwitchBackgroundJob(ctx context.Context) *twitchBackgroundJob {
	return &twitchBackgroundJob{ctx: ctx}
}

func (t *twitchBackgroundJob) Start() {
	twitch.Start(t.ctx)
}
