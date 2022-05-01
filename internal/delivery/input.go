package delivery

import "makarov.dev/bot/internal/delivery/web"

func StartAllInputs() {
	startAllBackgroundJobs()
	go web.StartWeb()
}
