package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mattcarp12/transit-pulse/backend/internal/transit"
)

var (
	transitClient *transit.Client
	staticStops   map[string]transit.Stop
	staticTrips   map[string]transit.Trip
	staticShapes  map[string][]transit.ShapePoint
	s3Client      *s3.Client
	bucketName    string
	isLocal       bool
)

func init() {
	log.Println("COLD START: Initializing Container...")

	// 1. Check if we are running locally
	isLocal = os.Getenv("LOCAL_MODE") == "true"

	bucketName = os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" && !isLocal {
		log.Fatal("CRITICAL ERROR: S3_BUCKET_NAME is not set")
	}

	// 2. Only initialize the AWS SDK if we are NOT in local mode
	if !isLocal {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatalf("Failed to load AWS config: %v", err)
		}
		s3Client = s3.NewFromConfig(cfg)
	}

	// 3. Initialize the Transit Client and fetch static data
	transitClient = transit.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Downloading static GTFS schedule into memory...")
	staticData, err := transitClient.FetchStaticData(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch static data: %v", err)
	}
	staticShapes = staticData.Shapes
	staticStops = staticData.Stops
	staticTrips = staticData.Trips

	if !isLocal {
		log.Println("Uploading static route_shapes.json to S3...")
		shapesBytes, err := json.Marshal(staticShapes)
		if err == nil {
			_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
				Bucket:      aws.String(bucketName),
				Key:         aws.String("route_shapes.json"),
				Body:        bytes.NewReader(shapesBytes),
				ContentType: aws.String("application/json"),
				//Cache for 24 hours
				CacheControl: aws.String("max-age=86400"),
			})
			if err != nil {
				log.Printf("Warning: Failed to upload route shapes: %v", err)
			} else {
				log.Println("Successfully uploaded route_shapes.json")
			}
		}
	}

	log.Printf("Cold start complete. Loaded %d stops, %d trips.", len(staticStops), len(staticTrips))
}

func HandleRequest(ctx context.Context) error {
	log.Println("Invocation started: Fetching live GTFS-RT feed...")

	feed, err := transitClient.FetchTripUpdates(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch live data: %v", err)
	}

	networkState := transit.BuildNetworkState(feed, staticStops, staticTrips)

	payloadBytes, err := json.MarshalIndent(networkState, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// If we're running locally, just print the JSON to the console instead of uploading to S3
	if isLocal {
		log.Println("Running in LOCAL MODE. Outputting JSON to console:")
		fmt.Println(string(payloadBytes))
		return nil
	}

	// PRODUCTION: Upload to S3
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(bucketName),
		Key:          aws.String("live_data.json"),
		Body:         bytes.NewReader(payloadBytes),
		ContentType:  aws.String("application/json"),
		CacheControl: aws.String("max-age=15"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %v", err)
	}

	log.Println("Successfully uploaded live_data.json to S3.")
	return nil
}

func main() {
	// If local, manually call the handler instead of waiting for Lambda to invoke it
	if isLocal {
		log.Println("Starting local execution...")
		if err := HandleRequest(context.Background()); err != nil {
			log.Fatalf("Local execution failed: %v", err)
		}
		log.Println("Local execution completed successfully.")
		return
	}

	// In production, Lambda will invoke HandleRequest
	lambda.Start(HandleRequest)
}
