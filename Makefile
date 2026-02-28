.PHONY: build test vet lint fmt check clean serve migrate deploy

VERCEL_ENV ?= production
VERCEL_TOKEN ?=
VERCEL_AUTH :=
ifneq ($(strip $(VERCEL_TOKEN)),)
VERCEL_AUTH := --token $(VERCEL_TOKEN)
endif

build:
	go build -o bin/supost .

test:
	go test ./... -race -coverprofile=coverage.out

vet:
	go vet ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

check: fmt vet build test
	@echo "All checks passed."

serve:
	go run . serve

migrate:
	@echo "Apply migrations to your database:"
	@for f in migrations/*.sql; do echo "  psql $$DATABASE_URL -f $$f"; done

clean:
	rm -rf bin/ coverage.out

deploy:
	vercel pull --yes --environment=$(VERCEL_ENV) --cwd . $(VERCEL_AUTH)
	vercel deploy --prod --yes --cwd . $(VERCEL_AUTH)
