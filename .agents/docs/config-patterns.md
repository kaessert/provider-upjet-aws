# Config Patterns — provider-upjet-aws

Hand-written Go configuration that drives Upjet code generation for every AWS CRD.

---

## Table of Contents

1. [Config Directory Structure](#config-directory-structure)
2. [How It All Fits Together](#how-it-all-fits-together)
3. [External Name Patterns](#external-name-patterns)
4. [Resource Configuration Pattern](#resource-configuration-pattern)
5. [Global Overrides](#global-overrides)
6. [Common Utilities](#common-utilities)
7. [Group & Kind Mapping](#group--kind-mapping)
8. [Service Examples](#service-examples)
9. [How to Add a New Resource](#how-to-add-a-new-resource)
10. [Common Gotchas](#common-gotchas)

---

## Config Directory Structure

```
config/
├── externalname.go            # ~3,700 lines — External name mapping (CRITICAL)
├── externalnamenottested.go   # External names not yet tested
├── groups.go                  # ~320 lines — Group/Kind mapping overrides
├── overrides.go               # ~250 lines — Global resource overrides
├── registry_common.go         # Provider schema, skip list, resource lists
├── registry_cluster.go        # GetProvider() — assembles cluster-scoped provider
├── registry_namespaced.go     # GetProviderNamespaced() — assembles namespace-scoped provider
├── field-rename.yaml          # Field renaming conversions (embedded)
├── schema.json                # Terraform provider schema (embedded)
├── provider-metadata.yaml     # Provider metadata (embedded)
│
├── cluster/                   # Cluster-scoped resource configurations
│   ├── provider.go            # init() — registers all cluster service configs
│   ├── registry.go            # Configurator type and ProviderConfiguration global
│   ├── common/
│   │   ├── common.go          # Shared utilities (ARN extractor, password gen, etc.)
│   │   ├── common_test.go
│   │   └── apis/
│   │       ├── extractor.go           # IntegrationIDPrefixed() for API Gateway
│   │       └── lambda/extractor.go    # FunctionInvokeARN() for Lambda
│   ├── ec2/config.go
│   ├── rds/config.go
│   ├── s3/config.go
│   ├── iam/config.go
│   ├── eks/config.go
│   ├── lambda/config.go
│   └── {service}/config.go    # ~100+ services
│
└── namespaced/                # Mirror of cluster/ for namespace-scoped provider
    ├── provider.go
    ├── registry.go
    ├── common/...             # Same utilities, different import path
    └── {service}/config.go    # Same configs, different import path
```

**Key insight**: `cluster/` and `namespaced/` are parallel trees. Both contain the same service configurations but produce different providers — one for cluster-scoped CRDs and one for namespace-scoped CRDs.

---

## How It All Fits Together

### Provider Assembly Flow

```
GetProvider() in registry_cluster.go
  │
  ├─ 1. Load provider schema (schema.json)
  ├─ 2. Apply defaultResourceOptions to EVERY resource:
  │     GroupKindOverrides()       ← config/groups.go
  │     KindOverrides()            ← config/groups.go
  │     RegionRequired()           ← config/overrides.go
  │     TagsAllRemoval()           ← config/overrides.go
  │     IdentifierAssignedByAWS()  ← config/overrides.go (default external name)
  │     KnownReferencers()         ← config/overrides.go
  │     ResourceConfigurator()     ← per-resource external name from externalname.go
  │     NamePrefixRemoval()        ← config/overrides.go
  │     DocumentationForTags()     ← config/overrides.go
  │     AddExternalTagsField()     ← config/overrides.go
  │
  ├─ 3. Apply service-specific configs:
  │     for _, configure := range cluster.ProviderConfiguration {
  │         configure(pc)  // each service's Configure() function
  │     }
  │
  └─ 4. pc.ConfigureResources()  → Upjet generates CRDs
```

### Service Registration Pattern

Each service registers its `Configure` function via `init()`:

```go
// config/cluster/provider.go
func init() {
    ProviderConfiguration.AddConfig(acm.Configure)
    ProviderConfiguration.AddConfig(ec2.Configure)
    ProviderConfiguration.AddConfig(eks.Configure)
    ProviderConfiguration.AddConfig(iam.Configure)
    ProviderConfiguration.AddConfig(rds.Configure)
    ProviderConfiguration.AddConfig(s3.Configure)
    // ... 100+ services
}
```

The registry itself is minimal:

```go
// config/cluster/registry.go
type Configure func(provider *config.Provider)

type Configurator []Configure

func (c *Configurator) AddConfig(conf Configure) {
    *c = append(*c, conf)
}

var ProviderConfiguration = Configurator{}
```

---

## External Name Patterns

`config/externalname.go` is the single most important configuration file. It maps every Terraform resource to a strategy for deriving the Crossplane external name (the identifier stored in `metadata.annotations[crossplane.io/external-name]`).

Resources are split across three maps based on their Terraform reconciliation architecture:

- `TerraformPluginSDKExternalNameConfigs` — SDK-based (no-fork), most resources
- `TerraformPluginFrameworkExternalNameConfigs` — Framework-based (newer)
- `CLIReconciledExternalNameConfigs` — CLI-based (legacy)

### Pattern 1: IdentifierFromProvider

**AWS assigns the ID.** The provider reads back whatever ID AWS returns. This is the most common pattern — used for any resource with an AWS-generated ID (ARNs, UUIDs, etc.).

```go
// AWS generates an opaque ID
"aws_sqs_queue":          config.IdentifierFromProvider,
"aws_rds_cluster":        config.IdentifierFromProvider,
"aws_acm_certificate":    config.IdentifierFromProvider,
"aws_amplify_app":        config.IdentifierFromProvider,
"aws_api_gateway_rest_api": config.IdentifierFromProvider,
```

**When to use**: The resource has an AWS-generated ID (instance IDs, ARNs, UUIDs) and you don't control the identifier.

### Pattern 2: NameAsIdentifier

**The `name` field IS the ID.** The user-provided name becomes both the Terraform ID and the Crossplane external name.

```go
// User-provided name is the Terraform ID
"aws_iam_role":                      config.NameAsIdentifier,
"aws_autoscaling_group":             config.NameAsIdentifier,
"aws_appmesh_mesh":                  config.NameAsIdentifier,
"aws_elasticache_serverless_cache":  config.NameAsIdentifier,
"aws_opensearchserverless_access_policy": config.NameAsIdentifier,
```

**When to use**: Terraform uses the `name` field as the resource's primary identifier (e.g., `terraform import aws_iam_role my-role`).

### Pattern 3: ParameterAsIdentifier

**A specific parameter field (not `name`) is the ID.** The named field becomes the Terraform identifier.

```go
// A specific field serves as the Terraform ID
"aws_lambda_function":       config.ParameterAsIdentifier("function_name"),
"aws_s3_directory_bucket":   config.ParameterAsIdentifier("bucket"),
"aws_osis_pipeline":         config.ParameterAsIdentifier("pipeline_name"),
"aws_accessanalyzer_analyzer": config.ParameterAsIdentifier("analyzer_name"),
"aws_dms_certificate":       config.ParameterAsIdentifier("certificate_id"),
"aws_location_tracker":      config.ParameterAsIdentifier("tracker_name"),
```

**When to use**: The Terraform resource uses a specific field (not `name`) as the import key. Check `terraform import` docs for the resource.

### Pattern 4: TemplatedStringAsIdentifier

**Composite IDs built from a template.** Used when Terraform's import ID combines multiple values.

```go
// Simple: parameter as the full ID
"aws_dynamodb_resource_policy": config.TemplatedStringAsIdentifier("",
    "{{ .parameters.resource_arn }}"),

// Two-part: analyzer_name/rule_name
"aws_accessanalyzer_archive_rule": config.TemplatedStringAsIdentifier("rule_name",
    "{{ .parameters.analyzer_name }}/{{ .external_name }}"),

// Comma-separated compound key
"aws_msk_single_scram_secret_association": config.TemplatedStringAsIdentifier("",
    "{{ .parameters.cluster_arn }},{{ .parameters.secret_arn }}"),

// Three-part with account metadata
"aws_opensearchserverless_security_config": config.TemplatedStringAsIdentifier("name",
    "{{ .parameters.type }}/{{ .setup.client_metadata.account_id }}/{{ .external_name }}"),

// ARN template using helper function
"aws_batch_job_queue": config.TemplatedStringAsIdentifier("name",
    fullARNTemplate("batch", "job-queue/{{ .external_name }}")),
// Expands to: "arn:{{ .setup.client_metadata.partition }}:batch:{{ .setup.configuration.region }}:{{ .setup.client_metadata.account_id }}:job-queue/{{ .external_name }}"

// Colon-separated compound key
"aws_glue_catalog_table_optimizer": config.TemplatedStringAsIdentifier("name",
    "{{ .parameters.catalog_id }}:{{ .parameters.database_name }}:{{ .external_name }}"),

// Three-part comma-separated
"aws_networkmanager_link_association": config.TemplatedStringAsIdentifier("",
    "{{ .parameters.global_network_id }},{{ .parameters.link_id }},{{ .parameters.device_id }}"),
```

**Template variables available:**
- `{{ .external_name }}` — The Crossplane external name (metadata annotation)
- `{{ .parameters.<field> }}` — Any field from `spec.forProvider`
- `{{ .setup.configuration.region }}` — Provider region
- `{{ .setup.client_metadata.account_id }}` — AWS account ID
- `{{ .setup.client_metadata.partition }}` — AWS partition (`aws`, `aws-cn`, `aws-us-gov`)

**ARN helper functions:**

```go
// Full regional ARN
func fullARNTemplate(service, resource string) string {
    // → "arn:{{partition}}:SERVICE:{{region}}:{{account_id}}:RESOURCE"
}

// Regionless ARN (e.g., IAM)
func regionlessARNTemplate(service, resource string) string {
    // → "arn:{{partition}}:SERVICE::{{account_id}}:RESOURCE"
}
```

**When to use**: The `terraform import` command requires a composite ID with slashes, colons, commas, or ARN format.

### Pattern 5: Custom Functions

**Complex logic that can't be expressed as a template.** Custom functions override individual callbacks on the `ExternalName` struct.

#### Default stub for empty external names

Used when the provider needs a non-empty ID during creation before AWS assigns the real one:

```go
func identifierFromProviderWithDefaultStub(defaultstub string) config.ExternalName {
    e := config.IdentifierFromProvider
    e.GetIDFn = func(_ context.Context, externalName string,
        _ map[string]any, _ map[string]any) (string, error) {
        if len(externalName) == 0 {
            return defaultstub, nil
        }
        return externalName, nil
    }
    return e
}

// Usage
"aws_bedrock_inference_profile": identifierFromProviderWithDefaultStub("bedrock12345"),
"aws_bedrockagent_agent":       identifierFromProviderWithDefaultStub("STUB123456"),
```

#### IAM Policy — ARN construction + name extraction

The IAM policy external name is the policy name, but the Terraform ID is the full ARN:

```go
func iamPolicy() config.ExternalName {
    e := config.NameAsIdentifier
    e.GetIDFn = func(_ context.Context, externalName string,
        parameters map[string]any, setup map[string]any) (string, error) {
        path, ok := parameters["path"]
        if !ok {
            path = "/"
        }
        accountID := setup["client_metadata"].(map[string]string)["account_id"]
        partition := setup["client_metadata"].(map[string]string)["partition"]
        return fmt.Sprintf("arn:%s:iam::%s:policy%s%s",
            partition, accountID, path, externalName), nil
    }
    e.GetExternalNameFn = func(tfstate map[string]any) (string, error) {
        id, ok := tfstate["id"]
        if !ok {
            return "", errors.New("id attribute missing from state file")
        }
        arnSlice := strings.Split(id.(string), "/")
        return arnSlice[len(arnSlice)-1], nil // Extract last part of ARN
    }
    return e
}
```

#### KMS Alias — prefix handling

The KMS alias has an `alias/` prefix in Terraform but users shouldn't need to type it:

```go
func kmsAlias() config.ExternalName {
    e := config.NameAsIdentifier
    e.SetIdentifierArgumentFn = func(base map[string]interface{}, externalName string) {
        if _, ok := base["name"]; !ok {
            if !strings.HasPrefix(externalName, "alias/") {
                base["name"] = fmt.Sprintf("alias/%s", externalName)
            } else {
                base["name"] = externalName
            }
        }
    }
    // GetExternalNameFn strips "alias/" prefix
    // GetIDFn adds "alias/" prefix back
    return e
}
```

#### Lambda Function URL — conditional qualifier

```go
func lambdaFunctionURL() config.ExternalName {
    e := config.IdentifierFromProvider
    e.GetIDFn = func(ctx context.Context, externalName string,
        parameters map[string]interface{}, _ map[string]interface{}) (string, error) {
        functionName, ok := parameters["function_name"]
        if !ok {
            return "", errors.New("function_name cannot be empty")
        }
        qualifier := parameters["qualifier"]
        if qualifier == nil || qualifier == "" {
            return functionName.(string), nil
        }
        return fmt.Sprintf("%s/%s", functionName.(string), qualifier.(string)), nil
    }
    return e
}
```

#### EKS OIDC Identity Provider — nested field + composite key

The most complex pattern — sets identifier inside a nested `oidc` block:

```go
func eksOIDCIdentityProvider() config.ExternalName {
    return config.ExternalName{
        SetIdentifierArgumentFn: func(base map[string]interface{}, externalName string) {
            if arr, ok := base["oidc"].([]interface{}); ok && len(arr) == 1 {
                if m, ok := arr[0].(map[string]interface{}); ok {
                    m["identity_provider_config_name"] = externalName
                }
            }
        },
        GetExternalNameFn: func(tfstate map[string]interface{}) (string, error) {
            if id, ok := tfstate["id"]; ok {
                return strings.Split(id.(string), ":")[1], nil
            }
            return "", errors.New("there is no id in tfstate")
        },
        GetIDFn: func(_ context.Context, externalName string,
            parameters map[string]interface{}, _ map[string]interface{}) (string, error) {
            cl, ok := parameters["cluster_name"]
            if !ok {
                return "", errors.New("cluster_name cannot be empty")
            }
            return fmt.Sprintf("%s:%s", cl.(string), externalName), nil
        },
        OmittedFields: []string{
            "oidc.identity_provider_config_name",
            "oidc.identity_provider_config_name_prefix",
        },
    }
}
```

#### API Gateway — formatted composite identifiers

A variadic helper for API Gateway resources with multi-part IDs:

```go
// apiGatewayFormattedIdentifier("aggr", "rest_api_id", "resource_id", "http_method")
// Reads from tfstate using "/" separator, parameters using "-" separator
func apiGatewayFormattedIdentifier(prefix string, keys ...string) config.ExternalName {
    // ...
}

// Used for:
"aws_api_gateway_gateway_response": apiGatewayFormattedIdentifier("aggr", ...),
"aws_api_gateway_integration":     apiGatewayFormattedIdentifier("agi", ...),
"aws_api_gateway_method":          apiGatewayFormattedIdentifier("agm", ...),
```

### ExternalName Callback Reference

| Callback | Purpose |
|----------|---------|
| `GetIDFn` | Convert external name → Terraform ID for import/state lookup |
| `GetExternalNameFn` | Extract external name from Terraform state after creation |
| `SetIdentifierArgumentFn` | Set the identifier field(s) in Terraform arguments before apply |
| `OmittedFields` | Fields to omit from CRD spec (because they're derived from external name) |

---

## Resource Configuration Pattern

Each service's `config.go` provides per-resource configuration via `AddResourceConfigurator`:

```go
// config/cluster/{service}/config.go
func Configure(p *config.Provider) {
    p.AddResourceConfigurator("aws_instance", func(r *config.Resource) {
        // 1. Cross-resource references
        r.References["subnet_id"] = config.Reference{
            TerraformName: "aws_subnet",
        }

        // 2. Late initialization suppression
        r.LateInitializer = config.LateInitializer{
            IgnoredFields: []string{"subnet_id", "network_interface"},
        }

        // 3. Async operations
        r.UseAsync = true

        // 4. Connection details extraction
        r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
            conn := map[string][]byte{}
            if a, ok := attr["endpoint"].(string); ok {
                conn["endpoint"] = []byte(a)
            }
            return conn, nil
        }

        // 5. Custom diff handling
        r.TerraformCustomDiff = func(diff *terraform.InstanceDiff,
            _ *terraform.InstanceState,
            _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
            if diff != nil && diff.Attributes != nil {
                delete(diff.Attributes, "some_noisy_field.#")
            }
            return diff, nil
        }

        // 6. Move fields to status
        config.MoveToStatus(r.TerraformResource, "field_name")
    })
}
```

### Configuration Capabilities

#### Cross-Resource References

Wire CRD fields to reference other Crossplane resources:

```go
// Simple reference — field value is the Terraform ID of the target resource
r.References["vpc_id"] = config.Reference{
    TerraformName: "aws_vpc",
}

// Reference with ARN extraction — target's ARN is the value
r.References["role_arn"] = config.Reference{
    TerraformName: "aws_iam_role",
    Extractor:     common.PathARNExtractor,
}

// Reference with custom Go field names (for plural/list fields)
r.References["vpc_config.subnet_ids"] = config.Reference{
    TerraformName:     "aws_subnet",
    RefFieldName:      "SubnetIDRefs",
    SelectorFieldName: "SubnetIDSelector",
}

// Reference using a specific field from the target resource
r.References["principal_arn"] = config.Reference{
    TerraformName: "aws_eks_access_entry",
    Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("principal_arn",true)`,
}

// Replace the entire References map (EKS Cluster example)
r.References = config.References{
    "role_arn": {
        TerraformName: "aws_iam_role",
        Extractor:     common.PathARNExtractor,
    },
    "vpc_config.subnet_ids": {
        TerraformName:     "aws_subnet",
        RefFieldName:      "SubnetIDRefs",
        SelectorFieldName: "SubnetIDSelector",
    },
}

// Delete an auto-generated reference that doesn't make sense
delete(r.References, "event_source_arn")
```

#### Late Initialization

Suppress auto-population of fields that conflict with each other:

```go
// EC2 Instance: network_interface and subnet_id are mutually exclusive
r.LateInitializer = config.LateInitializer{
    IgnoredFields: []string{
        "subnet_id",
        "network_interface",
        "private_ip",
        "source_dest_check",
        "vpc_security_group_ids",
        "associate_public_ip_address",
        "ipv6_addresses",
        "ipv6_address_count",
        "cpu_core_count",
        "cpu_threads_per_core",
        "cpu_options",
        "root_block_device",
    },
}

// Append to existing (IAM Role)
r.LateInitializer.IgnoredFields = append(r.LateInitializer.IgnoredFields,
    "managed_policy_arns", "inline_policy")

// Conditional ignored fields (EKS Node Group — only ignore when set)
r.LateInitializer = config.LateInitializer{
    IgnoredFields: []string{"release_version", "version"},
    ConditionalIgnoredFields: []string{"scaling_config"},
}
```

#### Connection Details

Extract sensitive values into a Kubernetes Secret:

```go
// RDS Instance — extract endpoint, username, port into connection secret
r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
    conn := map[string][]byte{}
    if a, ok := attr["endpoint"].(string); ok {
        conn["endpoint"] = []byte(a)
    }
    if a, ok := attr["reader_endpoint"].(string); ok {
        conn["reader_endpoint"] = []byte(a)
    }
    if a, ok := attr["master_username"].(string); ok {
        conn["master_username"] = []byte(a)
    }
    if a, ok := attr["port"]; ok {
        conn["port"] = []byte(fmt.Sprintf("%v", a))
    }
    return conn, nil
}

// IAM Access Key — username and secret
r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
    conn := map[string][]byte{}
    if a, ok := attr["id"].(string); ok {
        conn["username"] = []byte(a)
    }
    if a, ok := attr["secret"].(string); ok {
        conn["password"] = []byte(a)
    }
    return conn, nil
}
```

#### Custom Diff Handling

Suppress spurious diffs that cause unnecessary reconciliation:

```go
// EC2 Instance — delete noisy computed block diffs
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff,
    _ *terraform.InstanceState,
    _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
    if diff != nil && diff.Attributes != nil {
        delete(diff.Attributes, "enclave_options.#")
        delete(diff.Attributes, "metadata_options.#")
        delete(diff.Attributes, "maintenance_options.#")
        delete(diff.Attributes, "cpu_options.#")
        delete(diff.Attributes, "network_interface.#")
        delete(diff.Attributes, "capacity_reservation_specification.#")
        delete(diff.Attributes, "ephemeral_block_device.#")
        delete(diff.Attributes, "secondary_private_ips.#")
        delete(diff.Attributes, "private_dns_name_options.#")
    }
    return diff, nil
}

// RDS Instance — suppress engine version diff when new < old (downgrades not allowed)
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff,
    _ *terraform.InstanceState,
    _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
    if diff == nil || diff.Destroy {
        return diff, nil
    }
    if evDiff, ok := diff.Attributes["engine_version"]; ok &&
        evDiff.Old != "" && evDiff.New != "" {
        c := utils.CompareEngineVersions(evDiff.New, evDiff.Old)
        if c <= 0 {
            delete(diff.Attributes, "engine_version")
        }
    }
    return diff, nil
}

// S3 BucketAnalyticsConfiguration — delete expected_bucket_owner diff
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff,
    state *terraform.InstanceState,
    config *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
    if diff == nil || diff.Empty() || diff.Destroy || diff.Attributes == nil {
        return diff, nil
    }
    delete(diff.Attributes, "expected_bucket_owner")
    return diff, nil
}
```

#### MoveToStatus

Relocate fields from `spec` to `status` when they're managed by separate sub-resources:

```go
// S3 Bucket — fields managed by separate CRDs (BucketVersioning, BucketACL, etc.)
config.MoveToStatus(r.TerraformResource,
    "acceleration_status", "acl", "grant", "cors_rule", "lifecycle_rule",
    "logging", "object_lock_configuration", "policy",
    "replication_configuration", "request_payer",
    "server_side_encryption_configuration", "versioning", "website", "arn")

// EC2 Instance — security_groups is superseded by vpc_security_group_ids
config.MoveToStatus(r.TerraformResource, "security_groups")
```

#### Schema Modifications

Add custom fields or remove problematic ones:

```go
// RDS Instance — add auto_generate_password toggle
r.TerraformResource.Schema["auto_generate_password"] = &schema.Schema{
    Type:        schema.TypeBool,
    Optional:    true,
    Description: "If true, the password will be auto-generated and stored in the Secret.",
}
r.InitializerFns = append(r.InitializerFns,
    common.PasswordGenerator(
        "spec.forProvider.masterPasswordSecretRef",
        "spec.forProvider.autoGeneratePassword",
    ))

// Lambda Function — delete filename (local-only Terraform concept)
delete(r.TerraformResource.Schema, "filename")
```

#### Terraform Configuration Injector

Inject values from the provider config into Terraform parameters:

```go
// S3 Bucket — inject region and force_destroy default
r.TerraformConfigurationInjector = func(jsonMap map[string]any, params map[string]any) error {
    params["region"] = jsonMap["region"]
    if _, ok := jsonMap["forceDestroy"]; !ok {
        params["force_destroy"] = false
    }
    return nil
}
```

#### UseAsync

Enable async reconciliation for long-running operations:

```go
r.UseAsync = true  // Used for EC2 Instance, RDS Cluster, EKS Cluster, etc.
```

---

## Global Overrides

`config/overrides.go` defines `ResourceOption` functions applied to **every** resource during provider assembly.

### RegionRequired()

Makes the `region` field required for resources that have it in their Terraform schema. Also injects `"us-west-1"` as the default region in generated examples.

```go
func RegionRequired() config.ResourceOption {
    return func(r *config.Resource) {
        if s, ok := r.TerraformResource.Schema["region"]; ok {
            s.Required = true
            s.Optional = false
            s.Computed = false
        }
    }
}
```

### TagsAllRemoval()

Removes the Terraform-specific `tags_all` field (used for provider-wide default tags, not supported in Crossplane):

```go
func TagsAllRemoval() config.ResourceOption {
    return func(r *config.Resource) {
        if t, ok := r.TerraformResource.Schema["tags_all"]; ok {
            t.Computed = true
            t.Optional = false
        }
    }
}
```

### NamePrefixRemoval()

Removes `name_prefix` from all resources (Terraform-only feature):

```go
func NamePrefixRemoval() config.ResourceOption {
    return func(r *config.Resource) {
        r.ExternalName.OmittedFields = append(r.ExternalName.OmittedFields, "name_prefix")
    }
}
```

### IdentifierAssignedByAWS()

Sets `IdentifierFromProvider` as the **default** external name for all resources. Service-specific configs override this:

```go
func IdentifierAssignedByAWS() config.ResourceOption {
    return func(r *config.Resource) {
        r.ExternalName = config.IdentifierFromProvider
    }
}
```

### KnownReferencers()

Auto-adds cross-resource references for well-known field name patterns. Applied to every resource, so individual service configs don't need to repeat these:

```go
func KnownReferencers() config.ResourceOption {
    return func(r *config.Resource) {
        for k, s := range r.TerraformResource.Schema {
            if (s.Computed && !s.Optional) || s.Sensitive {
                continue // skip status-only and sensitive fields
            }
            switch {
            case strings.HasSuffix(k, "role_arn"):
                r.References[k] = config.Reference{
                    TerraformName: "aws_iam_role",
                    Extractor:     common.PathARNExtractor,
                }
            case strings.HasSuffix(k, "security_group_ids"):
                r.References[k] = config.Reference{
                    TerraformName:     "aws_security_group",
                    RefFieldName:      "..IDRefs",
                    SelectorFieldName: "..IDSelector",
                }
            case r.ShortGroup == "glue" && k == "database_name":
                r.References["database_name"] = config.Reference{
                    TerraformName: "aws_glue_catalog_database",
                }
            }
            switch k {
            case "vpc_id":        // → aws_vpc
            case "subnet_ids":    // → aws_subnet (with Refs/Selector)
            case "subnet_id":     // → aws_subnet
            case "iam_roles":     // → aws_iam_role (with Refs/Selector)
            case "security_group_id": // → aws_security_group
            case "kms_key_id":    // → aws_kms_key
            case "kms_key_arn":   // → aws_kms_key
            case "kms_key":       // → aws_kms_key
            }
        }
    }
}
```

**Summary of auto-references:**

| Field Pattern | Target Resource | Notes |
|---------------|----------------|-------|
| `*_role_arn` | `aws_iam_role` | ARN extraction |
| `*_security_group_ids` | `aws_security_group` | Plural with Refs/Selector |
| `vpc_id` | `aws_vpc` | |
| `subnet_ids` | `aws_subnet` | Plural with Refs/Selector |
| `subnet_id` | `aws_subnet` | |
| `iam_roles` | `aws_iam_role` | Plural with Refs/Selector |
| `security_group_id` | `aws_security_group` | |
| `kms_key_id` | `aws_kms_key` | |
| `kms_key_arn` | `aws_kms_key` | |
| `kms_key` | `aws_kms_key` | |
| `database_name` (Glue only) | `aws_glue_catalog_database` | |

### AddExternalTagsField()

Adds tag initialization for resources that have a `tags` field:

```go
func AddExternalTagsField() config.ResourceOption {
    return func(r *config.Resource) {
        if s, ok := r.TerraformResource.Schema["tags"]; ok && s.Type == schema.TypeMap {
            r.InitializerFns = append(r.InitializerFns, config.TagInitializer)
        }
    }
}
```

### Application Order

These are applied as `defaultResourceOptions` in `GetProvider()`:

```go
defaultResourceOptions := []config.ResourceOption{
    GroupKindOverrides(),
    KindOverrides(),
    RegionRequired(),
    TagsAllRemoval(),
    IdentifierAssignedByAWS(),  // Sets default external name
    KnownReferencers(),         // Auto-wires common references
    ResourceConfigurator(),     // Applies per-resource external name from externalname.go
    NamePrefixRemoval(),
    DocumentationForTags(),
    injectFieldRenamingConversionFunctions(),
    injectPluginFrameworkCustomStateEmptyCheck(),
    AddExternalTagsField(),
}
```

**Important**: `IdentifierAssignedByAWS()` runs before `ResourceConfigurator()`, so every resource starts with `IdentifierFromProvider` as default, then gets overridden by whatever is in `externalname.go`.

---

## Common Utilities

`config/cluster/common/common.go` provides shared functions used across service configs.

### ARNExtractor()

Extracts the ARN from `status.atProvider.arn` — used as a reference extractor:

```go
func ARNExtractor() reference.ExtractValueFn {
    return func(mg xpresource.Managed) string {
        paved, err := fieldpath.PaveObject(mg)
        if err != nil {
            return ""
        }
        r, err := paved.GetString("status.atProvider.arn")
        if err != nil {
            return ""
        }
        return r
    }
}

// Used via the constant:
const PathARNExtractor = SelfPackagePath + ".ARNExtractor()"

// In service configs:
r.References["role_arn"] = config.Reference{
    TerraformName: "aws_iam_role",
    Extractor:     common.PathARNExtractor,
}
```

### TerraformID()

Returns the raw Terraform resource ID — used when the reference value is the Terraform ID itself:

```go
func TerraformID() reference.ExtractValueFn {
    return func(mr xpresource.Managed) string {
        tr, ok := mr.(resource.Terraformed)
        if !ok {
            return ""
        }
        return tr.GetID()
    }
}

const PathTerraformIDExtractor = SelfPackagePath + ".TerraformID()"
```

### PasswordGenerator()

Auto-generates passwords and stores them in Kubernetes Secrets:

```go
func PasswordGenerator(secretRefFieldPath, toggleFieldPath string) config.NewInitializerFn

// Usage in RDS:
r.InitializerFns = append(r.InitializerFns,
    common.PasswordGenerator(
        "spec.forProvider.masterPasswordSecretRef",
        "spec.forProvider.autoGeneratePassword",
    ))
```

**Flow**: Checks if secret exists → if empty and toggle is true → generates password → creates/updates Secret with owner reference.

### RemovePolicyVersion()

Removes the `"Version"` key from a JSON IAM policy document (for normalization):

```go
func RemovePolicyVersion(p string) (string, error)
```

### RemoveDiffIfEmpty()

Suppresses diffs where both old and new values are empty strings:

```go
func RemoveDiffIfEmpty(keys []string) config.CustomDiff

// Usage:
r.TerraformCustomDiff = common.RemoveDiffIfEmpty([]string{"description", "policy"})
```

### Custom Extractors (apis/ subdirectory)

```go
// config/cluster/common/apis/extractor.go
func IntegrationIDPrefixed() reference.ExtractValueFn
// Returns "integrations/" + external name (for API Gateway v2 Routes)

// config/cluster/common/apis/lambda/extractor.go
func FunctionInvokeARN() reference.ExtractValueFn
// Returns the Lambda function's invoke ARN from Status.AtProvider.InvokeArn
```

---

## Group & Kind Mapping

`config/groups.go` controls how Terraform resource names map to Crossplane API groups and kinds.

### GroupMap — Group/Kind Calculator

Maps Terraform resource names to `(group, kind)` pairs. The `ReplaceGroupWords` helper strips N prefix words from the Terraform name:

```go
func ReplaceGroupWords(group string, count int) GroupKindCalculator {
    return func(resource string) (string, string) {
        words := strings.Split(strings.TrimPrefix(resource, "aws_"), "_")
        snakeKind := strings.Join(words[count:], "_")
        return group, name.NewFromSnake(snakeKind).Camel
    }
}

var GroupMap = map[string]GroupKindCalculator{
    // Strip 2 words: "aws_api_gateway_rest_api" → group: apigateway, kind: RestAPI
    "aws_api_gateway_rest_api": ReplaceGroupWords("apigateway", 2),

    // Strip 2 words: "aws_route53_resolver_rule" → group: route53resolver, kind: Rule
    "aws_route53_resolver_rule": ReplaceGroupWords("route53resolver", 2),

    // Map ALB resources to elbv2 group
    "aws_alb_listener": ReplaceGroupWords("elbv2", 0),
    // ...300+ entries
}
```

### KindMap — Direct Kind Overrides

For cases where the calculated kind doesn't match what you want:

```go
var KindMap = map[string]string{
    "aws_autoscaling_group":                    "AutoscalingGroup",
    "aws_cloudformation_type":                  "CloudFormationType",
    "aws_cloudtrail":                           "Trail",
    "aws_config_configuration_recorder_status": "AWSConfigurationRecorderStatus",
}
```

### Applied as ResourceOptions

```go
func GroupKindOverrides() config.ResourceOption {
    return func(r *config.Resource) {
        if f, ok := GroupMap[r.Name]; ok {
            r.ShortGroup, r.Kind = f(r.Name)
        }
    }
}

func KindOverrides() config.ResourceOption {
    return func(r *config.Resource) {
        if k, ok := KindMap[r.Name]; ok {
            r.Kind = k
        }
    }
}
```

---

## Service Examples

### EC2 — Full-featured configuration

```go
func Configure(p *config.Provider) {
    p.AddResourceConfigurator("aws_instance", func(r *config.Resource) {
        r.References["vpc_security_group_ids"] = config.Reference{
            TerraformName:     "aws_security_group",
            RefFieldName:      "VPCSecurityGroupIDRefs",
            SelectorFieldName: "VPCSecurityGroupIDSelector",
        }
        r.References["block_device_mappings.ebs.kms_key_id"] = config.Reference{
            TerraformName: "aws_kms_key",
            Extractor:     common.PathARNExtractor,
        }
        r.LateInitializer = config.LateInitializer{
            IgnoredFields: []string{
                "subnet_id", "network_interface", "private_ip",
                "source_dest_check", "vpc_security_group_ids",
                "associate_public_ip_address", "cpu_options", "root_block_device",
            },
        }
        r.UseAsync = true
        config.MoveToStatus(r.TerraformResource, "security_groups")
        r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
            if diff != nil && diff.Attributes != nil {
                delete(diff.Attributes, "enclave_options.#")
                delete(diff.Attributes, "metadata_options.#")
                // ... more noisy field suppression
            }
            return diff, nil
        }
    })
}
```

### RDS — Connection details + password generation + version comparison

```go
p.AddResourceConfigurator("aws_db_instance", func(r *config.Resource) {
    r.References["kms_key_id"] = config.Reference{
        TerraformName: "aws_kms_key",
        Extractor:     common.PathARNExtractor,
    }
    // Auto-generate password support
    r.TerraformResource.Schema["auto_generate_password"] = &schema.Schema{
        Type:     schema.TypeBool,
        Optional: true,
    }
    r.InitializerFns = append(r.InitializerFns,
        common.PasswordGenerator(
            "spec.forProvider.masterPasswordSecretRef",
            "spec.forProvider.autoGeneratePassword",
        ))
    // Extract connection details
    r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
        conn := map[string][]byte{}
        if a, ok := attr["endpoint"].(string); ok { conn["endpoint"] = []byte(a) }
        if a, ok := attr["master_username"].(string); ok { conn["master_username"] = []byte(a) }
        if a, ok := attr["port"]; ok { conn["port"] = []byte(fmt.Sprintf("%v", a)) }
        return conn, nil
    }
    // Suppress downgrade diffs (AWS doesn't allow RDS version downgrades)
    r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
        if evDiff, ok := diff.Attributes["engine_version"]; ok {
            if utils.CompareEngineVersions(evDiff.New, evDiff.Old) <= 0 {
                delete(diff.Attributes, "engine_version")
            }
        }
        return diff, nil
    }
    r.UseAsync = true
})
```

### S3 — MoveToStatus for sub-resource CRDs

```go
p.AddResourceConfigurator("aws_s3_bucket", func(r *config.Resource) {
    // These fields are managed by separate CRDs:
    // BucketVersioning, BucketACL, BucketLifecycleConfiguration, etc.
    config.MoveToStatus(r.TerraformResource,
        "acceleration_status", "acl", "grant", "cors_rule",
        "lifecycle_rule", "logging", "object_lock_configuration",
        "policy", "replication_configuration", "request_payer",
        "server_side_encryption_configuration", "versioning",
        "website", "arn")
    r.TerraformConfigurationInjector = func(jsonMap map[string]any, params map[string]any) error {
        params["region"] = jsonMap["region"]
        if _, ok := jsonMap["forceDestroy"]; !ok {
            params["force_destroy"] = false
        }
        return nil
    }
})
```

### EKS — Full reference replacement + conditional late init

```go
p.AddResourceConfigurator("aws_eks_cluster", func(r *config.Resource) {
    r.References = config.References{
        "role_arn": {
            TerraformName: "aws_iam_role",
            Extractor:     common.PathARNExtractor,
        },
        "vpc_config.subnet_ids": {
            TerraformName:     "aws_subnet",
            RefFieldName:      "SubnetIDRefs",
            SelectorFieldName: "SubnetIDSelector",
        },
        "vpc_config.security_group_ids": {
            TerraformName:     "aws_security_group",
            RefFieldName:      "SecurityGroupIDRefs",
            SelectorFieldName: "SecurityGroupIDSelector",
        },
    }
    r.UseAsync = true
})

p.AddResourceConfigurator("aws_eks_node_group", func(r *config.Resource) {
    r.LateInitializer = config.LateInitializer{
        IgnoredFields:            []string{"release_version", "version"},
        ConditionalIgnoredFields: []string{"scaling_config"},
    }
})
```

### Lambda — Schema deletion + reference cleanup

```go
p.AddResourceConfigurator("aws_lambda_function", func(r *config.Resource) {
    // "filename" is a local-only Terraform concept, remove it
    delete(r.TerraformResource.Schema, "filename")
    r.References["vpc_config.security_group_ids"] = config.Reference{
        TerraformName:     "aws_security_group",
        RefFieldName:      "SecurityGroupIDRefs",
        SelectorFieldName: "SecurityGroupIDSelector",
    }
    r.LateInitializer = config.LateInitializer{
        IgnoredFields: []string{"source_code_hash"},
    }
})

p.AddResourceConfigurator("aws_lambda_event_source_mapping", func(r *config.Resource) {
    // Remove auto-generated references that don't make sense
    delete(r.References, "event_source_arn")
    delete(r.References, "source_access_configuration.uri")
})
```

### IAM — Append to late init + policy ARN extraction

```go
p.AddResourceConfigurator("aws_iam_role", func(r *config.Resource) {
    r.LateInitializer.IgnoredFields = append(r.LateInitializer.IgnoredFields,
        "managed_policy_arns", "inline_policy")
})

p.AddResourceConfigurator("aws_iam_role_policy_attachment", func(r *config.Resource) {
    r.References["policy_arn"] = config.Reference{
        TerraformName: "aws_iam_policy",
        Extractor:     common.PathARNExtractor,
    }
})
```

---

## How to Add a New Resource

### Step 1: Add External Name Entry

In `config/externalname.go`, add an entry to the appropriate map:

```go
// In TerraformPluginSDKExternalNameConfigs (or the appropriate map)
var TerraformPluginSDKExternalNameConfigs = map[string]config.ExternalName{
    // ... existing entries ...
    "aws_new_service_resource": config.IdentifierFromProvider,
}
```

**Choose the right pattern:**
- Check `terraform import aws_new_service_resource` docs for the ID format
- Single AWS-generated ID → `IdentifierFromProvider`
- User-provided name → `NameAsIdentifier`
- Specific field → `ParameterAsIdentifier("field_name")`
- Composite ID → `TemplatedStringAsIdentifier("name_field", "template")`

### Step 2: Add Group Mapping (if needed)

In `config/groups.go`, add an entry if the default group/kind derivation isn't correct:

```go
var GroupMap = map[string]GroupKindCalculator{
    // ... existing entries ...
    "aws_new_service_resource": ReplaceGroupWords("newservice", 2),
}
```

### Step 3: Create/Update Service Config

Create `config/cluster/{service}/config.go`:

```go
package service

import "github.com/crossplane/upjet/v2/pkg/config"

func Configure(p *config.Provider) {
    p.AddResourceConfigurator("aws_new_service_resource", func(r *config.Resource) {
        // Add references
        r.References["vpc_id"] = config.Reference{
            TerraformName: "aws_vpc",
        }

        // Handle late init conflicts
        r.LateInitializer = config.LateInitializer{
            IgnoredFields: []string{"conflicting_field"},
        }

        // Enable async if long-running
        r.UseAsync = true
    })
}
```

### Step 4: Register in provider.go

In `config/cluster/provider.go`:

```go
import "github.com/upbound/provider-aws/v2/config/cluster/newservice"

func init() {
    // ... existing registrations ...
    ProviderConfiguration.AddConfig(newservice.Configure)
}
```

### Step 5: Mirror for Namespaced Provider

Repeat steps 3-4 in the `config/namespaced/` tree.

### Step 6: Generate and Test

```bash
make generate          # Generate CRDs and controllers
make test              # Run tests
make reviewable        # Full CI check
```

---

## Common Gotchas

### External Name Oscillation

**Symptom**: Resource keeps reconciling every cycle, external name changes.

**Cause**: The `GetExternalNameFn` extracts a different value than what `GetIDFn` produces. For example, the Terraform state ID is `arn:aws:iam::123456:policy/my-policy` but `GetExternalNameFn` returns the full ARN instead of just `my-policy`.

**Fix**: Ensure `GetExternalNameFn` and `GetIDFn` are inverses. If `GetIDFn(externalName) → tfID`, then `GetExternalNameFn(tfState) → externalName`.

### Circular References

**Symptom**: Two resources reference each other, creating an unresolvable dependency.

**Cause**: Resource A has a reference to Resource B, and Resource B has a reference back to Resource A.

**Fix**: Delete one direction of the reference:
```go
delete(r.References, "circular_field")
```

### Late Init Conflicts

**Symptom**: Resource oscillates between two states, alternately setting and unsetting fields.

**Cause**: Two mutually exclusive fields both get late-initialized. For example, EC2 Instance has both `subnet_id` and `network_interface` — AWS returns both but only one should be set.

**Fix**: Add conflicting fields to `LateInitializer.IgnoredFields`:
```go
r.LateInitializer = config.LateInitializer{
    IgnoredFields: []string{"subnet_id", "network_interface"},
}
```

### MoveToStatus Without Sub-Resource CRD

**Symptom**: Field is invisible to users — it's in status but there's no way to set it.

**Cause**: A field was moved to status (expecting a separate CRD to manage it) but the sub-resource CRD doesn't exist.

**Fix**: Either create the sub-resource CRD or don't move the field to status. S3 Bucket is the canonical example — `versioning`, `acl`, `cors_rule`, etc. are all managed by separate CRDs like `BucketVersioning`, `BucketACL`, `BucketCORSConfiguration`.

### Noisy Diffs on Computed Blocks

**Symptom**: Resource shows changes on every reconcile even when nothing changed. Plan diff shows changes to `foo.#` or `foo.0.bar`.

**Cause**: Terraform computed blocks generate diff entries even when values haven't changed.

**Fix**: Add a `TerraformCustomDiff` that deletes the noisy attributes:
```go
r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, ...) (*terraform.InstanceDiff, error) {
    if diff != nil && diff.Attributes != nil {
        delete(diff.Attributes, "noisy_block.#")
    }
    return diff, nil
}
```

### Sensitive Fields Not Appearing in Connection Secret

**Symptom**: Connection secret is empty or missing expected keys.

**Cause**: The `AdditionalConnectionDetailsFn` uses the wrong attribute name, or the attribute isn't populated yet when the function runs.

**Fix**: Check the actual Terraform state attribute names (they use snake_case). The function receives the raw Terraform attributes map:
```go
// Wrong: using Crossplane field names
attr["masterUsername"]  // ✗
// Right: using Terraform attribute names
attr["master_username"] // ✓
```

### Skip List

Some resources are explicitly excluded from code generation in `config/registry_common.go`:

```go
var skipList = []string{
    "aws_waf_rule_group$",              // Too big CRD schema
    "aws_wafregional_rule_group$",      // Too big CRD schema
    "aws_ecs_tag$",                     // Tags managed by ECS resources
    "aws_alb$",                         // Identical with aws_lb
    "aws_iam_policy_attachment$",       // Identical with aws_iam_*_policy_attachment
    "aws_location_map$",               // Failure with unknown reason
    "aws_rds_reserved_instance",       // Expense of testing
    // ...
}
```

If your resource isn't being generated, check if it's in the skip list.
