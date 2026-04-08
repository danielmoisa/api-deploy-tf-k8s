# API DevOps - S3 File Upload Service

A Go-based HTTP server that provides file upload functionality to Amazon S3 (or LocalStack for local development). The service includes REST API endpoints, AWS S3 integration with custom endpoint support, and infrastructure-as-code deployment with Terraform.

## Features

- **REST API** with POST `/upload` endpoint for multipart file uploads
- **Health Check** endpoint at GET `/health`
- **AWS S3 Integration** using AWS SDK v2
- **LocalStack Support** for local development and testing without AWS credentials
- **Custom S3 Endpoints** for flexible deployment scenarios
- **Infrastructure as Code** with Terraform for automated deployment
- **Environment-based Configuration** via `.env` file

## Prerequisites

- Go 1.24 or later
- Docker (for LocalStack)
- AWS credentials (for production) or LocalStack (for development)

## Setup

### Local Development with LocalStack

1. **Start LocalStack:**
   ```bash
   docker run -p 4566:4566 localstack/localstack
   ```

2. **Create S3 bucket in LocalStack:**
   ```bash
   aws s3 mb s3://my-bucket --endpoint-url http://localhost:4566
   ```

3. **Configure environment variables:**
   Copy `.env.example` to `.env` and update with your LocalStack settings:
   ```bash
   cp app/.env.example app/.env
   ```

4. **Run the application:**
   ```bash
   cd app
   go run main.go
   ```

### Production with AWS

1. **Set up AWS credentials:**
   - Configure AWS credentials in `~/.aws/credentials` or use IAM roles

2. **Configure environment variables in `.env`:**
   ```
   AWS_ACCESS_KEY_ID=your-access-key
   AWS_SECRET_ACCESS_KEY=your-secret-key
   S3_BUCKET_NAME=your-bucket-name
   AWS_REGION=us-east-1
   # Remove or omit S3_ENDPOINT for AWS
   PORT=8080
   ```

3. **Run the application:**
   ```bash
   cd app
   go run main.go
   ```

## API Endpoints

### Upload File
**POST** `/upload`

Upload a file to S3 using multipart form data.

**Request:**
```bash
curl -X POST -F "file=@path/to/file.txt" http://localhost:8080/upload
```

**Response:**
```json
{
  "status": "ok",
  "key": "1712546754123_file.txt",
  "bucket": "my-bucket"
}
```

**Optional Parameters:**
- `key`: Custom S3 object key (defaults to `<timestamp>_<filename>`)

### Health Check
**GET** `/health`

Check service health.

**Response:**
```json
{
  "status": "healthy"
}
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AWS_ACCESS_KEY_ID` | Yes | - | AWS access key (use "test" for LocalStack) |
| `AWS_SECRET_ACCESS_KEY` | Yes | - | AWS secret key (use "test" for LocalStack) |
| `S3_BUCKET_NAME` | Yes | - | S3 bucket name |
| `AWS_REGION` | No | `us-east-1` | AWS region |
| `S3_ENDPOINT` | No | - | Custom S3 endpoint (e.g., `http://localhost:4566` for LocalStack) |
| `PORT` | No | `8080` | Server port |

## Deployment

The project includes Terraform configuration for infrastructure deployment. See the Terraform files in the root directory for deployment details.

## Project Structure

```
api-devops/
├── app/                    # Go application
│   ├── main.go            # Application entry point
│   ├── go.mod             # Go module definition
│   ├── go.sum             # Go dependencies
│   ├── .env               # Environment variables (local)
│   └── .env.example       # Environment variables template
├── main.tf                # Terraform configuration
└── README.md              # This file
```
