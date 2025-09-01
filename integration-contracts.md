# Integration Contracts between Java and Golang Components

## Message Bus Protocol

```protobuf
syntax = "proto3";

package workloadorch;

// Command message from Java orchestrator to Golang scheduler
message OrchestratorCommand {
    string command_id = 1;
    uint64 timestamp = 2;
    oneof command {
        ScheduleWorkload schedule = 3;
        UpdateTarget target = 4;
        ModifyContext context = 5;
    }
}

// High-performance workload scheduling request
message ScheduleWorkload {
    string workload_id = 1;
    string target_id = 2;
    bytes configuration = 3;  // Protocol buffers encoded config
    uint32 priority = 4;
    uint64 deadline_ms = 5;
}

// Target state updates
message UpdateTarget {
    string target_id = 1;
    TargetState state = 2;
    map<string, string> capabilities = 3;
    ResourceLimits resources = 4;
}

// Context modifications
message ModifyContext {
    string context_id = 1;
    repeated string hierarchies = 2;
    map<string, string> properties = 3;
}

// Resource limits for targets
message ResourceLimits {
    uint32 cpu_millicores = 1;
    uint64 memory_bytes = 2;
    uint32 max_workloads = 3;
}

enum TargetState {
    UNKNOWN = 0;
    READY = 1;
    BUSY = 2;
    MAINTENANCE = 3;
    ERROR = 4;
}

// Performance metrics from Golang back to Java
message PerformanceMetrics {
    string component_id = 1;
    uint64 timestamp = 2;
    repeated Metric metrics = 3;
}

message Metric {
    string name = 1;
    double value = 2;
    string unit = 3;
}
```

## State Store Schema

```sql
-- Shared state schema for both Java and Golang components

-- Workload definitions (managed by Java)
CREATE TABLE workload_definitions (
    id VARCHAR(64) PRIMARY KEY,
    version VARCHAR(32) NOT NULL,
    schema_id VARCHAR(64) NOT NULL,
    configuration JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Target states (accessed by both, primarily Golang)
CREATE TABLE target_states (
    id VARCHAR(64) PRIMARY KEY,
    state VARCHAR(32) NOT NULL,
    capabilities JSONB NOT NULL,
    current_workloads INT NOT NULL,
    resource_usage JSONB NOT NULL,
    last_heartbeat TIMESTAMP NOT NULL,
    last_updated TIMESTAMP NOT NULL
);

-- Execution contexts (managed by Golang)
CREATE TABLE execution_contexts (
    id VARCHAR(64) PRIMARY KEY,
    hierarchies JSONB NOT NULL,
    properties JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Performance metrics (written by Golang, read by Java)
CREATE TABLE performance_metrics (
    component_id VARCHAR(64) NOT NULL,
    metric_name VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(32) NOT NULL,
    PRIMARY KEY (component_id, metric_name, timestamp)
);
```

## API Endpoints

### Golang Performance Engine REST API

```yaml
openapi: 3.0.0
info:
  title: Workload Orchestration Performance Engine API
  version: 1.0.0

paths:
  /v1/scheduler/health:
    get:
      summary: Get scheduler health status
      responses:
        200:
          description: Scheduler health information
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthStatus'

  /v1/scheduler/metrics:
    get:
      summary: Get real-time scheduler metrics
      responses:
        200:
          description: Current scheduler metrics
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SchedulerMetrics'

  /v1/targets/{targetId}/status:
    get:
      summary: Get target status
      parameters:
        - name: targetId
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: Target status information
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TargetStatus'

components:
  schemas:
    HealthStatus:
      type: object
      properties:
        status:
          type: string
          enum: [HEALTHY, DEGRADED, UNHEALTHY]
        queue_depth:
          type: integer
        active_workers:
          type: integer
        error_rate:
          type: number

    SchedulerMetrics:
      type: object
      properties:
        jobs_per_second:
          type: integer
        avg_latency_ms:
          type: number
        queue_depth:
          type: integer
        error_count:
          type: integer
        active_targets:
          type: integer

    TargetStatus:
      type: object
      properties:
        target_id:
          type: string
        state:
          type: string
        current_workloads:
          type: integer
        resource_usage:
          type: object
          properties:
            cpu_usage:
              type: number
            memory_usage:
              type: number
        last_heartbeat:
          type: string
          format: date-time
```

## Error Handling

1. **Transient Failures**
   - Retry with exponential backoff
   - Maximum retry attempts: 3
   - Base delay: 100ms
   - Max delay: 1s

2. **Circuit Breaking**
   - Error threshold: 50% of requests
   - Minimum requests: 20
   - Window size: 10s
   - Half-open after: 5s

3. **Error Codes**
   ```
   1000-1999: Job Scheduling Errors
   2000-2999: Target Management Errors
   3000-3999: Context Handling Errors
   4000-4999: Resource Errors
   5000-5999: System Errors
   ```

## Performance SLAs

1. **Job Scheduling**
   - Latency: < 1ms (p99)
   - Throughput: 10,000+ jobs/second
   - Error rate: < 0.01%

2. **State Updates**
   - Latency: < 10ms (p99)
   - Consistency delay: < 100ms

3. **Metrics Collection**
   - Collection interval: 1s
   - Aggregation delay: < 5s