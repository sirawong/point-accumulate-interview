# Configuration variables
CSV_FOLDER ?= ./csv_files
UPLOAD_URL ?= http://localhost:8080/api/v1/point/accumulate/upload
FILE_KEY ?= csv_files
OUTPUT_FILE ?= output.zip
SCRIPT_UPLOAD = ./scripts/upload.sh
SCRIPT_CLEAN_DATA = ./scripts/clear_customers.sh

.PHONY: upload up build down clean test

# Docker compose commands
up:
	docker-compose up -d

build:
	docker-compose up -d --build

up-mongo:
	docker-compose up mongo -d

down:
	docker-compose down

# Make script executable and upload CSV files
upload:
	@chmod +x $(SCRIPT_UPLOAD)
	@./$(SCRIPT_UPLOAD) "$(CSV_FOLDER)" "$(UPLOAD_URL)" "$(FILE_KEY)" "$(OUTPUT_FILE)"

clean:
	@echo "Clearing customers collection in pointdb (connecting to Docker)..."
	@chmod +x $(SCRIPT_CLEAN_DATA)
	@bash ./$(SCRIPT_CLEAN_DATA)
	@echo "✅ Done."

test:
	@go test ./... -v
