# This file exists for people unfamiliar with `Taskfile.yml`.
# Here are duplicated the main commands from there.

run:
	go run ./cmd/trading-robot

up:
	docker-compose -f ./deploy/docker-compose.yml up --build

down:
	docker-compose -f ./deploy/docker-compose.yml down \
	--rmi local \
	--volumes \
	--remove-orphans \
	--timeout 60; \
	docker-compose -f ./deploy/docker-compose.yml rm -f
