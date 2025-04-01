package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	pb "grpc/proto"

	"github.com/redis/go-redis/v9"

	"slices"

	"google.golang.org/grpc"
)

// Define a Redis client in the coordinatorServer struct
type coordinatorServer struct {
	pb.UnimplementedTaskCoordinatorServer
	mu              sync.Mutex
	workers         []string
	workerChannels  map[string]chan *pb.TaskRequest
	workerAvailable map[string]bool
	redisClient     *redis.Client // Redis client for task queue to store tasks in Redis so tasks can be retained and continue processing in case of worker or coordinator server down
	cond            *sync.Cond
	lastWorkerIdx   int
}

// Initialize Redis client
func newCoordinatorServer() *coordinatorServer {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "PaS5w0rD",
	})

	return &coordinatorServer{
		workers:         []string{},
		workerChannels:  make(map[string]chan *pb.TaskRequest),
		workerAvailable: make(map[string]bool),
		redisClient:     rdb,
	}
}

func removeWorker(workers []string, workerID string) []string {
	for i, id := range workers {
		if id == workerID {
			return slices.Delete(workers, i, i+1)
		}
	}
	return workers
}

func (s *coordinatorServer) SubmitTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	// Check if task is already in the queue
	exists, err := s.redisClient.LPos(ctx, "taskQueue", req.TaskId, redis.LPosArgs{}).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if exists != 0 {
		return &pb.TaskResponse{Success: false, Message: "Task already exists in the queue"}, nil
	}
	// Convert task to JSON for storing in Redis
	taskJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Store task in Redis list (LPUSH or RPUSH based on FIFO/LIFO preference)
	err = s.redisClient.RPush(ctx, "taskQueue", taskJSON).Err()
	if err != nil {
		return nil, err
	}

	// Signal that a new task is available
	s.cond.Broadcast()

	return &pb.TaskResponse{Success: true, Message: "Task queued"}, nil
}

func (s *coordinatorServer) RegisterWorker(stream pb.TaskCoordinator_RegisterWorkerServer) error {
	var workerID string
	taskChan := make(chan *pb.TaskRequest) // Unbuffered channel

	// First message from the worker to register
	status, err := stream.Recv()
	if err != nil {
		log.Printf("Failed to receive worker status: %v", err)
		return err
	}

	s.mu.Lock()
	workerID = status.WorkerId
	s.workers = append(s.workers, workerID)
	s.workerChannels[workerID] = taskChan
	s.workerAvailable[workerID] = true // Worker starts as available intially
	log.Printf("Worker %s registered", workerID)
	s.mu.Unlock()

	// Signal that a new worker is available
	s.cond.Broadcast()

	// Handle incoming worker status sent by it when it is available
	go func() {
		for {
			status, err := stream.Recv()
			if err != nil {
				log.Printf("Worker %s disconnected: %v", workerID, err)
				s.mu.Lock()
				delete(s.workerChannels, workerID)
				delete(s.workerAvailable, workerID)
				s.workers = removeWorker(s.workers, workerID)
				s.mu.Unlock()
				close(taskChan)
				s.cond.Broadcast() // Signal worker removal
				return
			}
			s.mu.Lock()
			s.workerAvailable[workerID] = status.Available
			s.mu.Unlock()
			s.cond.Broadcast() // Signal availability change
		}
	}()

	// Send tasks to worker from its channel
	for task := range taskChan {
		if err := stream.Send(task); err != nil {
			return err
		}
	}
	return nil
}

func (s *coordinatorServer) TaskCompleted(ctx context.Context, req *pb.TaskCompletion) (*pb.TaskResponse, error) {
	log.Printf("Task %s completed with success: %v. Task Result is %d", req.TaskId, req.Success, req.Result)
	s.mu.Lock()
	if _, exists := s.workerAvailable[req.WorkerId]; exists {
		s.workerAvailable[req.WorkerId] = true // Mark worker as available again after task completion
		log.Printf("Worker %s is now available for new tasks", req.WorkerId)
	}
	s.mu.Unlock()

	// Signal that a worker is free
	s.cond.Broadcast()

	return &pb.TaskResponse{Success: true, Message: "Completion noted"}, nil
}

func (s *coordinatorServer) dispatchTasks() {
	go func() {
		ctx := context.Background()
		for {
			s.mu.Lock()
			for s.redisClient.LLen(ctx, "taskQueue").Val() == 0 || !s.hasAvailableWorker() {
				s.cond.Wait()
			}

			// Find an available worker
			workerID := s.getNextAvailableWorker()
			if workerID == "" {
				s.mu.Unlock()
				continue
			}

			// Fetch the next task from Redis
			taskJSON, err := s.redisClient.LPop(ctx, "taskQueue").Result()
			if err != nil {
				s.mu.Unlock()
				log.Println("Error retrieving task from Redis:", err)
				continue
			}

			// Convert JSON to TaskRequest
			var task pb.TaskRequest
			if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
				s.mu.Unlock()
				log.Println("Error unmarshaling task:", err)
				continue
			}

			// Send task to worker
			taskChan := s.workerChannels[workerID]
			s.workerAvailable[workerID] = false
			s.mu.Unlock()

			select {
			case taskChan <- &task:
				log.Printf("Dispatched task %s (Num1: %d, Num2: %d) to worker %s", task.TaskId, task.GetNum1(), task.GetNum2(), workerID)
			default:
				// Worker channel full, requeue task in Redis
				s.mu.Lock()
				requeueTask, _ := json.Marshal(&pb.TaskRequest{
					TaskId: task.TaskId,
					Num1:   task.Num1,
					Num2:   task.Num2,
				})
				s.redisClient.RPush(ctx, "taskQueue", requeueTask)
				s.workerAvailable[workerID] = true
				s.mu.Unlock()
				log.Printf("Worker %s channel full, task %s requeued", workerID, task.TaskId)
			}
		}
	}()
}

// Helper func to check if any worker is available
func (s *coordinatorServer) hasAvailableWorker() bool {
	for _, available := range s.workerAvailable {
		if available {
			return true
		}
	}
	return false
}

// Helper func to get the next available worker in round robin order
func (s *coordinatorServer) getNextAvailableWorker() string {
	if len(s.workers) == 0 {
		return ""
	}

	maxRetries := 3                         // Number of retries before giving up
	retryInterval := 100 * time.Millisecond // Time between retries

	for range maxRetries {
		for range s.workers {
			s.lastWorkerIdx = (s.lastWorkerIdx + 1) % len(s.workers)
			workerID := s.workers[s.lastWorkerIdx]
			if s.workerAvailable[workerID] {
				return workerID
			}
		}
		// No available worker found, wait before retrying
		time.Sleep(retryInterval)
	}

	return "" //  return empty string if no worker is available after retries
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := newCoordinatorServer()
	server.cond = sync.NewCond(&server.mu) // initialize condition variable

	pb.RegisterTaskCoordinatorServer(grpcServer, server)

	// Start dispatching tasks in seperate goroutine if tasks and workers available else it will dispatch when tasks and workers are available
	server.dispatchTasks()

	log.Println("Coordinator server is running on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
