# Azure Workload Orchestration SDK - JavaScript Example

This project demonstrates how to use the Azure Workload Orchestration SDK in JavaScript/Node.js to manage workload orchestration resources including schemas, solution templates, targets, and configurations.

## Overview

The Azure Workload Orchestration service helps manage and deploy applications across distributed edge environments. This sample demonstrates the complete workflow from creating schemas and solution templates to deploying and configuring solutions on targets.

## Prerequisites

- Node.js (version 14 or higher)
- npm or yarn package manager
- Azure subscription with appropriate permissions
- Azure CLI (for authentication setup)

## Installation

1. Clone the repository and navigate to the JavaScript directory:
```bash
cd js
```

2. Install dependencies:
```bash
npm install
```

## Required Dependencies

The project uses the following key dependencies:

- `@azure/arm-workloadorchestration` - Azure Workload Orchestration SDK
- `@azure/identity` - Azure authentication
- `dotenv` - Environment variable management
- `axios` - HTTP client for REST API calls

## Configuration

### Default Configuration

The script uses the following default configuration (can be modified in `main.js`):

```javascript
const LOCATION = "eastus2euap";
const RESOURCE_GROUP = "sdkexamples";
const CONTEXT_RESOURCE_GROUP = "Mehoopany";
const CONTEXT_NAME = "Mehoopany-Context";
const SINGLE_CAPABILITY_NAME = "sdkexamples-soap";
```

## Authentication

The sample supports multiple authentication methods through `DefaultAzureCredential`:

1. **Environment variables** (recommended for automation)
2. **Azure CLI** (for local development)
3. **Managed Identity** (for Azure-hosted applications)
4. **Visual Studio Code** (for local development)

### Setting up Azure CLI Authentication

```bash
az login
az account set --subscription "your-subscription-id"
```

## Usage

Run the complete workflow:

```bash
node main.js
```

## Workflow Steps

The sample demonstrates a complete end-to-end workflow:

### 1. Context Management
- Fetches or creates an Azure context with capabilities
- Generates random capabilities for demonstration
- Manages capability hierarchies (country, region, factory, line)

### 2. Resource Creation
- **Schema Creation**: Creates a configuration schema with validation rules
- **Schema Version**: Creates a specific version of the schema
- **Solution Template**: Creates a reusable solution template
- **Solution Template Version**: Creates a versioned instance with Helm specifications
- **Target Creation**: Creates a deployment target with extended location

### 3. Configuration Management
- Uses REST API calls to set dynamic configurations
- Validates configuration values against the schema
- Supports various data types (float, string, boolean)

### 4. Deployment Workflow
- **Review**: Validates the solution template against the target
- **Publish**: Makes the solution version available for deployment
- **Install**: Deploys the solution to the target

## Key Features

### Retry Logic
The sample includes robust retry logic with exponential backoff:

```javascript
async function retryOperation(operation, maxAttempts = 3, delaySeconds = 30) {
    // Implements exponential backoff retry pattern
}
```

### Random Version Generation
Generates semantic versions for resources:

```javascript
function generateRandomSemanticVersion(includePrerelease = false, includeBuild = false) {
    // Generates versions like "2.15.43" or "1.0.0-alpha.5"
}
```

### Configuration Schema
Example schema with multiple data types:

```yaml
rules:
  configs:
    ErrorThreshold: { type: float, required: true, editableAt: [line], editableBy: [OT] }
    HealthCheckEndpoint: { type: string, required: false, editableAt: [line], editableBy: [OT] }
    EnableLocalLog: { type: boolean, required: true, editableAt: [line], editableBy: [OT] }
    AgentEndpoint: { type: string, required: true, editableAt: [line], editableBy: [OT] }
    HealthCheckEnabled: { type: boolean, required: false, editableAt: [line], editableBy: [OT] }
    ApplicationEndpoint: { type: string, required: true, editableAt: [line], editableBy: [OT] }
    TemperatureRangeMax: { type: float, required: true, editableAt: [line], editableBy: [OT] }
```

### Helm Integration
The sample includes Helm chart deployment specifications:

```javascript
specification: {
    components: [{
        name: "helmcomponent",
        type: "helm.v3",
        properties: { 
            chart: { 
                repo: "ghcr.io/eclipse-symphony/tests/helm/simple-chart", 
                version: "0.3.0", 
                wait: true, 
                timeout: "5m" 
            } 
        }
    }]
}
```

## Error Handling

The sample includes comprehensive error handling:

- **Retry mechanisms** for transient failures
- **Graceful degradation** when optional operations fail
- **Detailed error logging** with stack traces
- **Fallback strategies** for capability generation

## API Documentation

### Core Functions

#### `createSchema(client, resourceGroupName)`
Creates a new configuration schema with random versioning.

#### `createSchemaVersion(client, resourceGroupName, schemaName)`
Creates a versioned instance of a schema with validation rules.

#### `createSolutionTemplate(client, resourceGroupName, capabilities)`
Creates a reusable solution template with specified capabilities.

#### `createSolutionTemplateVersion(client, resourceGroupName, solutionTemplateName, schemaName, schemaVersion)`
Creates a versioned solution template with Helm specifications.

#### `createTarget(client, resourceGroupName, capabilities)`
Creates a deployment target with extended location support.

#### `reviewTarget(client, resourceGroupName, targetName, solutionTemplateVersionId)`
Reviews and validates a solution template against a target.

#### `publishTarget(client, resourceGroupName, targetName, solutionVersionId)`
Publishes a solution version for deployment.

#### `installTarget(client, resourceGroupName, targetName, solutionVersionId)`
Installs a solution on the specified target.

### Configuration API Functions

#### `createConfigurationApiCall(credential, subscriptionId, resourceGroup, configName, solutionName, configValues)`
Creates dynamic configuration using REST API calls.

#### `getConfigurationApiCall(credential, subscriptionId, resourceGroup, configName, solutionName)`
Retrieves configuration values for verification.

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify Azure CLI login: `az account show`
   - Check environment variables are correctly set
   - Ensure proper permissions in Azure subscription

2. **Resource Group Not Found**
   - Create resource groups before running the sample
   - Verify resource group names in configuration

3. **Extended Location Errors**
   - Ensure the custom location exists and is accessible
   - Check the extended location path format

4. **Configuration API Failures**
   - These are often expected and the sample continues gracefully
   - Check Azure region support for configuration APIs

## Output Example

The sample produces detailed console output showing each step:

```
==================================================
STEP 1: Managing Azure Context
==================================================
DEBUG: Fetching existing context: Mehoopany-Context
DEBUG: Found 5 existing capabilities.
DEBUG: Generated single random capability: sdkexamples-soap-7432

===> FINAL CAPABILITY FOR THIS RUN: sdkexamples-soap-7432 

==================================================
STEP 2: Creating Azure Resources
==================================================
Creating schema 'sdkexamples-schema-v3.12.67'...
Schema created successfully: sdkexamples-schema-v3.12.67
...