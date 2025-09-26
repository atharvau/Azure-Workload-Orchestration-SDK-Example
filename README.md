# Azure Workload Orchestration SDK Examples

This repository contains a collection of sample applications demonstrating the end-to-end usage of the Azure Workload Orchestration SDK for various programming languages. Each example showcases a common workflow for creating and managing Azure Workload Orchestration resources.

The repository is located at: [https://github.com/atharvau/Azure-Workload-Orchestration-SDK-Example](https://github.com/atharvau/Azure-Workload-Orchestration-SDK-Example)

## What These Examples Do

Each language-specific example implements a complete workflow that performs the following operations:

1.  **Authentication**: Connects to Azure using `DefaultAzureCredential`.
2.  **Context Management**: Fetches an existing Azure Context, adds a new randomly generated capability, and updates the context resource.
3.  **Schema Creation**: Creates a new Schema and a corresponding Schema Version.
4.  **Solution Template**: Creates a Solution Template and a version for it, linking it to the created schema.
5.  **Target Creation**: Deploys a Target resource with a specific configuration.
6.  **Configuration API Call**: Sets dynamic configuration values for the target by making a direct REST API call.
7.  **Deployment Workflow**:
    *   **Reviews** the solution on the target.
    *   **Publishes** the reviewed solution.
    *   **Installs** the published solution.


## SDKs and Libraries

The following table lists the official SDK libraries used in these examples.

| Language     | Library URL                                                                                                                              |
| :----------- | :--------------------------------------------------------------------------------------------------------------------------------------- |
| **Go**       | [pkg.go.dev/github.com/Azure/azure-sdk-for-go/.../armworkloadorchestration](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/workloadorchestration/armworkloadorchestration) |
| **Java**     | [search.maven.org/artifact/com.azure.resourcemanager/azure-resourcemanager-workloadorchestration](https://search.maven.org/artifact/com.azure.resourcemanager/azure-resourcemanager-workloadorchestration) |
| **JavaScript** | [npmjs.com/package/@azure/arm-workloadorchestration](https://www.npmjs.com/package/@azure/arm-workloadorchestration)                     |
| **Python**   | [pypi.org/project/azure-mgmt-workloadorchestration](https://pypi.org/project/azure-mgmt-workloadorchestration/)                               |
| **.NET**     | [nuget.org/packages/Azure.ResourceManager.WorkloadOrchestration](https://www.nuget.org/packages/Azure.ResourceManager.WorkloadOrchestration) |

## How to Run the Examples

Each language has its own detailed `README.md` with specific instructions for configuration, dependency installation, and execution.

- [Go README](./golang/README.md)
- [Java README](./java/README.md)
- [JavaScript README](./js/README.md)
- [Python README](./python/README.md)
- [.NET README](./net/README.md)

