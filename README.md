# Point Accumulation Service

A Go-based service that processes customer purchase records and calculates accumulated points based on predefined rules for different store branches and product categories.

## Overview

This service processes daily CSV files containing customer purchase records and generates point summary reports. Each branch can have different point accumulation rules based on product categories and purchase amounts.

## Features

- Process multiple CSV files containing purchase records
- Calculate points based on configurable rules (ratio, fixed points, percentage)
- Generate daily point summary reports sorted by highest points
- REST API for file uploads
- MongoDB storage for rules and customer data
- Docker containerization

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Make (convenience commands)

### Setup

1. Clone the repository
2. Start the services:
   ```bash
   make up
   # or
   docker-compose up -d
   ```

3. The service will be available at `http://localhost:8080`

### Upload Purchase Records

#### Using Make
Place your CSV files in the `csv_files` directory and run:
```bash
make upload
```
This command executes a script that sends all CSV files to the service for processing.

## Input Format

CSV files should be named in date format (`yyyy-mm-dd.csv`) and contain:

```csv
customer_id,product_id,category_id,category_name,branch_id,purchased_amount,currency
U000001,123123,CT1001,BEVERAGE,BR3456,220,THB
U000002,123124,CT1001,BEVERAGE,BR3444,400,THB
```

## Output

The service returns a ZIP file containing the point summary report:

- **Response Format**: ZIP file containing CSV report
- **File Location**: Saved as `output.zip` in project root directory
- **Report Content**: Point summary CSV with customer data

Example of the summary CSV content inside the ZIP file:

```csv
customer_id,points,last_purchase_date
U000001,27,2025-01-02
U000003,4,2025-01-02
U000002,4,2025-01-01
```

Results are sorted by:
1. Highest points (descending)
2. Latest purchase date (descending) for ties

## Point Calculation Rules

The system supports three types of rules:

- **RATIO**: Points per amount spent (e.g., 1 point per 100 THB)
- **FIXED_POINT**: Fixed points for minimum purchase (e.g., 2 points if > 200 THB)
- **PERCENTAGE**: Points as percentage of amount (e.g., 5% of purchase amount)

Default rules are automatically loaded into MongoDB on startup.

## Project Structure

```
├── cmd/api/          # Application entry point
├── internal/         # Internal application code
├── pkg/              # Shared packages
├── csv_files/        # Input CSV files
└── scripts/          # Utility scripts
```

## Commands

- `make up` - Start services
- `make down` - Stop services
- `make upload` - Upload CSV files from csv_files directory

## Configuration

Environment variables are configured in `.env`:

- `HTTP_SERVER_PORT`: API server port (default: 8080)
- `MONGO_URI`: MongoDB connection string
- `FILE_PATH`: Output file path pattern

## API Endpoints

- `POST /api/v1/point/accumulate/upload` - Upload CSV files for processing
