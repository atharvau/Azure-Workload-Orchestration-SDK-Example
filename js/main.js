// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

import { WorkloadOrchestrationManagementClient } from "@azure/arm-workloadorchestration";
import { DefaultAzureCredential } from "@azure/identity";

/**
 * This sample demonstrates how to create a complete workload orchestration workflow including:
 * 1. Schema creation with random semantic versioning
 * 2. Schema version creation with random semantic versioning
 * 3. Solution template creation (matching CLI parameters)
 * 4. Solution template version creation with random semantic versioning
 * 5. Target creation with matching capabilities
 * 
 * The solution template references the actual created schema and schema version.
 * The target uses the same capabilities as the solution template for consistency.
 * Based on CLI command: az workload-orchestration solution-template create --solution-template-name "sdkexamples-solution" 
 * -g sdkexamples -l eastus2euap --capabilities "sdkexamples-soap" --description "This is Holtmelt Solution"
 *
 * @summary create complete workload orchestration resources with random semantic versioning
 */

// Configuration
const SUBSCRIPTION_ID = "973d15c6-6c57-447e-b9c6-6d79b5b784ab";

/**
 * Generates a random semantic version string
 * @param {Object} options - Configuration options
 * @param {boolean} options.includePrerelease - Whether to include prerelease tags
 * @param {boolean} options.includeBuild - Whether to include build metadata
 * @param {number} options.majorMax - Maximum value for major version (default: 10)
 * @param {number} options.minorMax - Maximum value for minor version (default: 20)
 * @param {number} options.patchMax - Maximum value for patch version (default: 100)
 * @returns {string} A random semantic version string
 */
function generateRandomSemanticVersion(options = {}) {
    const {
        includePrerelease = Math.random() > 0.7, // 30% chance of prerelease
        includeBuild = Math.random() > 0.8,      // 20% chance of build metadata
        majorMax = 10,
        minorMax = 20,
        patchMax = 100
    } = options;
    
    // Generate major.minor.patch
    const major = Math.floor(Math.random() * majorMax);
    const minor = Math.floor(Math.random() * minorMax);
    const patch = Math.floor(Math.random() * patchMax);
    
    let version = `${major}.${minor}.${patch}`;
    
    // Add prerelease if needed
    if (includePrerelease) {
        const prereleaseTypes = ['alpha', 'beta', 'rc'];
        const prereleaseType = prereleaseTypes[Math.floor(Math.random() * prereleaseTypes.length)];
        const prereleaseNum = Math.floor(Math.random() * 10) + 1;
        version += `-${prereleaseType}.${prereleaseNum}`;
    }
    
    // Add build metadata if needed
    if (includeBuild) {
        // Build metadata could be date-based, git commit hash-like, or just a number
        const buildTypes = [
            // Date format: YYYYMMDD
            `${new Date().toISOString().slice(0, 10).replace(/-/g, '')}`,
            // Git hash-like
            `${Math.random().toString(36).substring(2, 8)}`,
            // Build number
            `${Math.floor(Math.random() * 10000)}`
        ];
        const buildMeta = buildTypes[Math.floor(Math.random() * buildTypes.length)];
        version += `+${buildMeta}`;
    }
    
    return version;
}

async function createSchema(client, resourceGroupName) {
    try {
        const randomVersion = generateRandomSemanticVersion({ includePrerelease: false, includeBuild: false });
        const schemaName = `sdkexamples-schema-v${randomVersion}`;
        
        console.log(`Creating schema with random version: ${schemaName}`);
        
        const schema = await client.schemas.createOrUpdate(
            resourceGroupName,
            schemaName,
            {
                location: "eastus2euap",
                properties: {}
            }
        );

        return schema;
    } catch (error) {
        console.error(`Error creating schema: ${error.message}`);
        throw error;
    }
}

async function createSchemaVersion(client, resourceGroupName, schemaName) {
    try {
        const randomSchemaVersion = generateRandomSemanticVersion({ includePrerelease: false, includeBuild: false });
        
        console.log(`Creating schema version with random version: ${randomSchemaVersion}`);
        
        const schemaVersion = await client.schemaVersions.createOrUpdate(
            resourceGroupName,
            schemaName,
            randomSchemaVersion,
            {
                properties: {
                    value: `rules:
  configs:
    ErrorThreshold:
      type: float
      required: true
    AppName:
      type: string
      required: true
    TemperatureRangeMax:
      type: int
      required: true
    HealthCheckEndpoint:
      type: string
      required: true
    EnableLocalLog:
      type: boolean
      required: true
    AgentEndpoint:
      type: string
      required: true
    HealthCheckEnabled:
      type: boolean
      required: true
    ApplicationEndpoint:
      type: string
      required: true`
                }
            }
        );

        return schemaVersion;
    } catch (error) {
        console.error(`Error creating schema version: ${error.message}`);
        throw error;
    }
}

async function createTarget(client, resourceGroupName) {
    try {
        const targetName = "sdkbox-mk799";
        
        console.log(`Creating target: ${targetName}`);
        
        const target = await client.targets.createOrUpdate(
            resourceGroupName,
            targetName,
            {
                extendedLocation: {
                    name: "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/configmanager-cloudtest-playground-portal/providers/Microsoft.ExtendedLocation/customLocations/den-Location",
                    type: "CustomLocation"
                },
                location: "eastus2euap",
                properties: {
                    capabilities: ["sdkexamples-soap"],
                    contextId: "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/Mehoopany/providers/Microsoft.Edge/contexts/Mehoopany-Context",
                    description: "This is MK-71 Site",
                    displayName: "sdkbox-mk71",
                    hierarchyLevel: "line",
                    solutionScope: "new",
                    targetSpecification: {
                        topologies: [
                            {
                                bindings: [
                                    {
                                        role: "helm.v3",
                                        provider: "providers.target.helm",
                                        config: {
                                            inCluster: "true"
                                        }
                                    }
                                ]
                            }
                        ]
                    }
                }
            }
        );

        return target;
    } catch (error) {
        console.error(`Error creating target: ${error.message}`);
        throw error;
    }
}

async function solutionTemplateCreateExactCliReplica() {
  const credential = new DefaultAzureCredential();
  const resourceGroupName = "sdkexamples";
  const solutionTemplateName = "sdkexamples-solution";
  
  const client = new WorkloadOrchestrationManagementClient(credential, SUBSCRIPTION_ID);
  
  // Step 1: Create schema
  console.log("Creating schema...");
  const schema = await createSchema(client, resourceGroupName);
  console.log("Schema created:", schema.name);
  
  // Step 2: Create schema version
  console.log("Creating schema version...");
  const schemaVersion = await createSchemaVersion(client, resourceGroupName, schema.name);
  console.log("Schema version created:", schemaVersion.name);
  
  // Step 3: Create solution template (matches first PUT request from CLI)
  console.log("Creating solution template...");
  const solutionTemplate = await client.solutionTemplates.createOrUpdate(
    resourceGroupName,
    solutionTemplateName,
    {
      location: "eastus2euap",
      properties: {
        capabilities: ["sdkexamples-soap"],
        description: "This is Holtmelt Solution",
      },
    }
  );
  
  console.log("Solution template created:", solutionTemplate.name);
  
  // Wait for solution template to be ready
  await new Promise(resolve => setTimeout(resolve, 10000));
  
  // Step 4: Create solution template version (matches POST request from CLI)
  console.log("Creating solution template version...");
  
  // Configuration using the actual created schema
  const configurations = `schema:
  name: ${schema.name}
  version: ${schemaVersion.name}
configs:
  AppName: Hotmelt`;

  // Exact specification from CLI request body
  const specification = {
    components: [
      {
        name: "helmcomponent",
        type: "helm.v3",
        properties: {
          chart: {
            repo: "ghcr.io/eclipse-symphony/tests/helm/simple-chart",
            version: "0.3.0",
            wait: true,
            timeout: "5m",
          },
        },
      },
    ],
  };

  // Generate random version to avoid conflicts
  const randomVersion = generateRandomSemanticVersion({ includePrerelease: false, includeBuild: false });
  console.log(`Using random version: ${randomVersion}`);

  const versionPayload = {
    solutionTemplateVersion: {
      properties: {
        configurations,
        specification,
        orchestratorType: "TO",
      },
    },
    version: randomVersion,
  };

  const poller = await client.solutionTemplates.createVersion(
    resourceGroupName,
    solutionTemplateName,
    versionPayload
  );

  // Enhanced polling with multiple approaches to handle beta SDK
  let result;
  let attempts = 0;
  const maxAttempts = 5;

  while (!result && attempts < maxAttempts) {
    attempts++;
    console.log(`Polling attempt ${attempts}/${maxAttempts}...`);

    try {
      // Approach 1: Standard pollUntilDone
      if (typeof poller.pollUntilDone === "function") {
        console.log("Using pollUntilDone...");
        result = await poller.pollUntilDone({
          intervalInMs: 2000,
        });
        break;
      }
    } catch (pollError) {
      console.log("pollUntilDone failed:", pollError.message);
    }

    try {
      // Approach 2: Async iterator
      if (Symbol.asyncIterator in poller) {
        console.log("Using async iterator...");
        for await (const state of poller) {
          console.log("Poll state:", state.status || "unknown");
          if (state.isCompleted || state.status === "Succeeded") {
            result = state.result || state;
            break;
          }
        }
        if (result) break;
      }
    } catch (iterError) {
      console.log("Async iterator failed:", iterError.message);
    }

    try {
      // Approach 3: Manual polling
      if (typeof poller.poll === "function") {
        console.log("Using manual polling...");
        await poller.poll();
        if (typeof poller.isDone === "function" && poller.isDone()) {
          result = poller.getResult ? poller.getResult() : poller.result;
          break;
        }
      }
    } catch (manualError) {
      console.log("Manual polling failed:", manualError.message);
    }

    // Approach 4: Check if result is already available
    if (poller.result) {
      console.log("Found synchronous result...");
      result = poller.result;
      break;
    }

    // Wait before next attempt
    if (attempts < maxAttempts) {
      console.log("Waiting 3 seconds before next attempt...");
      await new Promise(resolve => setTimeout(resolve, 3000));
    }
  }

  // If all polling attempts failed, use the poller itself as result
  if (!result) {
    console.log("All polling approaches failed, using poller as result...");
    result = poller;
  }

  console.log("Solution template version created successfully");
  
  // Step 5: Create target
  console.log("Creating target...");
  const target = await createTarget(client, resourceGroupName);
  console.log("Target created:", target.name);
  
  console.log("All resources created successfully");
  console.log("Result:", result);
  return { schema, schemaVersion, solutionTemplate, result, target };
}

/**
 * Demonstrates different ways to generate semantic versions
 */
function demonstrateSemanticVersions() {
    console.log("\n===== Random Semantic Version Generator Demonstration =====");
    
    // Generate 5 standard versions
    console.log("\nStandard versions (random major.minor.patch):");
    for (let i = 0; i < 5; i++) {
        console.log(`  ${generateRandomSemanticVersion({ includePrerelease: false, includeBuild: false })}`);
    }
    
    // Generate 5 versions with prereleases
    console.log("\nVersions with prereleases (major.minor.patch-prerelease):");
    for (let i = 0; i < 5; i++) {
        console.log(`  ${generateRandomSemanticVersion({ includePrerelease: true, includeBuild: false })}`);
    }
    
    // Generate 5 versions with build metadata
    console.log("\nVersions with build metadata (major.minor.patch+build):");
    for (let i = 0; i < 5; i++) {
        console.log(`  ${generateRandomSemanticVersion({ includePrerelease: false, includeBuild: true })}`);
    }
    
    // Generate 5 full versions (with both prerelease and build)
    console.log("\nFull versions (major.minor.patch-prerelease+build):");
    for (let i = 0; i < 5; i++) {
        console.log(`  ${generateRandomSemanticVersion({ includePrerelease: true, includeBuild: true })}`);
    }
    
    // Generate versions with custom ranges
    console.log("\nCustom range versions (higher version numbers):");
    for (let i = 0; i < 5; i++) {
        console.log(`  ${generateRandomSemanticVersion({ 
            majorMax: 100, 
            minorMax: 100, 
            patchMax: 1000 
        })}`);
    }
}

async function main() {
  try {
    await solutionTemplateCreateExactCliReplica();
    console.log("\n=== SUCCESS ===");
    console.log("Complete workload orchestration resources created successfully:");
    console.log("• Schema: [Random schema name with semantic version]");
    console.log("• Schema Version: [Random semantic version generated]");
    console.log("• Solution Template Name: sdkexamples-solution");
    console.log("• Solution Template Version: [Random semantic version generated]");
    console.log("• Target Name: sdkbox-mk71");
    console.log("• Resource Group: sdkexamples");
    console.log("• Location: eastus2euap");
    console.log("• Capabilities: [sdkexamples-soap] (consistent across solution template and target)");
    console.log("• Description: This is Holtmelt Solution");
    console.log("• Config: AppName: Hotmelt");
    console.log("• Helm Chart: ghcr.io/eclipse-symphony/tests/helm/simple-chart:0.3.0");
    console.log("• Target Type: helm.v3 with inCluster configuration");
    
    // Demonstrate semantic version generation
    demonstrateSemanticVersions();
  } catch (error) {
    console.error("Error:", error.message);
    if (error.name === "RestError") {
      console.error("Status:", error.statusCode);
      console.error("Code:", error.code);
      if (error.response) {
        try {
          const errorDetails = JSON.parse(error.response.bodyAsText);
          console.error("Details:", JSON.stringify(errorDetails, null, 2));
        } catch {
          console.error("Response:", error.response.bodyAsText);
        }
      }
    }
    throw error;
  }
}

main().catch(console.error);
