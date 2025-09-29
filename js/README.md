# Azure Workload Orchestration JavaScript SDK Example

This Node.js application demonstrates an end-to-end workflow for using the Azure Workload Orchestration service. It utilizes the `@azure/arm-workloadorchestration` SDK to create and manage various Workload Orchestration resources, including contexts, schemas, solution templates, and targets.

The application performs the following key operations:
1.  **Context Management**: Fetches an existing Azure Context, adds a new randomly generated capability, and updates the context.
2.  **Schema Creation**: Creates a new Schema and a corresponding Schema Version.
3.  **Solution Template**: Creates a Solution Template and a version for it, linking it to the created schema.
4.  **Target Creation**: Deploys a Target resource.
5.  **Configuration**: Sets configuration values for the target by making a direct REST API call using `axios`.
6.  **Deployment Workflow**: Reviews, publishes, and installs the solution on the target.

## Prerequisites

- Node.js (version 14 or later)
- npm (Node Package Manager)
- An active Azure Subscription.
- Azure credentials configured for authentication. The application uses `DefaultAzureCredential`, which supports multiple authentication methods. The recommended approach is to set the following environment variables:
  - `AZURE_CLIENT_ID`: Your application's client ID.
  - `AZURE_TENANT_ID`: Your Azure Active Directory tenant ID.
  - `AZURE_CLIENT_SECRET`: Your application's client secret.
  - `AZURE_SUBSCRIPTION_ID`: Your Azure Subscription ID.

  Alternatively, you can authenticate by logging in via Azure CLI (`az login`).

## Configuration

The application is configured via environment variables. You can create a `.env` file in the `js` directory to store these values. The most important variable is `AZURE_SUBSCRIPTION_ID`.

**Example `.env` file:**
```
AZURE_SUBSCRIPTION_ID="your-subscription-id"
AZURE_CLIENT_ID="your-client-id"
AZURE_TENANT_ID="your-tenant-id"
AZURE_CLIENT_SECRET="your-client-secret"
```

If `AZURE_SUBSCRIPTION_ID` is not set, the script will fall back to the hardcoded value in `main.js`.

## How to Run

1.  **Navigate to the directory**:
    ```sh
    cd js
    ```

2.  **Install dependencies**:
    This command will download the necessary packages defined in `package.json`.
    ```sh
    npm install
    ```

3.  **Run the application**:
    ```sh
    node main.js
    ```

## Sample Output

Below is a sample output from a successful run of the application.

```
PS C:\Users\audapure\Projects\ConfigManager\SDK\sdktester\js> node .\main.js
Successfully authenticated with Azure.

==================================================
STEP 1: Managing Azure Context
==================================================
DEBUG: Fetching existing context: Mehoopany-Context
DEBUG: Found 372 existing capabilities.
DEBUG: Generated single random capability: sdkexamples-soap-7773
Capability merge complete. Total unique capabilities: 373
Creating/updating context 'Mehoopany-Context'...
Context management completed successfully: Mehoopany-Context

===> FINAL CAPABILITY FOR THIS RUN: sdkexamples-soap-7773

Waiting 30 seconds after capability selection...

==================================================
STEP 2: Creating Azure Resources
==================================================
Creating schema 'sdkexamples-schema-v9.16.82'...
Schema created successfully: sdkexamples-schema-v9.16.82
Creating schema version '0.0.82' for schema 'sdkexamples-schema-v9.16.82'...
Schema version created successfully: 0.0.82
Creating solution template 'sdkexamples-solution1'...
Solution template created successfully: sdkexamples-solution1
Creating solution template version '6.1.9'...
Solution template version created successfully with ID: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/solutionTemplates/sdkexamples-solution1/versions/6.1.9
Creating target 'sdkbox-m23'...
Target created successfully: sdkbox-m23

==================================================
STEP 3: Setting Configuration via API
==================================================
Making PUT call to Configuration API...
Configuration API PUT call successful. Status: 200

Verifying configuration...
Making GET call to Configuration API: https://management.azure.com/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-m23Config/DynamicConfigurations/sdkexamples-solution1/versions/version1?api-version=2024-06-01-preview
Configuration GET call successful. Status: 200
Retrieved Configuration Data: {
  "id": "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-m23Config/DynamicConfigurations/sdkexamples-solution1/versions/version1",
  "name": "version1",
  "type": "microsoft.edge/configurations/dynamicconfigurations/versions",
  "systemData": {
    "createdBy": "cba491bc-48c0-44a6-a6c7-23362a7f54a9",
    "createdByType": "Application",
    "createdAt": "2025-09-11T03:54:39.3730928Z",
    "lastModifiedBy": "audapure@microsoft.com",
    "lastModifiedByType": "User",
    "lastModifiedAt": "2025-09-26T04:22:33.548318Z"
  },
  "properties": {
    "values": "ErrorThreshold: 35.3
HealthCheckEndpoint: http://localhost:8080/health
EnableLocalLog: true
AgentEndpoint: http://localhost:8080/agent
HealthCheckEnabled: true
ApplicationEndpoint: http://localhost:8080/app
TemperatureRangeMax: 100.5
",
    "provisioningState": "Succeeded"
  }
}

==================================================
STEP 4: Review, Publish, and Install
==================================================
Starting review for target sdkbox-m23 with template version ID: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/solutionTemplates/sdkexamples-solution1/versions/6.1.9
Listing all solution versions for solution 'sdkexamples-solution1' on target 'sdkbox-m23'...
------------------------------------
Found matching solution version: sdkbox-m23-6.1.9.1
  Extracted Review ID: 1a1be17a-a0c2-433f-9b8c-c09146f41ed4
  Revision: undefined
  State: InReview
Returning ID for further steps: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution1/versions/sdkbox-m23-6.1.9.1
Publishing solution version to target sdkbox-m23...
Publish operation completed successfully.
Installing solution on target sdkbox-m23...
Install operation completed for target sdkbox-m23.

==================================================
Workflow finished successfully!
==================================================
```
