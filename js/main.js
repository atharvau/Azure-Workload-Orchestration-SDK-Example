import { config } from 'dotenv';
import { DefaultAzureCredential } from '@azure/identity';
import { WorkloadOrchestrationManagementClient } from '@azure/arm-workloadorchestration';
import axios from 'axios';

// Load environment variables from .env file
config();

// --- Configuration ---
const LOCATION = "eastus2euap";
const SUBSCRIPTION_ID = process.env.AZURE_SUBSCRIPTION_ID || "973d15c6-6c57-447e-b9c6-6d79b5b784ab";
const RESOURCE_GROUP = "sdkexamples";
const CONTEXT_RESOURCE_GROUP = "Mehoopany";
const CONTEXT_NAME = "Mehoopany-Context";
const SINGLE_CAPABILITY_NAME = "sdkexamples-soap";

// --- Helper Functions ---

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

async function retryOperation(operation, maxAttempts = 3, delaySeconds = 30) {
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
        try {
            return await operation();
        } catch (e) {
            if (attempt === maxAttempts) {
                console.error(`Operation failed after ${maxAttempts} attempts.`);
                throw e;
            }
            console.log(`Attempt ${attempt} failed: ${e.message}`);
            console.log(`Waiting ${delaySeconds} seconds before retrying...`);
            await sleep(delaySeconds * 1000);
            delaySeconds *= 2; // Exponential backoff
        }
    }
}

function generateRandomSemanticVersion(includePrerelease = false, includeBuild = false) {
    const major = Math.floor(Math.random() * 11);
    const minor = Math.floor(Math.random() * 21);
    const patch = Math.floor(Math.random() * 101);
    let version = `${major}.${minor}.${patch}`;
    if (includePrerelease) {
        const prereleaseTypes = ['alpha', 'beta', 'rc'];
        const prereleaseType = prereleaseTypes[Math.floor(Math.random() * prereleaseTypes.length)];
        const prereleaseNum = Math.floor(Math.random() * 10) + 1;
        version += `-${prereleaseType}.${prereleaseNum}`;
    }
    if (includeBuild) {
        const buildNum = Math.floor(Math.random() * 10000) + 1;
        version += `+${buildNum}`;
    }
    return version;
}

// --- SDK Interaction Functions ---

async function createSchema(client, resourceGroupName) {
    const version = generateRandomSemanticVersion();
    const schemaName = `sdkexamples-schema-v${version}`;
    console.log(`Creating schema '${schemaName}'...`);
    return await client.schemas.createOrUpdate(resourceGroupName, schemaName, {
        location: LOCATION,
        properties: {}
    });
}

async function createSchemaVersion(client, resourceGroupName, schemaName) {
    const version = generateRandomSemanticVersion();
    const schemaVersionName = version;
    console.log(`Creating schema version '${schemaVersionName}' for schema '${schemaName}'...`);
    const schemaValue = `rules:
  configs:
    ErrorThreshold: { type: float, required: true, editableAt: [line], editableBy: [OT] }
    HealthCheckEndpoint: { type: string, required: false, editableAt: [line], editableBy: [OT] }
    EnableLocalLog: { type: boolean, required: true, editableAt: [line], editableBy: [OT] }
    AgentEndpoint: { type: string, required: true, editableAt: [line], editableBy: [OT] }
    HealthCheckEnabled: { type: boolean, required: false, editableAt: [line], editableBy: [OT] }
    ApplicationEndpoint: { type: string, required: true, editableAt: [line], editableBy: [OT] }
    TemperatureRangeMax: { type: float, required: true, editableAt: [line], editableBy: [OT] }`;
    
    return await client.schemaVersions.createOrUpdate(resourceGroupName, schemaName, schemaVersionName, {
        properties: { value: schemaValue }
    });
}

async function createSolutionTemplate(client, resourceGroupName, capabilities) {
    const solutionTemplateName = "sdkexamples-solution1";
    console.log(`Creating solution template '${solutionTemplateName}'...`);
    return await client.solutionTemplates.createOrUpdate(resourceGroupName, solutionTemplateName, {
        location: LOCATION,
        properties: {
            capabilities: capabilities || [SINGLE_CAPABILITY_NAME],
            description: "This is Holtmelt Solution with random capabilities"
        }
    });
}

async function createSolutionTemplateVersion(client, resourceGroupName, solutionTemplateName, schemaName, schemaVersion) {
    const version = generateRandomSemanticVersion(false, false);
    const solutionTemplateVersionName = version;
    console.log(`Creating solution template version '${solutionTemplateVersionName}'...`);
    const configurationsStr = `schema:
  name: ${schemaName}
  version: ${schemaVersion}
configs:
  AppName: Hotmelt
  TemperatureRangeMax: \${{$val(TemperatureRangeMax)}}
  ErrorThreshold: \${{$val(ErrorThreshold)}}
  HealthCheckEndpoint: \${{$val(HealthCheckEndpoint)}}
  EnableLocalLog: \${{$val(EnableLocalLog)}}
  AgentEndpoint: \${{$val(AgentEndpoint)}}
  HealthCheckEnabled: \${{$val(HealthCheckEnabled)}}
  ApplicationEndpoint: \${{$val(ApplicationEndpoint)}}`;
    
    const result = await client.solutionTemplates.createVersion(resourceGroupName, solutionTemplateName, {
        version: solutionTemplateVersionName,
        solutionTemplateVersion: {
            properties: {
                configurations: configurationsStr,
                specification: {
                    components: [{
                        name: "helmcomponent",
                        type: "helm.v3",
                        properties: { chart: { repo: "ghcr.io/eclipse-symphony/tests/helm/simple-chart", version: "0.3.0", wait: true, timeout: "5m" } }
                    }]
                },
                orchestratorType: "TO"
            }
        }
    });

    
    // The correct solution template version ID should be constructed from the resource path
    const solutionTemplateVersionId = `/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${resourceGroupName}/providers/Microsoft.Edge/solutionTemplates/${solutionTemplateName}/versions/${solutionTemplateVersionName}`;
    
    // Add the ID to the result object for consistency
    result.id = solutionTemplateVersionId;
    result.name = solutionTemplateVersionName;
    
    return result;
}

async function createTarget(client, resourceGroupName, capabilities) {
    const targetName = "sdkbox-m23";
    console.log(`Creating target '${targetName}'...`);
    
    const createOperation = async () => {
        return await client.targets.createOrUpdate(resourceGroupName, targetName, {
            extendedLocation: {
                name: `/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/configmanager-cloudtest-playground-portal/providers/Microsoft.ExtendedLocation/customLocations/den-Location`,
                type: "CustomLocation"
            },
            location: LOCATION,
            properties: {
                capabilities: capabilities || [SINGLE_CAPABILITY_NAME],
                contextId: `/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${CONTEXT_RESOURCE_GROUP}/providers/Microsoft.Edge/contexts/${CONTEXT_NAME}`,
                description: "This is MK-71 Site with random capabilities",
                displayName: "sdkbox-mk71",
                hierarchyLevel: "line",
                solutionScope: "new",
                targetSpecification: {
                    topologies: [{ bindings: [{ role: "helm.v3", provider: "providers.target.helm", config: { inCluster: "true" } }] }]
                }
            }
        });
    };
    
    return await retryOperation(createOperation);
}

async function reviewTarget(client, resourceGroupName, targetName, solutionTemplateVersionId) {
    console.log(`Starting review for target ${targetName} with template version ID: ${solutionTemplateVersionId}`);
    
    // First, trigger the review process. This creates a new solution version.
    const reviewOperation = async () => {
        return await client.targets.reviewSolutionVersion(resourceGroupName, targetName, {
            solutionDependencies: [],
            solutionInstanceName: targetName, // Assuming solution instance name is same as target name
            solutionTemplateVersionId: solutionTemplateVersionId
        });
    };
    const reviewResult = await retryOperation(reviewOperation);

    // Now, list all solution versions for the relevant solution to find the one we just created.
    // Now, list all solution versions for the relevant solution to find the one we just created.
    const solutionName = "sdkexamples-solution1"; // This should match the solution template name
    console.log(`Listing all solution versions for solution '${solutionName}' on target '${targetName}'...`);

    const solutionVersions = client.solutionVersions.listBySolution(
        resourceGroupName,
        targetName,
        solutionName
    );

    const solutionVersionList = [];
    for await (const version of solutionVersions) {
        solutionVersionList.push(version);
        console.log("------------------------------------");
        console.log(`Found Solution Version: ${version.name}`);
        console.log(`  ID: ${version.id}`);
        if (version.properties) {
            console.log(`  State: ${version.properties.state}`);
            console.log(`  Provisioning State: ${version.properties.provisioningState}`);
            console.log(`  Template Version ID: ${version.properties.solutionTemplateVersionId}`);
        }
    }
    console.log("------------------------------------");

    // Filter to find the entry that matches the solutionTemplateVersionId we used for the review
    const matchingVersion = solutionVersionList.find(version => 
        version.properties && version.properties.solutionTemplateVersionId === solutionTemplateVersionId
    );

    if (matchingVersion) {
        console.log(`Found matching solution version: ${matchingVersion.name}`);
        console.log(`  Extracted Review ID: ${matchingVersion.properties.reviewId}`);
        console.log(`  Revision: ${matchingVersion.properties.revision}`);
        console.log(`  State: ${matchingVersion.properties.state}`);
        
        // Return the full ID of the solution version for publish/install
        console.log(`Returning ID for further steps: ${matchingVersion.id}`);
        return matchingVersion.id;
    } else {
        console.error(`No matching solution version found for solutionTemplateVersionId: ${solutionTemplateVersionId}`);
        console.log("Available solution template version IDs found on target:");
        
        // Fallback to original behavior if no match found, though this is less reliable.
        if (reviewResult.id) {
            console.warn(`Falling back to ID from initial review response: ${reviewResult.id}`);
            return reviewResult.id;
        }
        throw new Error("Could not find a matching solution version ID after review and no fallback ID was available.");
    }
}

async function publishTarget(client, resourceGroupName, targetName, solutionVersionId) {
    console.log(`Publishing solution version to target ${targetName}...`);
    const publishOperation = async () => {
        const result = await client.targets.publishSolutionVersion(resourceGroupName, targetName, {
            solutionVersionId: solutionVersionId
        });
        console.log("Publish operation completed successfully.");
        return result;
    };
    return await retryOperation(publishOperation);
}

async function installTarget(client, resourceGroupName, targetName, solutionVersionId) {
    console.log(`Installing solution on target ${targetName}...`);
    const installOperation = async () => {
        await client.targets.installSolution(resourceGroupName, targetName, {
            solutionVersionId: solutionVersionId
        });
        console.log(`Install operation completed for target ${targetName}.`);
        return { message: "Installation complete." };
    };
    return await retryOperation(installOperation);
}

async function createConfigurationApiCall(credential, subscriptionId, resourceGroup, configName, solutionName, configValues) {
    const token = await credential.getToken("https://management.azure.com/.default");
    const url = `https://management.azure.com/subscriptions/${subscriptionId}/resourceGroups/${resourceGroup}/providers/Microsoft.Edge/configurations/${configName}/DynamicConfigurations/${solutionName}/versions/version1?api-version=2024-06-01-preview`;
    const headers = { "Authorization": `Bearer ${token.token}`, "Content-Type": "application/json" };
    const valuesString = Object.entries(configValues).map(([key, value]) => `${key}: ${String(value).toLowerCase()}`).join("\n") + "\n";
    const requestBody = { properties: { values: valuesString, provisioningState: "Succeeded" } };

    console.log(`Making PUT call to Configuration API...`);
    try {
        const response = await axios.put(url, requestBody, { headers });
        console.log("Configuration API PUT call successful. Status:", response.status);
        return response;
    } catch (e) {
        console.error(`Error calling Configuration API: ${e.response?.data ? JSON.stringify(e.response.data) : e.message}`);
        throw e;
    }
}

async function getConfigurationApiCall(credential, subscriptionId, resourceGroup, configName, solutionName) {
    try {
        const token = await credential.getToken("https://management.azure.com/.default");
        const url = `https://management.azure.com/subscriptions/${subscriptionId}/resourceGroups/${resourceGroup}/providers/Microsoft.Edge/configurations/${configName}/DynamicConfigurations/${solutionName}/versions/version1?api-version=2024-06-01-preview`;
        console.log(`Making GET call to Configuration API: ${url}`);
        const response = await axios.get(url, { headers: { "Authorization": `Bearer ${token.token}` } });
        console.log(`Configuration GET call successful. Status: ${response.status}`);
        console.log("Retrieved Configuration Data:", JSON.stringify(response.data, null, 2));
        return response;
    } catch (e) {
        console.error(`Configuration GET API call failed. Status: ${e.response?.status}, Response: ${e.response?.data ? JSON.stringify(e.response.data) : e.message}`);
        return null;
    }
}

async function getExistingContext(client, resourceGroupName, contextName) {
    try {
        console.log(`DEBUG: Fetching existing context: ${contextName}`);
        const context = await client.contexts.get(resourceGroupName, contextName);
        const existingCapabilities = context.properties?.capabilities || [];
        console.log(`DEBUG: Found ${existingCapabilities.length} existing capabilities.`);
        return existingCapabilities;
    } catch (e) {
        if (e.statusCode === 404) {
            console.log("DEBUG: Context not found, will create a new one.");
            return [];
        }
        console.error(`DEBUG: Error fetching context: ${e.message}`);
        throw e;
    }
}

function generateSingleRandomCapability() {
    const capabilityTypes = ["shampoo", "soap"];
    const capType = capabilityTypes[Math.floor(Math.random() * capabilityTypes.length)];
    const randomSuffix = Math.floor(1000 + Math.random() * 9000);
    const capability = { name: `sdkexamples-${capType}-${randomSuffix}`, description: `SDK generated ${capType} manufacturing capability` };
    console.log(`DEBUG: Generated single random capability: ${capability.name}`);
    return capability;
}

function mergeCapabilitiesWithUniqueness(existingCapabilities, newCapabilities) {
    const merged = new Map();
    [...existingCapabilities, ...newCapabilities].forEach(cap => {
        if (cap.name && !merged.has(cap.name)) {
            merged.set(cap.name, { name: cap.name, description: cap.description });
        }
    });
    const mergedArray = Array.from(merged.values());
    console.log(`Capability merge complete. Total unique capabilities: ${mergedArray.length}`);
    return mergedArray;
}

async function createOrUpdateContextWithHierarchies(client, resourceGroupName, contextName, capabilities) {
    const contextOperation = async () => {
        const hierarchies = [
            { name: "country", description: "Country level hierarchy" },
            { name: "region", description: "Regional level hierarchy" },
            { name: "factory", description: "Factory level hierarchy" },
            { name: "line", description: "Production line hierarchy" }
        ];
        console.log(`Creating/updating context '${contextName}'...`);
        return await client.contexts.createOrUpdate(resourceGroupName, contextName, {
            location: LOCATION,
            properties: { capabilities, hierarchies }
        });
    };
    return await retryOperation(contextOperation);
}

async function manageAzureContext(client) {
    try {
        const existingCapabilities = await getExistingContext(client, CONTEXT_RESOURCE_GROUP, CONTEXT_NAME);
        const newCapability = generateSingleRandomCapability();
        const mergedCapabilities = mergeCapabilitiesWithUniqueness(existingCapabilities, [newCapability]);
        const contextResult = await createOrUpdateContextWithHierarchies(client, CONTEXT_RESOURCE_GROUP, CONTEXT_NAME, mergedCapabilities);
        console.log(`Context management completed successfully: ${contextResult.name}`);
        return contextResult;
    } catch (e) {
        console.error(`Error in context management workflow: ${e.message}`);
        throw e;
    }
}

// --- Main Execution ---

async function main() {
    try {
        if (!SUBSCRIPTION_ID) throw new Error("AZURE_SUBSCRIPTION_ID environment variable not set.");
        const credential = new DefaultAzureCredential();
        await credential.getToken("https://management.azure.com/.default");
        console.log("Successfully authenticated with Azure.");
        
        const workloadClient = new WorkloadOrchestrationManagementClient(credential, SUBSCRIPTION_ID);

        // STEP 1: Manage Azure context
        console.log("\n" + "=".repeat(50) + "\nSTEP 1: Managing Azure Context\n" + "=".repeat(50));
        let capabilities = [];
        try {
            const contextResult = await manageAzureContext(workloadClient);
            const contextCapabilities = contextResult.properties?.capabilities || [];
            if (contextCapabilities.length > 0) {
                const lastCap = contextCapabilities[contextCapabilities.length - 1];
                if (lastCap.name) capabilities = [lastCap.name];
            }
        } catch (e) {
            console.error(`Context management failed, generating fallback capability: ${e.message}`);
        }

        if (capabilities.length === 0) {
            const newCapability = generateSingleRandomCapability();
            capabilities = [newCapability.name];
            console.log(`Using generated fallback capability: ${capabilities[0]}`);
        }
        
        console.log(`\n===> FINAL CAPABILITY FOR THIS RUN: ${capabilities[0]} \n`);
        console.log("Waiting 30 seconds after capability selection...");
        await sleep(30000);

        // STEP 2: Create Resources
        console.log("\n" + "=".repeat(50) + "\nSTEP 2: Creating Azure Resources\n" + "=".repeat(50));

        const schema = await createSchema(workloadClient, RESOURCE_GROUP);
        console.log(`Schema created successfully: ${schema.name}`);

        const schemaVersion = await createSchemaVersion(workloadClient, RESOURCE_GROUP, schema.name);
        console.log(`Schema version created successfully: ${schemaVersion.name}`);
        
        const solutionTemplate = await createSolutionTemplate(workloadClient, RESOURCE_GROUP, capabilities);
        console.log(`Solution template created successfully: ${solutionTemplate.name}`);

        const solutionTemplateVersionResult = await createSolutionTemplateVersion(workloadClient, RESOURCE_GROUP, solutionTemplate.name, schema.name, schemaVersion.name);
        const solutionTemplateVersionId = solutionTemplateVersionResult.id;
        if (!solutionTemplateVersionId) throw new Error("Failed to get ID from created solution template version.");
        console.log(`Solution template version created successfully with ID: ${solutionTemplateVersionId}`);
        
        const target = await createTarget(workloadClient, RESOURCE_GROUP, capabilities);
        console.log(`Target created successfully: ${target.name}`);

        // STEP 3: Configuration API Call
        console.log("\n" + "=".repeat(50) + "\nSTEP 3: Setting Configuration via API\n" + "=".repeat(50));
        
        try {
            const configName = `${target.name}Config`;
            const solutionName = "sdkexamples-solution1";
            const configValues = { ErrorThreshold: 35.3, HealthCheckEndpoint: "http://localhost:8080/health", EnableLocalLog: true, AgentEndpoint: "http://localhost:8080/agent", HealthCheckEnabled: true, ApplicationEndpoint: "http://localhost:8080/app", TemperatureRangeMax: 100.5 };
            
            await createConfigurationApiCall(credential, SUBSCRIPTION_ID, RESOURCE_GROUP, configName, solutionName, configValues);
            
            console.log("\nVerifying configuration...");
            await getConfigurationApiCall(credential, SUBSCRIPTION_ID, RESOURCE_GROUP, configName, solutionName);
        } catch(e) {
            console.warn(`Configuration API call failed, but continuing workflow: ${e.message}`);
        }

        // STEP 4: Review, Publish, and Install
        console.log("\n" + "=".repeat(50) + "\nSTEP 4: Review, Publish, and Install\n" + "=".repeat(50));

        const solutionVersionId = await reviewTarget(workloadClient, RESOURCE_GROUP, target.name, solutionTemplateVersionId);
        await publishTarget(workloadClient, RESOURCE_GROUP, target.name, solutionVersionId);
        await installTarget(workloadClient, RESOURCE_GROUP, target.name, solutionVersionId);
        
        console.log("\n" + "=".repeat(50) + "\nWorkflow finished successfully!\n" + "=".repeat(50));

    } catch (e) {
        console.error(`\nFATAL ERROR: An unexpected error occurred and stopped the workflow.`);
        console.error(e.message);
        if(e.stack) console.error(e.stack);
        if(e.details) console.error("Error Details:", JSON.stringify(e.details, null, 2));
    }
}

main();
