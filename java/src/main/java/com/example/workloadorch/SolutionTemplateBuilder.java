package com.example.workloadorch;

import com.azure.core.management.Region;
import com.azure.resourcemanager.workloadorchestration.WorkloadOrchestrationManager;
import com.azure.resourcemanager.workloadorchestration.models.*;
import java.util.*;

/**
 * Builder class for creating solution templates with proper validation and error handling.
 */
public class SolutionTemplateBuilder {
    private String name;
    private String region;
    private String resourceGroup;
    private Map<String, String> tags;
    private SolutionTemplateProperties properties;

    public SolutionTemplateBuilder() {
        this.tags = new HashMap<>();
    }

    /**
     * Sets the name of the solution template.
     * @param name Template name
     * @return Builder instance
     * @throws IllegalArgumentException if name is null or empty
     */
    public SolutionTemplateBuilder withName(String name) {
        if (name == null || name.trim().isEmpty()) {
            throw new IllegalArgumentException("Template name cannot be null or empty");
        }
        this.name = name;
        return this;
    }

    /**
     * Sets the region for the solution template.
     * @param region Azure region
     * @return Builder instance
     * @throws IllegalArgumentException if region is invalid
     */
    public SolutionTemplateBuilder withRegion(String region) {
        if (region == null || region.trim().isEmpty()) {
            throw new IllegalArgumentException("Region cannot be null or empty");
        }
        try {
            Region.fromName(region); // Validates if region is valid
            this.region = region;
            return this;
        } catch (Exception e) {
            throw new IllegalArgumentException("Invalid region: " + region, e);
        }
    }

    /**
     * Sets the resource group for the solution template.
     * @param resourceGroup Resource group name
     * @return Builder instance
     * @throws IllegalArgumentException if resource group is null or empty
     */
    public SolutionTemplateBuilder withResourceGroup(String resourceGroup) {
        if (resourceGroup == null || resourceGroup.trim().isEmpty()) {
            throw new IllegalArgumentException("Resource group cannot be null or empty");
        }
        this.resourceGroup = resourceGroup;
        return this;
    }

    /**
     * Adds a tag to the solution template.
     * @param key Tag key
     * @param value Tag value
     * @return Builder instance
     * @throws IllegalArgumentException if key is null or empty
     */
    public SolutionTemplateBuilder withTag(String key, String value) {
        if (key == null || key.trim().isEmpty()) {
            throw new IllegalArgumentException("Tag key cannot be null or empty");
        }
        this.tags.put(key, value);
        return this;
    }

    /**
     * Sets the properties for the solution template.
     * @param properties Template properties
     * @return Builder instance
     * @throws IllegalArgumentException if properties is null
     */
    public SolutionTemplateBuilder withProperties(SolutionTemplateProperties properties) {
        if (properties == null) {
            throw new IllegalArgumentException("Properties cannot be null");
        }
        this.properties = properties;
        return this;
    }

    /**
     * Creates the solution template using the specified manager.
     * @param manager WorkloadOrchestrationManager instance
     * @return Created SolutionTemplate
     * @throws IllegalStateException if required fields are missing
     */
    public SolutionTemplate create(WorkloadOrchestrationManager manager) {
        validateRequiredFields();
        
        return manager.solutionTemplates()
            .define(name)
            .withRegion(Region.fromName(region))
            .withExistingResourceGroup(resourceGroup)
            .withTags(tags)
            .withProperties(properties)
            .create();
    }

    private void validateRequiredFields() {
        List<String> missingFields = new ArrayList<>();
        
        if (name == null) missingFields.add("name");
        if (region == null) missingFields.add("region");
        if (resourceGroup == null) missingFields.add("resourceGroup");
        if (properties == null) missingFields.add("properties");
        
        if (!missingFields.isEmpty()) {
            throw new IllegalStateException(
                "Missing required fields: " + String.join(", ", missingFields));
        }
    }
}
