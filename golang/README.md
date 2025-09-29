# Azure Workload Orchestration Go SDK Example

This Go application demonstrates an end-to-end workflow for using the Azure Workload Orchestration service. It utilizes the `azure-sdk-for-go` to create and manage various Workload Orchestration resources, including contexts, schemas, solution templates, and targets.

The application performs the following key operations:
1.  **Context Management**: Fetches an existing Azure Context, adds a new randomly generated capability, and updates the context.
2.  **Schema Creation**: Creates a new Schema and a corresponding Schema Version.
3.  **Solution Template**: Creates a Solution Template and a version for it, linking it to the created schema.
4.  **Target Creation**: Deploys a Target resource.
5.  **Configuration**: Sets configuration values for the target by making a direct REST API call.
6.  **Deployment Workflow**: Reviews, publishes, and installs the solution on the target.

## Prerequisites

- Go (version 1.18 or later)
- An active Azure Subscription.
- Azure credentials configured for authentication. The application uses `azidentity.NewDefaultAzureCredential`, which supports multiple authentication methods. The recommended approach is to set the following environment variables:
  - `AZURE_CLIENT_ID`: Your application's client ID.
  - `AZURE_TENANT_ID`: Your Azure Active Directory tenant ID.
  - `AZURE_CLIENT_SECRET`: Your application's client secret.

  Alternatively, you can authenticate by logging in via Azure CLI (`az login`).

## Configuration

Before running the application, you must update the hardcoded constants in `main.go` to match your Azure environment:

```go
const (
	LOCATION               = "eastus2euap"
	SUBSCRIPTION_ID        = "YOUR_SUBSCRIPTION_ID" // Replace with your Subscription ID
	RESOURCE_GROUP         = "sdkexamples"
	CONTEXT_RESOURCE_GROUP = "Mehoopany"
	CONTEXT_NAME           = "Mehoopany-Context"
	SINGLE_CAPABILITY_NAME = "sdkexamples-soap"
)
```

## How to Run

1.  **Navigate to the directory**:
    ```sh
    cd golang
    ```

2.  **Install dependencies**:
    This command will download the necessary Azure SDK modules defined in `go.mod`.
    ```sh
    go mod tidy
    ```

3.  **Run the application**:
    ```sh
    go run main.go
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
