# Azure Workload Orchestration SDK - Go Implementation

This implementation demonstrates the usage of Azure Workload Orchestration SDK in Go, providing functionality for schema management, solution templates, and target operations.

## Prerequisites

- Go 1.18+
- Azure subscription
- Azure CLI installed and authenticated

## Project Structure

```
golang/
├── main.go              # Main application entry point
├── go.mod              # Go module definition
├── go.sum              # Module dependency checksums
├── version.txt         # SDK version information
└── README.md           # This documentation
```

## Features

- Azure Workload Orchestration SDK integration using `azure-sdk-for-go`
- Schema creation and management
- Schema version control
- Solution template management
- Target operations
- Error handling with retries
- Comprehensive logging
- Strong type safety and compile-time checks
- Concurrent operation support
- Resource cleanup management

## Setup and Installation

1. **Install Dependencies**
   ```bash
   go mod download
   ```

   Or let Go automatically install them:
   ```bash
   go get github.com/Azure/azure-sdk-for-go/sdk/azidentity
   go get github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/workloadorchestration/armworkloadorchestration
   ```

## Configuration

Update these values in your code or configuration:
```go
const (
    subscriptionID = "your-subscription-id"
    resourceGroup  = "your-resource-group"
    location       = "eastus"
)
```

## Usage Examples

### Basic Authentication and Client Setup

```go
import (
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/workloadorchestration/armworkloadorchestration"
)

func main() {
    cred, err := azidentity.NewDefaultAzureCredential(nil)
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }

    client, err := armworkloadorchestration.NewClient(subscriptionID, cred, nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
}
```

### Creating a Schema

```go
func createSchema(ctx context.Context, client *armworkloadorchestration.Client) (*armworkloadorchestration.Schema, error) {
    schema := &armworkloadorchestration.Schema{
        Properties: &armworkloadorchestration.SchemaProperties{
            Description: to.StringPtr("Test Schema"),
            Version:    to.StringPtr("1.0.0"),
        },
    }

    result, err := client.CreateOrUpdate(ctx, resourceGroup, "test-schema", schema, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create schema: %v", err)
    }

    return &result.Schema, nil
}
```

### Managing Solution Templates

```go
func createSolutionTemplate(ctx context.Context, client *armworkloadorchestration.Client) error {
    template := &armworkloadorchestration.SolutionTemplate{
        Properties: &armworkloadorchestration.SolutionTemplateProperties{
            SchemaReference: &armworkloadorchestration.Reference{
                Name:    to.StringPtr("test-schema"),
                Version: to.StringPtr("1.0.0"),
            },
        },
    }

    _, err := client.CreateOrUpdateTemplate(ctx, resourceGroup, "test-template", template, nil)
    return err
}
```

## Error Handling

The implementation includes robust error handling:

```go
func withRetry(ctx context.Context, op func() error) error {
    maxRetries := 3
    backoff := time.Second

    for i := 0; i < maxRetries; i++ {
        err := op()
        if err == nil {
            return nil
        }

        if !isRetryableError(err) {
            return err
        }

        if i < maxRetries-1 {
            time.Sleep(backoff)
            backoff *= 2 // Exponential backoff
        }
    }
    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

## Best Practices

1. **Error Handling**
   - Use explicit error checking
   - Implement retry mechanisms for transient failures
   - Provide detailed error context

2. **Resource Management**
   - Use `defer` for cleanup
   - Implement proper context handling
   - Close connections and resources properly

3. **Performance**
   - Use goroutines for concurrent operations
   - Implement connection pooling
   - Cache frequently used data

4. **Security**
   - Use secure credential management
   - Implement proper logging (avoid sensitive data)
   - Use HTTPS for all connections

## Common Issues and Solutions

1. **Authentication Failures**
   - Verify Azure role assignments
   - Check credential expiration
   - Ensure proper permissions

2. **Resource Creation Failures**
   - Verify resource name uniqueness
   - Check resource quota limits
   - Validate input parameters

3. **Network Issues**
   - Implement proper timeouts
   - Use connection pooling
   - Add retry logic for transient failures

## Testing

The implementation includes:

```go
func TestSchemaCreation(t *testing.T) {
    // Test code examples
}
```

Run tests using:
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Run tests: `go test ./...`
4. Commit your changes
5. Push to the branch
6. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.