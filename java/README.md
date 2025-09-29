# Azure Workload Orchestration Java SDK Example

This Java application demonstrates an end-to-end workflow for using the Azure Workload Orchestration service. It utilizes the `azure-sdk-for-java` to create and manage various Workload Orchestration resources, including contexts, schemas, solution templates, and targets.

The application performs the following key operations:
1.  **Context Management**: Fetches an existing Azure Context, adds a new randomly generated capability, and updates the context.
2.  **Schema Creation**: Creates a new Schema and a corresponding Schema Version.
3.  **Solution Template**: Creates a Solution Template and a version for it, linking it to the created schema.
4.  **Target Creation**: Deploys a Target resource.
5.  **Configuration**: Sets configuration values for the target by making a direct REST API call.
6.  **Deployment Workflow**: Reviews, publishes, and installs the solution on the target.

## Prerequisites

- Java (version 11 or later)
- Apache Maven
- An active Azure Subscription.
- Azure credentials configured for authentication. The application uses `DefaultAzureCredentialBuilder`, which supports multiple authentication methods. The recommended approach is to set the following environment variables:
  - `AZURE_CLIENT_ID`: Your application's client ID.
  - `AZURE_TENANT_ID`: Your Azure Active Directory tenant ID.
  - `AZURE_CLIENT_SECRET`: Your application's client secret.

  Alternatively, you can authenticate by logging in via Azure CLI (`az login`).

## Configuration

Before running the application, you may need to update the hardcoded constants in `WorkloadOrchestrationSample.java` to match your Azure environment, especially if you are not setting `AZURE_SUBSCRIPTION_ID` and `AZURE_TENANT_ID` as environment variables.

```java
public class WorkloadOrchestrationSample {

    // Configuration
    private static final String LOCATION = "eastus2euap";
    private static final String TENANT_ID = System.getenv().getOrDefault("AZURE_TENANT_ID", "YOUR_TENANT_ID");
    private static final String SUBSCRIPTION_ID = System.getenv().getOrDefault("AZURE_SUBSCRIPTION_ID", "YOUR_SUBSCRIPTION_ID");
    private static final String RESOURCE_GROUP = "sdkexamples";
    private static final String CONTEXT_RESOURCE_GROUP = "Mehoopany";
    private static final String CONTEXT_NAME = "Mehoopany-Context";
    private static final String SINGLE_CAPABILITY_NAME = "sdkexamples-soap";
    //...
}
```

## How to Run

1.  **Navigate to the directory**:
    ```sh
    cd java
    ```

2.  **Compile the application**:
    This command will download dependencies from `pom.xml` and compile the source code.
    ```sh
    mvn compile
    ```

3.  **Run the application**:
    ```sh
    mvn exec:java
    ```

## Sample Output

Below is a sample output from a successful run of the application.

```
PS C:\Users\audapure\Projects\ConfigManager\SDK\sdktester\java> mvn exec:java
[INFO] Scanning for projects...
[INFO]
[INFO] -------< com.azure.resourcemanager:workloadorchestration-sample >-------
[INFO] Building workloadorchestration-sample 1
[INFO]   from pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO]
[INFO] --- exec:3.5.1:java (default-cli) @ workloadorchestration-sample ---
Authenticating with Azure...
Successfully authenticated with Azure.
==================================================
STEP 1: Managing Azure Context with Random Capabilities
==================================================
Fetching existing context: Mehoopany-Context in resource group Mehoopany
Added new capability: sdkexamples-shampoo-4515
Creating/updating context 'Mehoopany-Context'...
FINAL CAPABILITY SELECTION: sdkexamples-shampoo-4515
==================================================

Waiting 30 seconds after capability selection...
Continuing with resource creation...

==================================================
STEP 2: Creating Azure Resources
==================================================
Schema created successfully: sdkexamples-schema-v6.8.52
Schema version created successfully: 3.2.22
Proceeding with solution template and target creation...

Solution template created successfully: sdkexamples-solution123
Solution template version created successfully: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/solutionTemplates/sdkexamples-solution123/versions/5.4.74
Target created successfully: sdkbox-m23
==================================================
STEP 3: Setting Configuration Values via Configuration API
==================================================
Making PUT call to Configuration API: https://management.azure.com/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-m23Config/DynamicConfigurations/sdkexamples-solution123/versions/version1?api-version=2024-06-01-preview
Request body: {
  "properties": {
    "values": "AgentEndpoint: http://localhost:8080/agent\r\nTemperatureRangeMax: 100.5\r\nApplicationEndpoint: http://localhost:8080/app\r\nHealthCheckEndpoint: http://localhost:8080/health\r\nHealthCheckEnabled: true\r\nEnableLocalLog: true\r\nErrorThreshold: 35.3\r\n",
    "provisioningState": "Succeeded"
  }
}
Configuration API call successful. Status: 200
Response: {"id":"/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-m23Config/DynamicConfigurations/sdkexamples-solution123/versions/version1","name":"version1","type":"microsoft.edge/configurations/dynamicconfigurations/versions","systemData":{"createdBy":"cba491bc-48c0-44a6-a6c7-23362a7f54a9","createdByType":"Application","createdAt":"2025-09-24T12:31:27.8673274Z","lastModifiedBy":"audapure@microsoft.com","lastModifiedByType":"User","lastModifiedAt":"2025-09-26T04:15:19.1040485Z"},"properties":{"values":"AgentEndpoint: http://localhost:8080/agent\r\nTemperatureRangeMax: 100.5\r\nApplicationEndpoint: http://localhost:8080/app\r\nHealthCheckEndpoint: http://localhost:8080/health\r\nHealthCheckEnabled: true\r\nEnableLocalLog: true\r\nErrorThreshold: 35.3\r\n","provisioningState":"Succeeded"}}

==================================================
STEP 3.1: Getting Configuration to verify values
==================================================
Making GET call to Configuration API: https://management.azure.com/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-m23Config/DynamicConfigurations/sdkexamples-solution123/versions/version1?api-version=2024-06-01-preview
Configuration GET API call successful. Status: 200
Parsed Configuration Data:
{
  "id": "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/configurations/sdkbox-m23Config/DynamicConfigurations/sdkexamples-solution123/versions/version1",
  "name": "version1",
  "type": "microsoft.edge/configurations/dynamicconfigurations/versions",
  "systemData": {
    "createdBy": "cba491bc-48c0-44a6-a6c7-23362a7f54a9",
    "createdByType": "Application",
    "createdAt": "2025-09-24T12:31:27.8673274Z",
    "lastModifiedBy": "audapure@microsoft.com",
    "lastModifiedByType": "User",
    "lastModifiedAt": "2025-09-26T04:15:19.1040485Z"
  },
  "properties": {
    "values": "AgentEndpoint: http://localhost:8080/agent\r\nTemperatureRangeMax: 100.5\r\nApplicationEndpoint: http://localhost:8080/app\r\nHealthCheckEndpoint: http://localhost:8080/health\r\nHealthCheckEnabled: true\r\nEnableLocalLog: true\r\nErrorThreshold: 35.3\r\n",
    "provisioningState": "Succeeded"
  }
}
==================================================
STEP 4: Review Target Deployment
==================================================
Starting review for target sdkbox-m23
Listing all solution versions for solution: sdkexamples-solution123
------------------------------------
Found Solution Version: sdkbox-m23-2.15.28.1
  ID: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution123/versions/sdkbox-m23-2.15.28.1
  State: InReview
  Provisioning State: Succeeded
------------------------------------
Found Solution Version: sdkbox-m23-8.11.26.1
  ID: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution123/versions/sdkbox-m23-8.11.26.1
  State: InReview
  Provisioning State: Succeeded
------------------------------------
Found Solution Version: sdkbox-m23-5.0.50.1
  ID: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution123/versions/sdkbox-m23-5.0.50.1
  State: Undeployed
  Provisioning State: Succeeded
------------------------------------
Found Solution Version: sdkbox-m23-6.6.61.1
  ID: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution123/versions/sdkbox-m23-6.6.61.1
  State: Deployed
  Provisioning State: Succeeded
------------------------------------
Found Solution Version: sdkbox-m23-5.4.74.1
  ID: /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution123/versions/sdkbox-m23-5.4.74.1
  State: InReview
  Provisioning State: Succeeded
------------------------------------
All solution versions JSON:
Found matching solution version: sdkbox-m23-5.4.74.1
Extracted reviewId: 8a162ade-c17a-4951-ac62-9192efab4b63
Revision: 1
State: InReview
==================================================
STEP 5: Publish and Install Solution
==================================================
The workflow has completed the following steps:
? Context management with capabilities
? Schema creation
? Solution template creation
? Target creation
? Configuration API calls
? Target review

Proceeding with publish and install operations...
Publishing solution version /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution123/versions/sdkbox-m23-5.4.74.1 to target sdkbox-m23
Publish operation completed successfully.
Installing solution version /subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/sdkexamples/providers/Microsoft.Edge/targets/sdkbox-m23/solutions/sdkexamples-solution123/versions/sdkbox-m23-5.4.74.1 on target sdkbox-m23
Install operation completed successfully.
```

