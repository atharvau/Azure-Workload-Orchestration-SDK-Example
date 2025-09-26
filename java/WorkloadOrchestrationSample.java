import com.azure.core.credential.TokenCredential;
import com.azure.core.credential.TokenRequestContext;
import com.azure.core.http.rest.PagedIterable;
import com.azure.core.management.AzureEnvironment;
import com.azure.core.management.exception.ManagementException;
import com.azure.core.management.profile.AzureProfile;
import com.azure.core.util.BinaryData;
import com.azure.identity.DefaultAzureCredentialBuilder;
import com.azure.resourcemanager.workloadorchestration.WorkloadOrchestrationManager;
import com.azure.resourcemanager.workloadorchestration.models.*;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonObject;

import java.io.FileWriter;
import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.function.Supplier;
import java.util.stream.Collectors;
import java.util.Optional;
import java.time.OffsetDateTime;
import com.google.gson.TypeAdapter;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonWriter;
import java.io.IOException;

public class WorkloadOrchestrationSample {

    // Configuration
    private static final String LOCATION = "eastus2euap";
    private static final String TENANT_ID = System.getenv().getOrDefault("AZURE_TENANT_ID", "33e01921-4d64-4f8c-a055-5bda89d835b9");
    private static final String SUBSCRIPTION_ID = System.getenv().getOrDefault("AZURE_SUBSCRIPTION_ID", "973d15c6-6c57-447e-b9c6-6d79b5b784ab");
    private static final String RESOURCE_GROUP = "sdkexamples";
    private static final String CONTEXT_RESOURCE_GROUP = "Mehoopany";
    private static final String CONTEXT_NAME = "Mehoopany-Context";
    private static final String SINGLE_CAPABILITY_NAME = "sdkexamples-soap";

    private static final Random RANDOM = new Random();
    private static final Gson GSON = new GsonBuilder()
        .setPrettyPrinting()
        .registerTypeAdapter(OffsetDateTime.class, new TypeAdapter<OffsetDateTime>() {
            @Override
            public void write(JsonWriter out, OffsetDateTime value) throws IOException {
                out.value(value.toString());
            }

            @Override
            public OffsetDateTime read(JsonReader in) throws IOException {
                return OffsetDateTime.parse(in.nextString());
            }
        }.nullSafe())
        .create();

    public static void main(String[] args) {
        try {
            System.out.println("Authenticating with Azure...");
            TokenCredential credential = new DefaultAzureCredentialBuilder().build();
            AzureProfile profile = new AzureProfile(TENANT_ID, SUBSCRIPTION_ID, AzureEnvironment.AZURE);

            WorkloadOrchestrationManager manager = WorkloadOrchestrationManager
                .configure()
                .authenticate(credential, profile);
            System.out.println("Successfully authenticated with Azure.");

            // STEP 1: Manage Azure context
            System.out.println("==================================================");
            System.out.println("STEP 1: Managing Azure Context with Random Capabilities");
            System.out.println("==================================================");
            List<String> capabilities = manageAzureContext(manager);
            System.out.printf("FINAL CAPABILITY SELECTION: %s%n", capabilities.get(0));
            System.out.println("==================================================");

            System.out.println("\nWaiting 30 seconds after capability selection...");
            TimeUnit.SECONDS.sleep(30);
            System.out.println("Continuing with resource creation...\n");


            System.out.println("==================================================");
            System.out.println("STEP 2: Creating Azure Resources");
            System.out.println("==================================================");

            // Create Schema and Version
            Schema schema = createSchema(manager, RESOURCE_GROUP);
            System.out.printf("Schema created successfully: %s%n", schema.name());
            SchemaVersion schemaVersion = createSchemaVersion(manager, RESOURCE_GROUP, schema.name());
            System.out.printf("Schema version created successfully: %s%n", schemaVersion.name());

            System.out.println("Proceeding with solution template and target creation...\n");

            // Create Solution Template and Version
            SolutionTemplate solutionTemplate = createSolutionTemplate(manager, RESOURCE_GROUP, capabilities);
            System.out.printf("Solution template created successfully: %s%n", solutionTemplate.name());

            String solutionTemplateVersion = createSolutionTemplateVersion(manager, RESOURCE_GROUP, solutionTemplate.name(), schema.name(), schemaVersion.name());
            System.out.printf("Solution template version created successfully: %s%n", solutionTemplateVersion);

            // Create Target
            Target target = createTarget(manager, RESOURCE_GROUP, capabilities);
            System.out.printf("Target created successfully: %s%n", target.name());

            System.out.println("==================================================");
            System.out.println("STEP 3: Setting Configuration Values via Configuration API");
            System.out.println("==================================================");
            String configName = target.name() + "Config";
            String solutionNameForConfig = "sdkexamples-solution123";
            String configVersion = "1.0.0";
            Map<String, Object> configValues = new HashMap<>();
            configValues.put("ErrorThreshold", 35.3);
            configValues.put("HealthCheckEndpoint", "http://localhost:8080/health");
            configValues.put("EnableLocalLog", true);
            configValues.put("AgentEndpoint", "http://localhost:8080/agent");
            configValues.put("HealthCheckEnabled", true);
            configValues.put("ApplicationEndpoint", "http://localhost:8080/app");
            configValues.put("TemperatureRangeMax", 100.5);

            createConfigurationApiCall(credential, SUBSCRIPTION_ID, RESOURCE_GROUP, configName, solutionNameForConfig, configVersion, configValues);

            System.out.println("\n==================================================");
            System.out.println("STEP 3.1: Getting Configuration to verify values");
            System.out.println("==================================================");
            getConfigurationApiCall(credential, SUBSCRIPTION_ID, RESOURCE_GROUP, configName, solutionNameForConfig, configVersion);


            System.out.println("==================================================");
            System.out.println("STEP 4: Review Target Deployment");
            System.out.println("==================================================");
            String solutionVersionId = reviewTarget(manager, RESOURCE_GROUP, target.name(), solutionTemplateVersion);


            System.out.println("==================================================");
            System.out.println("STEP 5: Publish and Install Solution");
            System.out.println("==================================================");
            System.out.println("The workflow has completed the following steps:");
            System.out.println("✓ Context management with capabilities");
            System.out.println("✓ Schema creation");
            System.out.println("✓ Solution template creation");
            System.out.println("✓ Target creation");
            System.out.println("✓ Configuration API calls");
            System.out.println("✓ Target review");
            System.out.println("\nProceeding with publish and install operations...");

            // Publish and Install
            publishTarget(manager, RESOURCE_GROUP, target.name(), solutionVersionId);
            installTarget(manager, RESOURCE_GROUP, target.name(), solutionVersionId);
            
            System.out.println("==================================================");
            System.out.println("STEP 6: Getting Solution Version (Java equivalent)");
            System.out.println("==================================================");
            // getSolutionVersion(manager, RESOURCE_GROUP, target.name(), "sdkexamples-solution123", "sdkbox-m7738-7.1.73.1");

        } catch (Exception e) {
            System.err.println("An unexpected error occurred: " + e.getMessage());
            e.printStackTrace();
        }
    }
    
    private static <T> T retryOperation(Supplier<T> operation) throws InterruptedException {
        int maxAttempts = 3;
        int delaySeconds = 30;
        for (int attempt = 0; attempt < maxAttempts; attempt++) {
            try {
                return operation.get();
            } catch (Exception e) {
                if (attempt == maxAttempts - 1) {
                    throw new RuntimeException("Operation failed after " + maxAttempts + " attempts", e);
                }
                System.out.printf("Attempt %d failed: %s%n", attempt + 1, e.getMessage());
                System.out.printf("Waiting %d seconds before retrying...%n", delaySeconds);
                TimeUnit.SECONDS.sleep(delaySeconds);
                delaySeconds *= 2; // Exponential backoff
            }
        }
        throw new IllegalStateException("Should not reach here");
    }

    private static String generateRandomSemanticVersion() {
        return String.format("%d.%d.%d", RANDOM.nextInt(11), RANDOM.nextInt(21), RANDOM.nextInt(101));
    }

    private static Schema createSchema(WorkloadOrchestrationManager manager, String resourceGroupName) {
        String schemaName = "sdkexamples-schema-v" + generateRandomSemanticVersion();
        return manager.schemas().define(schemaName)
            .withRegion(LOCATION)
            .withExistingResourceGroup(resourceGroupName)
            .withProperties(new SchemaProperties())
            .create();
    }

    private static SchemaVersion createSchemaVersion(WorkloadOrchestrationManager manager, String resourceGroupName, String schemaName) {
        String schemaVersionName = generateRandomSemanticVersion();
        String schemaValue = "rules:\n" +
            "  configs:\n" +
            "    ErrorThreshold:\n" +
            "      type: float\n" +
            "      required: true\n" +
            "      editableAt:\n" +
            "        - line\n" +
            "      editableBy:\n" +
            "        - OT\n" +
            "    HealthCheckEndpoint:\n" +
            "      type: string\n" +
            "      required: false\n" +
            "      editableAt:\n" +
            "        - line\n" +
            "      editableBy:\n" +
            "        - OT\n" +
            "    EnableLocalLog:\n" +
            "      type: boolean\n" +
            "      required: false\n" +
            "      editableAt:\n" +
            "        - line\n" +
            "      editableBy:\n" +
            "        - OT\n" +
            "    AgentEndpoint:\n" +
            "      type: string\n" +
            "      required: false\n" +
            "      editableAt:\n" +
            "        - line\n" +
            "      editableBy:\n" +
            "        - OT\n" +
            "    HealthCheckEnabled:\n" +
            "      type: boolean\n" +
            "      required: false\n" +
            "      editableAt:\n" +
            "        - line\n" +
            "      editableBy:\n" +
            "        - OT\n" +
            "    ApplicationEndpoint:\n" +
            "      type: string\n" +
            "      required: true\n" +
            "      editableAt:\n" +
            "        - line\n" +
            "      editableBy:\n" +
            "        - OT\n" +
            "    TemperatureRangeMax:\n" +
            "      type: float\n" +
            "      required: true\n" +
            "      editableAt:\n" +
            "        - line\n" +
            "      editableBy:\n" +
            "        - OT";

        return manager.schemaVersions().define(schemaVersionName)
            .withExistingSchema(resourceGroupName, schemaName)
            .withProperties(new SchemaVersionProperties().withValue(schemaValue))
            .create();
    }

    private static SolutionTemplate createSolutionTemplate(WorkloadOrchestrationManager manager, String resourceGroupName, List<String> capabilities) {
        String solutionTemplateName = "sdkexamples-solution123";
        return manager.solutionTemplates().define(solutionTemplateName)
            .withRegion(LOCATION)
            .withExistingResourceGroup(resourceGroupName)
            .withProperties(new SolutionTemplateProperties()
                .withCapabilities(capabilities)
                .withDescription("This is Holtmelt Solution with random capabilities"))
            .create();
    }

    private static String createSolutionTemplateVersion(WorkloadOrchestrationManager manager, String resourceGroupName, String solutionTemplateName, String schemaName, String schemaVersion) throws InterruptedException {
        return retryOperation(() -> {
            String version = generateRandomSemanticVersion();
            String configurationsStr = String.format(
                "schema:\n" +
                "  name: %s\n" +
                "  version: %s\n" +
                "configs:\n" +
                "  AppName: Hotmelt\n" +
                "  TemperatureRangeMax: ${{$val(TemperatureRangeMax)}}\n" +
                "  ErrorThreshold: ${{$val(ErrorThreshold)}}\n" +
                "  HealthCheckEndpoint: ${{$val(HealthCheckEndpoint)}}\n" +
                "  EnableLocalLog: ${{$val(EnableLocalLog)}}\n" +
                "  AgentEndpoint: ${{$val(AgentEndpoint)}}\n" +
                "  HealthCheckEnabled: ${{$val(HealthCheckEnabled)}}\n" +
                "  ApplicationEndpoint: ${{$val(ApplicationEndpoint)}}",
                schemaName, schemaVersion);

            Map<String, Object> helmComponent = new HashMap<>();
            helmComponent.put("name", "helmcomponent");
            helmComponent.put("type", "helm.v3");
            Map<String, Object> chartProps = new HashMap<>();
            chartProps.put("repo", "ghcr.io/eclipse-symphony/tests/helm/simple-chart");
            chartProps.put("version", "0.3.0");
            chartProps.put("wait", true);
            chartProps.put("timeout", "5m");
            Map<String, Object> helmProps = new HashMap<>();
            helmProps.put("chart", chartProps);
            helmComponent.put("properties", helmProps);
            
            Map<String, BinaryData> specification = new HashMap<>();
            specification.put("components", BinaryData.fromObject(List.of(helmComponent)));

            SolutionTemplateVersion result = manager.solutionTemplates().createVersion(resourceGroupName, solutionTemplateName,
                new com.azure.resourcemanager.workloadorchestration.fluent.models.SolutionTemplateVersionWithUpdateTypeInner()
                    .withVersion(version)
                    .withSolutionTemplateVersion(
                        new com.azure.resourcemanager.workloadorchestration.fluent.models.SolutionTemplateVersionInner()
                            .withProperties(new SolutionTemplateVersionProperties()
                                .withConfigurations(configurationsStr)
                                .withSpecification(specification)
                                .withOrchestratorType(OrchestratorType.TO))));

                String solutionTemplateVersionId = String.format("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Edge/solutionTemplates/%s/versions/%s",
                    SUBSCRIPTION_ID, resourceGroupName, solutionTemplateName, version);
           
            return solutionTemplateVersionId;
        });
    }
    
    private static Target createTarget(WorkloadOrchestrationManager manager, String resourceGroupName, List<String> capabilities) throws InterruptedException {
        return retryOperation(() -> {
            String targetName = "sdkbox-m23";
            
            Map<String, Object> bindingConfig = Map.of("inCluster", "true");
            Map<String, Object> binding = Map.of("role", "helm.v3", "provider", "providers.target.helm", "config", bindingConfig);
            Map<String, Object> topology = Map.of("bindings", List.of(binding));
            Map<String, BinaryData> targetSpec = new HashMap<>();
            targetSpec.put("topologies", BinaryData.fromObject(List.of(topology)));

            return manager.targets().define(targetName)
                .withRegion(LOCATION)
                .withExistingResourceGroup(resourceGroupName)
                .withExtendedLocation(new ExtendedLocation()
                    .withName("/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/configmanager-cloudtest-playground-portal/providers/Microsoft.ExtendedLocation/customLocations/den-Location")
                    .withType(ExtendedLocationType.CUSTOM_LOCATION))
                .withProperties(new TargetProperties()
                    .withCapabilities(capabilities)
                    .withContextId(String.format("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Edge/contexts/%s", SUBSCRIPTION_ID, CONTEXT_RESOURCE_GROUP, CONTEXT_NAME))
                    .withDescription("This is MK-71 Site with random capabilities")
                    .withDisplayName("sdkbox-mk71")
                    .withHierarchyLevel("line")
                    .withSolutionScope("new")
                    .withTargetSpecification(targetSpec))
                .create(com.azure.core.util.Context.NONE);
        });
    }

    private static String reviewTarget(WorkloadOrchestrationManager manager, String resourceGroupName, String targetName, String solutionTemplateVersionId) throws InterruptedException {
         return retryOperation(() -> {
            System.out.printf("Starting review for target %s%n", targetName);
            SolutionVersion reviewResult = manager.targets().reviewSolutionVersion(
                resourceGroupName,
                targetName,
                new SolutionTemplateParameter()
                    .withSolutionDependencies(Collections.emptyList())
                    .withSolutionInstanceName(targetName)
                    .withSolutionTemplateVersionId(solutionTemplateVersionId)
            );

         try {
            final String solutionName = "sdkexamples-solution123";

            // The equivalent of 'await client.solutionVersions.get(...)'
            PagedIterable<SolutionVersion> solutionVersions = manager.solutionVersions()
                .listBySolution(resourceGroupName, targetName, solutionName);

            System.out.println("Listing all solution versions for solution: " + solutionName);

            // You can now iterate through the results directly
            for (SolutionVersion version : solutionVersions) {
                System.out.println("------------------------------------");
                System.out.printf("Found Solution Version: %s%n", version.name());
                System.out.printf("  ID: %s%n", version.id());
                
                if (version.properties() != null) {
                    System.out.printf("  State: %s%n", version.properties().state());
                    System.out.printf("  Provisioning State: %s%n", version.properties().provisioningState());
                }
            }
            System.out.println("------------------------------------");

            List<SolutionVersion> solutionVersionList = solutionVersions.stream().collect(Collectors.toList());
            System.out.println("All solution versions JSON:");
            
            // Filter to find the entry that matches solutionTemplateVersionId and extract reviewId
            Optional<SolutionVersion> matchingVersion = solutionVersionList.stream()
                .filter(version -> version.properties() != null && 
                        solutionTemplateVersionId.equals(version.properties().solutionTemplateVersionId()))
                .findFirst();
            
            if (matchingVersion.isPresent()) {
                SolutionVersion version = matchingVersion.get();
                String reviewId = version.properties().reviewId();
                System.out.printf("Found matching solution version: %s%n", version.name());
                System.out.printf("Extracted reviewId: %s%n", reviewId);
                System.out.printf("Revision: %s%n", version.properties().revision());
                System.out.printf("State: %s%n", version.properties().state());
                
                // Return the full ID of the solution version
                return version.id();
            } else {
                System.out.printf("No matching solution version found for solutionTemplateVersionId: %s%n", solutionTemplateVersionId);
                System.out.println("Available solution template version IDs:");
                solutionVersionList.forEach(v -> {
                    if (v.properties() != null && v.properties().solutionTemplateVersionId() != null) {
                        System.out.printf("  - %s%n", v.properties().solutionTemplateVersionId());
                    }
                });
                // Fallback to original behavior if no match found
                return reviewResult.id();
            }
            
        } catch (ManagementException e) {
            System.err.printf("Error retrieving solution versions: %s%n", e.getMessage());
        }
        System.out.println("------------------------------------");
        return reviewResult.id();
        });
    }    private static void publishTarget(WorkloadOrchestrationManager manager, String resourceGroupName, String targetName, String solutionVersionId) throws InterruptedException {
        retryOperation(() -> {
            System.out.printf("Publishing solution version %s to target %s%n", solutionVersionId, targetName);
            manager.targets().publishSolutionVersion(
                resourceGroupName,
                targetName,
                new SolutionVersionParameter().withSolutionVersionId(solutionVersionId)
            );
            System.out.println("Publish operation completed successfully.");
            return null; // Supplier must return something
        });
    }

    private static void installTarget(WorkloadOrchestrationManager manager, String resourceGroupName, String targetName, String solutionVersionId) throws InterruptedException {
        retryOperation(() -> {
            System.out.printf("Installing solution version %s on target %s%n", solutionVersionId, targetName);
            manager.targets().installSolution(
                resourceGroupName,
                targetName,
                new InstallSolutionParameter().withSolutionVersionId(solutionVersionId)
            );
             System.out.println("Install operation completed successfully.");
            return null; // Supplier must return something
        });
    }
    
    private static void createConfigurationApiCall(TokenCredential credential, String subscriptionId, String resourceGroup, String configName, String solutionName, String version, Map<String, Object> configValues) throws IOException, InterruptedException {
        String token = credential.getToken(new TokenRequestContext().addScopes("https://management.azure.com/.default")).block().getToken();
        String url = String.format(
            "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Edge/configurations/%s/DynamicConfigurations/%s/versions/version1?api-version=2024-06-01-preview",
            subscriptionId, resourceGroup, configName, solutionName);

        StringBuilder valuesBuilder = new StringBuilder();
        for (Map.Entry<String, Object> entry : configValues.entrySet()) {
            valuesBuilder.append(String.format("%s: %s%n", entry.getKey(), entry.getValue().toString().toLowerCase()));
        }

        JsonObject properties = new JsonObject();
        properties.addProperty("values", valuesBuilder.toString());
        properties.addProperty("provisioningState", "Succeeded");

        JsonObject requestBodyJson = new JsonObject();
        requestBodyJson.add("properties", properties);
        String requestBody = GSON.toJson(requestBodyJson);

        System.out.printf("Making PUT call to Configuration API: %s%n", url);
        System.out.printf("Request body: %s%n", requestBody);
        
        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(url))
            .header("Authorization", "Bearer " + token)
            .header("Content-Type", "application/json")
            .PUT(HttpRequest.BodyPublishers.ofString(requestBody))
            .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

        if (response.statusCode() >= 200 && response.statusCode() < 300) {
            System.out.printf("Configuration API call successful. Status: %d%n", response.statusCode());
            System.out.printf("Response: %s%n", response.body());
        } else {
            throw new IOException(String.format("Configuration API call failed. Status: %d, Response: %s", response.statusCode(), response.body()));
        }
    }

    private static void getConfigurationApiCall(TokenCredential credential, String subscriptionId, String resourceGroup, String configName, String solutionName, String version) throws IOException, InterruptedException {
        String token = credential.getToken(new TokenRequestContext().addScopes("https://management.azure.com/.default")).block().getToken();
        String url = String.format(
            "https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Edge/configurations/%s/DynamicConfigurations/%s/versions/version1?api-version=2024-06-01-preview",
            subscriptionId, resourceGroup, configName, solutionName);
            
        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(url))
            .header("Authorization", "Bearer " + token)
            .header("Content-Type", "application/json")
            .GET()
            .build();

        System.out.printf("Making GET call to Configuration API: %s%n", url);
        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

        if (response.statusCode() == 200) {
            System.out.printf("Configuration GET API call successful. Status: %d%n", response.statusCode());
            System.out.println("Parsed Configuration Data:");
            System.out.println(GSON.toJson(GSON.fromJson(response.body(), JsonObject.class)));
        } else {
             System.out.printf("Configuration GET API call failed. Status: %d, Response: %s%n", response.statusCode(), response.body());
        }
    }
    
    private static List<Capability> getExistingCapabilities(WorkloadOrchestrationManager manager, String resourceGroupName, String contextName) {
        try {
            System.out.printf("Fetching existing context: %s in resource group %s%n", contextName, resourceGroupName);
            ContextModel context = manager.contexts().getByResourceGroup(resourceGroupName, contextName);
            return context.properties().capabilities();
        } catch (ManagementException e) {
            if (e.getResponse().getStatusCode() == 404) {
                System.out.println("Context not found, will create a new one.");
                return new ArrayList<>();
            }
            throw e;
        }
    }
    
    private static List<String> manageAzureContext(WorkloadOrchestrationManager manager) throws InterruptedException {
        List<Capability> existingCapabilities = getExistingCapabilities(manager, CONTEXT_RESOURCE_GROUP, CONTEXT_NAME);
        Set<String> existingNames = existingCapabilities.stream().map(Capability::name).collect(Collectors.toSet());

        String[] capabilityTypes = {"shampoo", "soap"};
        String capType = capabilityTypes[RANDOM.nextInt(capabilityTypes.length)];
        String newCapabilityName = String.format("sdkexamples-%s-%d", capType, 1000 + RANDOM.nextInt(9000));
        
        List<Capability> mergedCapabilities = new ArrayList<>(existingCapabilities);
        
        if (!existingNames.contains(newCapabilityName)) {
            mergedCapabilities.add(new Capability()
                .withName(newCapabilityName)
                .withDescription("SDK generated " + capType + " manufacturing capability")
                .withState(ResourceState.ACTIVE));
             System.out.printf("Added new capability: %s%n", newCapabilityName);
        } else {
             System.out.printf("Capability %s already exists, skipping addition.%n", newCapabilityName);
        }

        createOrUpdateContextWithHierarchies(manager, CONTEXT_RESOURCE_GROUP, CONTEXT_NAME, mergedCapabilities);
        
        return List.of(newCapabilityName);
    }
    
    private static void createOrUpdateContextWithHierarchies(WorkloadOrchestrationManager manager, String resourceGroupName, String contextName, List<Capability> capabilities) throws InterruptedException {
        retryOperation(() -> {
            System.out.printf("Creating/updating context '%s'...%n", contextName);
            manager.contexts().define(contextName)
                .withRegion(LOCATION)
                .withExistingResourceGroup(resourceGroupName)
                .withProperties(new ContextProperties()
                    .withCapabilities(capabilities)
                    .withHierarchies(Arrays.asList(
                        new Hierarchy().withName("country").withDescription("Country level hierarchy"),
                        new Hierarchy().withName("region").withDescription("Regional level hierarchy"),
                        new Hierarchy().withName("factory").withDescription("Factory level hierarchy"),
                        new Hierarchy().withName("line").withDescription("Production line hierarchy")
                    )))
                .create();
            return null;
        });
    }
}
