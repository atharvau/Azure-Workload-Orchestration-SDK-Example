# Azure Workload Orchestration SDK - Java Implementation

This implementation demonstrates the usage of Azure Workload Orchestration SDK in Java, providing robust functionality for schema management, solution templates, and workload orchestration with comprehensive error handling and logging.

## Prerequisites

- Java 11 or later
- Maven 3.6 or later
- Azure subscription with appropriate permissions
- Azure CLI installed and authenticated

## Project Structure

```
java/
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── com/
│   │   │       └── example/
│   │   │           └── workloadorch/
│   │   │               ├── SolutionTemplateBuilder.java
│   │   │               ├── ValidationUtils.java
│   │   │               ├── VersionManager.java
│   │   │               └── WorkloadOrchestrationDemo.java
│   │   └── resources/
│   │       └── logback.xml
│   └── test/
│       └── java/
│           └── com/
│               └── example/
│                   └── workloadorch/
│                       ├── ValidationTests.java
│                       └── SolutionTemplateBuilderTests.java
├── pom.xml
├── run.sh
└── README.md
```

## Features

- Azure Workload Orchestration SDK integration
- Schema creation and management
- Schema version control
- Solution template management
- Target operations
- Robust error handling with configurable retries
- Comprehensive logging using SLF4J with Logback
- Unit and integration testing
- Maven-based build system
- Automated deployment scripts

## Setup and Installation

1. **Build Project**
   ```bash
   mvn clean package
   ```

2. **Run Tests**
   ```bash
   mvn test
   ```

## Configuration

### Application Configuration

1. Update constants in `WorkloadOrchestrationDemo.java`:
   ```java
   private static final String SUBSCRIPTION_ID = "<your-subscription-id>";
   private static final String RESOURCE_GROUP = "rgconfigurationmanager";
   private static final String SCHEMA_NAME = "testSchema";
   private static final String SCHEMA_VERSION = "1.0.0";
   ```

2. Logging Configuration (`logback.xml`):
   ```xml
   <configuration>
       <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
           <!-- Console logging configuration -->
       </appender>
       <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
           <!-- File logging configuration -->
       </appender>
   </configuration>
   ```

## Usage Examples

### Basic Client Setup

```java
public class WorkloadOrchestrationDemo {
    private static final Duration TIMEOUT = Duration.ofSeconds(30);
    private final WorkloadOrchestrationClient client;

    public WorkloadOrchestrationDemo() {
        DefaultAzureCredential credential = new DefaultAzureCredentialBuilder().build();
        client = new WorkloadOrchestrationClientBuilder()
            .credential(credential)
            .subscriptionId(SUBSCRIPTION_ID)
            .buildClient();
    }
}
```

### Creating a Schema

```java
public Schema createSchema(String schemaName, String version) {
    SchemaProperties properties = new SchemaProperties()
        .setDescription("Test Schema")
        .setVersion(version);
    
    return client.getSchemas()
        .define(schemaName)
        .withRegion(REGION)
        .withProperties(properties)
        .create();
}
```

### Managing Solution Templates

```java
public SolutionTemplate createTemplate(String templateName, Schema schema) {
    SolutionTemplateProperties properties = new SolutionTemplateProperties()
        .setSchemaReference(new Reference()
            .setName(schema.name())
            .setVersion(schema.properties().version()));
    
    return client.getSolutionTemplates()
        .define(templateName)
        .withProperties(properties)
        .create();
}
```

## Error Handling

The implementation includes comprehensive error handling:

```java
public class RetryableOperation<T> {
    private static final int MAX_RETRIES = 3;
    private static final Duration INITIAL_DELAY = Duration.ofSeconds(1);

    public T execute(Supplier<T> operation) {
        int attempts = 0;
        Duration delay = INITIAL_DELAY;

        while (true) {
            try {
                return operation.get();
            } catch (Exception e) {
                if (!isRetryable(e) || attempts >= MAX_RETRIES) {
                    throw e;
                }
                sleep(delay);
                delay = delay.multipliedBy(2); // Exponential backoff
                attempts++;
            }
        }
    }
}
```

## Logging

Comprehensive logging is implemented using SLF4J with Logback:

- **ERROR**: Critical failures requiring immediate attention
- **WARN**: Recoverable issues and retry attempts
- **INFO**: Normal operation events
- **DEBUG**: Detailed troubleshooting information

Example log output:
```
2025-09-01 09:15:23,456 [main] INFO  c.e.w.WorkloadOrchestrationDemo - Initializing client
2025-09-01 09:15:24,789 [main] DEBUG c.e.w.SolutionTemplateBuilder - Creating template: test-template
2025-09-01 09:15:25,123 [main] WARN  c.e.w.RetryableOperation - Retrying operation after transient failure
```

## Best Practices

1. **Authentication**
   - Use managed identities in production
   - Implement proper credential rotation
   - Handle token expiration gracefully

2. **Error Handling**
   - Implement retry mechanisms for transient failures
   - Log errors with context
   - Use custom exceptions for business logic

3. **Resource Management**
   - Implement proper resource cleanup
   - Use try-with-resources for AutoCloseable resources
   - Handle concurrent access properly

4. **Testing**
   - Write unit tests for business logic
   - Implement integration tests
   - Use mocks for external services

## Common Issues and Solutions

1. **Authentication Failures**
   - Verify Azure role assignments
   - Ensure proper network access
   - Check service principal permissions

2. **Resource Creation Failures**
   - Validate input parameters
   - Check resource name uniqueness
   - Verify resource quotas

3. **Connection Issues**
   - Implement proper timeouts
   - Use connection pooling
   - Add retry logic

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
mvn test

# Run specific test class
mvn test -Dtest=ValidationTests

# Run with coverage
mvn clean test jacoco:report
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `mvn clean test`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.