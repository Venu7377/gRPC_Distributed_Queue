# gRPC Distributed Queue

## Simulation Steps

1. Start the coordination server.
2. Start the required number of workers.
3. Simulate tasks using the client by setting the `taskCount` variable to specify the number of tasks to initiate.

**Note:** Steps 2 and 3 can be performed in any order.If no workers are available, tasks will be queued and will be dispatched when a worker or workers become ready.
