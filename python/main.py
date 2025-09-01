import os
from azure.identity import DefaultAzureCredential
from azure.mgmt.workloadorchestration import WorkloadOrchestrationMgmtClient
from azure.core.exceptions import HttpResponseError

# Configuration
LOCATION = "eastus2euap"
SUBSCRIPTION_ID = "973d15c6-6c57-447e-b9c6-6d79b5b784ab"
RESOURCE_GROUP = "ConfigManager-CloudTest-Playground-Portal"

def get_next_version():
    try:
        with open('version.txt', 'r') as f:
            version = int(f.read().strip())
    except FileNotFoundError:
        version = 0
    
    with open('version.txt', 'w') as f:
        f.write(str(version + 1))
    
    return version

def create_schema(client, resource_group_name, subscription_id):
    try:
        version = get_next_version()
        schema_name = f"test-schema-v{version}"
        
        schema_result = client.schemas.begin_create_or_update(
            resource_group_name=resource_group_name,
            schema_name=schema_name,
            resource={
                "location": LOCATION,
                "properties": {}
            }
        ).result()
        return schema_result
    except Exception as e:
        print(f"Error creating schema: {e}")
        raise

def create_schema_version(client, resource_group_name, schema_name):
    try:
        version = get_next_version()
        schema_version_name = f"1.0.{version}"
        schema_version_result = client.schema_versions.begin_create_or_update(
            resource_group_name=resource_group_name,
            schema_name=schema_name,
            schema_version_name=schema_version_name,
            resource={
                "properties": {
                    "value": "rules:\n  configs:\n      ErrorThreshold:\n        type: float\n        required: true\n  "
                }
            }
        ).result()
        return schema_version_result
    except Exception as e:
        print(f"Error creating schema version: {e}")
        raise

def create_solution_template(client, resource_group_name):
    try:
        version = get_next_version()
        solution_template_name = f"my-solution-template-v{version}"
        solution_template_result = client.solution_templates.begin_create_or_update(
            resource_group_name=resource_group_name,
            solution_template_name=solution_template_name,
            resource={
                "location": LOCATION,
                "properties": {
                    "capabilities": ["sdkbox-soap"],
                    "description": "This is Test Solution"
                }
            }
        ).result()
        return solution_template_result
    except Exception as e:
        print(f"Error creating solution template: {e}")
        raise

def create_solution_template_version(client, resource_group_name, solution_template_name, schema_name, schema_version):
    try:
        version = get_next_version()
        solution_template_version_name = f"1.0.{version}"
        configurations_str = f"schema:\n  name: {schema_name}\n  version: {schema_version}\nconfigs:\n  AppName: Hotmelt\n  TemperatureRangeMax: ${{$val(TemperatureRangeMax)}}\n  ErrorThreshold: ${{$val(ErrorThreshold)}}\n  HealthCheckEndpoint: ${{$val(HealthCheckEndpoint)}}\n  EnableLocalLog: ${{$val(EnableLocalLog)}}\n  AgentEndpoint: ${{$val(AgentEndpoint)}}\n  HealthCheckEnabled: ${{$val(HealthCheckEnabled)}}\n  ApplicationEndpoint: ${{$val(ApplicationEndpoint)}}\n"
        solution_template_version_result = client.solution_templates.begin_create_version(
            resource_group_name=resource_group_name,
            solution_template_name=solution_template_name,
            body={
                "solutionTemplateVersion": {
                    "properties": {
                        "configurations": configurations_str,
                        "specification": {
                            "components": [
                                {
                                    "name": "helmcomponent",
                                    "type": "helm.v3",
                                    "properties": {
                                        "chart": {
                                            "repo": "ghcr.io/eclipse-symphony/tests/helm/simple-chart",
                                            "version": "0.3.0",
                                            "wait": True,
                                            "timeout": "5m"
                                        }
                                    }
                                }
                            ]
                        },
                        "orchestratorType": "TO"
                    }
                },
                "version": solution_template_version_name
            }
        ).result()
        return solution_template_version_result
    except Exception as e:
        print(f"Error creating solution template version: {e}")
        raise

def create_target(client, resource_group_name):
    try:
        target_name = "sdkbox-mk71"
        target_result = client.targets.begin_create_or_update(
            resource_group_name=resource_group_name,
            target_name=target_name,
            resource={
                "extendedLocation": {
                    "name": "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/configmanager-cloudtest-playground-portal/providers/Microsoft.ExtendedLocation/customLocations/den-Location",
                    "type": "CustomLocation"
                },
                "location": LOCATION,
                "properties": {
                    "capabilities": ["sdkbox-soap"],
                    "contextId": "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/Mehoopany/providers/Microsoft.Edge/contexts/Mehoopany-Context",
                    "description": "This is MK-71 Site",
                    "displayName": "sdkbox-mk71",
                    "hierarchyLevel": "site",
                    "solutionScope": "new",
                    "targetSpecification": {
                        "topologies": [
                            {
                                "bindings": [
                                    {
                                        "role": "helm.v3",
                                        "provider": "providers.target.helm",
                                        "config": {
                                            "inCluster": "true"
                                        }
                                    }
                                ]
                            }
                        ]
                    }
                }
            }
        ).result()
        return target_result
    except Exception as e:
        print(f"Error creating target: {e}")
        raise

def main():
    """
    This script authenticates with Azure and creates various resources.
    It uses DefaultAzureCredential for authentication.
    """
    try:
        # Get subscription ID from environment variable
        subscription_id = SUBSCRIPTION_ID
        if not subscription_id:
            print("Error: AZURE_SUBSCRIPTION_ID environment variable not set.")
            return

        # Authenticate using DefaultAzureCredential
        credential = DefaultAzureCredential()

        # Create the management client
        workload_client = WorkloadOrchestrationMgmtClient(credential, subscription_id)

        print("Successfully authenticated with Azure.")
        
        resource_group_name = RESOURCE_GROUP

        try:
            # Create a new schema
            print(f"Creating schema in resource group: {resource_group_name}")
            schema = create_schema(workload_client, resource_group_name, subscription_id)
            print(f"Schema created successfully: {schema.name}")

            # Create a new schema version
            print(f"Creating schema version for schema: {schema.name}")
            schema_version = create_schema_version(workload_client, resource_group_name, schema.name)
            print(f"Schema version created successfully: {schema_version.name}")

            # Create a new solution template
            print(f"Creating solution template in resource group: {resource_group_name}")
            solution_template = create_solution_template(workload_client, resource_group_name)
            print(f"Solution template created successfully: {solution_template.name}")

            # Create a new solution template version
            print(f"Creating solution template version for template: {solution_template.name}")
            solution_template_version = create_solution_template_version(workload_client, resource_group_name, solution_template.name, schema.name, schema_version.name)
            print(f"Solution template version created successfully: {solution_template_version.name}")
        except Exception as e:
            print(f"An error occurred during resource creation: {e}")
            return

        # Create a new target
        print(f"Creating target in resource group: {resource_group_name}")
        target = create_target(workload_client, resource_group_name)
        print(f"Target created successfully: {target.name}")



    except HttpResponseError as e:
        print(f"An HTTP error occurred: {e.message}")
    except Exception as e:
        print(f"An unexpected error occurred: {e}")

if __name__ == "__main__":
    main()
