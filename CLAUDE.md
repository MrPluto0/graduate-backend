# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go backend system for network device management and task scheduling, implementing a Lyapunov-based resource allocation algorithm for distributed task scheduling across communication devices. The system manages network topology, tasks, and real-time scheduling with performance metrics tracking.

## Development Commands

### Build and Run

```bash
# Install dependencies
go mod tidy

# Run the application
go run cmd/server/main.go

# Build binary
go build -o ./tmp/main.exe ./cmd/server

# Hot reload with Air (recommended for development)
air
```

### Swagger Documentation

```bash
# Generate Swagger docs (run after API changes)
swag init -g cmd/server/main.go -o ./docs

# Access Swagger UI
# http://localhost:8080/swagger/index.html
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/algorithm/...
```

## Architecture Overview

### Core Algorithm Engine (`internal/algorithm/`)

The heart of the system is a **Lyapunov drift-plus-penalty** based task scheduling algorithm that operates in discrete time slots.

**Key Components:**

1. **System** (`system.go`): Singleton managing the entire scheduling system
   - Loads network topology from database
   - Runs scheduling loop (1 second per time slot)
   - Coordinates between Graph, TaskManager, and State

2. **Graph** (`graph.go`): Network topology and routing
   - Floyd-Warshall algorithm for shortest path calculation
   - Runs once at initialization (static topology)
   - Path computation for task data transfer

3. **TaskManager** (`task_manager.go`): Task lifecycle management
   - Manages task submission, tracking, and completion
   - Maps: `Tasks` (ID→Task), `UserTasks` (UserID→TaskIDs)
   - Active task filtering by status

4. **State** (`state.go`): Scheduling state snapshots
   - Contains `TaskSnapshots` for all active tasks in current time slot
   - Computes system-level metrics (delay, energy, queue lengths)
   - Used for optimization cost calculation

**Data Flow:**
```
User submits task → TaskManager → System starts loop →
Each time slot: Create State → Graph.schedule() iterates to find best assignment →
Update Task.MetricsHistory → Check completion → Next slot
```

**Scheduling Logic:**
- Each iteration tries random task-to-comm assignments
- Uses greedy randomized algorithm with queue-based resource allocation
- Cost function: `Drift + V × Penalty` where penalty = weighted sum of delay, energy, load
- Keeps best state across multiple iterations (early stopping when cost stabilizes)

### Data Model Separation (`internal/algorithm/define/`)

**Critical Design Pattern:**
- **Task**: Persistent object with full history (`MetricsHistory[]`, status, assignment)
- **TaskSnapshot**: Temporary compute context for one scheduling iteration (queues, resource fraction)
- **Never mix** persistent fields into TaskSnapshot or vice versa

### API Layer (`internal/api/`)

Standard layered architecture:
- **handlers/**: HTTP request handlers (thin layer)
- **middleware/**: JWT auth, logging, CORS
- **routes.go**: Route registration with middleware chain

All endpoints under `/api/v1/` with JWT authentication (`Authorization: Bearer <token>`)

### Data Layer

- **Repository** (`internal/repository/`): Data access abstraction
- **Service** (`internal/service/`): Business logic layer
- **Models** (`internal/models/`): Database models (GORM)
- **Storage**: SQLite database (`data.db`)

## Critical Implementation Rules

### Concurrency Safety

The algorithm runs in a goroutine with ticker-based scheduling. Always use:
```go
s.mutex.Lock()   // Write operations
s.mutex.RLock()  // Read operations
```

**Race Condition Risks:**
- `TaskManager.Tasks` accessed from multiple goroutines
- `System.CurrentState` read by API handlers during scheduling
- Always lock before accessing shared state

### Algorithm Constants (`internal/algorithm/constant/`)

All magic numbers must be defined here:
- `Slot`: Time slot duration (seconds)
- `Iters`: Max scheduling iterations per time slot
- `Rho`, `C`, `Kappa`: Task computation parameters
- `V`: Lyapunov penalty weight

### Performance Metrics Calculation

Metrics are computed at **TaskSnapshot** level then aggregated:

**Transfer delay:**
```go
delay = Σ(data_transferred / link_speed) for each path segment
```

**Compute delay:**
```go
delay = processed_data × Rho / (resource_fraction × C)
```

**Energy:** Similar formulas with power consumption parameters

**Always** update `Task.MetricsHistory` after each time slot via `TaskManager.syncFromState()`

### Task Status Lifecycle

```
TaskPending → TaskQueued → TaskComputing → TaskCompleted
                                ↓
                           TaskFailed
```

Status transitions happen in `TaskManager.syncFromState()` based on queue and processing state.

## Code Organization Principles

From `.github/copilot-instructions.md`:
- Keep code simple and readable
- Avoid over-engineering
- Comments only for core logic, not obvious code
- Use latest Go idioms

### Naming Conventions

- **Exported**: CamelCase (e.g., `NewSystem`, `TaskManager`)
- **Unexported**: camelCase for internal functions (e.g., `addTask`, `syncFromState`)
- **Constants**: UPPER_SNAKE_CASE in constant package

### Error Handling

Always return errors up the stack. Log at the point of handling, not at every return:
```go
func (s *System) SubmitTask(req TaskBase) (*Task, error) {
    if _, exists := s.UserMap[req.UserID]; !exists {
        return nil, fmt.Errorf("用户不存在: %d", req.UserID)  // Clear error message
    }
    // ...
}
```

## Known Issues and Limitations

From `improvement.md`:

1. **No dynamic topology updates**: Graph.Floyd runs once at initialization
2. **Deep copy overhead**: State copying in every iteration can be optimized
3. **No task priorities**: All tasks treated equally (no SLA/deadline support)
4. **No task cancellation**: Cannot stop running tasks
5. **No retry mechanism**: Failed tasks are not automatically retried

## Testing Guidelines

When writing tests:
- Use table-driven tests for multiple scenarios
- Mock repository layer for service tests
- Test concurrency with `go test -race`
- Test algorithm correctness with small synthetic workloads

## API Development

When adding new endpoints:
1. Define handler in `internal/api/handlers/`
2. Add route in `internal/api/routes.go`
3. Add Swagger annotations (`// @Summary`, `// @Param`, etc.)
4. Run `swag init` to regenerate docs
5. Test with Swagger UI or curl

**Authentication:**
All `/api/v1/*` routes (except `/auth/login`) require JWT:
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/...
```

## Database Migrations

Currently no formal migration system. Schema changes require:
1. Update model in `internal/models/`
2. GORM auto-migration in `pkg/database/connection.go`
3. Test with fresh `data.db`

## Configuration

Edit `configs/config.yaml` for:
- Server port (default: 8080)
- JWT secret and expiration
- Database connection (currently unused, using SQLite)

**JWT Secret:** Change from default in production!

## Debugging Tips

**Algorithm not running?**
- Check `System.IsRunning` in `/api/v1/algorithm/info`
- Verify tasks exist with `/api/v1/algorithm/tasks`
- Check logs for "所有数据处理完成"

**Scheduling slow?**
- Reduce `constant.Iters` (fewer random trials per slot)
- Check number of active tasks and comm devices (complexity: O(Iters × Tasks × Comms))

**Task stuck in pending?**
- Ensure algorithm is started (submit task triggers automatic start)
- Check user exists in `System.UserMap`
- Verify network topology loaded (`System.Graph` initialized)

## Architecture Diagrams

See [docs/architecture.md](docs/architecture.md) for detailed system architecture, data flow diagrams, and API endpoint listing.
