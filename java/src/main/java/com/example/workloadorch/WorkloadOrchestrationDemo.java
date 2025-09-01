package com.example.workloadorch;

import com.azure.core.exception.ResourceNotFoundException;
import com.azure.core.management.Region;
import com.azure.core.management.profile.AzureProfile;
import com.azure.core.util.BinaryData;
import com.azure.identity.DefaultAzureCredentialBuilder;
import com.azure.resourcemanager.workloadorchestration.WorkloadOrchestrationManager;
import com.azure.resourcemanager.workloadorchestration.models.*;
import com.azure.resourcemanager.workloadorchestration.fluent.models.*;
import com.azure.resourcemanager.workloadorchestration.implementation.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.*;
import java.nio.file.*;
import java.time.Duration;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.concurrent.TimeUnit;
import java.util.Arrays;
import java.util.Collections;
import java.util.Map;
import java.util.HashMap;

public class WorkloadOrchestrationDemo {
    private static final Logger logger = LoggerFactory.getLogger(WorkloadOrchestrationDemo.class);
    
    private static final String SUBSCRIPTION_ID = "973d15c6-6c57-447e-b9c6-6d79b5b784ab";
    private static final String RESOURCE_GROUP = "ConfigManager-CloudTest-Playground-Portal";
    private static final String LOCATION = "eastus2euap";
    private static final int MAX_RETRIES = 3;
    private static final Duration RETRY_DELAY = Duration.ofSeconds(5);

    private WorkloadOrchestrationManager manager;
    private VersionManager versionManager;

    public WorkloadOrchestrationDemo() {
        this.manager = createManager();
        this.versionManager = new VersionManager(manager, RESOURCE_GROUP);
        logger.info("Successfully initialized WorkloadOrchestrationManager");
    }

    protected WorkloadOrchestrationManager createManager() {
        AzureProfile profile = new AzureProfile("72f988bf-86f1-41af-91ab-2d7cd011db47",
                                              SUBSCRIPTION_ID,
                                              com.azure.core.management.AzureEnvironment.AZURE);
        return WorkloadOrchestrationManager.authenticate(
            new DefaultAzureCredentialBuilder().build(),
            profile);
    }


    public Schema createSchema() {
        String timestamp = LocalDateTime.now().format(DateTimeFormatter.ofPattern("MMddHHmm"));
        String schemaName = String.format("test-schema-%s", timestamp);
        logger.info("Attempting to create schema: {}", schemaName);
        
        for (int attempt = 1; attempt <= MAX_RETRIES; attempt++) {
            try {
                Schema schema = manager.schemas()
                    .define(schemaName)
                    .withRegion(Region.fromName(LOCATION))
                    .withExistingResourceGroup(RESOURCE_GROUP)
                    .withProperties(new SchemaProperties())
                    .create();

                logger.info("Successfully created schema: {}", schemaName);
                return schema;
                
            } catch (Exception e) {
                handleRetryableOperation(attempt, "schema", e);
            }
        }
        return null;
    }

    public SchemaVersion createSchemaVersion(String schemaName) {
        String schemaVersionName = versionManager.getNextSchemaVersion(schemaName);
        logger.info("Attempting to create schema version: {}", schemaVersionName);
        
        for (int attempt = 1; attempt <= MAX_RETRIES; attempt++) {
            try {
                SchemaVersion schemaVersion = manager.schemaVersions()
                    .define(schemaVersionName)
                    .withExistingSchema(RESOURCE_GROUP, schemaName)
                    .withProperties(new SchemaVersionProperties().withValue(
                        "rules:\n  configs:\n      ErrorThreshold:\n        type: float\n        required: true\n  "
                    ))
                    .create();

                logger.info("Successfully created schema version: {}", schemaVersionName);
                return schemaVersion;
                
            } catch (ResourceNotFoundException e) {
                logger.error("Schema {} not found. Please ensure schema exists before creating version", schemaName);
                throw e;
            } catch (Exception e) {
                handleRetryableOperation(attempt, "schema version", e);
            }
        }
        return null;
    }

    public SolutionTemplate createSolutionTemplate() {
        String timestamp = LocalDateTime.now().format(DateTimeFormatter.ofPattern("MMddHHmm"));
        String templateName = String.format("my-solution-template-%s", timestamp);
        logger.info("Attempting to create solution template: {}", templateName);
        
        for (int attempt = 1; attempt <= MAX_RETRIES; attempt++) {
            try {
                SolutionTemplateBuilder builder = new SolutionTemplateBuilder()
                    .withName(templateName)
                    .withRegion(LOCATION)
                    .withResourceGroup(RESOURCE_GROUP)
                    .withProperties(new SolutionTemplateProperties()
                        .withCapabilities(Arrays.asList("sdkbox-soap"))
                        .withDescription("This is Test Solution"));

                SolutionTemplate template = builder.create(manager);
                logger.info("Successfully created solution template: {}", templateName);
                return template;
                
            } catch (Exception e) {
                handleRetryableOperation(attempt, "solution template", e);
            }
        }
        return null;
    }

    public SolutionTemplateVersion createSolutionTemplateVersion(
            String templateName, String schemaName, String schemaVersion) {
        String versionName = versionManager.getNextTemplateVersion(templateName);
        templateName = ValidationUtils.validateRequired(templateName, "Template name");
        schemaName = ValidationUtils.validateRequired(schemaName, "Schema name");
        schemaVersion = ValidationUtils.validateVersion(schemaVersion);
        logger.info("Attempting to create solution template version: {}", versionName);
        
        String configStr = String.format(
            "schema:\n  name: %s\n  version: %s\n" +
            "configs:\n" +
            "  AppName: Hotmelt\n" +
            "  TemperatureRangeMax: ${$val(TemperatureRangeMax)}\n" +
            "  ErrorThreshold: ${$val(ErrorThreshold)}\n" +
            "  HealthCheckEndpoint: ${$val(HealthCheckEndpoint)}\n" +
            "  EnableLocalLog: ${$val(EnableLocalLog)}\n" +
            "  AgentEndpoint: ${$val(AgentEndpoint)}\n" +
            "  HealthCheckEnabled: ${$val(HealthCheckEnabled)}\n" +
            "  ApplicationEndpoint: ${$val(ApplicationEndpoint)}\n",
            schemaName, schemaVersion
        );

        Map<String, Object> chartProps = new HashMap<>();
        chartProps.put("repo", "ghcr.io/eclipse-symphony/tests/helm/simple-chart");
        chartProps.put("version", "0.3.0");
        chartProps.put("wait", true);
        chartProps.put("timeout", "5m");

        Map<String, Object> component = new HashMap<>();
        component.put("name", "helmcomponent");
        component.put("type", "helm.v3");
        component.put("properties", Map.of("chart", chartProps));

        for (int attempt = 1; attempt <= MAX_RETRIES; attempt++) {
            try {
                Map<String, BinaryData> specification = Collections.singletonMap(
                    "components", BinaryData.fromObject(Arrays.asList(component)));

                SolutionTemplateVersionProperties versionProps = new SolutionTemplateVersionProperties()
                    .withConfigurations(configStr)
                    .withSpecification(specification)
                    .withOrchestratorType(OrchestratorType.TO);

                SolutionTemplateVersion templateVersion = manager.solutionTemplates()
                    .createVersion(RESOURCE_GROUP, templateName,
                        new SolutionTemplateVersionWithUpdateTypeInner()
                            .withVersion("12.34.3")
                            .withSolutionTemplateVersion(
                                new SolutionTemplateVersionInner()
                                    .withProperties(versionProps)),
                        com.azure.core.util.Context.NONE);

                logger.info("Successfully created solution template version: {}", versionName);
                return templateVersion;
                
            } catch (Exception e) {
                handleRetryableOperation(attempt, "solution template version", e);
            }
        }
        return null;
    }

    public Target createTarget() {
        String targetName = "sdkbox-mk71";
        logger.info("Attempting to create target: {}", targetName);
        
        for (int attempt = 1; attempt <= MAX_RETRIES; attempt++) {
            try {
                ExtendedLocation extLocation = new ExtendedLocation()
                    .withName("/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/configmanager-cloudtest-playground-portal/providers/Microsoft.ExtendedLocation/customLocations/den-Location")
                    .withType(ExtendedLocationType.CUSTOM_LOCATION);

                Map<String, Object> bindingConfig = new HashMap<>();
                bindingConfig.put("inCluster", "true");

                Map<String, Object> binding = new HashMap<>();
                binding.put("role", "helm.v3");
                binding.put("provider", "providers.target.helm");
                binding.put("config", bindingConfig);

                Map<String, Object> topology = new HashMap<>();
                topology.put("bindings", Arrays.asList(binding));

                Target target = manager.targets()
                    .define(targetName)
                    .withRegion(Region.fromName(LOCATION))
                    .withExistingResourceGroup(RESOURCE_GROUP)
                    .withExtendedLocation(extLocation)
                    .withProperties(new TargetProperties()
                        .withCapabilities(Arrays.asList("sdkbox-soap"))
                        .withContextId("/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/Mehoopany/providers/Microsoft.Edge/contexts/Mehoopany-Context")
                        .withDescription("This is MK-71 Site")
                        .withDisplayName("sdkbox-mk71")
                        .withHierarchyLevel("line")
                        .withSolutionScope("new")
                        .withTargetSpecification(Collections.singletonMap(
                            "topologies", BinaryData.fromObject(Arrays.asList(topology))
                        )))
                    .create();

                logger.info("Successfully created target: {}", targetName);
                return target;
                
            } catch (Exception e) {
                handleRetryableOperation(attempt, "target", e);
            }
        }
        return null;
    }

    private void handleRetryableOperation(int attempt, String operationType, Exception e) {
        if (attempt == MAX_RETRIES) {
            logger.error("Failed to create {} after {} attempts: {}", operationType, MAX_RETRIES, e.getMessage());
            throw new RuntimeException(String.format("Failed to create %s", operationType), e);
        }
        
        logger.warn("Attempt {} failed to create {}: {}. Retrying...", attempt, operationType, e.getMessage());
        try {
            TimeUnit.MILLISECONDS.sleep(RETRY_DELAY.toMillis());
        } catch (InterruptedException ie) {
            Thread.currentThread().interrupt();
            throw new RuntimeException("Interrupted during retry delay", ie);
        }
    }

    public static void main(String[] args) {
        try {
            WorkloadOrchestrationDemo demo = new WorkloadOrchestrationDemo();
            logger.info("Starting resource creation process...");

            // Create schema
            Schema schema = demo.createSchema();
            if (schema == null) {
                throw new RuntimeException("Failed to create schema");
            }

            // Create schema version
            SchemaVersion schemaVersion = demo.createSchemaVersion(schema.name());
            if (schemaVersion == null) {
                throw new RuntimeException("Failed to create schema version");
            }

            SolutionTemplate template = demo.createSolutionTemplate();
            if (template == null) {
                throw new RuntimeException("Failed to create solution template");
            }

            // Create solution template version
            SolutionTemplateVersion templateVersion = demo.createSolutionTemplateVersion(
                template.name(), schema.name(), schemaVersion.name());
            if (templateVersion == null) {
                throw new RuntimeException("Failed to create solution template version");
            }

            // Create target
            Target target = demo.createTarget();
            if (target == null) {
                throw new RuntimeException("Failed to create target");
            }

            logger.info("Successfully created all resources");
        } catch (Exception e) {
            logger.error("Application failed: {}", e.getMessage(), e);
            System.exit(1);
        }
    }
}
