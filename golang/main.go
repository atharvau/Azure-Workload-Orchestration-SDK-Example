package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/workloadorchestration/armworkloadorchestration"
)

const (
	location        = "eastus2euap"
	subscriptionID  = "973d15c6-6c57-447e-b9c6-6d79b5b784ab"
	resourceGroup   = "ConfigManager-CloudTest-Playground-Portal"
	versionFilePath = "version.txt"
)

var (
	version = 0
)

func getNextVersion() int {
	data, err := os.ReadFile(versionFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("failed to read version file: %v", err)
		}
		version = 0
	} else {
		_, err := fmt.Sscanf(string(data), "%d", &version)
		if err != nil {
			log.Fatalf("failed to parse version: %v", err)
		}
	}

	version++

	err = os.WriteFile(versionFilePath, []byte(fmt.Sprintf("%d", version)), 0644)
	if err != nil {
		log.Fatalf("failed to write version file: %v", err)
	}

	return version
}

func main() {
	fmt.Println("Starting Go application...")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	ctx := context.Background()

	clientFactory, err := armworkloadorchestration.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	fmt.Println("Successfully authenticated with Azure.")

	// Get context information
	contextsClient := clientFactory.NewContextsClient()
	maxRetries := 3

	fmt.Printf("Getting context information for Mehoopany-Context...\n")
	
	var contextInfo armworkloadorchestration.ContextsClientGetResponse
	var getErr error
	
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("Retry attempt %d/%d\n", i+1, maxRetries)
		}
		
		contextInfo, getErr = contextsClient.Get(ctx, "Mehoopany", "Mehoopany-Context", nil)
		
		if getErr == nil {
			break
		}
		
		log.Printf("Attempt %d failed: %v", i+1, getErr)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if getErr != nil {
		log.Printf("Failed to get context info after %d attempts: %v", maxRetries, getErr)
	} else {
		fmt.Printf("Context Name: %s\n", *contextInfo.Name)
		fmt.Printf("Context ID: %s\n", *contextInfo.ID)
		fmt.Printf("Context Type: %s\n", *contextInfo.Type)
		fmt.Printf("Location: %s\n", *contextInfo.Location)
		fmt.Printf("Resource Group: Mehoopany\n")
		
		if contextInfo.Properties != nil {
			fmt.Println("\nCapabilities:")
			for i, cap := range contextInfo.Properties.Capabilities {
				fmt.Printf("%d. Name: %s\n", i+1, *cap)
			}
			
			fmt.Println("\nHierarchies:")
			for i, hier := range contextInfo.Properties.Hierarchies {
				fmt.Printf("%d. Name: %s\n", i+1, *hier)
			}
			
			if contextInfo.Properties.ProvisioningState != nil {
				fmt.Printf("\nProvisioning State: %s\n", *contextInfo.Properties.ProvisioningState)
			}
		}
	}
}

func createSchema(ctx context.Context, client *armworkloadorchestration.SchemasClient) *armworkloadorchestration.Schema {
	version := getNextVersion()
	schemaName := fmt.Sprintf("test-schema-v%d", version)

	fmt.Printf("Creating schema in resource group: %s\n", resourceGroup)

	poller, err := client.BeginCreateOrUpdate(ctx, resourceGroup, schemaName, armworkloadorchestration.Schema{
		Location: to.Ptr(location),
		Properties: &armworkloadorchestration.SchemaProperties{},
	}, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}

	fmt.Printf("Schema created successfully: %s\n", *res.Name)
	return &res.Schema
}

func createSchemaVersion(ctx context.Context, client *armworkloadorchestration.SchemaVersionsClient, schemaName string) *armworkloadorchestration.SchemaVersion {
	version := getNextVersion()
	schemaVersionName := fmt.Sprintf("1.0.%d", version)

	fmt.Printf("Creating schema version for schema: %s\n", schemaName)

	poller, err := client.BeginCreateOrUpdate(ctx, resourceGroup, schemaName, schemaVersionName, armworkloadorchestration.SchemaVersion{
		Properties: &armworkloadorchestration.SchemaVersionProperties{
			Value: to.Ptr("rules:\n  configs:\n      ErrorThreshold:\n        type: float\n        required: true\n  "),
		},
	}, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}

	fmt.Printf("Schema version created successfully: %s\n", *res.Name)
	return &res.SchemaVersion
}

func createSolutionTemplate(ctx context.Context, client *armworkloadorchestration.SolutionTemplatesClient) *armworkloadorchestration.SolutionTemplate {
	version := getNextVersion()
	solutionTemplateName := fmt.Sprintf("my-solution-template-v%d", version)

	fmt.Printf("Creating solution template in resource group: %s\n", resourceGroup)

	poller, err := client.BeginCreateOrUpdate(ctx, resourceGroup, solutionTemplateName, armworkloadorchestration.SolutionTemplate{
		Location: to.Ptr(location),
		Properties: &armworkloadorchestration.SolutionTemplateProperties{
			Capabilities: []*string{to.Ptr("sdkbox-soap")},
			Description:  to.Ptr("This is Test Solution"),
		},
	}, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}

	fmt.Printf("Solution template created successfully: %s\n", *res.Name)
	return &res.SolutionTemplate
}

func createSolutionTemplateVersion(ctx context.Context, client *armworkloadorchestration.SolutionTemplatesClient, solutionTemplateName string, schemaName string, schemaVersionName string) {
	version := getNextVersion()
	solutionTemplateVersionName := fmt.Sprintf("1.0.%d", version)

	fmt.Printf("Creating solution template version for template: %s\n", solutionTemplateName)

	versionBody := armworkloadorchestration.SolutionTemplateVersionWithUpdateType{
		SolutionTemplateVersion: &armworkloadorchestration.SolutionTemplateVersion{
			Properties: &armworkloadorchestration.SolutionTemplateVersionProperties{
				Configurations: to.Ptr(fmt.Sprintf("schema:\n  name: %s\n  version: %s\nconfigs:\n  AppName: Hotmelt\n  TemperatureRangeMax: 250\n  ErrorThreshold: 0.5\n  HealthCheckEndpoint: http://localhost:8080/health\n  EnableLocalLog: true\n  AgentEndpoint: http://localhost:8081\n  HealthCheckEnabled: true\n  ApplicationEndpoint: http://localhost:8082\n", schemaName, schemaVersionName)),
				Specification: map[string]interface{}{
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
				},
				OrchestratorType: to.Ptr(armworkloadorchestration.OrchestratorTypeTO),
			},
		},
		Version: &solutionTemplateVersionName,
	}
	poller, err := client.BeginCreateVersion(ctx, resourceGroup, solutionTemplateName, versionBody, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}

	fmt.Printf("Solution template version created successfully: %s\n", *res.Name)
}

func createTarget(ctx context.Context, client *armworkloadorchestration.TargetsClient) {
	version := getNextVersion()
	targetName := fmt.Sprintf("sdkbox-mk71-v%d", version)

	fmt.Printf("Creating target in resource group: %s\n", resourceGroup)

	poller, err := client.BeginCreateOrUpdate(ctx, resourceGroup, targetName, armworkloadorchestration.Target{
		Location: to.Ptr(location),
		ExtendedLocation: &armworkloadorchestration.ExtendedLocation{
			Name: to.Ptr("/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/configmanager-cloudtest-playground-portal/providers/Microsoft.ExtendedLocation/customLocations/den-Location"),
			Type: to.Ptr(armworkloadorchestration.ExtendedLocationTypeCustomLocation),
		},
		Properties: &armworkloadorchestration.TargetProperties{
			Capabilities: []*string{to.Ptr("sdkbox-soap")},
			ContextID:    to.Ptr("/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/Mehoopany/providers/Microsoft.Edge/contexts/Mehoopany-Context"),
			Description:  to.Ptr("Test target for workload orchestration"),
			DisplayName:  to.Ptr(targetName),
			HierarchyLevel: to.Ptr("Factory"),
			SolutionScope: to.Ptr("new"),
			TargetSpecification: map[string]interface{}{
				"topologies": []map[string]interface{}{
					{
						"bindings": []map[string]interface{}{
							{
								"role": "helm.v3",
								"provider": "providers.target.helm",
							},
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}

	fmt.Printf("Target created successfully: %s\n", *res.Name)
}