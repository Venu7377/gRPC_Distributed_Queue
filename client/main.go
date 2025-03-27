package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "grpc/proto"
)

var (
	grpcClient  pb.TaskCoordinatorClient
	taskCounter uint64 //  counter for task ids
)

func main() {
	// Connect to GRPC server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()
	grpcClient = pb.NewTaskCoordinatorClient(conn)
	if grpcClient == nil {
		log.Fatalf("grpcClient is nil")
	}
	taskCount := 1000 // Number of tasks to simulate and submit
	for range taskCount {
		// Submit tasks to the coordinator
		submitTaskHandler()
	}

}

// Handler to submit tasks
func submitTaskHandler() {
	// Task ID like: T1 ,T2 ...
	taskID := fmt.Sprintf("T%d", atomic.AddUint64(&taskCounter, 1))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now().UnixNano()
	num1 := int32(now % 100)
	num2 := int32((now / 100) % 100)

	resp, err := grpcClient.SubmitTask(ctx, &pb.TaskRequest{
		TaskId: taskID,
		Num1:   num1,
		Num2:   num2,
	})
	if err != nil {
		log.Printf("Failed to submit task %s: %v", taskID, err)
		return
	}

	log.Printf("Task %s Submitted. Message: %s", taskID, resp.Message)
}
