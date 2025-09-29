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
- Complete end-to-end workflow automation
- Schema creation and management
- Schema version control
- Solution template management
- Target operations with capability coordination
- Configuration management via REST API
- Automatic error handling and retries
- Comprehensive logging
- Azure Context management for organizational capabilities

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

## Usage

The application demonstrates a complete Azure Workload Orchestration workflow:

1. **Context Management**: Adds new capabilities to an existing Azure Context
2. **Schema Creation**: Creates schema and schema version with configuration rules  
3. **Solution Template**: Creates solution template and deployable version
4. **Target Creation**: Sets up target environment for deployments
5. **Configuration**: Sets dynamic configuration values via REST API
6. **Deployment Workflow**: Reviews, publishes, and installs the solution

The main workflow is implemented in `main.py` and runs automatically when executed.

## Running the Application

```bash
# Navigate to the python directory
cd python

# Install dependencies
pip install -r requirements.txt

# Run the application
python main.py
```


## Output

```
STEP 1: Managing Azure Context with Random Capabilities
DEBUG: Fetching existing context: Mehoopany-Context
DEBUG: Generated single random capability: sdkexamples-soap-1182
CAPABILITY MERGE PROCESS

DEBUG: PROCESSING NEW CAPABILITIES...
  ADDED NEW[0]: sdkexamples-soap-1182

DEBUG: MERGE RESULTS VALIDATION
  Initial existing count: 368
  New capabilities count: 1
  Final merged count: 369
  Unique names count: 369
VALIDATION PASSED - Proceeding with 369 capabilities
Capabilities saved to context-capabilities.json
Creating/updating context: Mehoopany-Context
Context management completed successfully: Mehoopany-Context
Waiting 30 seconds for context propagation...
Verifying capability in context...
DEBUG: Extracting capability from context result...
DEBUG: Found 369 capabilities in context
SELECTED CAPABILITY FOR ALL RESOURCES: sdkexamples-soap-1182
DEBUG: This capability will be used consistently across:
  - Solution Template
  - Target
  - All other resource operations

FINAL CAPABILITY SELECTION: sdkexamples-soap-1182
Verifying capability exists in context...
Capability sdkexamples-soap-1182 verified in context
STEP 2: Creating Azure Resources
Creating schema in resource group: sdkexamples
Schema created successfully: sdkexamples-schema-v2.12.13
Creating schema version for schema: sdkexamples-schema-v2.12.13
Schema version created successfully: 8.1.27
Proceeding with solution template and target creation...
Creating solution template in resource group: sdkexamples
Solution template created successfully: sdkexamples-solution1
Creating solution template version for template: sdkexamples-solution1
Solution template version created successfully
Successfully extracted solution template version ID: 7a8e5772-899c-4128-b3a5-80ec414e4b9f*4E0CAA57E1E3D1EE525B9EB955CC9EE2A9ECEB932427359A68A069F24C434EB1
Creating target in resource group: sdkexamples
Target provisioning completed successfully. Final provisioning state: Succeeded
Target created successfully: sdkbox-mk799jyjsdd
STEP 3: Setting Configuration Values via Configuration API
Calling Configuration API with:
  Config Name: sdkbox-mk799jyjsddConfig
  Solution Name: sdkexamples-solution1
  Version: 1.0.0
  Configuration Values:
    HealthCheckEndpoint: http://localhost:8080/health
    EnableLocalLog: true
    AgentEndpoint: http://localhost:8080/agent
    HealthCheckEnabled: true
    ApplicationEndpoint: http://localhost:8080/app
    TemperatureRangeMax: 100.5
    ErrorThreshold: 35.3

Debug: Request URL:
https://management.azure.com/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-mk799jyjsddConfig/DynamicConfigurations/sdkexamples-solution1/versions/version1?api-version=2024-06-01-preview
Making PUT call to Configuration API: https://management.azure.com/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-mk799jyjsddConfig/DynamicConfigurations/sdkexamples-solution1/versions/version1?api-version=2024-06-01-preview
Request body: {"properties":{"provisioningState":"Succeeded","values":"HealthCheckEndpoint: http://localhost:8080/health\nEnableLocalLog: true\nAgentEndpoint: http://localhost:8080/agent\nHealthCheckEnabled: true\nApplicationEndpoint: http://localhost:8080/app\nTemperatureRangeMax: 100.5\nErrorThreshold: 35.3\n"}}

Debug: Response Details:
- Status Code: 200

Debug: Response Body:
{"id":"/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-mk799jyjsddConfig/DynamicConfigurations/sdkexamples-solution1/versions/version1","name":"version1","type":"microsoft.edge/configurations/dynamicconfigurations/versions","systemData":{"createdBy":"cba491bc-48c0-44a6-a6c7-23362a7f54a9","createdByType":"Application","createdAt":"2025-09-11T03:59:03.781696Z","lastModifiedBy":"audapure@microsoft.com","lastModifiedByType":"User","lastModifiedAt":"2025-09-26T04:00:01.7664664Z"},"properties":{"values":"HealthCheckEndpoint: http://localhost:8080/health\nEnableLocalLog: true\nAgentEndpoint: http://localhost:8080/agent\nHealthCheckEnabled: true\nApplicationEndpoint: http://localhost:8080/app\nTemperatureRangeMax: 100.5\nErrorThreshold: 35.3\n","provisioningState":"Succeeded"}}
Configuration API call successful. Status: 200
Configuration API call completed successfully

STEP 3.1: Getting Configuration to verify values
Making GET call to Configuration API: https://management.azure.com/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-mk799jyjsddConfig/DynamicConfigurations/sdkexamples-solution1/versions/version1?api-version=2024-06-01-preview
Configuration GET API call successful. Status: 200
Retrieved Configuration Response: {"id":"/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-mk799jyjsddConfig/DynamicConfigurations/sdkexamples-solution1/versions/version1","name":"version1","type":"microsoft.edge/configurations/dynamicconfigurations/versions","systemData":{"createdBy":"cba491bc-48c0-44a6-a6c7-23362a7f54a9","createdByType":"Application","createdAt":"2025-09-11T03:59:03.781696Z","lastModifiedBy":"audapure@microsoft.com","lastModifiedByType":"User","lastModifiedAt":"2025-09-26T04:00:01.7664664Z"},"properties":{"values":"HealthCheckEndpoint: http://localhost:8080/health\nEnableLocalLog: true\nAgentEndpoint: http://localhost:8080/agent\nHealthCheckEnabled: true\nApplicationEndpoint: http://localhost:8080/app\nTemperatureRangeMax: 100.5\nErrorThreshold: 35.3\n","provisioningState":"Succeeded"}}
Parsed Configuration Data:
{
  "id": "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-mk799jyjsddConfig/DynamicConfigurations/sdkexamples-solution1/versions/version1",
  "name": "version1",
  "properties": {
    "provisioningState": "Succeeded",
    "values": "HealthCheckEndpoint: http://localhost:8080/health\nEnableLocalLog: true\nAgentEndpoint: http://localhost:8080/agent\nHealthCheckEnabled: true\nApplicationEndpoint: http://localhost:8080/app\nTemperatureRangeMax: 100.5\nErrorThreshold: 35.3\n"
  },
  "systemData": {
    "createdAt": "2025-09-11T03:59:03.781696Z",
    "createdBy": "cba491bc-48c0-44a6-a6c7-23362a7f54a9",
    "createdByType": "Application",
    "lastModifiedAt": "2025-09-26T04:00:01.7664664Z",
    "lastModifiedBy": "audapure@microsoft.com",
    "lastModifiedByType": "User"
  },
  "type": "microsoft.edge/configurations/dynamicconfigurations/versions"
}
Configuration Values: HealthCheckEndpoint: http://localhost:8080/health
EnableLocalLog: true
AgentEndpoint: http://localhost:8080/agent
HealthCheckEnabled: true
ApplicationEndpoint: http://localhost:8080/app
TemperatureRangeMax: 100.5
ErrorThreshold: 35.3

STEP 4: Review Target Deployment
Using solution template version ID: 7a8e5772-899c-4128-b3a5-80ec414e4b9f*4E0CAA57E1E3D1EE525B9EB955CC9EE2A9ECEB932427359A68A069F24C434EB1
Starting review for target sdkbox-mk799jyjsdd
Review completed for target sdkbox-mk799jyjsdd
STEP 5: Publish and Install Solution
The workflow has completed the following steps:
✓ Context management with capabilities
✓ Schema creation
✓ Solution template creation
✓ Target creation
✓ Configuration API calls
✓ Target review

TARGET INFORMATION:
  Name: sdkbox-mk799jyjsdd
  Resource Group: sdkexamples
  Capabilities: [sdkexamples-soap-1182]

CONFIGURATION COMPLETED:
  Config Name: sdkbox-mk799jyjsddConfig
  Solution Name: sdkexamples-solution1

Proceeding with publish and install operations...
Publishing solution version to target sdkbox-mk799jyjsdd
Publish operation completed successfully
Installing solution version on target sdkbox-mk799jyjsdd
Install operation completed successfully

WORKFLOW COMPLETED SUCCESSFULLY!
```
