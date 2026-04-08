package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	S3BucketName       string
	S3Endpoint         string
	Port               string
}

func loadConfig() Config {
	cfg := Config{
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          getEnvOrDefault("AWS_REGION", "us-east-1"),
		S3BucketName:       os.Getenv("S3_BUCKET_NAME"),
		S3Endpoint:         os.Getenv("S3_ENDPOINT"),
		Port:               getEnvOrDefault("PORT", "8080"),
	}

	if cfg.AWSAccessKeyID == "" || cfg.AWSSecretAccessKey == "" || cfg.S3BucketName == "" {
		log.Fatal("Missing required env vars: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, S3_BUCKET_NAME")
	}

	return cfg
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func newS3Client(cfg Config) *s3.Client {
	customEndpoint := os.Getenv("AWS_ENDPOINT")

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if customEndpoint != "" {
			o.BaseEndpoint = aws.String(customEndpoint)
			// CRITICAL FIX: Tell LocalStack to use Path-Style
			o.UsePathStyle = true
		}
	})
}

// uploadHandler handles POST /upload
// Accepts multipart/form-data with a "file" field.
// Optionally accepts a "key" field to specify the S3 object key;
// defaults to <timestamp>_<original_filename>.
func uploadHandler(s3Client *s3.Client, bucketName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 32 MB max in-memory, rest spills to disk
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing 'file' field: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Determine the S3 key
		key := r.FormValue("key")
		if key == "" {
			key = fmt.Sprintf("%d_%s", time.Now().UnixMilli(), filepath.Base(header.Filename))
		}

		_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(key),
			Body:          file,
			ContentLength: aws.Int64(header.Size),
			ContentType:   aws.String(header.Header.Get("Content-Type")),
		})
		if err != nil {
			http.Error(w, "Failed to upload to S3: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","key":%q,"bucket":%q}`, key, bucketName)
	}
}

// healthHandler handles GET /health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"status":"healthy"}`)
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	cfg := loadConfig()
	s3Client := newS3Client(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler(s3Client, cfg.S3BucketName))
	mux.HandleFunc("/health", healthHandler)

	addr := ":" + cfg.Port
	endpointInfo := cfg.AWSRegion
	if cfg.S3Endpoint != "" {
		endpointInfo += fmt.Sprintf(", endpoint: %s", cfg.S3Endpoint)
	}
	log.Printf("Server listening on %s  (bucket: %s, %s)", addr, cfg.S3BucketName, endpointInfo)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
