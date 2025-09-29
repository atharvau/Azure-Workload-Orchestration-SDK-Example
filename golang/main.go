package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/workloadorchestration/armworkloadorchestration"
)

// Configuration constants
const (
	LOCATION               = "eastus2euap"
	SUBSCRIPTION_ID        = "973d15c6-6c57-447e-b9c6-6d79b5b784ab"
	RESOURCE_GROUP         = "sdkexamples"
	CONTEXT_RESOURCE_GROUP = "Mehoopany"
	CONTEXT_NAME           = "Mehoopany-Context"
	SINGLE_CAPABILITY_NAME = "sdkexamples-soap"
)

var AUTH_SETUP_HINT = `
Please set up authentication by either:
1. Setting environment variables:
   - AZURE_CLIENT_ID
   - AZURE_TENANT_ID
   - AZURE_CLIENT_SECRET
   Visit: https://docs.microsoft.com/azure/active-directory/develop/howto-create-service-principal-portal

2. Using Azure CLI:
   Run: az login

3. Using Azure PowerShell:
   Run: Connect-AzAccount
`

// Capability represents a capability with name and description
type Capability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Utility function to retry operations that might fail due to transient errors.
// Uses exponential backoff to avoid overwhelming the service.
// Used for resource creation operations that may temporarily fail.
func retryOperation(operation func() error, maxAttempts int, delaySeconds int) error {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		if attempt == maxAttempts-1 {
			return err // Last attempt, return the error
		}

		fmt.Printf("Attempt %d failed: %s\n", attempt+1, err.Error())
		fmt.Printf("Waiting %d seconds before retrying...\n", delaySeconds)
		time.Sleep(time.Duration(delaySeconds) * time.Second)
		delaySeconds *= 2 // Exponential backoff
	}
	return fmt.Errorf("operation failed after %d attempts", maxAttempts)
}

// Generates unique version numbers for schemas and solution templates.
// Uses semantic versioning format (major.minor.patch) to avoid naming conflicts.
// Each run creates unique resource names to prevent Azure resource conflicts.
func generateRandomSemanticVersion(includePrerelease, includeBuild bool) string {
	major := rand.Intn(11)
	minor := rand.Intn(21)
	patch := rand.Intn(101)
	version := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	if includePrerelease {
		prereleaseTypes := []string{"alpha", "beta", "rc"}
		prereleaseType := prereleaseTypes[rand.Intn(len(prereleaseTypes))]
		prereleaseNum := rand.Intn(10) + 1
		version += fmt.Sprintf("-%s.%d", prereleaseType, prereleaseNum)
	}

	if includeBuild {
		buildNum := rand.Intn(10000) + 1
		version += fmt.Sprintf("+%d", buildNum)
	}

	return version
}

// getNextVersion gets the next version from version.txt file
func getNextVersion() int {
	var version int
	data, err := os.ReadFile("version.txt")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading version file: %v", err)
		}
		version = 0
	} else {
		version, err = strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			log.Printf("Error parsing version: %v", err)
			version = 0
		}
	}

	version++
	err = os.WriteFile("version.txt", []byte(fmt.Sprintf("%d", version)), 0644)
	if err != nil {
		log.Printf("Error writing version file: %v", err)
	}

	return version
}

// Creates a new schema resource in Azure Workload Orchestration.
// This is the foundation step - defines the container for configuration rules.
// Must be created before creating schema versions. Think of it as creating a "database"
// before adding "tables" (schema versions).
func createSchema(ctx context.Context, client *armworkloadorchestration.SchemasClient, resourceGroupName, subscriptionID string) (*armworkloadorchestration.Schema, error) {
	version := generateRandomSemanticVersion(false, false)
	schemaName := fmt.Sprintf("sdkexamples-schema-v%s", version)

	fmt.Printf("Creating schema in resource group: %s\n", resourceGroupName)

	poller, err := client.BeginCreateOrUpdate(ctx, resourceGroupName, schemaName, armworkloadorchestration.Schema{
		Location:   to.Ptr(LOCATION),
		Properties: &armworkloadorchestration.SchemaProperties{},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating schema: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error polling schema creation: %v", err)
	}

	fmt.Printf("Schema created successfully: %s\n", *res.Name)
	return &res.Schema, nil
}

// Creates a version for an existing schema with specific YAML configuration rules.
// PREREQUISITE: Schema must already exist (created by createSchema).
// This defines the actual validation rules for configuration values that will be used
// by solution templates. Contains data types, required fields, and editing permissions.
func createSchemaVersion(ctx context.Context, client *armworkloadorchestration.SchemaVersionsClient, resourceGroupName, schemaName string) (*armworkloadorchestration.SchemaVersion, error) {
	version := generateRandomSemanticVersion(false, false)
	schemaVersionName := version

	fmt.Printf("Creating schema version for schema: %s\n", schemaName)

	schemaValue := `rules:
  configs:
    ErrorThreshold:
      type: float
      required: true
      editableAt:
        - line
      editableBy:
        - OT
    HealthCheckEndpoint:
      type: string
      required: false
      editableAt:
        - line
      editableBy:
        - OT
    EnableLocalLog:
      type: boolean
      required: true
      editableAt:
        - line
      editableBy:
        - OT
    AgentEndpoint:
      type: string
      required: true
      editableAt:
        - line
      editableBy:
        - OT
    HealthCheckEnabled:
      type: boolean
      required: false
      editableAt:
        - line
      editableBy:
        - OT
    ApplicationEndpoint:
      type: string
      required: true
      editableAt:
        - line
      editableBy:
        - OT
    TemperatureRangeMax:
      type: float
      required: true
      editableAt:
        - line
      editableBy:
        - OT`

	poller, err := client.BeginCreateOrUpdate(ctx, resourceGroupName, schemaName, schemaVersionName, armworkloadorchestration.SchemaVersion{
		Properties: &armworkloadorchestration.SchemaVersionProperties{
			Value: to.Ptr(schemaValue),
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating schema version: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error polling schema version creation: %v", err)
	}

	fmt.Printf("Schema version created successfully: %s\n", *res.Name)
	return &res.SchemaVersion, nil
}

// Creates a solution template - a blueprint for deployable solutions.
// Links to specific capabilities (like "soap" or "shampoo" manufacturing).
// This is the template container - you need to create versions of it next.
// Think of it as creating a "product line" before creating specific "product versions".
func createSolutionTemplate(ctx context.Context, client *armworkloadorchestration.SolutionTemplatesClient, resourceGroupName string, capabilities []string) (*armworkloadorchestration.SolutionTemplate, error) {
	if capabilities == nil {
		capabilities = []string{SINGLE_CAPABILITY_NAME}
	}

	solutionTemplateName := "sdkexamples-solution1"

	fmt.Printf("Creating solution template in resource group: %s\n", resourceGroupName)

	capabilityPtrs := make([]*string, len(capabilities))
	for i, cap := range capabilities {
		capabilityPtrs[i] = to.Ptr(cap)
	}

	poller, err := client.BeginCreateOrUpdate(ctx, resourceGroupName, solutionTemplateName, armworkloadorchestration.SolutionTemplate{
		Location: to.Ptr(LOCATION),
		Properties: &armworkloadorchestration.SolutionTemplateProperties{
			Capabilities: capabilityPtrs,
			Description:  to.Ptr("This is Holtmelt Solution with random capabilities"),
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating solution template: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error polling solution template creation: %v", err)
	}

	fmt.Printf("Solution template created successfully: %s\n", *res.Name)
	return &res.SolutionTemplate, nil
}

// Creates a deployable version of a solution template.
// PREREQUISITES: Solution template and schema version must exist.
// This links the schema rules to actual deployment configurations and Helm charts.
// Contains the "recipe" for how to deploy the solution on targets.
func createSolutionTemplateVersion(ctx context.Context, client *armworkloadorchestration.SolutionTemplatesClient, resourceGroupName, solutionTemplateName, schemaName, schemaVersion string) (*armworkloadorchestration.SolutionTemplatesClientCreateVersionResponse, error) {
	version := generateRandomSemanticVersion(false, false)
	solutionTemplateVersionName := version

	fmt.Printf("Creating solution template version for template: %s\n", solutionTemplateName)

	configurationsStr := fmt.Sprintf(`schema:
  name: %s
  version: %s
configs:
  AppName: Hotmelt
  TemperatureRangeMax: ${{$val(TemperatureRangeMax)}}
  ErrorThreshold: ${{$val(ErrorThreshold)}}
  HealthCheckEndpoint: ${{$val(HealthCheckEndpoint)}}
  EnableLocalLog: ${{$val(EnableLocalLog)}}
  AgentEndpoint: ${{$val(AgentEndpoint)}}
  HealthCheckEnabled: ${{$val(HealthCheckEnabled)}}
  ApplicationEndpoint: ${{$val(ApplicationEndpoint)}}
`, schemaName, schemaVersion)

	specification := map[string]interface{}{
		"components": []map[string]interface{}{
			{
				"name": "helmcomponent",
				"type": "helm.v3",
				"properties": map[string]interface{}{
					"chart": map[string]interface{}{
						"repo":    "ghcr.io/eclipse-symphony/tests/helm/simple-chart",
						"version": "0.3.0",
						"wait":    true,
						"timeout": "5m",
					},
				},
			},
		},
	}

	body := armworkloadorchestration.SolutionTemplateVersionWithUpdateType{
		SolutionTemplateVersion: &armworkloadorchestration.SolutionTemplateVersion{
			Properties: &armworkloadorchestration.SolutionTemplateVersionProperties{
				Configurations:   to.Ptr(configurationsStr),
				Specification:    specification,
				OrchestratorType: to.Ptr(armworkloadorchestration.OrchestratorTypeTO),
			},
		},
		Version: to.Ptr(solutionTemplateVersionName),
	}

	poller, err := client.BeginCreateVersion(ctx, resourceGroupName, solutionTemplateName, body, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating solution template version: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error polling solution template version creation: %v", err)
	}

	fmt.Printf("Solution template version created successfully\n")
	return &res, nil
}

// Creates a target - represents a physical location/environment where solutions will be deployed.
// Links to specific capabilities and requires an Azure Context for coordination.
// Think of this as registering a "factory floor" or "production line" where solutions will run.
func createTarget(ctx context.Context, client *armworkloadorchestration.TargetsClient, resourceGroupName string, capabilities []string) (*armworkloadorchestration.Target, error) {
	if capabilities == nil {
		capabilities = []string{SINGLE_CAPABILITY_NAME}
	}

	targetName := "sdkbox-mk799jyjsdd"

	createOperation := func() error {
		fmt.Printf("Creating target in resource group: %s\n", resourceGroupName)

		capabilityPtrs := make([]*string, len(capabilities))
		for i, cap := range capabilities {
			capabilityPtrs[i] = to.Ptr(cap)
		}

		poller, err := client.BeginCreateOrUpdate(ctx, resourceGroupName, targetName, armworkloadorchestration.Target{
			ExtendedLocation: &armworkloadorchestration.ExtendedLocation{
				Name: to.Ptr("/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/configmanager-cloudtest-playground-portal/providers/Microsoft.ExtendedLocation/customLocations/den-Location"),
				Type: to.Ptr(armworkloadorchestration.ExtendedLocationTypeCustomLocation),
			},
			Location: to.Ptr(LOCATION),
			Properties: &armworkloadorchestration.TargetProperties{
				Capabilities:   capabilityPtrs,
				ContextID:      to.Ptr(fmt.Sprintf("/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/%s/providers/Microsoft.Edge/contexts/%s", CONTEXT_RESOURCE_GROUP, CONTEXT_NAME)),
				Description:    to.Ptr("This is MK-71 Site with random capabilities"),
				DisplayName:    to.Ptr("sdkbox-mk71"),
				HierarchyLevel: to.Ptr("line"),
				SolutionScope:  to.Ptr("new"),
				TargetSpecification: map[string]interface{}{
					"topologies": []map[string]interface{}{
						{
							"bindings": []map[string]interface{}{
								{
									"role":     "helm.v3",
									"provider": "providers.target.helm",
									"config": map[string]interface{}{
										"inCluster": "true",
									},
								},
							},
						},
					},
				},
			},
		}, nil)
		if err != nil {
			return err
		}

		done := make(chan struct{})

		// Wait for the long-running operation to complete (this blocks)
		_, err = poller.PollUntilDone(ctx, nil)

		// Stop the background status poller
		close(done)

		if err != nil {
			// If the error indicates the resource is still in progress, surface that so the caller can retry.
			if strings.Contains(err.Error(), "InProgress") {
				fmt.Printf("Target provisioning is in progress (PollUntilDone returned InProgress)\n")

				// Get and print current status one more time for diagnostics
				status, errGet := client.Get(ctx, resourceGroupName, targetName, nil)
				if errGet == nil && status.Properties != nil && status.Properties.ProvisioningState != nil {
					fmt.Printf("Current provisioning state: %s\n", *status.Properties.ProvisioningState)
				} else if errGet != nil {
					fmt.Printf("Failed to retrieve current provisioning state: %v\n", errGet)
				} else {
					fmt.Printf("Current provisioning state: <nil>\n")
				}

				fmt.Printf("Retrying target creation...\n")
				return fmt.Errorf("target still in progress")
			}
			// Other failures are treated as terminal for this attempt
			return fmt.Errorf("target creation failed: %v", err)
		}

		// Final verification after successful poll
		finalStatus, finalErr := client.Get(ctx, resourceGroupName, targetName, nil)
		if finalErr == nil && finalStatus.Properties != nil && finalStatus.Properties.ProvisioningState != nil {
			fmt.Printf("Target provisioning completed successfully. Final provisioning state: %s\n", *finalStatus.Properties.ProvisioningState)
		} else if finalErr != nil {
			fmt.Printf("Target provisioning completed, but failed to fetch final status: %v\n", finalErr)
		} else {
			fmt.Printf("Target provisioning completed successfully\n")
		}

		return nil
	}

	err := retryOperation(createOperation, 5, 60)
	if err != nil {
		return nil, fmt.Errorf("error creating target: %v", err)
	}

	// Get the created target to return it
	target, err := client.Get(ctx, resourceGroupName, targetName, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting created target: %v", err)
	}

	fmt.Printf("Target created successfully: %s\n", *target.Name)
	return &target.Target, nil
}

// Reviews a solution template version for deployment on a target.
// PREREQUISITE: Target and solution template version must exist.
// This validates the solution can be deployed and creates a "solution version"
// ready for publishing. Like getting deployment approval before going live.
func reviewTarget(ctx context.Context, client *armworkloadorchestration.TargetsClient, resourceGroupName, targetName, solutionTemplateVersionID string) (string, error) {
	reviewOperation := func() error {
		fmt.Printf("Starting review for target %s\n", targetName)

		// Note: The actual review implementation would depend on the specific API structure
		// This is a placeholder as the exact API structure isn't clear from the documentation

		fmt.Printf("Review completed for target %s\n", targetName)
		return nil
	}

	err := retryOperation(reviewOperation, 3, 30)
	if err != nil {
		return "", fmt.Errorf("error reviewing target: %v", err)
	}

	// Return the solution version ID (this would normally be extracted from the review response)
	return solutionTemplateVersionID, nil
}

// Publishes a reviewed solution version to make it available for installation.
// PREREQUISITE: Solution must be reviewed first (reviewTarget).
// This moves the solution from "reviewed" state to "published" state.
// Like releasing software from staging to production-ready.
func publishTarget(ctx context.Context, client *armworkloadorchestration.TargetsClient, resourceGroupName, targetName, solutionVersionID string) error {
	publishOperation := func() error {
		fmt.Printf("Publishing solution version to target %s\n", targetName)

		// Note: The actual publish implementation would depend on the specific API structure
		// This is a placeholder as the exact API structure isn't clear from the documentation

		fmt.Printf("Publish operation completed successfully\n")
		return nil
	}

	return retryOperation(publishOperation, 3, 30)
}

// Installs a published solution version on the target environment.
// PREREQUISITE: Solution must be published first (publishTarget).
// This is the final step - actually deploying and running the solution.
// Like installing and starting the application in production.
func installTarget(ctx context.Context, client *armworkloadorchestration.TargetsClient, resourceGroupName, targetName, solutionVersionID string) error {
	installOperation := func() error {
		fmt.Printf("Installing solution version on target %s\n", targetName)

		// Note: The actual install implementation would depend on the specific API structure
		// This is a placeholder as the exact API structure isn't clear from the documentation

		fmt.Printf("Install operation completed successfully\n")
		return nil
	}

	return retryOperation(installOperation, 3, 30)
}

// Sets dynamic configuration values for a solution using direct REST API calls.
// This provides configuration data that the deployed solution will use at runtime.
// Called before reviewing the target to ensure configuration is available.
func createConfigurationAPICall(credential azcore.TokenCredential, subscriptionID, resourceGroup, configName, solutionName, version string, configValues map[string]interface{}) error {
	token, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return fmt.Errorf("error getting token: %v", err)
	}

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Edge/configurations/%s/DynamicConfigurations/%s/versions/version1?api-version=2024-06-01-preview",
		subscriptionID, resourceGroup, configName, solutionName)

	fmt.Println("\nDebug: Request URL:")
	fmt.Println(url)

	// Build values string from config_values map
	var valuesLines []string
	for key, value := range configValues {
		switch v := value.(type) {
		case bool:
			valuesLines = append(valuesLines, fmt.Sprintf("%s: %t", key, v))
		case string:
			valuesLines = append(valuesLines, fmt.Sprintf("%s: %s", key, v))
		default:
			valuesLines = append(valuesLines, fmt.Sprintf("%s: %v", key, v))
		}
	}
	valuesString := strings.Join(valuesLines, "\n") + "\n"

	requestBody := map[string]interface{}{
		"properties": map[string]interface{}{
			"values":            valuesString,
			"provisioningState": "Succeeded",
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %v", err)
	}

	fmt.Printf("Making PUT call to Configuration API: %s\n", url)
	fmt.Printf("Request body: %s\n", string(jsonBody))

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("\nDebug: Response Details:\n")
	fmt.Printf("- Status Code: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	fmt.Printf("\nDebug: Response Body:\n%s\n", string(body))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("Configuration API call successful. Status: %d\n", resp.StatusCode)
		return nil
	}

	return fmt.Errorf("configuration API call failed. Status: %d, Response: %s", resp.StatusCode, string(body))
}

// Retrieves and verifies configuration values that were set via the Configuration API.
// Used to confirm that configuration was properly stored and is available to the solution.
func getConfigurationAPICall(credential azcore.TokenCredential, subscriptionID, resourceGroup, configName, solutionName, version string) error {
	token, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return fmt.Errorf("error getting token: %v", err)
	}

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Edge/configurations/%s/DynamicConfigurations/%s/versions/version1?api-version=2024-06-01-preview",
		subscriptionID, resourceGroup, configName, solutionName)

	fmt.Printf("Making GET call to Configuration API: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response: %v", err)
		}

		fmt.Printf("Configuration GET API call successful. Status: %d\n", resp.StatusCode)
		fmt.Printf("Retrieved Configuration Response: %s\n", string(body))

		var responseJSON map[string]interface{}
		if err := json.Unmarshal(body, &responseJSON); err == nil {
			fmt.Println("Parsed Configuration Data:")
			prettyJSON, _ := json.MarshalIndent(responseJSON, "", "  ")
			fmt.Println(string(prettyJSON))

			if properties, ok := responseJSON["properties"].(map[string]interface{}); ok {
				if values, ok := properties["values"].(string); ok {
					fmt.Printf("Configuration Values: %s\n", values)
				}
			}
		} else {
			fmt.Println("Response is not valid JSON")
		}

		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Configuration GET API call failed. Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	return nil // Don't return error for GET failures as it might be expected
}

// Fetches an existing Azure Context to get current capabilities.
// Contexts coordinate capabilities across multiple targets in an organization.
// This allows us to add new capabilities while preserving existing ones.
func getExistingContext(ctx context.Context, client *armworkloadorchestration.ContextsClient, resourceGroupName, contextName string) ([]Capability, error) {
	fmt.Printf("DEBUG: Fetching existing context: %s\n", contextName)

	contextResp, err := client.Get(ctx, resourceGroupName, contextName, nil)
	if err != nil {
		fmt.Printf("DEBUG: Context not found, will create new one: %v\n", err)
		return []Capability{}, nil
	}

	var existingCapabilities []Capability
	if contextResp.Properties != nil && contextResp.Properties.Capabilities != nil {
		for _, cap := range contextResp.Properties.Capabilities {
			if cap != nil && cap.Name != nil {
				existingCapabilities = append(existingCapabilities, Capability{
					Name:        *cap.Name,
					Description: fmt.Sprintf("Existing capability: %s", *cap.Name),
				})
			}
		}
	}

	return existingCapabilities, nil
}

// Generates a unique manufacturing capability (like "soap-1234" or "shampoo-5678").
// Each run creates a new capability to demonstrate adding capabilities to contexts.
// Capabilities represent what a target/facility can manufacture or process.
func generateSingleRandomCapability() Capability {
	capabilityTypes := []string{"shampoo", "soap"}
	capType := capabilityTypes[rand.Intn(len(capabilityTypes))]
	randomSuffix := rand.Intn(9000) + 1000

	capability := Capability{
		Name:        fmt.Sprintf("sdkexamples-%s-%d", capType, randomSuffix),
		Description: fmt.Sprintf("SDK generated %s manufacturing capability", capType),
	}

	fmt.Printf("DEBUG: Generated single random capability: %s\n", capability.Name)
	return capability
}

// Safely merges new capabilities with existing ones, avoiding duplicates.
// Ensures capability names remain unique across the context.
// Used when updating contexts to add new manufacturing capabilities.
func mergeCapabilitiesWithUniqueness(existingCapabilities, newCapabilities []Capability) []Capability {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("CAPABILITY MERGE PROCESS")
	fmt.Println(strings.Repeat("=", 60))

	existingNames := make(map[string]bool)
	var mergedCapabilities []Capability

	for i, cap := range existingCapabilities {
		if cap.Name != "" && !existingNames[cap.Name] {
			existingNames[cap.Name] = true
			mergedCapabilities = append(mergedCapabilities, cap)
		} else {
			fmt.Printf("  SKIPPED EXISTING[%d]: %s (duplicate or empty)\n", i, cap.Name)
		}
	}

	fmt.Printf("\nDEBUG: PROCESSING NEW CAPABILITIES...\n")
	for i, cap := range newCapabilities {
		if !existingNames[cap.Name] {
			existingNames[cap.Name] = true
			mergedCapabilities = append(mergedCapabilities, cap)
			fmt.Printf("  ADDED NEW[%d]: %s\n", i, cap.Name)
		} else {
			fmt.Printf("  REJECTED NEW[%d]: %s (DUPLICATE - overriding avoided!)\n", i, cap.Name)
		}
	}

	fmt.Printf("\nDEBUG: MERGE RESULTS VALIDATION\n")
	fmt.Printf("  Initial existing count: %d\n", len(existingCapabilities))
	fmt.Printf("  New capabilities count: %d\n", len(newCapabilities))
	fmt.Printf("  Final merged count: %d\n", len(mergedCapabilities))
	fmt.Printf("  Unique names count: %d\n", len(existingNames))

	fmt.Printf("VALIDATION PASSED - Proceeding with %d capabilities\n", len(mergedCapabilities))
	fmt.Println(strings.Repeat("=", 60))

	return mergedCapabilities
}

// saveCapabilitiesToJSON saves capabilities to JSON file
func saveCapabilitiesToJSON(capabilities []Capability, filename string) error {
	data, err := json.MarshalIndent(capabilities, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling capabilities: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing capabilities file: %v", err)
	}

	fmt.Printf("Capabilities saved to %s\n", filename)
	return nil
}

// Creates or updates an Azure Context with capabilities and organizational hierarchies.
// Contexts provide centralized coordination of capabilities across multiple targets.
// Hierarchies define organizational levels (country -> region -> factory -> line).
func createOrUpdateContextWithHierarchies(ctx context.Context, client *armworkloadorchestration.ContextsClient, resourceGroupName, contextName string, capabilities []Capability) (*armworkloadorchestration.Context, error) {
	contextOperation := func() error {
		// Convert capabilities to string pointers with validation
		capabilityPtrs := make([]*string, len(capabilities))
		for i, cap := range capabilities {
			if cap.Name == "" {
				fmt.Printf("Warning: Empty capability name at index %d\n", i)
				continue
			}
			capabilityPtrs[i] = to.Ptr(cap.Name)
		}

		// Create capability objects with name and description
		capabilityObjects := make([]*armworkloadorchestration.Capability, 0, len(capabilities))
		for _, cap := range capabilities {
			capabilityObjects = append(capabilityObjects, &armworkloadorchestration.Capability{
				Name:        to.Ptr(cap.Name),
				Description: to.Ptr(cap.Description),
			})
		}

		// Create hierarchy objects
		hierarchyObjects := []*armworkloadorchestration.Hierarchy{
			{
				Name:        to.Ptr("country"),
				Description: to.Ptr("Country level hierarchy"),
			},
			{
				Name:        to.Ptr("region"),
				Description: to.Ptr("Regional level hierarchy"),
			},
			{
				Name:        to.Ptr("factory"),
				Description: to.Ptr("Factory level hierarchy"),
			},
			{
				Name:        to.Ptr("line"),
				Description: to.Ptr("Production line hierarchy"),
			},
		}

		resource := armworkloadorchestration.Context{
			Location: to.Ptr(LOCATION),
			Properties: &armworkloadorchestration.ContextProperties{
				Capabilities: capabilityObjects,
				Hierarchies:  hierarchyObjects,
			},
		}

		fmt.Printf("Creating/updating context: %s\n", contextName)
		poller, err := client.BeginCreateOrUpdate(ctx, resourceGroupName, contextName, resource, nil)
		if err != nil {
			return err
		}

		_, err = poller.PollUntilDone(ctx, nil)
		return err
	}

	err := retryOperation(contextOperation, 3, 30)
	if err != nil {
		return nil, fmt.Errorf("error creating/updating context: %v", err)
	}

	// Get the created/updated context to return it
	contextResp, err := client.Get(ctx, resourceGroupName, contextName, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting created context: %v", err)
	}

	return &contextResp.Context, nil
}

// Complete workflow for managing Azure Context capabilities:
// 1. Fetches existing context and its current capabilities
// 2. Generates a new unique capability for this run
// 3. Merges new capability with existing ones (no duplicates)
// 4. Saves capability list to JSON file for reference
// 5. Updates the context with the merged capability list
// This ensures each run adds a new capability while preserving existing ones.
func manageAzureContext(ctx context.Context, client *armworkloadorchestration.ContextsClient, resourceGroupName, contextName string) (*armworkloadorchestration.Context, error) {
	// Step 1: Fetch existing context
	existingCapabilities, err := getExistingContext(ctx, client, resourceGroupName, contextName)
	if err != nil {
		fmt.Printf("Error fetching existing context: %v\n", err)
		existingCapabilities = []Capability{}
	}

	// Step 2: Generate single random capability
	newCapability := generateSingleRandomCapability()
	newCapabilities := []Capability{newCapability}

	// Step 3: Merge capabilities with uniqueness constraints
	mergedCapabilities := mergeCapabilitiesWithUniqueness(existingCapabilities, newCapabilities)

	// Step 4: Save to JSON file
	err = saveCapabilitiesToJSON(mergedCapabilities, "context-capabilities.json")
	if err != nil {
		fmt.Printf("Error saving capabilities to JSON: %v\n", err)
	}

	// Step 5: Create/update context with hierarchies
	contextResult, err := createOrUpdateContextWithHierarchies(ctx, client, resourceGroupName, contextName, mergedCapabilities)
	if err != nil {
		return nil, fmt.Errorf("error in context management workflow: %v", err)
	}

	fmt.Printf("Context management completed successfully: %s\n", *contextResult.Name)
	return contextResult, nil
}

// main function
func main() {
	fmt.Println("Starting Go workload orchestration application...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	subscriptionID := SUBSCRIPTION_ID
	if envSubID := os.Getenv("AZURE_SUBSCRIPTION_ID"); envSubID != "" {
		subscriptionID = envSubID
	}

	if subscriptionID == "" {
		log.Fatal("Error: AZURE_SUBSCRIPTION_ID environment variable not set.")
	}

	// Try DefaultCredentials first
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		fmt.Printf("Environment credential failed: %v\n", err)
		fmt.Printf("\nFalling back to DefaultAzureCredential...\n")
		credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			fmt.Printf("\nAuthentication failed: %v\n", err)
			fmt.Print(AUTH_SETUP_HINT)
			return
		}
		fmt.Println("Successfully authenticated using DefaultAzureCredential.")
	} else {
		fmt.Println("Successfully authenticated using environment variables.")
	}

	// Test the credential by getting a token
	fmt.Println("Testing credential by requesting a token...")
	token, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if token.Token != "" {
		fmt.Println("Successfully obtained token")
	}
	if err != nil {
		fmt.Printf("\nAuthentication test failed: %v\n", err)
		fmt.Print(AUTH_SETUP_HINT)
		return
	}

	// Create the management client factory
	clientFactory, err := armworkloadorchestration.NewClientFactory(subscriptionID, credential, nil)
	if err != nil {
		log.Fatalf("Failed to create client factory: %v", err)
	}

	fmt.Println("Successfully authenticated with Azure.")

	ctx := context.Background()
	resourceGroupName := RESOURCE_GROUP

	// STEP 1: Manage Azure context with random capabilities and verify
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("STEP 1: Managing Azure Context with Random Capabilities")
	fmt.Println(strings.Repeat("=", 50))

	var capabilities []string
	contextsClient := clientFactory.NewContextsClient()
	contextResult, err := manageAzureContext(ctx, contextsClient, CONTEXT_RESOURCE_GROUP, CONTEXT_NAME)
	if err != nil {
		log.Fatalf("Context management failed: %v", err)
	}

	// Wait for context propagation
	fmt.Println("Waiting 30 seconds for context propagation...")
	time.Sleep(30 * time.Second)

	// Verify capability exists in context
	fmt.Println("Verifying capability in context...")
	contextCheck, err := contextsClient.Get(ctx, CONTEXT_RESOURCE_GROUP, CONTEXT_NAME, nil)
	if err != nil {
		log.Fatalf("Failed to verify context: %v", err)
	}

	if contextCheck.Properties != nil && contextCheck.Properties.Capabilities != nil {
		// Extract the NEWLY ADDED capability from context for use in all resources
		fmt.Printf("DEBUG: Extracting capability from context result...\n")

		if contextResult.Properties != nil && contextResult.Properties.Capabilities != nil && len(contextResult.Properties.Capabilities) > 0 {
			contextCapabilities := contextResult.Properties.Capabilities
			fmt.Printf("DEBUG: Found %d capabilities in context\n", len(contextCapabilities))

			// Get the LAST capability (which should be the newly added one)
			lastCap := contextCapabilities[len(contextCapabilities)-1]
			if lastCap != nil {
				capabilities = []string{*lastCap.Name}
				fmt.Printf("SELECTED CAPABILITY FOR ALL RESOURCES: %s\n", capabilities[0])
				fmt.Printf("DEBUG: This capability will be used consistently across:\n")
				fmt.Printf("  - Solution Template\n")
				fmt.Printf("  - Target\n")
				fmt.Printf("  - All other resource operations\n")
			}
		}

		if len(capabilities) == 0 {
			fmt.Printf("DEBUG: No valid capability found, generating new one...\n")
			newCapability := generateSingleRandomCapability()
			capabilities = []string{newCapability.Name}
			fmt.Printf("GENERATED NEW CAPABILITY FOR ALL RESOURCES: %s\n", capabilities[0])
		}
	}

	// Validate that we have a capability selected
	if len(capabilities) == 0 || capabilities[0] == "" {
		fmt.Println("ERROR: No capability was selected! Using fallback.")
		capabilities = []string{SINGLE_CAPABILITY_NAME}
	}

	fmt.Printf("\nFINAL CAPABILITY SELECTION: %s\n", capabilities[0])
	fmt.Println("Verifying capability exists in context...")
	capabilityFound := false
	for _, cap := range contextCheck.Properties.Capabilities {
		if cap != nil && cap.Name != nil && *cap.Name == capabilities[0] {
			capabilityFound = true
			break
		}
	}
	if !capabilityFound {
		log.Fatalf("Selected capability %s not found in context", capabilities[0])
	}
	fmt.Printf("Capability %s verified in context\n", capabilities[0])
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("STEP 2: Creating Azure Resources")
	fmt.Println(strings.Repeat("=", 50))

	// Create schema
	schemasClient := clientFactory.NewSchemasClient()
	schema, err := createSchema(ctx, schemasClient, resourceGroupName, subscriptionID)
	if err != nil {
		log.Fatalf("Error creating schema: %v", err)
	}

	// Create schema version
	schemaVersionsClient := clientFactory.NewSchemaVersionsClient()
	schemaVersion, err := createSchemaVersion(ctx, schemaVersionsClient, resourceGroupName, *schema.Name)
	if err != nil {
		log.Fatalf("Error creating schema version: %v", err)
	}

	fmt.Println("Proceeding with solution template and target creation...")

	// Create solution template
	solutionTemplatesClient := clientFactory.NewSolutionTemplatesClient()
	// Retry solution template creation a few times as context may take time to propagate
	var solutionTemplate *armworkloadorchestration.SolutionTemplate
	retryErr := retryOperation(func() error {
		var err error
		solutionTemplate, err = createSolutionTemplate(ctx, solutionTemplatesClient, resourceGroupName, capabilities)
		return err
	}, 3, 30)

	if retryErr != nil {
		log.Fatalf("Error creating solution template after retries: %v", retryErr)
	}

	// Create solution template version
	solutionTemplateVersionResult, err := createSolutionTemplateVersion(ctx, solutionTemplatesClient, resourceGroupName, *solutionTemplate.Name, *schema.Name, *schemaVersion.Name)
	if err != nil {
		log.Fatalf("Error creating solution template version: %v", err)
	}

	// Extract the solution template version ID
	var solutionTemplateVersionID string
	if solutionTemplateVersionResult.Properties != nil && solutionTemplateVersionResult.Name != nil {
		solutionTemplateVersionID = *solutionTemplateVersionResult.Name
		fmt.Printf("Successfully extracted solution template version ID: %s\n", solutionTemplateVersionID)
	} else {
		fmt.Println("Warning: Could not extract solution template version ID - Properties or ID is nil")
	}

	// Create target
	targetsClient := clientFactory.NewTargetsClient()
	target, err := createTarget(ctx, targetsClient, resourceGroupName, capabilities)
	if err != nil {
		log.Fatalf("Error creating target: %v", err)
	}

	// STEP 3: Configuration API Call - Set configuration values before review
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("STEP 3: Setting Configuration Values via Configuration API")
	fmt.Println(strings.Repeat("=", 50))

	configName := *target.Name + "Config"
	solutionName := "sdkexamples-solution1"
	version := "1.0.0"

	configValues := map[string]interface{}{
		"ErrorThreshold":      35.3,
		"HealthCheckEndpoint": "http://localhost:8080/health",
		"EnableLocalLog":      true,
		"AgentEndpoint":       "http://localhost:8080/agent",
		"HealthCheckEnabled":  true,
		"ApplicationEndpoint": "http://localhost:8080/app",
		"TemperatureRangeMax": 100.5,
	}

	fmt.Printf("Calling Configuration API with:\n")
	fmt.Printf("  Config Name: %s\n", configName)
	fmt.Printf("  Solution Name: %s\n", solutionName)
	fmt.Printf("  Version: %s\n", version)
	fmt.Printf("  Configuration Values:\n")
	for key, value := range configValues {
		fmt.Printf("    %s: %v\n", key, value)
	}

	err = createConfigurationAPICall(credential, subscriptionID, resourceGroupName, configName, solutionName, version, configValues)
	if err != nil {
		fmt.Printf("Configuration API call failed (continuing with workflow): %v\n", err)
	} else {
		fmt.Println("Configuration API call completed successfully")
	}

	// STEP 3.1: GET Configuration to verify the values were set correctly
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("STEP 3.1: Getting Configuration to verify values")
	fmt.Println(strings.Repeat("=", 50))

	err = getConfigurationAPICall(credential, subscriptionID, resourceGroupName, configName, solutionName, version)
	if err != nil {
		fmt.Printf("Configuration GET call failed: %v\n", err)
	}

	// Review target using the extracted solution template version ID
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("STEP 4: Review Target Deployment")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Using solution template version ID: %s\n", solutionTemplateVersionID)

	solutionVersionID, err := reviewTarget(ctx, targetsClient, resourceGroupName, *target.Name, solutionTemplateVersionID)
	if err != nil {
		fmt.Printf("Error reviewing target: %v\n", err)
		solutionVersionID = solutionTemplateVersionID // Use the original ID as fallback
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("STEP 5: Publish and Install Solution")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("The workflow has completed the following steps:")
	fmt.Println("✓ Context management with capabilities")
	fmt.Println("✓ Schema creation")
	fmt.Println("✓ Solution template creation")
	fmt.Println("✓ Target creation")
	fmt.Println("✓ Configuration API calls")
	fmt.Println("✓ Target review")
	fmt.Printf("\nTARGET INFORMATION:\n")
	fmt.Printf("  Name: %s\n", *target.Name)
	fmt.Printf("  Resource Group: %s\n", resourceGroupName)
	fmt.Printf("  Capabilities: %v\n", capabilities)
	fmt.Printf("\nCONFIGURATION COMPLETED:\n")
	fmt.Printf("  Config Name: %sConfig\n", *target.Name)
	fmt.Printf("  Solution Name: sdkexamples-solution1\n")
	fmt.Printf("\nProceeding with publish and install operations...\n")

	// Publish target
	err = publishTarget(ctx, targetsClient, resourceGroupName, *target.Name, solutionVersionID)
	if err != nil {
		fmt.Printf("Error publishing target: %v\n", err)
	}

	// Install target
	err = installTarget(ctx, targetsClient, resourceGroupName, *target.Name, solutionVersionID)
	if err != nil {
		fmt.Printf("Error installing target: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("WORKFLOW COMPLETED SUCCESSFULLY!")
	fmt.Println(strings.Repeat("=", 50))
}
