build:
	docker build -t cr.yandex/crpbjrij9hsbb1q5fj6s/bot .

push:
	docker push cr.yandex/crpbjrij9hsbb1q5fj6s/bot

deploy: build push