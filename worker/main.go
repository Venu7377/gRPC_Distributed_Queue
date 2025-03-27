package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Worker struct {
	id          string
	isAvailable bool
	mu          sync.Mutex // to access safely
}

func (w *Worker) SetAvailability(available bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.isAvailable = available
}

func (w *Worker) IsAvailable() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.isAvailable
}

func (w *Worker) run() {
	// retry to handle connection issues
	for {
		err := w.connectAndWork()
		if err != nil {
			log.Printf("Worker %s encountered error: %v. Reconnecting in 5 seconds...", w.id, err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}

func (w *Worker) connectAndWork() error {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTaskCoordinatorClient(conn)
	stream, err := client.RegisterWorker(context.Background())
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	// send initial worker status with its id to register
	if err := stream.Send(&pb.WorkerStatus{WorkerId: w.id, Available: true}); err != nil {
		return fmt.Errorf("failed to send initial worker status: %v", err)
	}
	log.Printf("Worker %s registered with server", w.id)

	// send worker status updates (heartbeat) every 500ms in a separate goroutine when it is available
	go func() {
		lastSentAvailable := w.IsAvailable()
		for {
			currentAvailable := w.IsAvailable()
			if currentAvailable != lastSentAvailable { // only send to stream on status change
				if err := stream.Send(&pb.WorkerStatus{WorkerId: w.id, Available: currentAvailable}); err != nil {
					log.Printf("Worker %s failed to send status: %v", w.id, err)
					return
				}
				lastSentAvailable = currentAvailable
				log.Printf("Worker %s sent availability update: %v", w.id, currentAvailable)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Receive and process tasks
	for {
		task, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream closed: %v", err)
		}

		w.SetAvailability(false)
		log.Printf("Worker %s processing task %s", w.id, task.TaskId)

		result := doSum(w, task)

		reportTaskCompletion(client, w, task, result)
	}
}

func main() {
	workerID := fmt.Sprintf("Worker%04d", 1000+rand.Intn(9000))
	w := &Worker{id: workerID}
	w.SetAvailability(true)
	log.Printf("Starting %s", workerID)
	w.run()
}

func doSum(worker *Worker, task *pb.TaskRequest) int32 {
	// just simulating long processing using sleep
	// sleepTime := 10 * time.Second // 10s
	// log.Printf("Worker %s will take %v to complete task %s", worker.id, sleepTime, task.TaskId)
	// time.Sleep(sleepTime)
	log.Printf("Worker %s processing task", worker.id)
	return task.Num1 + task.Num2
}

func reportTaskCompletion(client pb.TaskCoordinatorClient, w *Worker, task *pb.TaskRequest, result int32) {
	// Report task completion with WorkerId
	_, err := client.TaskCompleted(context.Background(), &pb.TaskCompletion{
		TaskId:   task.TaskId,
		WorkerId: w.id,
		Success:  true,
		Result:   result,
	})
	if err != nil {
		log.Printf("Worker %s failed to report completion for task %s: %v", w.id, task.TaskId, err)
	}

	w.SetAvailability(true)
	log.Printf("Worker %s completed task %s (Result : %d) and is now available", w.id, task.TaskId, result)
}
