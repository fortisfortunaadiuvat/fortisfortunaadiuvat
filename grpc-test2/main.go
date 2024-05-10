package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbGen "proto_gen"
)

func main() {
	// Define command-line flags
	serverAddr := flag.String("server", "10.41.76.196:50051", "Address and port of the gRPC server")
	flag.Parse()

	// Establish a connection to the gRPC server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	// Create a gRPC client
	client := pbGen.NewTestappClient(conn)

	// Create a context
	ctx := context.Background()

	var (
		avgDur time.Duration
		minDur = time.Duration(math.MaxInt64)
		maxDur = time.Duration(math.MinInt64)
	)

	// Loop for 100 iterations
	for i := 0; i < 100; i++ {
		// Record the start time
		t0 := time.Now()

		// Call the gRPC method
		_, err := client.Ping(ctx, &pbGen.TestappPingRequest{
			Content: "Hello, Tokopedia",
		})
		if err != nil {
			fmt.Printf("Error calling Ping: %v\n", err)
			return
		}

		// Calculate the duration and update min, max, and average
		d0 := time.Since(t0)
		if minDur > d0 {
			minDur = d0
		}
		if maxDur < d0 {
			maxDur = d0
		}
		avgDur += d0
	}

	// Calculate average duration in microseconds
	avgDurMicro := avgDur / 100

	// Print the results
	fmt.Printf("Average Duration: %d microseconds\n", avgDurMicro.Microseconds())
	fmt.Printf("Minimum Duration: %d microseconds\n", minDur.Microseconds())
	fmt.Printf("Maximum Duration: %d microseconds\n", maxDur.Microseconds())
}

