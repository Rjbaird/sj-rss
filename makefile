include .env

# ==========================================
## Makefile for the project
# ==========================================

# ==========================================
## BUILD
# ==========================================

## build/app: Build the cmd/app binary
.PHONY: build/app
build/app:
	@echo "Building cmd/app"
	go build -o ./bin/app ./cmd/app


# ==========================================
## RUN
# ==========================================

## run/app: Run the cmd/app binary
.PHONY: run/app
run/app:
	@echo "Running cmd/app"
	go run ./cmd/app