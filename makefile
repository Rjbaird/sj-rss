include .env

# ==========================================
## Makefile for the project
# ==========================================

# ==========================================
## BUILD
# ==========================================

## build/views: Build html, css and js files
.PHONY: build/views
build/views:
	@echo "Building views"
	cd views && npm run build


## build/app: Build the cmd/app binary
.PHONY: build/app
build/app: build/views
	@echo "Building cmd/app"
	go build -o ./bin/app ./cmd/app


# ==========================================
## RUN
# ==========================================

## run/views: Run npm commands for html, css and js files during development
.PHONY: run/views
run/views:
	@echo "Running views"
	cd views && npm run dev

## run/app: Run the cmd/app binary
.PHONY: run/app
run/app:
	@echo "Running cmd/app"
	go run ./cmd/app


# ==========================================
## DEVELOPMENT
# ==========================================

## dev: Run the cmd/app binary and npm commands for html, css and js files during development
.PHONY: dev
dev:
	@echo "Running dev"
	@make -j 2 run/views run/app