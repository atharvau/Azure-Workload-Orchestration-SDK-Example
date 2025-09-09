import os
import random
import time
import json
import requests
from datetime import datetime
from azure.identity import DefaultAzureCredential, EnvironmentCredential
from azure.mgmt.workloadorchestration import WorkloadOrchestrationMgmtClient
from azure.core.exceptions import HttpResponseError

def retry_operation(operation, max_attempts=3, delay_seconds=30):
    """
    Retry an operation with exponential backoff
    """
    for attempt in range(max_attempts):
        try:
            return operation()
        except Exception as e:
            if attempt == max_attempts - 1:  # Last attempt
                raise  # Re-raise the last exception
            print(f"Attempt {attempt + 1} failed: {str(e)}")
            print(f"Waiting {delay_seconds} seconds before retrying...")
            time.sleep(delay_seconds)
            delay_seconds *= 2  # Exponential backoff

# Configuration
LOCATION = "eastus2euap"
SUBSCRIPTION_ID = os.getenv("AZURE_SUBSCRIPTION_ID", "973d15c6-6c57-447e-b9c6-6d79b5b784ab")
RESOURCE_GROUP = "sdkexamples"
CONTEXT_RESOURCE_GROUP = "Mehoopany"  # Hardcoded resource group for context
CONTEXT_NAME = "Mehoopany-Context"    # Hardcoded context name

# Single capability configuration - ensures consistency across all resources
SINGLE_CAPABILITY_NAME = "sdkexamples-soap"

# Authentication setup hints
AUTH_SETUP_HINT = """
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
"""

def generate_random_semantic_version(include_prerelease=False, include_build=False):
    """
    Generate a random semantic version string
    """
    major = random.randint(0, 10)
    minor = random.randint(0, 20)
    patch = random.randint(0, 100)
    version = f"{major}.{minor}.{patch}"
    
    if include_prerelease:
        prerelease_types = ['alpha', 'beta', 'rc']
        prerelease_type = random.choice(prerelease_types)
        prerelease_num = random.randint(1, 10)
        version += f"-{prerelease_type}.{prerelease_num}"
    
    if include_build:
        build_num = random.randint(1, 10000)
        version += f"+{build_num}"
    
    return version

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
        version = generate_random_semantic_version()
        schema_name = f"sdkexamples-schema-v{version}"
        
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
        # Use semantic versioning
        version = generate_random_semantic_version()
        schema_version_name = version
        schema_version_result = client.schema_versions.begin_create_or_update(
            resource_group_name=resource_group_name,
            schema_name=schema_name,
            schema_version_name=schema_version_name,
            resource={
                "properties": {
                    "value": """rules:
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
        - OT"""
                }
            }
        ).result()
        return schema_version_result
    except Exception as e:
        print(f"Error creating schema version: {e}")
        raise

def create_solution_template(client, resource_group_name, capabilities=None):
    try:
        if capabilities is None:
            capabilities = [SINGLE_CAPABILITY_NAME]
        
        solution_template_name = "sdkexamples-solution1"
        solution_template_result = client.solution_templates.begin_create_or_update(
            resource_group_name=resource_group_name,
            solution_template_name=solution_template_name,
            resource={
                "location": LOCATION,
                "properties": {
                    "capabilities": capabilities,
                    "description": "This is Holtmelt Solution with random capabilities"
                }
            }
        ).result()
        return solution_template_result
    except Exception as e:
        print(f"Error creating solution template: {e}")
        raise

def create_solution_template_version(client, resource_group_name, solution_template_name, schema_name, schema_version):
    try:
        # Generate a clean version number without pre-release or build info
        version = generate_random_semantic_version(include_prerelease=False, include_build=False)
        solution_template_version_name = version
        configurations_str = f"""schema:
  name: {schema_name}
  version: {schema_version}
configs:
  AppName: Hotmelt
  TemperatureRangeMax: ${{{{$val(TemperatureRangeMax)}}}}
  ErrorThreshold: ${{{{$val(ErrorThreshold)}}}}
  HealthCheckEndpoint: ${{{{$val(HealthCheckEndpoint)}}}}
  EnableLocalLog: ${{{{$val(EnableLocalLog)}}}}
  AgentEndpoint: ${{{{$val(AgentEndpoint)}}}}
  HealthCheckEnabled: ${{{{$val(HealthCheckEnabled)}}}}
  ApplicationEndpoint: ${{{{$val(ApplicationEndpoint)}}}}
"""
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

def create_target(client, resource_group_name, capabilities=None):
    # Process capabilities at the function level to avoid scoping issues
    if capabilities is None:
        capabilities = [SINGLE_CAPABILITY_NAME]
    
    def create_operation():
        target_name = "sdkbox-mk799"
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
                    "capabilities": capabilities,
                    "contextId": f"/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/{CONTEXT_RESOURCE_GROUP}/providers/Microsoft.Edge/contexts/{CONTEXT_NAME}",
                    "description": "This is MK-71 Site with random capabilities",
                    "displayName": "sdkbox-mk71",
                    "hierarchyLevel": "line",
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

    try:
        return retry_operation(create_operation)
    except Exception as e:
        print(f"Error creating target: {e}")
        raise

def review_target(client, resource_group_name, target_name, solution_template_version_id):
    """
    Review a target deployment using a solution template version
    """
    def review_operation():
        print(f"Starting review for target {target_name}")
        review_result = client.targets.begin_review_solution_version(
            resource_group_name=resource_group_name,
            target_name=target_name,
            body={
                "solutionDependencies": [],
                "solutionInstanceName": target_name,
                "solutionTemplateVersionId": solution_template_version_id
            }
        ).result()
        return review_result

    try:
        review_result = retry_operation(review_operation)
        print(review_result)

        # Handle response that might be wrapped in _data
        if hasattr(review_result, '_data'):
            response_dict = review_result._data
        elif hasattr(review_result, '__dict__'):
            response_dict = review_result.__dict__
        else:
            response_dict = dict(review_result)

        # Extract solutionTemplateVersionId from the nested structure
        try:
            if '_data' in response_dict:
                response_dict = response_dict['_data']
            
            properties = response_dict.get('properties', {})
            version_id = properties.get('id')
            
            if version_id:
                print(f"Found solution version ID: {version_id}")
                return version_id
            
            # Debug output for troubleshooting
            print("Debug - Full response:", response_dict)
            print("Debug - Properties:", properties)
            print("Debug - Nested properties:", nested_props)
            raise ValueError("Could not find solutionTemplateVersionId in review response")
        except Exception as e:
            print(f"Debug - Error: {str(e)}")
            raise ValueError(f"Error extracting solutionTemplateVersionId: {str(e)}")
    except Exception as e:
        print(f"Error reviewing target: {e}")
        raise

def publish_target(client, resource_group_name, target_name, solution_version_id):
    """
    Publish a solution version to a target
    """
    def publish_operation():
        print(f"Publishing solution version to target {target_name}")
        publish_result = client.targets.begin_publish_solution_version(
            resource_group_name=resource_group_name,
            target_name=target_name,
            body={
                "solutionVersionId": solution_version_id
            }
        ).result()
        print("Publish operation completed successfully")
        return publish_result

    try:
        return retry_operation(publish_operation)
    except Exception as e:
        print(f"Error publishing to target: {e}")
        raise

def install_target(client, resource_group_name, target_name, solution_version_id):
    """
    Install a published solution version on a target
    """
    def install_operation():
        print(f"Installing solution version on target {target_name}")
        install_result = client.targets.begin_install_solution(
            resource_group_name=resource_group_name,
            target_name=target_name,
            body={
                "solutionVersionId": solution_version_id
            }
        ).result()
        
        # Store the install job ID from the response
        install_job_id = install_result.job_id if hasattr(install_result, 'job_id') else None
        print(f"Install operation completed. Job ID: {install_job_id}")
        return install_result

    try:
        return retry_operation(install_operation)
    except Exception as e:
        print(f"Error installing on target: {e}")
        raise

def create_configuration_api_call(credential, subscription_id, resource_group, config_name, solution_name, version, config_values):
    """
    Make PUT call to Azure Configuration API to set configuration values
    """
    try:
        # Get bearer token from DefaultAzureCredential
        token = credential.get_token("https://management.azure.com/.default")
        bearer_token = token.token
        
        # Construct the API URL with correct API version (matching CLI format)
        url = f"https://management.azure.com/subscriptions/{subscription_id}/resourceGroups/{resource_group}/providers/Microsoft.Edge/configurations/{config_name}/DynamicConfigurations/{solution_name}/versions/version1?api-version=2024-06-01-preview"
        
        print("\nDebug: Request URL:")
        print(url)
        
        headers = {
            "Authorization": f"Bearer {bearer_token}",
            "Content-Type": "application/json"
        }
        print("\nDebug: Request Headers:")
        print(f"- Content-Type: {headers['Content-Type']}")
        print("- Authorization: Bearer [token-hidden]")
        
        # Build values string from config_values dictionary matching CLI format
        values_lines = []
        for key, value in config_values.items():
            if isinstance(value, bool):
                # Convert boolean to lowercase string (true/false, not "true"/"false")
                values_lines.append(f"{key}: {str(value).lower()}")
            elif isinstance(value, str):
                # String values without quotes in YAML format
                values_lines.append(f"{key}: {value}")
            else:
                # Numbers and other types
                values_lines.append(f"{key}: {value}")
        
        values_string = "\n".join(values_lines) + "\n"
        
        # Request body with all configuration values
        request_body = {
            "properties": {
                "values": values_string,
                "provisioningState": "Succeeded"
            }
        }
        
        print(f"Making PUT call to Configuration API: {url}")
        print(f"Request body: {json.dumps(request_body, indent=2)}")
        
        response = requests.put(url, headers=headers, json=request_body)
        
        print("\nDebug: Response Details:")
        print(f"- Status Code: {response.status_code}")
        print("- Response Headers:")
        for key, value in response.headers.items():
            print(f"  {key}: {value}")
        
        try:
            response_json = response.json()
            print("\nDebug: Response Body (JSON):")
            print(json.dumps(response_json, indent=2))
        except json.JSONDecodeError:
            print("\nDebug: Response Body (Raw):")
            print(response.text)
        
        if response.status_code in [200, 201, 202]:
            print(f"Configuration API call successful. Status: {response.status_code}")
            print(f"Response: {response.text}")
            return response
        else:
            print(f"Configuration API call failed. Status: {response.status_code}")
            print(f"Response: {response.text}")
            response.raise_for_status()
            
    except Exception as e:
        print(f"Error calling Configuration API: {e}")
        raise

def get_configuration_api_call(credential, subscription_id, resource_group, config_name, solution_name, version):
    """
    Make GET call to Azure Configuration API to retrieve current configuration values
    """
    try:
        # Get bearer token from DefaultAzureCredential
        token = credential.get_token("https://management.azure.com/.default")
        bearer_token = token.token
        
        # Construct the API URL (same as PUT but for GET)
        url = f"https://management.azure.com/subscriptions/{subscription_id}/resourceGroups/{resource_group}/providers/Microsoft.Edge/configurations/{config_name}/DynamicConfigurations/{solution_name}/versions/version1?api-version=2024-06-01-preview"
        
        headers = {
            "Authorization": f"Bearer {bearer_token}",
            "Content-Type": "application/json"
        }
        
        print(f"Making GET call to Configuration API: {url}")
        
        response = requests.get(url, headers=headers)
        
        if response.status_code in [200]:
            print(f"Configuration GET API call successful. Status: {response.status_code}")
            print(f"Retrieved Configuration Response: {response.text}")
            
            # Try to parse and pretty print the JSON response
            try:
                response_json = response.json()
                print("Parsed Configuration Data:")
                print(json.dumps(response_json, indent=2))
                
                # Extract and display the actual values if they exist
                if 'properties' in response_json and 'values' in response_json['properties']:
                    print(f"Configuration Values: {response_json['properties']['values']}")
                    
            except json.JSONDecodeError:
                print("Response is not valid JSON")
                
            return response
        else:
            print(f"Configuration GET API call failed. Status: {response.status_code}")
            print(f"Response: {response.text}")
            # Don't raise exception for GET failures as it might be expected
            return None
            
    except Exception as e:
        print(f"Error calling Configuration GET API: {e}")
        return None

def get_existing_context(client, resource_group_name, context_name):
    """
    Fetch existing Azure context and return its capabilities
    """
    try:
        print(f"DEBUG: Fetching existing context: {context_name}")
        context = client.contexts.get(
            resource_group_name=resource_group_name,
            context_name=context_name
        )
        
        existing_capabilities = []
        if hasattr(context, 'properties') and hasattr(context.properties, 'capabilities'):
            existing_capabilities = context.properties.capabilities
        
        print(f"DEBUG: Found {len(existing_capabilities)} existing capabilities")
        if existing_capabilities:
            print("DEBUG: Existing capabilities details:")
            for i, cap in enumerate(existing_capabilities):
                print(f"  [{i}] Type: {type(cap)}")
                if isinstance(cap, dict):
                    print(f"      Name: {cap.get('name', 'N/A')}")
                    print(f"      Description: {cap.get('description', 'N/A')}")
                    print(f"      State: {cap.get('state', 'N/A')}")
                else:
                    print(f"      Name: {getattr(cap, 'name', 'N/A')}")
                    print(f"      Description: {getattr(cap, 'description', 'N/A')}")
                    print(f"      State: {getattr(cap, 'state', 'N/A')}")
        
        return existing_capabilities
        
    except HttpResponseError as e:
        if e.status_code == 404:
            print("DEBUG: Context not found, will create new one")
            return []
        else:
            print(f"DEBUG: Error fetching context: {e}")
            raise
    except Exception as e:
        print(f"DEBUG: Error fetching context: {e}")
        return []

def generate_single_random_capability():
    """
    Generate a single random Shampoo or Soap capability
    """
    capability_types = ["shampoo", "soap"]
    cap_type = random.choice(capability_types)
    random_suffix = random.randint(1000, 9999)
    
    capability = {
        "name": f"sdkexamples-{cap_type}-{random_suffix}",
        "description": f"SDK generated {cap_type} manufacturing capability"
    }
    
    print(f"DEBUG: Generated single random capability: {capability['name']}")
    return capability

def merge_capabilities_with_uniqueness(existing_capabilities, new_capabilities):
    """
    Merge capabilities ensuring no duplicates by name with comprehensive debugging
    """
    print("=" * 60)
    print("DEBUGGING CAPABILITY MERGE PROCESS")
    print("=" * 60)
    
    # Debug: Show what's coming in
    print(f"DEBUG: EXISTING CAPABILITIES - Count: {len(existing_capabilities)}")
    if existing_capabilities:
        for i, cap in enumerate(existing_capabilities):
            print(f"  EXISTING[{i}]:")
            print(f"    Type: {type(cap)}")
            if isinstance(cap, dict):
                print(f"    Name: {cap.get('name', 'N/A')}")
                print(f"    Description: {cap.get('description', 'N/A')}")
                print(f"    Has State: {'state' in cap}")
            else:
                print(f"    Name: {getattr(cap, 'name', 'N/A')}")
                print(f"    Description: {getattr(cap, 'description', 'N/A')}")
                print(f"    Has State: {hasattr(cap, 'state')}")
    
    print(f"\nDEBUG: NEW CAPABILITIES - Count: {len(new_capabilities)}")
    if new_capabilities:
        for i, cap in enumerate(new_capabilities):
            print(f"  NEW[{i}]:")
            print(f"    Type: {type(cap)}")
            if isinstance(cap, dict):
                print(f"    Name: {cap.get('name', 'N/A')}")
                print(f"    Description: {cap.get('description', 'N/A')}")
                print(f"    Has State: {'state' in cap}")
            else:
                print(f"    Name: {getattr(cap, 'name', 'N/A')}")
                print(f"    Description: {getattr(cap, 'description', 'N/A')}")
                print(f"    Has State: {hasattr(cap, 'state')}")
    
    # Process existing capabilities
    existing_names = set()
    merged_capabilities = []
    
    print(f"\nDEBUG: PROCESSING EXISTING CAPABILITIES...")
    for i, cap in enumerate(existing_capabilities):
        if isinstance(cap, dict):
            cap_name = cap.get('name', '')
            cap_desc = cap.get('description', '')
        else:
            cap_name = getattr(cap, 'name', '')
            cap_desc = getattr(cap, 'description', '')
        
        if cap_name and cap_name not in existing_names:
            existing_names.add(cap_name)
            # Convert to dict format without state field
            capability_dict = {
                "name": cap_name,
                "description": cap_desc
            }
            merged_capabilities.append(capability_dict)
            print(f"  ADDED EXISTING[{i}]: {cap_name}")
        else:
            print(f"  SKIPPED EXISTING[{i}]: {cap_name} (duplicate or empty)")
    
    print(f"\nDEBUG: PROCESSING NEW CAPABILITIES...")
    # Process new capabilities
    for i, cap in enumerate(new_capabilities):
        cap_name = cap.get('name', '') if isinstance(cap, dict) else getattr(cap, 'name', '')
        
        if cap_name not in existing_names:
            existing_names.add(cap_name)
            # Ensure new capabilities don't have state field
            capability_dict = {
                "name": cap.get('name', '') if isinstance(cap, dict) else getattr(cap, 'name', ''),
                "description": cap.get('description', '') if isinstance(cap, dict) else getattr(cap, 'description', '')
            }
            merged_capabilities.append(capability_dict)
            print(f"  ADDED NEW[{i}]: {cap_name}")
        else:
            print(f"  REJECTED NEW[{i}]: {cap_name} (DUPLICATE - overriding avoided!)")
    
    print(f"\nDEBUG: MERGE RESULTS VALIDATION")
    print(f"  Initial existing count: {len(existing_capabilities)}")
    print(f"  New capabilities count: {len(new_capabilities)}")
    print(f"  Final merged count: {len(merged_capabilities)}")
    print(f"  Unique names count: {len(existing_names)}")
    
    # Validation check
    expected_max = len(existing_capabilities) + len(new_capabilities)
    if len(merged_capabilities) > expected_max:
        print(f"ERROR: Merged count ({len(merged_capabilities)}) exceeds maximum expected ({expected_max})")
        print("STOPPING EXECUTION - CAPABILITY MERGE VALIDATION FAILED")
        raise ValueError(f"Capability merge validation failed: count {len(merged_capabilities)} > {expected_max}")
    
    # Final validation: ensure all merged capabilities have required fields
    print(f"\nDEBUG: FINAL CAPABILITY VALIDATION")
    for i, cap in enumerate(merged_capabilities):
        if not isinstance(cap, dict):
            print(f"ERROR: Capability[{i}] is not dict: {type(cap)}")
            raise ValueError(f"Invalid capability format at index {i}")
        if not cap.get('name'):
            print(f"ERROR: Capability[{i}] missing name: {cap}")
            raise ValueError(f"Capability missing name at index {i}")
        if 'state' in cap:
            print(f"WARNING: Capability[{i}] contains deprecated 'state' field: {cap['name']}")
    
    print(f"VALIDATION PASSED - Proceeding with {len(merged_capabilities)} capabilities")
    print("=" * 60)
    
    return merged_capabilities

def save_capabilities_to_json(capabilities, filename="context-capabilities.json"):
    """
    Save capabilities to JSON file
    """
    try:
        with open(filename, 'w') as f:
            json.dump(capabilities, f, indent=2)
        print(f"Capabilities saved to {filename}")
    except Exception as e:
        print(f"Error saving capabilities to JSON: {e}")
        raise

def create_or_update_context_with_hierarchies(client, resource_group_name, context_name, capabilities):
    """
    Create or update Azure context with capabilities and hierarchies
    """
    def context_operation():
        hierarchies = [
            {"name": "country", "description": "Country level hierarchy"},
            {"name": "region", "description": "Regional level hierarchy"},
            {"name": "factory", "description": "Factory level hierarchy"},
            {"name": "line", "description": "Production line hierarchy"}
        ]
        
        resource = {
            "location": LOCATION,
            "properties": {
                "capabilities": capabilities,
                "hierarchies": hierarchies
            }
        }
        
        print(f"Creating/updating context: {context_name}")
        context_result = client.contexts.begin_create_or_update(
            resource_group_name=resource_group_name,
            context_name=context_name,
            resource=resource
        ).result()
        
        context_result = client.contexts.begin_create_or_update(
            resource_group_name=resource_group_name,
            context_name=context_name,
            resource=resource
        ).result()
        
        return context_result
    
    try:
        return retry_operation(context_operation)
    except Exception as e:
        print(f"Error creating/updating context: {e}")
        raise

def manage_azure_context(client, resource_group_name=CONTEXT_RESOURCE_GROUP, context_name=CONTEXT_NAME):
    """
    Complete context management workflow:
    1. Fetch existing context
    2. Generate single random capability
    3. Merge with uniqueness
    4. Save to JSON
    5. Update context
    """
    try:
        # Step 1: Fetch existing context
        existing_capabilities = get_existing_context(client, resource_group_name, context_name)
        
        # Step 2: Generate single random capability  
        new_capability = generate_single_random_capability()
        new_capabilities = [new_capability]  # Convert to list for merge function
        
        # Step 3: Merge capabilities with uniqueness constraints
        merged_capabilities = merge_capabilities_with_uniqueness(existing_capabilities, new_capabilities)
        
        # Step 4: Save to JSON file
        save_capabilities_to_json(merged_capabilities)
        
        # Step 5: Create/update context with hierarchies
        context_result = create_or_update_context_with_hierarchies(
            client, resource_group_name, context_name, merged_capabilities
        )
        
        print(f"Context management completed successfully: {context_result.name}")
        return context_result
        
    except Exception as e:
        print(f"Error in context management workflow: {e}")
        raise

def main():
    """
    This script authenticates with Azure and creates various resources.
    It uses DefaultAzureCredential for authentication.
    """
    try:
        subscription_id = SUBSCRIPTION_ID
        if not subscription_id:
            print("Error: AZURE_SUBSCRIPTION_ID environment variable not set.")
            return

        # Try DefaultCredentials first
        try:
            credential = DefaultAzureCredential()
            # Test the credential by getting a token
            credential.get_token("https://management.azure.com/.default")
            print("Successfully authenticated using environment variables.")
        except Exception as e:
            print("Environment credential failed:", str(e))
            print("\nFalling back to DefaultAzureCredential...")
            try:
                credential = DefaultAzureCredential()
                credential.get_token("https://management.azure.com/.default")
                print("Successfully authenticated using DefaultAzureCredential.")
            except Exception as auth_error:
                print("\nAuthentication failed:")
                print(str(auth_error))
                print(AUTH_SETUP_HINT)
                return

        # Create the management client with subscription ID
        workload_client = WorkloadOrchestrationMgmtClient(credential, subscription_id)

        print("Successfully authenticated with Azure.")
        
        resource_group_name = RESOURCE_GROUP

        # STEP 1: Manage Azure context with random capabilities
        print("=" * 50)
        print("STEP 1: Managing Azure Context with Random Capabilities")
        print("=" * 50)
        try:
            # Use hardcoded values for context management
            context_result = manage_azure_context(workload_client)
            
            # Extract the NEWLY ADDED capability from context for use in all resources
            capabilities = None
            print(f"DEBUG: Extracting capability from context result...")
            
            if hasattr(context_result, 'properties') and hasattr(context_result.properties, 'capabilities'):
                context_capabilities = context_result.properties.capabilities
                print(f"DEBUG: Found {len(context_capabilities)} capabilities in context")
                
                if context_capabilities:
                    # Get the LAST capability (which should be the newly added one)
                    last_cap = context_capabilities[-1]
                    print(f"DEBUG: Last capability type: {type(last_cap)}")
                    print(f"DEBUG: Last capability data: {last_cap}")
                    
                    cap_name = last_cap.get('name', '') if isinstance(last_cap, dict) else getattr(last_cap, 'name', '')
                    if cap_name:
                        capabilities = [cap_name]
                        print(f"SELECTED CAPABILITY FOR ALL RESOURCES: {capabilities[0]}")
                        print(f"DEBUG: This capability will be used consistently across:")
                        print(f"  - Solution Template")
                        print(f"  - Target")
                        print(f"  - All other resource operations")
                    else:
                        print("DEBUG: Could not extract capability name from last capability")
                else:
                    print("DEBUG: No capabilities found in context")
            else:
                print("DEBUG: Context result has no capabilities property")
            
            if not capabilities:
                # Generate a single random capability if none found in context
                print("DEBUG: No valid capability found, generating new one...")
                new_capability = generate_single_random_capability()
                capabilities = [new_capability['name']]
                print(f"GENERATED NEW CAPABILITY FOR ALL RESOURCES: {capabilities[0]}")
        except Exception as e:
            print(f"Context management failed, generating new random capability: {e}")
            new_capability = generate_single_random_capability()
            capabilities = [new_capability['name']]
            print(f"FALLBACK CAPABILITY FOR ALL RESOURCES: {capabilities[0]}")
            
        # Validate that we have a capability selected
        if not capabilities or not capabilities[0]:
            print("ERROR: No capability was selected! Using fallback.")
            capabilities = [SINGLE_CAPABILITY_NAME]
            
        print(f"\nFINAL CAPABILITY SELECTION: {capabilities[0]}")
        print("=" * 60)


        print("=" * 50)
        print("STEP 2: Creating Azure Resources")
        print("=" * 50)
        try:
            # Create a new schema
            print(f"Creating schema in resource group: {resource_group_name}")
            schema = create_schema(workload_client, resource_group_name, subscription_id)
            print(f"Schema created successfully: {schema.name}")

            # Create a new schema version
            print(f"Creating schema version for schema: {schema.name}")
            schema_version = create_schema_version(workload_client, resource_group_name, schema.name)
            print(f"Schema version created successfully: {schema_version.name}")

        except Exception as e:
            print(f"An error occurred during resource creation: {e}")
            return

        print("Proceeding with solution template and target creation...")
        print()

        try:
            # Create a new solution template with the same random capability
            print(f"Creating solution template in resource group: {resource_group_name}")
            print(f"Using capability: {capabilities}")
            solution_template = create_solution_template(workload_client, resource_group_name, capabilities)
            print(f"Solution template created successfully: {solution_template.name}")

            # Create a new solution template version
            print(f"Creating solution template version for template: {solution_template.name}")
            solution_template_version_result = create_solution_template_version(workload_client, resource_group_name, solution_template.name, schema.name, schema_version.name)
            print(f"Solution template version created successfully: {solution_template_version_result}")

            # Extract the solution template version ID from the response properties
            if (hasattr(solution_template_version_result, 'properties') and
                hasattr(solution_template_version_result.properties, 'solutionTemplateVersionId')):
                solution_template_version_id = solution_template_version_result.properties.solutionTemplateVersionId
            else:
                solution_template_version_id = solution_template_version_result.properties.get('solutionTemplateVersionId')

            # Create a new target with the same random capability
            print(f"Creating target in resource group: {resource_group_name}")
            print(f"Using capability: {capabilities}")
            target = create_target(workload_client, resource_group_name, capabilities)
            print(f"Target created successfully: {target.name}")

        except Exception as e:
            print(f"An error occurred during target creation: {e}")
            return

        # STEP 3: Configuration API Call - Set configuration values before review
        print("=" * 50)
        print("STEP 3: Setting Configuration Values via Configuration API")
        print("=" * 50)
        try:
            # Configuration parameters for the API call
            config_name = target.name + "Config"  # Configuration name should be targetName+Config
            solution_name = "sdkexamples-solution1"  # Use hardcoded solution template name
            version = "1.0.0"  # Configuration version
            
            # Configuration values matching the schema
            config_values = {
                "ErrorThreshold": 35.3,
                "HealthCheckEndpoint": "http://localhost:8080/health",
                "EnableLocalLog": True,
                "AgentEndpoint": "http://localhost:8080/agent",
                "HealthCheckEnabled": True,
                "ApplicationEndpoint": "http://localhost:8080/app",
                "TemperatureRangeMax": 100.5
            }
            
            print(f"Calling Configuration API with:")
            print(f"  Config Name: {config_name}")
            print(f"  Solution Name: {solution_name}")
            print(f"  Version: {version}")
            print(f"  Configuration Values:")
            for key, value in config_values.items():
                print(f"    {key}: {value}")
            
            config_response = create_configuration_api_call(
                credential,
                subscription_id,
                resource_group_name,
                config_name,
                solution_name,
                version,
                config_values
            )
            print("Configuration API call completed successfully")
            
            # STEP 3.1: GET Configuration to verify the values were set correctly
            print("\n" + "=" * 50)
            print("STEP 3.1: Getting Configuration to verify values")
            print("=" * 50)
            try:
                get_response = get_configuration_api_call(
                    credential,
                    subscription_id,
                    resource_group_name,
                    config_name,
                    solution_name,
                    version
                )
                if get_response:
                    print("Configuration GET call completed successfully")
                else:
                    print("Configuration GET call returned no data")
            except Exception as get_error:
                print(f"Configuration GET call failed: {get_error}")
            
        except Exception as e:
            print(f"Configuration API call failed (continuing with workflow): {e}")
            # Continue with the workflow even if Configuration API fails

        # Review target using the extracted solution template version ID
        print("=" * 50)
        print("STEP 4: Review Target Deployment")
        print("=" * 50)
        print(f"Using solution template version ID: {solution_template_version_id}")

        solution_version_id = review_target(
            workload_client,
            resource_group_name,
            target.name,
            solution_template_version_id
        )

        print("=" * 50)
        print("STEP 5: Publish and Install Solution")
        print("=" * 50)
        print("The workflow has completed the following steps:")
        print("✓ Context management with capabilities")
        print("✓ Schema creation")
        print("✓ Solution template creation")
        print("✓ Target creation")
        print("✓ Configuration API calls")
        print("✓ Target review")
        print()
        print("TARGET INFORMATION:")
        print(f"  Name: {target.name}")
        print(f"  Resource Group: {resource_group_name}")
        print(f"  Capabilities: {capabilities}")
        print()
        print("CONFIGURATION COMPLETED:")
        print(f"  Config Name: {target.name}Config")
        print(f"  Solution Name: sdkexamples-solution1")
        print()
        print("Proceeding with publish and install operations...")

        # Publish target
        publish_result = publish_target(
            workload_client,
            resource_group_name,
            target.name,
            solution_version_id
        )

        # Install target
        install_result = install_target(
            workload_client,
            resource_group_name,
            target.name,
            solution_version_id
        )


    except HttpResponseError as e:
        print(f"An HTTP error occurred: {e.message}")
    except Exception as e:
        print(f"An unexpected error occurred: {e}")

if __name__ == "__main__":
    main()
