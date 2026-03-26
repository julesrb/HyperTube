.PHONY: up up-vpn down build logs api stream frontend

up:
	docker compose up --build

up-vpn:
	docker compose --profile vpn up --build

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

api:
	cd services/api && go run .

stream:
	cd services/torrent-stream && go run .

frontend:
	cd frontend && npm run dev

tidy:
	cd services/api && go mod tidy
	cd services/torrent-stream && go mod tidy
