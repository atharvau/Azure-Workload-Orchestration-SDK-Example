# Azure Workload Orchestration SDK Collection

This repository contains implementations of the Azure Workload Orchestration SDK in multiple programming languages, providing a consistent interface for managing Azure workload orchestration resources across different technology stacks.

## Available Implementations

| Language | Status | Key Features | Min Version Required |
|----------|---------|--------------|---------------------|
| [Java](./java/README.md) | ✅ Production Ready | Advanced logging, Comprehensive error handling | Java 11+ |
| [Python](./python/README.md) | ✅ Production Ready | Simple API, Easy setup | Python 3.9+ |
| [Go](./golang/README.md) | ✅ Production Ready | High performance, Strong typing | Go 1.18+ |

## Common Features

All implementations provide:
- Schema creation and management
- Schema version control
- Solution template management
- Target operations
- Error handling with retries
- Comprehensive logging
- Azure authentication integration

## Architecture

The SDKs follow a common architecture pattern, as detailed in our [Architecture Documentation](./architecture.md):

```mermaid
graph TB
    subgraph Client Application
        WOD[WorkloadOrchestrationDemo]
        Config[Configuration Manager]
        Metrics[Metrics Collector]
    end

    subgraph Core Services
        RM[Resource Manager]
        EH[Error Handler]
        RC[Resource Cleanup]
    end

    subgraph Azure Services
        Azure[Azure Workload Orchestration]
    end

    Client Application --> Core Services
    Core Services --> Azure Services
```

## Quick Start

1. Choose your preferred language implementation
2. Follow the language-specific README for setup instructions
3. Configure Azure credentials
4. Run the example code

## Implementation Comparison

### Java Implementation
- Best for: Enterprise applications
- Features comprehensive logging with SLF4J
- Includes robust error handling and retries
- Maven-based build system

### Python Implementation
- Best for: Quick prototypes and scripts
- Simple, straightforward API
- Easy to set up and run
- Great for automation tasks

### Go Implementation
- Best for: High-performance applications
- Strong type safety
- Excellent concurrency support
- Minimal external dependencies

## Common Patterns

All implementations follow these common patterns:

1. **Authentication**
   - DefaultAzureCredential usage
   - Secure credential management

2. **Error Handling**
   - Retry mechanisms for transient failures
   - Proper error propagation
   - Detailed error logging

3. **Resource Management**
   - Proper cleanup of resources
   - Version management
   - State tracking

4. **Configuration**
   - External configuration files
   - Override capabilities

## Getting Started

1. **Prerequisites**
   - Azure subscription
   - Azure CLI installed and authenticated
   - Language-specific tools installed

2. **Choose Implementation**
   - Select the appropriate language SDK
   - Follow language-specific README
   - Run example code

## Contributing

1. Choose the appropriate language implementation
2. Follow the common README template
3. Ensure all tests pass
4. Submit a pull request

## Testing Strategy

Each implementation includes:
- Unit tests
- Integration tests
- Performance benchmarks
- Mock services for testing

## Support

- File issues in the GitHub repository
- Check language-specific READMEs for known issues
- Reference the architecture documentation for design questions

## License

All implementations are licensed under the MIT License - see individual LICENSE files for details.