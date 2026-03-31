# Comprehensive Guide to Upjet AWS Provider Configuration Patterns

**Document Purpose**: Complete catalog of configuration patterns used in the AWS provider's `config/` directory, which controls code generation for all 100+ AWS resources.

**Last Updated**: Based on analysis of provider-aws/v2 config/

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [External Name Configuration Patterns](#external-name-configuration-patterns)
3. [Resource Configurator Patterns](#resource-configurator-patterns)
4. [Reference Configuration Patterns](#reference-configuration-patterns)
5. [Sensitivity and Connection Details](#sensitivity-and-connection-details)
6. [Late Initialization](#late-initialization)
7. [Custom Diff Handlers](#custom-diff-handlers)
8. [Terraform Configuration Injection](#terraform-configuration-injection)
9. [Mutual Exclusion Handling](#mutual-exclusion-handling)
10. [Service-Specific Examples](#service-specific-examples)
11. [Common Patterns and Best Practices](#common-patterns-and-best-practices)
12. [Common Pitfalls](#common-pitfalls)

---

## Architecture Overview

### Directory Structure

```
config/
├── externalname.go              # External name configurations (3600+ lines)
├── groups.go                    # Group/Kind mapping overrides
├── overrides.go                 # Global resource options
├── cluster/                     # Cluster-scoped resource configs
│   ├── provider.go             # Main provider init
│   ├── common/                 # Shared utilities
│   │   ├── common.go          # ARN/ID extractors, password generation
│   │   └── apis/              # Custom extractors
│   ├── ec2/config.go
│   ├── iam/config.go
│   ├── s3/config.go
│   └── [100+ services]/config.go
└── namespaced/                 # Namespaced resource configs (parallel to cluster/)
    └── [same structure]
```

### Two Architecture Variants

The provider supports two parallel architectures:
- **`config/cluster/`** - Cluster-scoped resources
- **`config/namespaced/`** - Namespace-scoped resources (same service configs, different instantiation)

Each has identical configuration patterns but separate initialization.

### Configuration Flow

```
1. External Name Lookup (externalname.go)
   ↓
2. Resource Configurator Called (service/config.go)
   ├─ Add references
   ├─ Set late initializer
   ├─ Configure custom diff
   ├─ Mark sensitive fields
   └─ Configure connection details
   ↓
3. Global Overrides Applied (overrides.go)
   ├─ Group/Kind assignment
   ├─ Required region
   ├─ Tags handling
   └─ Known referencers
   ↓
4. Code Generation
```

---

## External Name Configuration Patterns

**File**: `config/externalname.go` (3676 lines)

External name configuration determines how the Crossplane resource ID maps to/from the Terraform resource ID. This is THE most critical configuration.

### Pattern 1: IdentifierFromProvider (Provider-Assigned ID)

**Usage**: Resources where AWS assigns the ID (most common)

```go
"aws_sqs_queue": config.IdentifierFromProvider,
"aws_sns_topic": config.IdentifierFromProvider,
"aws_dynamodb_table": config.IdentifierFromProvider,
```

**Behavior**:
- Extract ID from Terraform state field `id`
- Pass raw ID to Terraform during import
- No transformation needed

**When to Use**: When AWS generates the ID and Terraform uses it directly

---

### Pattern 2: NameAsIdentifier

**Usage**: Resources where the user-defined name IS the Terraform ID

```go
"aws_iam_policy": config.NameAsIdentifier,
"aws_iam_role": config.NameAsIdentifier,
"aws_kms_key": config.NameAsIdentifier,
"aws_elasticache_serverless_cache": config.NameAsIdentifier,
```

**Behavior**:
- Use the `name` field value as the external name
- Omit `name` from Terraform parameters (used as ID instead)
- Automatically set `name` in Terraform config from external name

**Example**:
```yaml
apiVersion: iam.aws.upbound.io/v1beta1
kind: Policy
metadata:
  name: my-policy  # This becomes the external name AND the AWS name
spec:
  forProvider:
    # name is omitted; derived from metadata.name
```

---

### Pattern 3: ParameterAsIdentifier

**Usage**: Resources where a specific parameter field is the ID

```go
"aws_s3_directory_bucket": config.ParameterAsIdentifier("bucket"),
"aws_osis_pipeline": config.ParameterAsIdentifier("pipeline_name"),
```

**Behavior**:
- Extract ID from specified parameter field (e.g., `bucket`, `pipeline_name`)
- Use that field as external name
- Omit field from parameters (used as ID instead)

---

### Pattern 4: TemplatedStringAsIdentifier

**Usage**: Resources where ID is constructed from multiple fields + template

```go
"aws_batch_job_queue": config.TemplatedStringAsIdentifier(
    "name",  // parameter to use as identifier
    fullARNTemplate("batch", "job-queue/{{ .external_name }}")
),

"aws_msk_single_scram_secret_association": config.TemplatedStringAsIdentifier(
    "",
    "{{ .parameters.cluster_arn }},{{ .parameters.secret_arn }}"
),
```

**Template Variables**:
- `{{ .external_name }}` - The user-defined external name
- `{{ .parameters.FIELD }}` - Any parameter from spec.forProvider
- `{{ .setup.configuration.region }}` - Region
- `{{ .setup.client_metadata.account_id }}` - AWS account ID
- `{{ .setup.client_metadata.partition }}` - AWS partition (aws, aws-cn, etc.)

**Full ARN Template**:
```go
fullARNTemplate(service, resource)
// Expands to: arn:{{ partition }}:{{ service }}:{{ region }}:{{ account }}:{{ resource }}

// Example:
"aws_batch_job_queue": config.TemplatedStringAsIdentifier(
    "name",
    fullARNTemplate("batch", "job-queue/{{ .external_name }}")
),
// Result: arn:aws:batch:us-west-2:123456789012:job-queue/my-queue
```

**Regionless ARN Template**:
```go
regionlessARNTemplate(service, resource)
// Expands to: arn:{{ partition }}:{{ service }}::{{ account }}:{{ resource }}
// Used for IAM, etc. (no region)

// Example:
regionlessARNTemplate("iam", "role/{{ .external_name }}")
// Result: arn:aws:iam::123456789012:role/my-role
```

---

### Pattern 5: FormattedIdentifierFromProvider

**Usage**: ID built from spec fields only (no user input)

```go
FormattedIdentifierFromProvider("/", "account_id", "service_principal")
// Result: 123456789012/config.amazonaws.com

FormattedIdentifierFromProvider(",", "global_network_id", "attachment_id")
// Result: global-network-1234,attachment-5678
```

**Helper Function**:
```go
func FormattedIdentifierFromProvider(separator string, keys ...string) config.ExternalName
```

---

### Pattern 6: FormattedIdentifierUserDefinedNameLast

**Usage**: ID with spec fields + user name at END

```go
// Example: budget_id:budget_name
FormattedIdentifierUserDefinedNameLast(
    "name",  // parameter for external name
    ":",     // separator
    "budget_id",  // spec field
)
```

**Behavior**:
- Extracts `name` from parameters
- Uses it as external name (last component)
- Joins with spec field using separator
- Terraform ID: `budget_id:budget_name`
- External name: `budget_name`

---

### Pattern 7: FormattedIdentifierUserDefinedNameFirst

**Usage**: ID with user name at START, spec fields after

```go
// Example: budget_name:product_id
FormattedIdentifierUserDefinedNameFirst(
    "name",
    ":",
    "product_id",  // spec field
)
```

---

### Pattern 8: Custom ExternalName Functions

**Usage**: Complex, service-specific ID extraction

```go
func apiGatewayAccount() config.ExternalName {
    e := config.IdentifierFromProvider
    e.GetIDFn = func(ctx context.Context, externalName string, _ map[string]any, _ map[string]any) (string, error) {
        if len(externalName) == 0 {
            return "api-gateway-account", nil  // Magic constant for singleton
        }
        return externalName, nil
    }
    return e
}

"aws_api_gateway_account": apiGatewayAccount(),
```

**Custom Function Components**:

```go
type ExternalName struct {
    // Extract external name from Terraform state
    GetExternalNameFn func(tfstate map[string]any) (string, error)
    
    // Build Terraform ID from external name + parameters
    GetIDFn func(ctx context.Context, externalName string, parameters map[string]any, setup map[string]any) (string, error)
    
    // Set identifier in Terraform parameters from external name
    SetIdentifierArgumentFn func(base map[string]any, externalName string)
    
    // Fields to omit from Terraform parameters (used as ID instead)
    OmittedFields []string
    
    // Fields that together form the identifier
    IdentifierFields []string
    
    // Disable the automatic name initializer
    DisableNameInitializer bool
}
```

---

### Pattern 9: Composite Key Examples

**Multi-part IDs separated by special characters**:

```go
// ECS Cluster: Extract cluster name from ARN
"aws_ecs_cluster": func() {
    e := config.IdentifierFromProvider
    e.GetExternalNameFn = func(tfstate map[string]interface{}) (string, error) {
        // ID format: arn:aws:ecs:region:account:cluster/cluster-name
        parts := strings.Split(tfstate["id"].(string), "/")
        return parts[len(parts)-1], nil  // cluster-name
    }
    return e
}(),

// AppConfig Environment: account_id:application_id:environment_id
"aws_appconfig_environment": appConfigEnvironment(),

// Network Manager Link Association: global_network_id,link_id,device_id
"aws_networkmanager_link_association": 
    config.TemplatedStringAsIdentifier("", 
        "{{ .parameters.global_network_id }},{{ .parameters.link_id }},{{ .parameters.device_id }}"
    ),
```

---

### Pattern 10: Stub Values for Computed IDs

**Usage**: Resources where ID is computed but Terraform requires a value**

```go
"aws_bedrock_inference_profile": identifierFromProviderWithDefaultStub("bedrock12345"),
"aws_route53profiles_profile": identifierFromProviderWithDefaultStub("rp-stub123456"),

func identifierFromProviderWithDefaultStub(stub string) config.ExternalName {
    e := config.IdentifierFromProvider
    // During creation, provide stub; replace with actual ID from Terraform state
    return e
}
```

---

### Pattern 11: No External Name Configuration

**Some resources use defaults**:

```go
"aws_vpc": // Uses defaults from groups.go
```

---

## Resource Configurator Patterns

**Files**: `config/cluster/[SERVICE]/config.go`

Each service has a `Configure(p *config.Provider)` function that uses `p.AddResourceConfigurator()` to customize individual resources.

### Basic Structure

```go
package ec2

import "github.com/crossplane/upjet/v2/pkg/config"

func Configure(p *config.Provider) {
    p.AddResourceConfigurator("aws_instance", func(r *config.Resource) {
        // 1. Configure references
        r.References["subnet_id"] = config.Reference{
            TerraformName: "aws_subnet",
        }
        
        // 2. Configure late initialization
        r.LateInitializer = config.LateInitializer{
            IgnoredFields: []string{"availability_zone"},
        }
        
        // 3. Configure custom diff
        r.TerraformCustomDiff = customDiffFn
        
        // 4. Other options
        r.UseAsync = true
    })
}
```

---

## Reference Configuration Patterns

**Purpose**: Define how Crossplane resources reference other resources

### Pattern 1: Simple Reference (Single Field)

```go
r.References["subnet_id"] = config.Reference{
    TerraformName: "aws_subnet",
}
```

**Result**:
- Generates `SubnetIDRef` and `SubnetIDSelector` fields
- User can use `spec.forProvider.subnetIDRef.name` OR `spec.forProvider.subnetIDSelector.matchLabels`
- Automatically resolves to subnet ID

---

### Pattern 2: Array References (Multiple Resources)

```go
r.References["vpc_security_group_ids"] = config.Reference{
    TerraformName:     "aws_security_group",
    RefFieldName:      "VPCSecurityGroupIDRefs",      // Custom plural field name
    SelectorFieldName: "VPCSecurityGroupIDSelector",  // Custom plural selector
}
```

**Result**:
- Generates `spec.forProvider.vpcSecurityGroupIDRefs[]` (array of refs)
- Generates `spec.forProvider.vpcSecurityGroupIDSelector` (label selector)
- Resolves to multiple security group IDs

---

### Pattern 3: Nested Field References

```go
r.References["vpc_config.security_group_ids"] = config.Reference{
    TerraformName:     "aws_security_group",
    RefFieldName:      "SecurityGroupIDRefs",
    SelectorFieldName: "SecurityGroupIDSelector",
}

r.References["block_device_mappings.ebs.kms_key_id"] = config.Reference{
    TerraformName: "aws_kms_key",
}

r.References["network_interfaces.security_groups"] = config.Reference{
    TerraformName:     "aws_security_group",
    RefFieldName:      "SecurityGroupRefs",
    SelectorFieldName: "SecurityGroupSelector",
}
```

---

### Pattern 4: Custom Value Extractors

**Purpose**: Extract non-ID values for references (e.g., ARNs)**

```go
r.References["role"] = config.Reference{
    TerraformName: "aws_iam_role",
    Extractor:     common.PathARNExtractor,  // Get ARN instead of ID
}

r.References["policy_arn"] = config.Reference{
    TerraformName: "aws_iam_policy",
    Extractor:     common.PathARNExtractor,  // Get ARN
}

r.References["stream_arn"] = config.Reference{
    TerraformName: "aws_kinesis_stream",
    Extractor:     common.PathTerraformIDExtractor,  // Get Terraform ID
}
```

**Built-in Extractors** (in `config/cluster/common/apis/extractor.go`):

```go
common.PathARNExtractor              // Extract status.atProvider.arn
common.PathTerraformIDExtractor      // Extract resource ID directly
```

**Custom Extractors**:

```go
r.References["endpoint_config"] = config.Reference{
    TerraformName: "aws_custom_resource",
    Extractor: `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("arn",true)`,
}
```

---

### Pattern 5: Custom Extractor Functions

```go
func IntegrationIDPrefixed() reference.ExtractValueFn {
    return func(mg xpresource.Managed) string {
        // Add custom prefix to extracted value
        return "integrations/" + meta.GetExternalName(mg)
    }
}

"aws_apigatewayv2_integration": apis.IntegrationIDPrefixed(),
```

---

### Pattern 6: Deleting Auto-Generated References

```go
// Remove auto-generated reference (causing circular dependency)
delete(r.References, "lambda_function.lambda_function_arn")

// Remove reference that's too ambiguous
delete(r.References, "bucket")

// Remove because field is too broad
delete(r.References, "event_source_arn")  // Can be multiple types
```

---

## Sensitivity and Connection Details

### Pattern 1: AdditionalConnectionDetailsFn

**Purpose**: Extract non-secret sensitive data from resource attributes and expose in Secret

```go
r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
    conn := map[string][]byte{}
    
    if a, ok := attr["endpoint"].(string); ok {
        conn["endpoint"] = []byte(a)
    }
    if a, ok := attr["address"].(string); ok {
        conn["address"] = []byte(a)
        conn["host"] = []byte(a)  // Alias
    }
    if a, ok := attr["port"]; ok {
        conn["port"] = []byte(fmt.Sprintf("%v", a))
    }
    if a, ok := attr["username"].(string); ok {
        conn["username"] = []byte(a)
    }
    if a, ok := attr["password"].(string); ok {
        conn["password"] = []byte(a)
    }
    
    return conn, nil
}
```

**Example: RDS Database Connection Details**

```go
// aws_db_instance
r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
    conn := map[string][]byte{}
    if a, ok := attr["endpoint"].(string); ok {
        conn["endpoint"] = []byte(a)
    }
    if a, ok := attr["address"].(string); ok {
        conn["address"] = []byte(a)  // Also extract address component
        conn["host"] = []byte(a)
    }
    if a, ok := attr["username"].(string); ok {
        conn["username"] = []byte(a)
    }
    if a, ok := attr["port"]; ok {
        conn["port"] = []byte(fmt.Sprintf("%v", a))
    }
    if a, ok := attr["password"].(string); ok {
        conn["password"] = []byte(a)
    }
    return conn, nil
}
```

**Result**: Secret created with keys like:
```yaml
data:
  endpoint: "mydb.us-west-2.rds.amazonaws.com:3306"
  host: "mydb.us-west-2.rds.amazonaws.com"
  port: "3306"
  username: "admin"
  password: "auto-generated-or-provided"
```

---

### Pattern 2: Mark Fields as Sensitive in Schema

```go
p.AddResourceConfigurator("aws_cloudfront_function", func(r *config.Resource) {
    // Allow secret reference for code field
    r.TerraformResource.Schema["code"].Sensitive = true
})

p.AddResourceConfigurator("aws_cloudfront_public_key", func(r *config.Resource) {
    // Allow secret reference for key field
    r.TerraformResource.Schema["encoded_key"].Sensitive = true
})
```

---

## Late Initialization

**Purpose**: Determine which fields should be late-initialized from AWS state

### Pattern 1: IgnoredFields

```go
r.LateInitializer = config.LateInitializer{
    IgnoredFields: []string{
        "availability_zone",       // Conflicts with availability_zone_id
        "db_name",
        "name",
        "subnet_id",               // Conflicts with network_interface
        "source_dest_check",
        "vpc_security_group_ids",
    },
}
```

**Behavior**:
- Fields listed are NOT late-initialized from AWS state
- Typically used for fields that conflict with other fields or are rarely modified

**Example**: EC2 Instance
- AWS auto-assigns `availability_zone`
- But user might specify `availability_zone_id`
- These conflict, so ignore `availability_zone` in late init

---

### Pattern 2: ConditionalIgnoredFields

```go
r.LateInitializer = config.LateInitializer{
    IgnoredFields: []string{"enabled_cloudwatch_logs_exports"},
    ConditionalIgnoredFields: []string{"scaling_config"},  // Only if another field is set
}
```

---

## Custom Diff Handlers

**Purpose**: Suppress spurious diffs between desired and actual state

### Pattern 1: Basic Custom Diff

```go
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, state *terraform.InstanceState, config *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
    if diff == nil || diff.Empty() || diff.Destroy {
        return diff, nil
    }
    
    // Suppress diff if old and new are equivalent
    if diff.Attributes["policy"] != nil && 
       diff.Attributes["policy"].Old != "" && 
       diff.Attributes["policy"].New != "" {
        
        // Compare with AWS policy equivalence
        ok, err := awspolicy.PoliciesAreEquivalent(old, new)
        if ok {
            delete(diff.Attributes, "policy")
        }
    }
    
    return diff, nil
}
```

### Pattern 2: Remove Diff on Empty Values

```go
r.TerraformCustomDiff = common.RemoveDiffIfEmpty([]string{
    "volume_tags.%",  // Remove count diff if empty
})
```

**Helper Function**:

```go
func RemoveDiffIfEmpty(keys []string) config.CustomDiff {
    return func(diff *terraform.InstanceDiff, state *terraform.InstanceState, config *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
        // Skip on create
        if state == nil || state.Empty() {
            return diff, nil
        }
        if diff == nil || diff.Empty() {
            return diff, nil
        }
        
        // Remove diff for empty keys
        for _, key := range keys {
            if diff.Attributes[key] != nil && 
               diff.Attributes[key].Old == "" && 
               diff.Attributes[key].New == "" {
                delete(diff.Attributes, key)
            }
        }
        return diff, nil
    }
}
```

---

### Pattern 3: Complex Diff - Version Comparison (RDS)

```go
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, _ *terraform.InstanceState, _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
    if diff == nil || diff.Destroy {
        return diff, nil
    }
    
    // Ignore engine version diff if desired version < actual version
    // (AWS RDS doesn't allow downgrades)
    if evDiff, ok := diff.Attributes["engine_version"]; ok && evDiff.Old != "" && evDiff.New != "" {
        comparison := utils.CompareEngineVersions(evDiff.New, evDiff.Old)
        if comparison <= 0 {  // Desired <= Actual, suppress diff
            delete(diff.Attributes, "engine_version")
        }
    }
    return diff, nil
}
```

---

### Pattern 4: Policy Equivalence (SNS, SQS)

```go
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, _ *terraform.InstanceState, _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
    if diff == nil || diff.Attributes["policy"] == nil {
        return diff, nil
    }
    
    // Remove "Version" field from policies before comparison
    vOld, _ := common.RemovePolicyVersion(diff.Attributes["policy"].Old)
    vNew, _ := common.RemovePolicyVersion(diff.Attributes["policy"].New)
    
    // Check if semantically equivalent
    ok, _ := awspolicy.PoliciesAreEquivalent(vOld, vNew)
    if ok {
        delete(diff.Attributes, "policy")
    }
    
    return diff, nil
}
```

---

## Terraform Configuration Injection

**Purpose**: Set default values or transform user input before sending to Terraform

### Pattern 1: Basic TerraformConfigurationInjector

```go
r.TerraformConfigurationInjector = func(jsonMap map[string]any, params map[string]any) error {
    // Set default if not provided
    if _, ok := jsonMap["forceDestroy"]; !ok {
        params["force_destroy"] = false
    }
    
    // Pass region from state to params
    params["region"] = jsonMap["region"]
    
    return nil
}
```

**Pattern in S3**:

```go
r.TerraformConfigurationInjector = func(jsonMap map[string]any, params map[string]any) error {
    params["region"] = jsonMap["region"]
    
    // Default force_destroy to false if not specified
    if _, ok := jsonMap["forceDestroy"]; !ok {
        params["force_destroy"] = false
    }
    return nil
}
```

---

### Pattern 2: Disable Auto-Generation

```go
r.TerraformConfigurationInjector = func(_ map[string]any, params map[string]any) error {
    // Clear auto-generated fields
    params["name_prefix"] = ""
    return nil
}
```

---

### Pattern 3: Set Conditional Defaults

```go
r.TerraformConfigurationInjector = func(jsonMap map[string]any, params map[string]any) error {
    // Default ACL to private if not provided
    if _, ok := jsonMap["acl"]; !ok {
        params["acl"] = "private"
    }
    return nil
}
```

---

## Mutual Exclusion Handling

**Purpose**: Move fields to status that are managed by separate resources

### Pattern 1: MoveToStatus

```go
config.MoveToStatus(r.TerraformResource, 
    "acceleration_status",      // Managed by aws_s3_bucket_accelerate_configuration
    "acl",                      // Managed by aws_s3_bucket_acl
    "grant",                    // Managed by aws_s3_bucket_acl
    "cors_rule",                // Managed by aws_s3_bucket_cors_configuration
    "lifecycle_rule",           // Managed by aws_s3_bucket_lifecycle_configuration
    "logging",                  // Managed by aws_s3_bucket_logging
    "policy",                   // Managed by aws_s3_bucket_policy
    "website",                  // Managed by aws_s3_bucket_website_configuration
)
```

**Effect**:
- Moves listed fields from `spec.forProvider` to `status.atProvider`
- User CANNOT set via spec
- Prevents conflicts with separate Crossplane resources

**Examples**:

```go
// S3 Bucket
config.MoveToStatus(r.TerraformResource, 
    "ingress", "egress"  // Managed by aws_security_group_rule
)

// Security Group
config.MoveToStatus(r.TerraformResource, 
    "ingress", "egress"
)

// ECS Cluster
config.MoveToStatus(r.TerraformResource, 
    "capacity_providers"  // Managed by aws_ecs_cluster_capacity_providers
)

// RDS Cluster
config.MoveToStatus(r.TerraformResource, 
    "iam_roles"  // Managed by aws_rds_cluster_role_association
)
```

---

## Terraform Configuration Injection: Security Group Self Reference

```go
r.TerraformConfigurationInjector = func(jsonMap map[string]any, params map[string]any) error {
    // Default "self" to false if not specified
    if _, ok := jsonMap["self"]; !ok {
        params["self"] = false
    }
    return nil
}
```

---

## Service-Specific Examples

### EC2

**Key Patterns**:
- Heavy use of nested references (`root_block_device.kms_key_id`)
- Late initializer to handle conflicting fields
- Custom diff to suppress spurious diffs
- Move inline ingress/egress to status (use separate SecurityGroupRule)

```go
p.AddResourceConfigurator("aws_instance", func(r *config.Resource) {
    r.UseAsync = true
    r.References["subnet_id"] = config.Reference{
        TerraformName: "aws_subnet",
    }
    r.References["vpc_security_group_ids"] = config.Reference{
        TerraformName: "aws_security_group",
        RefFieldName: "VPCSecurityGroupIDRefs",
        SelectorFieldName: "VPCSecurityGroupIDSelector",
    }
    r.References["root_block_device.kms_key_id"] = config.Reference{
        TerraformName: "aws_kms_key",
    }
    r.LateInitializer = config.LateInitializer{
        IgnoredFields: []string{
            "subnet_id",
            "network_interface",
            "private_ip",
            "vpc_security_group_ids",
            "availability_zone",
        },
    }
    r.TerraformCustomDiff = common.RemoveDiffIfEmpty([]string{"volume_tags.%"})
    config.MoveToStatus(r.TerraformResource, "security_groups")
})
```

---

### RDS

**Key Patterns**:
- Custom version comparison (downgrades not allowed)
- Auto-password generation
- Complex connection details extraction
- Custom documentation

```go
p.AddResourceConfigurator("aws_db_instance", func(r *config.Resource) {
    r.UseAsync = true
    
    // Password auto-generation
    desc, _ := comments.New("If true, password will be auto-generated...")
    r.TerraformResource.Schema["auto_generate_password"] = &schema.Schema{
        Type: schema.TypeBool,
        Optional: true,
        Description: desc.String(),
    }
    r.InitializerFns = append(r.InitializerFns,
        common.PasswordGenerator(
            "spec.forProvider.passwordSecretRef",
            "spec.forProvider.autoGeneratePassword",
        ))
    
    // Connection details
    r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
        conn := map[string][]byte{}
        if a, ok := attr["endpoint"].(string); ok {
            conn["endpoint"] = []byte(a)
        }
        if a, ok := attr["address"].(string); ok {
            conn["address"] = []byte(a)
            conn["host"] = []byte(a)
        }
        if a, ok := attr["username"].(string); ok {
            conn["username"] = []byte(a)
        }
        if a, ok := attr["port"]; ok {
            conn["port"] = []byte(fmt.Sprintf("%v", a))
        }
        if a, ok := attr["password"].(string); ok {
            conn["password"] = []byte(a)
        }
        return conn, nil
    }
    
    // Version comparison diff
    r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, _ *terraform.InstanceState, _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
        if evDiff, ok := diff.Attributes["engine_version"]; ok && evDiff.Old != "" && evDiff.New != "" {
            c := utils.CompareEngineVersions(evDiff.New, evDiff.Old)
            if c <= 0 {
                delete(diff.Attributes, "engine_version")
            }
        }
        return diff, nil
    }
})
```

---

### S3

**Key Patterns**:
- Heavy use of MoveToStatus for sub-resources
- Custom external name (random RFC1123 subdomain)
- Connection details extraction
- Configuration injection

```go
p.AddResourceConfigurator("aws_s3_bucket", func(r *config.Resource) {
    // Random name generation
    r.MetaResource.ExternalName = registry.RandRFC1123Subdomain
    
    // Move sub-resources to status
    config.MoveToStatus(r.TerraformResource, 
        "acceleration_status", "acl", "grant", "cors_rule", "lifecycle_rule",
        "logging", "object_lock_configuration", "policy", "replication_configuration",
        "request_payer", "server_side_encryption_configuration", "versioning",
        "website", "arn")
    
    // Connection details
    r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
        conn := map[string][]byte{}
        if a, ok := attr["id"].(string); ok {
            conn["id"] = []byte(a)
        }
        if a, ok := attr["arn"].(string); ok {
            conn["arn"] = []byte(a)
        }
        if a, ok := attr["region"].(string); ok {
            conn["region"] = []byte(a)
        }
        return conn, nil
    }
    
    // Default values
    r.TerraformConfigurationInjector = func(jsonMap map[string]any, params map[string]any) error {
        params["region"] = jsonMap["region"]
        if _, ok := jsonMap["forceDestroy"]; !ok {
            params["force_destroy"] = false
        }
        return nil
    }
})
```

---

### IAM

**Key Patterns**:
- Simple name-based external names
- Limited use of references
- Custom documentation for conflict fields
- Mark as required when needed

```go
p.AddResourceConfigurator("aws_iam_policy", func(r *config.Resource) {
    // Ensure name is required (TF makes it optional with generation)
    r.MarkAsRequired("name")
})

p.AddResourceConfigurator("aws_iam_role", func(r *config.Resource) {
    r.MetaResource.ArgumentDocs["inline_policy"] = 
        "Configuration block defining exclusive set of IAM inline policies..."
    r.MetaResource.ArgumentDocs["managed_policy_arns"] = 
        "Set of exclusive IAM managed policy ARNs..."
    
    // Ignore managed policies in late init to avoid conflicts with separate resources
    r.LateInitializer.IgnoredFields = append(
        r.LateInitializer.IgnoredFields, 
        "managed_policy_arns", "inline_policy")
})
```

---

### EKS

**Key Patterns**:
- Custom extractors for cluster name references
- Use of Terraform ID extractor (for resources created before external name set)
- Conditional ignored fields
- Custom documentation

```go
p.AddResourceConfigurator("aws_eks_node_group", func(r *config.Resource) {
    r.References["cluster_name"] = config.Reference{
        TerraformName: "aws_eks_cluster",
        Extractor: "ExternalNameIfClusterActive()",  // Custom extractor
    }
    
    r.LateInitializer = config.LateInitializer{
        IgnoredFields: []string{
            "release_version",
            "version",
        },
        ConditionalIgnoredFields: []string{
            "scaling_config",  // Only if specified elsewhere
        },
    }
    
    r.MetaResource.ArgumentDocs["subnet_ids"] = 
        "- Identifiers of EC2 Subnets. Private subnets must have tag kubernetes.io/role/internal-elb..."
})
```

---

### DynamoDB

**Key Patterns**:
- Simple references
- Delete conflicting auto-generated references
- Path-based ARN extraction

```go
p.AddResourceConfigurator("aws_dynamodb_kinesis_streaming_destination", func(r *config.Resource) {
    r.References["table_name"] = config.Reference{
        TerraformName: "aws_dynamodb_table",
    }
    r.References["stream_arn"] = config.Reference{
        TerraformName: "aws_kinesis_stream",
        Extractor: common.PathTerraformIDExtractor,  // Get stream ARN
    }
})

p.AddResourceConfigurator("aws_dynamodb_table_item", func(r *config.Resource) {
    r.References["table_name"] = config.Reference{
        TerraformName: "aws_dynamodb_table",
    }
    delete(r.References, "hash_key")  // Too ambiguous
})
```

---

### Lambda

**Key Patterns**:
- Complex nested references
- Delete computed field references
- Custom late initializer (source_code_hash)
- Schema modification (delete filename)

```go
p.AddResourceConfigurator("aws_lambda_function", func(r *config.Resource) {
    r.References["s3_bucket"] = config.Reference{
        TerraformName: "aws_s3_bucket",
    }
    r.References["role"] = config.Reference{
        TerraformName: "aws_iam_role",
        Extractor: common.PathARNExtractor,
    }
    r.References["vpc_config.security_group_ids"] = config.Reference{
        TerraformName: "aws_security_group",
        RefFieldName: "SecurityGroupIDRefs",
        SelectorFieldName: "SecurityGroupIDSelector",
    }
    
    // Remove filename field (requires local file)
    delete(r.TerraformResource.Schema, "filename")
    
    r.LateInitializer = config.LateInitializer{
        IgnoredFields: []string{"source_code_hash"},
    }
})
```

---

### SecretsManager

**Key Patterns**:
- Complex custom diff (replica set handling)
- Move related fields to status (rotation, policy)
- Configuration injection

```go
p.AddResourceConfigurator("aws_secretsmanager_secret", func(r *config.Resource) {
    // Move to separate resources
    config.MoveToStatus(r.TerraformResource, 
        "rotation_rules", "rotation_lambda_arn",  // Use RotationRule
        "policy",  // Use SecretPolicy
    )
    
    // Complex custom diff for replica sets
    r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
        // [Complex logic to suppress spurious replica diffs]
        // Compares region, kms_key_id pairs between desired and current state
        // Suppresses diffs when semantically equivalent
        return diff, nil
    }
})
```

---

### ElastiCache

**Key Patterns**:
- Auth token auto-generation
- Connection details with multiple endpoints
- Schema modification to add custom field
- Version migration with custom converters

```go
p.AddResourceConfigurator("aws_elasticache_replication_group", func(r *config.Resource) {
    // Add custom field
    r.TerraformResource.Schema["auto_generate_auth_token"] = &schema.Schema{
        Type: schema.TypeBool,
        Optional: true,
    }
    
    // Add initializer for auto-generation
    r.InitializerFns = append(r.InitializerFns,
        common.PasswordGenerator(
            "spec.forProvider.authTokenSecretRef",
            "spec.forProvider.autoGenerateAuthToken",
        ))
    
    // Connection details
    r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
        conn := map[string][]byte{}
        if a, ok := attr["configuration_endpoint_address"].(string); ok {
            conn["configuration_endpoint_address"] = []byte(a)
        }
        if a, ok := attr["port"]; ok {
            conn["port"] = []byte(fmt.Sprintf("%v", a))
        }
        return conn, nil
    }
    
    // Version migration
    r.Version = "v1beta2"
    r.Conversions = append(r.Conversions,
        conversion.NewCustomConverter("v1beta1", "v1beta2", customConversionFn))
})
```

---

### SNS/SQS

**Key Patterns**:
- Policy equivalence checking
- ARN extraction for references
- Late initializer for policy conflicts
- Custom documentation

```go
p.AddResourceConfigurator("aws_sns_topic", func(r *config.Resource) {
    // Ignore policy in late init (managed by sns_topic_policy)
    r.LateInitializer.IgnoredFields = append(
        r.LateInitializer.IgnoredFields, "policy")
    
    // Custom diff for policy equivalence
    r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
        if diff.Attributes["policy"] == nil {
            return diff, nil
        }
        
        // Remove Version field and compare
        vOld, _ := common.RemovePolicyVersion(diff.Attributes["policy"].Old)
        vNew, _ := common.RemovePolicyVersion(diff.Attributes["policy"].New)
        ok, _ := awspolicy.PoliciesAreEquivalent(vOld, vNew)
        if ok {
            delete(diff.Attributes, "policy")
        }
        return diff, nil
    }
    
    // Connection details
    r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
        conn := map[string][]byte{}
        if a, ok := attr["arn"].(string); ok {
            conn["arn"] = []byte(a)
        }
        return conn, nil
    }
})
```

---

### Route53

**Key Patterns**:
- Custom diff for FQDN normalization
- Simple references to zones and health checks

```go
p.AddResourceConfigurator("aws_route53_record", func(r *config.Resource) {
    r.References["zone_id"] = config.Reference{
        TerraformName: "aws_route53_zone",
    }
    r.References["health_check_id"] = config.Reference{
        TerraformName: "aws_route53_health_check",
    }
    
    // Suppress diff for trailing dot normalization
    r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
        if nameDiff, ok := diff.Attributes["name"]; ok {
            if strings.TrimSuffix(nameDiff.New, ".") == strings.TrimSuffix(nameDiff.Old, ".") {
                delete(diff.Attributes, "name")
            }
        }
        return diff, nil
    }
})
```

---

### CloudFront

**Key Patterns**:
- Use async for slow operations
- Mark fields as sensitive for secret handling
- Delete problematic domain references

```go
p.AddResourceConfigurator("aws_cloudfront_distribution", func(r *config.Resource) {
    r.UseAsync = true
    delete(r.References, "origin.domain_name")  // Circular dependency
})

p.AddResourceConfigurator("aws_cloudfront_function", func(r *config.Resource) {
    // Mark code as sensitive to allow secret reference
    r.TerraformResource.Schema["code"].Sensitive = true
})
```

---

## Common Patterns and Best Practices

### 1. **When to Use References**

```go
// ✅ DO: Reference to another Crossplane resource
r.References["subnet_id"] = config.Reference{
    TerraformName: "aws_subnet",
}

// ✅ DO: Array reference with proper field names
r.References["vpc_security_group_ids"] = config.Reference{
    TerraformName: "aws_security_group",
    RefFieldName: "VPCSecurityGroupIDRefs",
    SelectorFieldName: "VPCSecurityGroupIDSelector",
}

// ✅ DO: Extract ARN when needed
r.References["role"] = config.Reference{
    TerraformName: "aws_iam_role",
    Extractor: common.PathARNExtractor,
}

// ❌ DON'T: Reference ambiguous fields
delete(r.References, "event_source_arn")  // Can be multiple types

// ❌ DON'T: Reference if it causes circular dependency
delete(r.References, "lambda_function.lambda_function_arn")
```

---

### 2. **When to Use Late Initializer**

```go
// ✅ DO: Ignore conflicting fields
r.LateInitializer = config.LateInitializer{
    IgnoredFields: []string{
        "availability_zone",      // Conflicts with availability_zone_id
        "availability_zone_id",
    },
}

// ✅ DO: Ignore fields managed by separate resources
r.LateInitializer.IgnoredFields = append(
    r.LateInitializer.IgnoredFields, 
    "managed_policy_arns",  // Use RolePolicyAttachment instead
)

// ❌ DON'T: Late initialize something user always provides
// (It will conflict with their spec)
```

---

### 3. **When to Use Custom Diff**

```go
// ✅ DO: Suppress equivalent but differently formatted values
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
    // Check policy equivalence (ignore Version field)
    ok, _ := awspolicy.PoliciesAreEquivalent(old, new)
    if ok {
        delete(diff.Attributes, "policy")
    }
    return diff, nil
}

// ✅ DO: Suppress FQDN trailing dot differences
if strings.TrimSuffix(new, ".") == strings.TrimSuffix(old, ".") {
    delete(diff.Attributes, "name")
}

// ✅ DO: Suppress version diffs when downgrades not allowed
if compareVersions(new, old) <= 0 {
    delete(diff.Attributes, "engine_version")
}

// ❌ DON'T: Use custom diff for user input validation
// Let Terraform handle that
```

---

### 4. **When to Use MoveToStatus**

```go
// ✅ DO: Move fields managed by separate resources
config.MoveToStatus(r.TerraformResource, 
    "ingress", "egress",  // Use SecurityGroupRule instead
    "inline_policy",      // Use RolePolicy instead
    "rotation_rules",     // Use SecretRotation instead
)

// ❌ DON'T: Move fields the user should control
// Only move if you have a separate resource type for it
```

---

### 5. **Custom Extraction for Values**

```go
// ✅ DO: Extract ARN for IAM resources
Extractor: common.PathARNExtractor

// ✅ DO: Extract Terraform ID for streaming resources
Extractor: common.PathTerraformIDExtractor

// ✅ DO: Use templated extractors
Extractor: `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("arn",true)`

// ❌ DON'T: Try to extract non-existent fields
// The field must exist in status.atProvider
```

---

### 6. **Async Operations**

```go
// ✅ DO: Mark slow operations as async
r.UseAsync = true  // For RDS, EKS, ECS, CloudFront, etc.

// ❌ DON'T: Mark fast operations as async
// Only use for operations that take > 1 minute
```

---

### 7. **Connection Details Best Practices**

```go
// ✅ DO: Extract all user-needed connection info
r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
    conn := map[string][]byte{}
    
    // Extract all readable values
    if a, ok := attr["endpoint"].(string); ok {
        conn["endpoint"] = []byte(a)
    }
    if a, ok := attr["port"]; ok {
        conn["port"] = []byte(fmt.Sprintf("%v", a))
    }
    
    // Type assert and convert properly
    return conn, nil
}

// ✅ DO: Provide aliases
if a, ok := attr["address"].(string); ok {
    conn["address"] = []byte(a)
    conn["host"] = []byte(a)  // Alias
}

// ❌ DON'T: Forget type assertions
// Type assertions can panic if wrong type
```

---

## Common Pitfalls

### 1. **External Name Oscillation**

**Problem**: External name changes between reconciliations

**Cause**: `GetExternalNameFn` extracts different value than `GetIDFn` builds

```go
// ❌ WRONG:
e.GetExternalNameFn = func(tfstate map[string]any) (string, error) {
    return tfstate["id"], nil  // Returns "arn:aws:..."
}
e.GetIDFn = func(ctx, name string, params, setup map[string]any) (string, error) {
    return name + "-id", nil  // Returns "name-id"
}
// Result: "arn:aws:..." != "name-id" → oscillation

// ✅ CORRECT:
e.GetExternalNameFn = func(tfstate map[string]any) (string, error) {
    // Extract just the name part
    return strings.Split(tfstate["id"], "/")[1], nil  // "name"
}
e.GetIDFn = func(ctx, name string, params, setup map[string]any) (string, error) {
    // Build full ID from name
    return "service/" + name, nil  // "service/name"
}
// Result: "name" == "name" ✓
```

---

### 2. **Circular Dependencies via References**

**Problem**: Reference creates circular import

```go
// ❌ WRONG:
r.References["lambda_function.lambda_function_arn"] = config.Reference{
    TerraformName: "aws_lambda_function",
}
// S3 bucket references Lambda, Lambda references S3 bucket

// ✅ CORRECT:
delete(r.References, "lambda_function.lambda_function_arn")
// User must provide ARN directly
```

---

### 3. **Late Initializer Conflicts**

**Problem**: Field conflicts between init and user spec

```go
// ❌ WRONG:
r.LateInitializer.IgnoredFields = []string{
    // Don't ignore common fields users provide
}

// ✅ CORRECT:
r.LateInitializer.IgnoredFields = []string{
    "availability_zone",  // Ignore because availability_zone_id used instead
    "name",              // Ignore only if user provides it elsewhere
}
```

---

### 4. **Reference to Computed-Only Field**

**Problem**: Referencing field that's only computed

```go
// ❌ WRONG:
r.References["security_group_ids"] = config.Reference{
    TerraformName: "aws_security_group",
}
// If security_group_ids is auto-computed and user can't set it

// ✅ CORRECT:
// Check schema - field must have Required: true or Optional: true
// Only reference fields users can actually set
```

---

### 5. **Missing Type Assertions in Connection Details**

**Problem**: Type assertion panic in connection detail extraction

```go
// ❌ WRONG:
if port, ok := attr["port"]; ok {
    // Don't convert properly
    conn["port"] = []byte(port.(string))  // Might panic if int
}

// ✅ CORRECT:
if port, ok := attr["port"]; ok {
    conn["port"] = []byte(fmt.Sprintf("%v", port))  // Works for any type
}
```

---

### 6. **Overly Complex Custom Diff**

**Problem**: Custom diff logic too complex, brittle

```go
// ❌ OVERCOMPLICATED:
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
    // 200+ lines of nested logic
}

// ✅ BETTER:
// Use existing helper functions
r.TerraformCustomDiff = common.RemoveDiffIfEmpty([]string{"field.%"})

// Or wrap complex logic in a helper
r.TerraformCustomDiff = customDiffHelperFunction()
```

---

### 7. **Wrong Extractor**

**Problem**: Reference uses wrong extractor, gets wrong value

```go
// ❌ WRONG:
r.References["role"] = config.Reference{
    TerraformName: "aws_iam_role",
    // Default extracts external name (role name)
}
// But Terraform expects ARN

// ✅ CORRECT:
r.References["role"] = config.Reference{
    TerraformName: "aws_iam_role",
    Extractor: common.PathARNExtractor,  // Extract ARN
}
```

---

### 8. **Forgetting Array Reference Field Names**

**Problem**: Auto-generated field names are wrong

```go
// ❌ WRONG:
r.References["subnet_ids"] = config.Reference{
    TerraformName: "aws_subnet",
    // Generated as SubnetIdsRef, SubnetIdsSelector (wrong plural)
}

// ✅ CORRECT:
r.References["subnet_ids"] = config.Reference{
    TerraformName: "aws_subnet",
    RefFieldName: "SubnetIDRefs",  // Proper plural
    SelectorFieldName: "SubnetIDSelector",
}
```

---

### 9. **MoveToStatus But Resource Exists**

**Problem**: Field moved to status but no separate resource type

```go
// ❌ WRONG:
config.MoveToStatus(r.TerraformResource, "some_field")
// But there's no aws_service_some_field resource

// ✅ CORRECT:
// Only move if:
// 1. A separate Crossplane resource manages it, OR
// 2. It's purely informational from AWS

// Otherwise user can't set it anywhere
```

---

### 10. **Sensitive Field Without Allowing Secret Ref**

**Problem**: Field is secret but can't take secret ref

```go
// ❌ WRONG:
// Field is sensitive but no way to provide via secret

// ✅ CORRECT:
r.TerraformResource.Schema["code"].Sensitive = true
// Allows: spec.forProvider.codeSecretRef: ...
```

---

## How to Configure a New Resource

### Step-by-Step Process

#### 1. **Determine External Name Strategy**

```go
// Check AWS Terraform provider docs for import format
// Example: "aws_foo_bar can be imported using the ID foo/bar"

// Determine which pattern fits:
- IdentifierFromProvider       // AWS generates ID
- NameAsIdentifier             // User name is ID
- ParameterAsIdentifier("field")  // Field is ID
- TemplatedStringAsIdentifier(...) // Composite ID
- Custom function              // Complex logic
```

#### 2. **Add to externalname.go**

```go
// In TerraformPluginFrameworkExternalNameConfigs or TerraformPluginSDKExternalNameConfigs map

"aws_foo_bar": config.NameAsIdentifier,
// OR
"aws_foo_bar": config.TemplatedStringAsIdentifier(
    "name",
    fullARNTemplate("service", "resource/{{ .external_name }}")
),
```

#### 3. **Create Service Config File**

```go
// config/cluster/[service]/config.go

package myservice

import "github.com/crossplane/upjet/v2/pkg/config"

func Configure(p *config.Provider) {
    p.AddResourceConfigurator("aws_foo_bar", func(r *config.Resource) {
        // Configure references
        r.References["foo_id"] = config.Reference{
            TerraformName: "aws_foo",
        }
        
        // Configure late init
        r.LateInitializer = config.LateInitializer{
            IgnoredFields: []string{"computed_field"},
        }
        
        // Configure diff
        r.TerraformCustomDiff = customDiffFn
        
        // Mark as async if slow
        r.UseAsync = true
    })
}
```

#### 4. **Register in Provider**

```go
// config/cluster/provider.go

func init() {
    ProviderConfiguration.AddConfig(myservice.Configure)
}
```

#### 5. **Test Mapping**

```bash
# Generate resources and verify external name works
make generate

# Check generated CRD has correct Group/Version/Kind
# Check references work
# Check late init doesn't conflict with user spec
```

---

## Resource Configuration Checklist

- [ ] External name pattern identified and added to externalname.go
- [ ] Service config file created (config/[service]/config.go)
- [ ] References configured for all valid cross-references
- [ ] Late initializer configured to avoid conflicts
- [ ] Custom diff added if needed to suppress spurious diffs
- [ ] Sensitive fields configured with connection details
- [ ] Mutually exclusive fields moved to status if needed
- [ ] Custom documentation added for confusing fields
- [ ] Schema modifications made if needed (delete filename, add custom fields)
- [ ] Async marked if operation takes > 1 minute
- [ ] Registered in provider.go init()

---

## Pattern Summary Table

| Pattern | Use Case | Complexity | File |
|---------|----------|-----------|------|
| IdentifierFromProvider | AWS-generated ID | ⭐ | externalname.go |
| NameAsIdentifier | User name is ID | ⭐ | externalname.go |
| ParameterAsIdentifier | Parameter is ID | ⭐ | externalname.go |
| TemplatedString | Composite ID | ⭐⭐ | externalname.go |
| References | Cross-resource links | ⭐ | service/config.go |
| LateInitializer | Avoid conflicts | ⭐⭐ | service/config.go |
| CustomDiff | Suppress spurious | ⭐⭐⭐ | service/config.go |
| ConnectionDetails | Extract conn info | ⭐⭐ | service/config.go |
| MoveToStatus | Separate sub-resources | ⭐ | service/config.go |
| PasswordGenerator | Auto-gen secrets | ⭐⭐ | service/config.go |

---

## Related Files and Concepts

- **Upjet Config Package**: `github.com/crossplane/upjet/v2/pkg/config`
- **Common Helpers**: `config/cluster/common/common.go`
- **Groups/Kinds**: `config/groups.go`
- **Global Options**: `config/overrides.go`
- **Generated Structs**: Output in `apis/[group]/[version]/`

---

## References

- [Crossplane AWS Provider GitHub](https://github.com/upbound/provider-aws)
- [Upjet Documentation](https://github.com/crossplane/upjet)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

