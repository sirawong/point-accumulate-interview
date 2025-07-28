#!/bin/bash

# Default values
DEFAULT_FOLDER="./csv_files"
DEFAULT_URL="http://localhost:8080/api/v1/point/accumulate/upload"
DEFAULT_FILE_KEY="csv_files"
DEFAULT_OUTPUT_FILE="output.zip"

# Configuration
FOLDER=${1:-$DEFAULT_FOLDER}
URL=${2:-$DEFAULT_URL}
FILE_KEY=${3:-$DEFAULT_FILE_KEY}
OUTPUT_FILE=${4:-$DEFAULT_OUTPUT_FILE}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

if [ ! -d "$FOLDER" ]; then
    print_error "Folder '$FOLDER' does not exist!"
    exit 1
fi

CSV_FILES=$(find "$FOLDER" -name "*.csv" -type f)

if [ -z "$CSV_FILES" ]; then
    print_warning "No CSV files found in folder '$FOLDER'"
    exit 0
fi

FILE_COUNT=$(echo "$CSV_FILES" | wc -l)
print_info "Found $FILE_COUNT CSV files in '$FOLDER'"

CURL_CMD="curl -X POST -o \"$OUTPUT_FILE\""

while IFS= read -r file; do
    CURL_CMD="$CURL_CMD -F \"${FILE_KEY}=@${file};type=text/csv\""
    print_info "Adding file: $(basename "$file")"
done <<< "$CSV_FILES"

CURL_CMD="$CURL_CMD \"$URL\""

print_info "Uploading to: $URL"
print_info "File key: $FILE_KEY"
print_info "Saving response to: $OUTPUT_FILE"
print_info "Executing curl command..."

eval $CURL_CMD

if [ $? -eq 0 ]; then
    print_success "Upload completed successfully! Response saved to '$OUTPUT_FILE'"
else
    print_error "Upload failed!"
    exit 1
fi
