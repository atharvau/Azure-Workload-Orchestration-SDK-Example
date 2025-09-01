# Azure Workload Orchestration SDK - Python Implementation

This implementation demonstrates the usage of Azure Workload Orchestration SDK in Python, providing a simple yet powerful interface for managing Azure workload orchestration resources with proper error handling and logging.

## Prerequisites

- Python 3.9+
- Azure subscription
- Azure CLI installed and authenticated
- pip (Python package manager)

## Project Structure

```
python/
├── main.py              # Main application entry point
├── test_targets.py      # Test implementations
├── requirements.txt     # Project dependencies
├── version.txt         # SDK version information
└── README.md           # This documentation
```

## Features

- Azure Workload Orchestration SDK integration using `azure-mgmt-workloadorchestration`
- Schema creation and management
- Schema version control
- Solution template management
- Target operations
- Automatic error handling and retries
- Comprehensive logging
- Async operation support
- Type hints for better code reliability
- Integration with Python's async/await syntax

## Setup and Installation

```bash
pip install -r requirements.txt
```

## Configuration

Create a configuration file or use environment variables:

```python
class Config:
    SUBSCRIPTION_ID = os.getenv("AZURE_SUBSCRIPTION_ID")
    RESOURCE_GROUP = "your-resource-group"
    LOCATION = "eastus"
    
    # Retry configuration
    MAX_RETRIES = 3
    RETRY_DELAY = 1  # seconds
    MAX_DELAY = 10   # seconds
```

## Usage Examples

### Basic Client Setup

```python
from azure.identity import DefaultAzureCredential
from azure.mgmt.workloadorchestration import WorkloadOrchestrationClient

class WorkloadOrchestrationManager:
    def __init__(self):
        self.credential = DefaultAzureCredential()
        self.client = WorkloadOrchestrationClient(
            credential=self.credential,
            subscription_id=Config.SUBSCRIPTION_ID
        )
```

### Creating a Schema

```python
async def create_schema(self, name: str, version: str) -> Schema:
    """Create a new schema with specified name and version.
    
    Args:
        name: Schema name
        version: Schema version
        
    Returns:
        Schema: Created schema object
        
    Raises:
        SchemaCreationError: If schema creation fails
    """
    try:
        schema_properties = {
            "description": "Test Schema",
            "version": version
        }
        
        poller = await self.client.schemas.begin_create_or_update(
            resource_group_name=Config.RESOURCE_GROUP,
            schema_name=name,
            parameters={"properties": schema_properties}
        )
        
        return await poller.result()
    except Exception as e:
        raise SchemaCreationError(f"Failed to create schema: {str(e)}")
```

### Managing Solution Templates

```python
async def create_solution_template(
    self, 
    name: str, 
    schema_ref: dict
) -> SolutionTemplate:
    """Create a new solution template.
    
    Args:
        name: Template name
        schema_ref: Schema reference dictionary
        
    Returns:
        SolutionTemplate: Created template object
    """
    template_properties = {
        "schemaReference": schema_ref,
        "description": "Test Template"
    }
    
    async with self.client.solution_templates as templates:
        return await templates.create_or_update(
            resource_group_name=Config.RESOURCE_GROUP,
            template_name=name,
            parameters={"properties": template_properties}
        )
```

## Error Handling

Implement robust error handling with retries:

```python
from typing import TypeVar, Callable, Awaitable
import asyncio
from tenacity import retry, stop_after_attempt, wait_exponential

T = TypeVar("T")

def with_retry(
    func: Callable[..., Awaitable[T]]
) -> Callable[..., Awaitable[T]]:
    """Decorator for retrying async operations with exponential backoff."""
    
    @retry(
        stop=stop_after_attempt(Config.MAX_RETRIES),
        wait=wait_exponential(
            multiplier=Config.RETRY_DELAY,
            max=Config.MAX_DELAY
        )
    )
    async def wrapper(*args, **kwargs) -> T:
        try:
            return await func(*args, **kwargs)
        except Exception as e:
            if not is_retryable_error(e):
                raise
            raise
    return wrapper
```

## Logging

Configure logging with Python's built-in logging module:

```python
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(),
        logging.FileHandler('workload_orch.log')
    ]
)

logger = logging.getLogger(__name__)
```

## Best Practices

1. **Type Hints**
   - Use type hints for better code maintainability
   - Enable mypy for static type checking
   - Document parameter and return types

2. **Async/Await**
   - Use async/await for I/O-bound operations
   - Implement proper connection pooling
   - Handle async context managers correctly

3. **Error Handling**
   - Use custom exceptions for business logic
   - Implement proper retry mechanisms
   - Log errors with context

4. **Resource Management**
   - Use context managers (`with` statements)
   - Implement proper cleanup
   - Handle connection pooling

## Common Issues and Solutions

1. **Authentication Failures**
   - Check Azure CLI authentication
   - Validate role assignments
   - Verify service principal permissions

2. **Resource Creation Failures**
   - Validate input parameters
   - Check resource name uniqueness
   - Verify resource quotas

3. **Network Issues**
   - Implement proper timeouts
   - Use connection pooling
   - Add retry logic

## Testing

Run tests using pytest:

```bash
# Install test dependencies
pip install pytest pytest-asyncio pytest-cov

# Run tests
pytest

# Run with coverage
pytest --cov=. tests/
```

Example test:
```python
import pytest
from unittest.mock import Mock, patch

@pytest.mark.asyncio
async def test_schema_creation():
    manager = WorkloadOrchestrationManager()
    schema = await manager.create_schema("test-schema", "1.0.0")
    assert schema.name == "test-schema"
    assert schema.properties.version == "1.0.0"
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Run tests: `pytest`
4. Implement type hints
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.