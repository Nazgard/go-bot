package background

import "context"

// Job интерфейс для всех периодических задач
type Job interface {
	// Start блокирует текущую горутину. Выполняет какую-либо логику в заданный период
	Start()
}

var jobs = make([]Job, 0)

// StartAllBackgroundJobs запускает все периодические задачи. Не блокирует текущую горутину
func StartAllBackgroundJobs(ctx context.Context) {
	appendJobs(ctx)
	for _, job := range jobs {
		go job.Start()
	}
}

func appendJobs(ctx context.Context) {
	kz := newKinozalBackgroundJob(ctx)
	jobs = append(jobs, kz)

	lf := newLostFilmBackgroundJob(ctx)
	jobs = append(jobs, lf)

	tg := newTelegramBackgroundJob(ctx)
	jobs = append(jobs, tg)

	h := newHealthBackgroundJob(ctx)
	jobs = append(jobs, h)

	t := newTwitchBackgroundJob(ctx)
	jobs = append(jobs, t)
}
