# Azure WorkloadOrchestration SDK Example (JavaScript)

This example demonstrates how to use the Azure WorkloadOrchestration SDK in JavaScript.

## Prerequisites

- Node.js (latest LTS version recommended)
- An Azure subscription
- Azure credentials with appropriate permissions

## Setup

1. Install dependencies:
```bash
npm install
```

2. Configure Azure Credentials:
   - The example uses `DefaultAzureCredential` from `@azure/identity`
   - Make sure you're logged in with Azure CLI or have appropriate environment variables set

3. Update Configuration:
   - Open `main.js` or `index.js`
   - Replace the `subscriptionId` with your Azure subscription ID
   - In `main.js`, update RESOURCE_GROUP and other constants as needed

## Running the Examples

1. Basic Client Initialization (index.js):
```bash
node index.js
```

2. Full Resource Management Example (main.js):
```bash
npm start
# or
node main.js
```

## Code Structure

- `main.js` - Complete example demonstrating schema, solution template, and target creation
- `index.js` - Basic example showing client initialization
- `package.json` - Project configuration and dependencies

## Features

The main.js example demonstrates:
- Schema creation and versioning
- Solution template creation and versioning
- Target creation
- Version management using local file storage
- Error handling and async operations

## Dependencies

- `@azure/arm-workloadorchestration` - Azure WorkloadOrchestration SDK
- `@azure/identity` - Azure authentication library

## Notes

- The examples use `DefaultAzureCredential` which attempts to authenticate using multiple methods
- Make sure you have appropriate permissions in your Azure subscription
- For production use, implement proper error handling and logging
- Version tracking is done using a local version.txt file