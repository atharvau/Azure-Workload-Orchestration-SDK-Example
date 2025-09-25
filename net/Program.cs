using Azure;
using Azure.Core;
using Azure.Core.Diagnostics;
using Azure.Core.Pipeline;
using Azure.Identity;
using Azure.ResourceManager;
using Azure.ResourceManager.Models;
using Azure.ResourceManager.Resources;
using Azure.ResourceManager.WorkloadOrchestration;
using Azure.ResourceManager.WorkloadOrchestration.Models;
using Microsoft.Extensions.Logging;
using System;
using System.Diagnostics.Tracing;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

// =================================================================================
// Custom HTTP Request/Response Logging Policy
// =================================================================================

public class HttpRequestResponseLoggingPolicy : HttpPipelinePolicy
{
    public override async ValueTask ProcessAsync(HttpMessage message, ReadOnlyMemory<HttpPipelinePolicy> pipeline)
    {
        var timestamp = DateTime.Now.ToString("HH:mm:ss.fff");
        
        // Log the outgoing request
        Console.WriteLine($"\n🚀 [{timestamp}] HTTP REQUEST SENT:");
        Console.WriteLine($"   Method: {message.Request.Method}");
        Console.WriteLine($"   URL: {message.Request.Uri}");
        Console.WriteLine($"   Headers:");
        foreach (var header in message.Request.Headers)
        {
            var value = header.Name.Contains("Authorization") ? "[REDACTED]" : header.Value;
            Console.WriteLine($"     {header.Name}: {value}");
        }
        
        if (message.Request.Content != null)
        {
            try
            {
                var content = message.Request.Content.ToString();
                if (!string.IsNullOrEmpty(content))
                {
                    Console.WriteLine($"   Content: {content}");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"   Content: [Could not read content: {ex.Message}]");
            }
        }
        
        // Process the request through the pipeline
        await ProcessNextAsync(message, pipeline);
        
        // Log the received response
        timestamp = DateTime.Now.ToString("HH:mm:ss.fff");
        Console.WriteLine($"\n📨 [{timestamp}] HTTP RESPONSE RECEIVED:");
        Console.WriteLine($"   Status: {message.Response.Status} {message.Response.ReasonPhrase}");
        Console.WriteLine($"   Headers:");
        foreach (var header in message.Response.Headers)
        {
            Console.WriteLine($"     {header.Name}: {header.Value}");
        }
        
        if (message.Response.Content != null)
        {
            try
            {
                var content = message.Response.Content.ToString();
                if (!string.IsNullOrEmpty(content))
                {
                    Console.WriteLine($"   Content: {content}");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"   Content: [Could not read content: {ex.Message}]");
            }
        }
        Console.WriteLine(new string('-', 80));
    }

    public override void Process(HttpMessage message, ReadOnlyMemory<HttpPipelinePolicy> pipeline)
    {
        ProcessAsync(message, pipeline).GetAwaiter().GetResult();
    }
}

public class Program
{
    // =================================================================================
    // Configuration
    // =================================================================================
    private const string Location = "eastus2euap";
    private const string ResourceGroupName = "sdkexamples";
    private const string ContextResourceGroupName = "Mehoopany";
    private const string ContextName = "Mehoopany-Context";

    public static async Task Main(string[] args)
    {
        string subscriptionId = Environment.GetEnvironmentVariable("AZURE_SUBSCRIPTION_ID") ?? "973d15c6-6c57-447e-b9c6-6d79b5b784ab";

        // =================================================================================
        // Azure SDK Logging Configuration
        // =================================================================================

        // Enable Azure SDK HTTP logging and tracing
        using AzureEventSourceListener listener = AzureEventSourceListener.CreateConsoleLogger(EventLevel.Verbose);

        // Configure logger factory for detailed logging
        using var loggerFactory = LoggerFactory.Create(builder =>
        {
            builder
                .AddConsole()
                .SetMinimumLevel(LogLevel.Debug);
        });

        Console.WriteLine("🔍 AZURE SDK DEBUG LOGGING ENABLED");
        Console.WriteLine("📊 Event Level: Verbose (all HTTP requests/responses will be logged)");
        Console.WriteLine("📋 Logger Level: Debug (detailed internal operations)");
        Console.WriteLine(new string('=', 80));

        // Authentication and Client Setup
        var credential = new DefaultAzureCredential();

        var clientOptions = new ArmClientOptions();
        // Enable diagnostics and retry logging
        clientOptions.Diagnostics.IsLoggingEnabled = true;
        clientOptions.Diagnostics.IsLoggingContentEnabled = true;
        clientOptions.Diagnostics.IsTelemetryEnabled = true;
        clientOptions.Diagnostics.IsDistributedTracingEnabled = true;
        clientOptions.Diagnostics.LoggedHeaderNames.Add("*");
        clientOptions.Diagnostics.LoggedQueryParameters.Add("*");

        clientOptions.SetApiVersion(new ResourceType("Microsoft.Edge/locations/operationStatuses"), "2023-07-01-preview");

        // Add custom HTTP request/response logging pipeline policy
        clientOptions.AddPolicy(new HttpRequestResponseLoggingPolicy(), Azure.Core.HttpPipelinePosition.PerCall);

        var armClient = new ArmClient(credential, default, clientOptions);

        Console.WriteLine("Successfully authenticated with Azure.");
        Console.WriteLine("🔧 ARM Client configured with full diagnostics logging enabled");

        SubscriptionResource subscription = armClient.GetSubscriptionResource(ResourceIdentifier.Root.AppendChildResource("subscriptions", subscriptionId));
        ResourceGroupResource resourceGroup = await subscription.GetResourceGroups().GetAsync(ResourceGroupName);
        ResourceGroupResource contextResourceGroup = await subscription.GetResourceGroups().GetAsync(ContextResourceGroupName);

        // =================================================================================
        //  Step 1: Manage Azure Context & Select a Capability
        // =================================================================================
        Console.WriteLine("\n" + new string('=', 50));
        Console.WriteLine("STEP 1: Managing Azure Context with Random Capabilities");
        Console.WriteLine(new string('=', 50));

        string selectedCapability;
        EdgeContextResource contextResult;

        try
        {
            // Full context management workflow
            contextResult = await ManageAzureContextAsync(contextResourceGroup, ContextName);

            // Extract the newly added capability for consistent use
            var lastCapability = contextResult.Data.Properties.Capabilities.LastOrDefault();
            if (lastCapability?.Name != null)
            {
                selectedCapability = lastCapability.Name;
                Console.WriteLine($"SELECTED CAPABILITY FOR ALL RESOURCES: {selectedCapability}");
            }
            else
            {
                throw new InvalidOperationException("Could not extract a capability from the updated context.");
            }
        }
        catch (Exception e)
        {
            Console.WriteLine($"Context management failed: {e.Message}. Generating a new random capability as a fallback.");
            var (name, _) = GenerateSingleRandomCapability();
            selectedCapability = name;
            Console.WriteLine($"FALLBACK CAPABILITY FOR ALL RESOURCES: {selectedCapability}");
        }

        Console.WriteLine($"\nFINAL CAPABILITY SELECTION: {selectedCapability}");
        Console.WriteLine(new string('=', 60));
        Console.WriteLine("\nWaiting 30 seconds after capability selection...");
        await Task.Delay(TimeSpan.FromSeconds(30));
        Console.WriteLine("Continuing with resource creation...\n");

        // =================================================================================
        //  Step 2: Create Schema, Solution Template, and Target
        // =================================================================================
        Console.WriteLine(new string('=', 50));
        Console.WriteLine("STEP 2: Creating Azure Resources");
        Console.WriteLine(new string('=', 50));

        EdgeSchemaResource schema;
        EdgeSchemaVersionResource schemaVersion;
        EdgeSolutionTemplateResource solutionTemplate;
        EdgeSolutionTemplateVersionResource solutionTemplateVersion;
        EdgeTargetResource target;

        try
        {
            // Create Schema
            Console.WriteLine($"Creating schema in resource group: {resourceGroup.Data.Name}");
            schema = await CreateSchemaAsync(resourceGroup);
            Console.WriteLine($"Schema created successfully: {schema.Data.Name}");

            // Create Schema Version
            Console.WriteLine($"Creating schema version for schema: {schema.Data.Name}");
            schemaVersion = await CreateSchemaVersionAsync(schema);
            Console.WriteLine($"Schema version created successfully: {schemaVersion.Data.Name}");

            // Create Solution Template
            Console.WriteLine($"Creating solution template with capability: {selectedCapability}");
            solutionTemplate = await CreateSolutionTemplateAsync(resourceGroup, selectedCapability);
            Console.WriteLine($"Solution template created successfully: {solutionTemplate.Data.Name}");

            // Create Solution Template Version
            Console.WriteLine($"Creating solution template version for template: {solutionTemplate.Data.Name}");
            solutionTemplateVersion = await CreateSolutionTemplateVersionAsync(solutionTemplate, schema.Data.Name, schemaVersion.Data.Name);
            // Note: The ID is available as solutionTemplateVersion.Id
            Console.WriteLine($"Solution template version created successfully with ID: {solutionTemplateVersion.Id}");

            // Create Target
            Console.WriteLine($"Creating target with capability: {selectedCapability}");
            target = await CreateTargetAsync(resourceGroup, selectedCapability, subscriptionId);
            Console.WriteLine($"Target created successfully: {target.Data.Name}");
        }
        catch (Exception e)
        {
            Console.WriteLine($"An error occurred during resource creation: {e}");
            return;
        }

        Console.WriteLine("\nWorkflow finished successfully!");
    }

    // =================================================================================
    // Helper Functions
    // =================================================================================

    // Generates a random semantic version string
    static string GenerateRandomSemanticVersion(bool includePrerelease = false, bool includeBuild = false)
    {
        var rand = new Random();
        string version = $"{rand.Next(0, 11)}.{rand.Next(0, 21)}.{rand.Next(0, 101)}";
        if (includePrerelease)
        {
            string[] prereleaseTypes = { "alpha", "beta", "rc" };
            version += $"-{prereleaseTypes[rand.Next(prereleaseTypes.Length)]}.{rand.Next(1, 11)}";
        }
        if (includeBuild)
        {
            version += $"+{rand.Next(1, 10001)}";
        }
        return version;
    }

    // Generates a single random capability
    static (string name, string description) GenerateSingleRandomCapability()
    {
        var rand = new Random();
        string[] capabilityTypes = { "shampoo", "soap" };
        string capType = capabilityTypes[rand.Next(capabilityTypes.Length)];
        int randomSuffix = rand.Next(1000, 10000);
        string name = $"sdkexamples-{capType}-{randomSuffix}";
        string description = $"SDK generated {capType} manufacturing capability";
        Console.WriteLine($"DEBUG: Generated single random capability: {name}");
        return (name, description);
    }

    // Full context management workflow
    static async Task<EdgeContextResource> ManageAzureContextAsync(ResourceGroupResource contextResourceGroup, string contextName)
    {
        EdgeContextCollection contextCollection = contextResourceGroup.GetEdgeContexts();

        // 1. Fetch existing context and its capabilities
        List<ContextCapability> existingCapabilities = new();
        if (await contextCollection.ExistsAsync(contextName))
        {
            EdgeContextResource existingContext = await contextCollection.GetAsync(contextName);
            if (existingContext.Data.Properties?.Capabilities != null)
            {
                existingCapabilities.AddRange(existingContext.Data.Properties.Capabilities);
            }
        }
        Console.WriteLine($"DEBUG: Found {existingCapabilities.Count} existing capabilities.");

        // 2. Generate a new random capability
        var (newName, newDescription) = GenerateSingleRandomCapability();
        var newCapability = new ContextCapability(newName, newDescription);

        // 3. Merge with uniqueness (ensuring no `state` field is sent back)
        var mergedCapabilities = existingCapabilities
            .Select(c => new ContextCapability(c.Name, c.Description)) // Re-create to remove output-only properties
            .ToList();
        if (!mergedCapabilities.Any(c => c.Name.Equals(newCapability.Name, StringComparison.OrdinalIgnoreCase)))
        {
            mergedCapabilities.Add(newCapability);
        }
        Console.WriteLine($"DEBUG: Merged capabilities count: {mergedCapabilities.Count}");

        // 4. Define hierarchies and create/update the context
        var hierarchies = new List<ContextHierarchy>
        {
            new("country", "Country level hierarchy"),
            new("region", "Regional level hierarchy"),
            new("factory", "Factory level hierarchy"),
            new("line", "Production line hierarchy")
        };

        var contextData = new EdgeContextData(new AzureLocation(Location))
        {
            Properties = new EdgeContextProperties(mergedCapabilities, hierarchies)
        };

        Console.WriteLine($"Creating/updating context: {contextName}");
        ArmOperation<EdgeContextResource> contextOperation = await contextCollection.CreateOrUpdateAsync(WaitUntil.Completed, contextName, contextData);
        Console.WriteLine($"Context management completed successfully: {contextOperation.Value.Data.Name}");

        return contextOperation.Value;
    }

    // Creates a new Schema
    static async Task<EdgeSchemaResource> CreateSchemaAsync(ResourceGroupResource resourceGroup)
    {
        var schemaCollection = resourceGroup.GetEdgeSchemas();
        string schemaName = $"sdkexamples-schema-v{GenerateRandomSemanticVersion()}";
        var schemaData = new EdgeSchemaData(new AzureLocation(Location))
        {
            Properties = new EdgeSchemaProperties()
        };
        ArmOperation<EdgeSchemaResource> operation = await schemaCollection.CreateOrUpdateAsync(WaitUntil.Completed, schemaName, schemaData);
        return operation.Value;
    }

    // Creates a new Schema Version
    static async Task<EdgeSchemaVersionResource> CreateSchemaVersionAsync(EdgeSchemaResource schema)
    {
        var versionCollection = schema.GetEdgeSchemaVersions();
        string schemaVersionName = GenerateRandomSemanticVersion();
        var schemaVersionData = new EdgeSchemaVersionData
        {
            Properties = new EdgeSchemaVersionProperties(
                """
                rules:
                  configs:
                    ErrorThreshold:
                      type: float
                      required: true
                      editableAt: [line]
                      editableBy: [OT]
                    HealthCheckEndpoint:
                      type: string
                      required: false
                      editableAt: [line]
                      editableBy: [OT]
                    EnableLocalLog:
                      type: boolean
                      required: true
                      editableAt: [line]
                      editableBy: [OT]
                    AgentEndpoint:
                      type: string
                      required: true
                      editableAt: [line]
                      editableBy: [OT]
                    HealthCheckEnabled:
                      type: boolean
                      required: false
                      editableAt: [line]
                      editableBy: [OT]
                    ApplicationEndpoint:
                      type: string
                      required: true
                      editableAt: [line]
                      editableBy: [OT]
                    TemperatureRangeMax:
                      type: float
                      required: true
                      editableAt: [line]
                      editableBy: [OT]
                """
            )
        };
        ArmOperation<EdgeSchemaVersionResource> operation = await versionCollection.CreateOrUpdateAsync(WaitUntil.Completed, schemaVersionName, schemaVersionData);
        return operation.Value;
    }

    // Creates a new Solution Template
    static async Task<EdgeSolutionTemplateResource> CreateSolutionTemplateAsync(ResourceGroupResource resourceGroup, string capability)
    {
        var templateCollection = resourceGroup.GetEdgeSolutionTemplates();
        string templateName = "sdkexamples-solution1547";
        var templateData = new EdgeSolutionTemplateData(new AzureLocation(Location))
        {
            Properties = new EdgeSolutionTemplateProperties("This is Holtmelt Solution with random capabilities", new[] { capability })
        };
        
        Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Starting Solution Template creation");
        Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Template Name = {templateName}");
        Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Location = {Location}");
        Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Capability = {capability}");
        Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Resource Group = {resourceGroup.Data.Name}");
        Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Subscription = {resourceGroup.Id.SubscriptionId}");
        
        int maxRetries = 1;
        TimeSpan delay = TimeSpan.FromSeconds(60);

        for (int i = 0; i < maxRetries; i++)
        {
            try
            {
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Calling CreateOrUpdateAsync with WaitUntil.Completed (Attempt {i + 1}/{maxRetries})");
                
                ArmOperation<EdgeSolutionTemplateResource> operation = await templateCollection.CreateOrUpdateAsync(WaitUntil.Completed, templateName, templateData);
                
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Operation completed successfully");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Final Status = {operation.GetRawResponse().Status}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_DEBUG: Resource ID = {operation.Value.Id}");
                
                return operation.Value;
            }
            catch (Azure.RequestFailedException ex) when (ex.ErrorCode == "SolutionTemplateCapabilityInvalid" && i < maxRetries - 1)
            {
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_WARN: Capability not yet propagated. Retrying in {delay.TotalSeconds} seconds...");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_WARN: Error: {ex.Message}");
                await Task.Delay(delay);
            }
            catch (Azure.RequestFailedException ex)
            {
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Azure.RequestFailedException during LRO polling");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: HTTP Status = {ex.Status}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Error Code = {ex.ErrorCode}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Message = {ex.Message}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Request ID = {ex.GetRawResponse()?.ClientRequestId ?? "Unknown"}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Response Headers Available = {ex.GetRawResponse()?.Headers != null}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Full Response Content = {ex.GetRawResponse()?.Content}");
                
                if (ex.Message.Contains("API version") || ex.Message.Contains("api-versions"))
                {
                    Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: API VERSION MISMATCH DETECTED!");
                    Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: This indicates the SDK is using an incompatible API version");
                    Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Current SDK version is trying to use 2025-06-01 but service supports 2023-07-01-preview");
                }
                
                throw;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Unexpected exception during LRO");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Exception Type = {ex.GetType().FullName}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Message = {ex.Message}");
                Console.WriteLine($"[{DateTime.Now:HH:mm:ss.fff}] LRO_ERROR: Stack Trace = {ex.StackTrace}");
                throw;
            }
        }
        throw new Exception("Failed to create solution template after multiple retries.");
    }

    // Creates a new Solution Template Version
    static async Task<EdgeSolutionTemplateVersionResource> CreateSolutionTemplateVersionAsync(EdgeSolutionTemplateResource template, string schemaName, string schemaVersion)
    {
        string version = GenerateRandomSemanticVersion();

        string configurationsStr = $@"schema:
  name: {schemaName}
  version: {schemaVersion}
configs:
  AppName: Hotmelt
  TemperatureRangeMax: ${{val(TemperatureRangeMax)}}
  ErrorThreshold: ${{val(ErrorThreshold)}}
  HealthCheckEndpoint: ${{val(HealthCheckEndpoint)}}
  EnableLocalLog: ${{val(EnableLocalLog)}}
  AgentEndpoint: ${{val(AgentEndpoint)}}
  HealthCheckEnabled: ${{val(HealthCheckEnabled)}}
  ApplicationEndpoint: ${{val(ApplicationEndpoint)}}
";

        var spec = new Dictionary<string, BinaryData>
        {
            {
                "components", BinaryData.FromString(
                    """
                    [
                        {
                            "name": "helmcomponent",
                            "type": "helm.v3",
                            "properties": {
                                "chart": {
                                    "repo": "ghcr.io/eclipse-symphony/tests/helm/simple-chart",
                                    "version": "0.3.0",
                                    "wait": true,
                                    "timeout": "5m"
                                }
                            }
                        }
                    ]
                    """
                )
            }
        };

        var versionProps = new EdgeSolutionTemplateVersionProperties(configurationsStr, spec)
        {
            OrchestratorType = SolutionVersionOrchestratorType.TO
        };

        var versionWithUpdate = new EdgeSolutionTemplateVersionWithUpdateType(new EdgeSolutionTemplateVersionData { Properties = versionProps })
        {
            Version = version
        };

        ArmOperation<EdgeSolutionTemplateVersionResource> operation = await template.CreateVersionAsync(WaitUntil.Completed, versionWithUpdate);
        return operation.Value;
    }

    // Creates a new Target
   static async Task<EdgeTargetResource> CreateTargetAsync(ResourceGroupResource resourceGroup, string capability, string subscriptionId)
    {
        var targetCollection = resourceGroup.GetEdgeTargets();
        string targetName = "sdkbox-m23";
        var topologies = new List<IDictionary<string, BinaryData>>
        {
            new Dictionary<string, BinaryData>
            {
                { "bindings", BinaryData.FromString(
                    """
                    [
                        {
                            "role": "helm.v3",
                            "provider": "providers.target.helm",
                            "config": { "inCluster": "true" }
                        }
                    ]
                    """
                )}
            }
        };

        var targetData = new EdgeTargetData(new AzureLocation(Location))
        {
            Properties = new EdgeTargetProperties(
                "This is MK-71 Site with random capabilities",
                "sdkbox-mk71",
                new ResourceIdentifier($"/subscriptions/{subscriptionId}/resourceGroups/{ContextResourceGroupName}/providers/Microsoft.Edge/contexts/{ContextName}"),
                new Dictionary<string, BinaryData> { { "topologies", BinaryData.FromObjectAsJson(topologies) } },
                new[] { capability },
                "line"
            )
            {
                SolutionScope = "new"
            }
        };

        ArmOperation<EdgeTargetResource> operation = await targetCollection.CreateOrUpdateAsync(WaitUntil.Completed, targetName, targetData);
        return operation.Value;
    }
}
