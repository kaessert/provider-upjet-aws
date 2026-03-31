# TF Business Logic Catalog

Extracted from `config/cluster/*/config.go`. Documents every resource with
`TerraformConfigurationInjector`, `TerraformCustomDiff`, or `UseAsync = true`,
along with the native equivalent approach for each.

---

## TerraformConfigurationInjector Resources

These resources inject extra key-value pairs into the Terraform configuration
map before plan/apply. In native controllers this logic belongs in `Observe()`
or `Create()`/`Update()` as late-initialization or defaulting.

---

### aws_cloudformation_stack_set_instance

**File:** `config/cluster/cloudformation/config.go`

#### TerraformConfigurationInjector
**What it does:** Copies the `region` field from the raw JSON map (Crossplane
annotations / external-name derivation) into the Terraform `params` map so that
TF sees the region parameter explicitly.

**Native equivalent:** In `Create()` and `Update()`, read the region from the
Crossplane managed resource spec (or from the provider region) and pass it
directly to the AWS SDK call. No late-init required because the region is a
required field that must be specified by the user.

---

### aws_security_group_rule

**File:** `config/cluster/ec2/config.go`

#### TerraformConfigurationInjector
**What it does:** If `self` is not set by the user (absent from the JSON map),
defaults `params["self"] = false`. This prevents Terraform drift caused by the
computed field being nil vs. `false`.

**Native equivalent:** Late-initialize `spec.forProvider.self` to `*false` in
`Observe()` if the field is nil. This is a standard Crossplane late-init
pattern using `lateInitialize()`.

---

### aws_ami_copy

**File:** `config/cluster/ec2/config.go`

#### TerraformConfigurationInjector
**What it does:** Always sets `params["ebs_block_device"] = []any{}` (empty
list) to prevent Terraform from trying to diff an unset EBS block device. Also
defaults `params["encrypted"] = false` if the user did not set `encrypted`.

**Native equivalent:**
1. The `ebs_block_device` workaround is not needed in native AWS SDK — EBS
   block devices are read from `DescribeImages` and compared directly; an
   unset field does not cause drift.
2. Late-initialize `spec.forProvider.encrypted` to `*false` in `Observe()` if
   the field is nil and the AWS state shows the value is false.

---

### aws_opensearchserverless_security_policy

**File:** `config/cluster/opensearchserverless/config.go`

#### TerraformConfigurationInjector
**What it does:** Uses `config.CanonicalizeJSONParameters("policy")`. This
unmarshals the `policy` JSON string in the Terraform params map, re-marshals it
via a canonical encoder (key ordering, whitespace normalization), and stores the
result back. This ensures that two semantically equivalent JSON documents
produce the same string, preventing spurious diffs.

**Native equivalent:** In `Observe()`, before comparing desired vs. actual
policy, canonicalize both strings using `json.Canonicalize()` (or an equivalent
that produces deterministic key ordering). Use
`awspolicy.PoliciesAreEquivalent()` or a simple re-marshal through
`encoding/json` to normalize both sides before string comparison.

---

### aws_opensearchserverless_lifecycle_policy

**File:** `config/cluster/opensearchserverless/config.go`

#### TerraformConfigurationInjector
**What it does:** Same as `aws_opensearchserverless_security_policy` — canonicalizes the
`policy` JSON string to prevent spurious diffs.

**Native equivalent:** Same approach — canonicalize both desired and actual
policy strings in `Observe()` before comparing.

---

### aws_opensearchserverless_access_policy

**File:** `config/cluster/opensearchserverless/config.go`

#### TerraformConfigurationInjector
**What it does:** Same as `aws_opensearchserverless_security_policy` — canonicalizes the
`policy` JSON string to prevent spurious diffs.

**Native equivalent:** Same approach — canonicalize both desired and actual
policy strings in `Observe()` before comparing.

---

### aws_s3_bucket

**File:** `config/cluster/s3/config.go`

#### TerraformConfigurationInjector
**What it does:** Two injections:
1. Copies `region` from the JSON map into `params["region"]` so Terraform sees
   the explicit region.
2. If `forceDestroy` is not set by the user, defaults
   `params["force_destroy"] = false`. This prevents Terraform from computing a
   drift on the `force_destroy` field and attempting to reconcile it.

**Native equivalent:**
1. Region: Pass region from the provider config or spec directly to AWS SDK
   calls. Not needed in native controller.
2. `force_destroy`: Late-initialize `spec.forProvider.forceDestroy` to `*false`
   in `Observe()` if the field is nil. This is the canonical Crossplane
   late-init pattern.

---

### aws_s3_object

**File:** `config/cluster/s3/config.go`

#### TerraformConfigurationInjector
**What it does:** If `acl` is not set by the user (absent from JSON map),
defaults `params["acl"] = "private"`. This prevents drift from the AWS default
ACL being applied and not matching an unset field.

**Native equivalent:** Late-initialize `spec.forProvider.acl` to `*"private"`
in `Observe()` if the field is nil and the AWS API reports the ACL as
`"private"`. Alternatively, use a CRD defaulting webhook with
`+kubebuilder:default="private"`.

---

### aws_secretsmanager_secret

**File:** `config/cluster/secretsmanager/config.go`

#### TerraformConfigurationInjector
**What it does:** Forces `params["name_prefix"] = ""`. This prevents the
`name_prefix` field (an alternative to `name` for TF-generated names) from
being set by TF's defaults, which would conflict with the externally-managed
name.

**Native equivalent:** `name_prefix` is a Terraform-specific concept for random
name generation. In native controllers the secret name comes directly from the
external-name annotation. Simply never set `name_prefix` in Create/Update API
calls. No late-init needed.

---

## TerraformCustomDiff Resources

These resources modify the computed Terraform diff to suppress spurious changes.
In native controllers this logic belongs in `Observe()` where we compare desired
vs. actual state and decide whether an update is needed.

---

### aws_backup_selection

**File:** `config/cluster/backup/config.go`

#### TerraformCustomDiff
**What it does:** Removes diff keys matching `condition.*.#` where both old and
new values are `"0"` and not NewComputed. This suppresses spurious diffs on
empty `condition` list counts that AWS API may return as zero even when the user
does not specify any conditions.

**Native equivalent:** In `Observe()`, when comparing `Conditions` lists: if
both desired and actual are nil or empty, treat them as equal regardless of
whether the AWS response returns a zero-length slice vs. nil.

---

### aws_dms_endpoint

**File:** `config/cluster/dms/config.go`

#### TerraformCustomDiff
**What it does:** Removes the `redshift_settings.#` diff attribute. The AWS API
populates a default `redshift_settings` block even when the user does not
specify one, causing a spurious diff on the count field.

**Native equivalent:** In `Observe()`, when comparing `RedshiftSettings`: if
the desired spec has no RedshiftSettings configured (nil), ignore the
RedshiftSettings returned by the AWS API rather than treating it as a drift.

---

### aws_instance

**File:** `config/cluster/ec2/config.go`

#### TerraformCustomDiff
**What it does:** Uses `common.RemoveDiffIfEmpty([]string{"volume_tags.%"})`.
Removes the `volume_tags.%` diff attribute if both old and new values are empty
strings. The `volume_tags.%` is a map count field that can appear as an
empty-string diff due to Terraform's internal map diffing.

**Native equivalent:** In `Observe()`, when comparing volume tags: if both
desired and actual tag maps are empty/nil, suppress the drift rather than
triggering an update.

---

### aws_spot_instance_request

**File:** `config/cluster/ec2/config.go`

#### TerraformCustomDiff
**What it does:** Removes several computed block count attributes from the diff:
`enclave_options.#`, `metadata_options.#`, `maintenance_options.#`,
`cpu_options.#`, `network_interface.#`,
`capacity_reservation_specification.#`, `ephemeral_block_device.#`,
`secondary_private_ips.#`, `private_dns_name_options.#`. These are all
AWS-computed nested blocks that get defaulted by the API when a spot instance is
requested, but users don't specify them explicitly.

**Native equivalent:** In `Observe()`, do not compare these computed-only nested
block counts. Only compare fields that the user explicitly set in spec. Use
`lateInitialize()` for fields that should be set from AWS state.

---

### aws_ami_copy

**File:** `config/cluster/ec2/config.go`

#### TerraformCustomDiff
**What it does:** Removes the `ebs_block_device.#` diff attribute. The AWS API
returns EBS block devices for a copied AMI that were inherited from the source,
causing drift when the user did not specify any `ebs_block_device` blocks.

**Native equivalent:** In `Observe()`, if `spec.forProvider.ebsBlockDevice` is
nil or empty, do not treat the AWS-returned EBS block device list as a drift.

---

### aws_ecs_service

**File:** `config/cluster/ecs/config.go`

#### TerraformCustomDiff
**What it does:** Suppresses spurious diffs when `task_definition` changes. If
the desired task definition (td.New, an ARN like
`arn:aws:ecs:us-west-1:123456:task-definition/svc:36`) shares the same family
and revision suffix as the current value (td.Old), the diff is suppressed. This
handles the case where Terraform sees the ARN-qualified form vs. the plain
family name, or vice versa.

**Native equivalent:** In `Observe()`, when comparing `TaskDefinition`: strip
the ARN prefix and revision suffix (`arn:aws:ecs:...:task-definition/FAMILY:REV`
→ `FAMILY:REV`) from both desired and actual before comparing. If the user
specified a bare family name, also compare against the family portion only.

---

### aws_elasticache_replication_group

**File:** `config/cluster/elasticache/config.go`

#### TerraformCustomDiff
**What it does:** Removes the `security_group_names.#` diff attribute. Legacy
EC2-Classic security group names are returned by the AWS API but are not
meaningful in a VPC context; the diff causes spurious updates.

**Native equivalent:** In `Observe()`, do not compare `SecurityGroupNames` (the
legacy EC2-Classic field). If the spec does not specify it, ignore the value
returned by the AWS API.

---

### aws_lb_listener

**File:** `config/cluster/elbv2/config.go`

#### TerraformCustomDiff
**What it does:** Handles the dual representations of a "forward" action in ALB
listeners. A user can specify either `default_action.0.target_group_arn`
(shortcut) or `default_action.0.forward` (full form). The AWS API populates
both, causing diff when only one is specified. The custom diff:
1. Removes diff entries where `default_action.*.forward.#` goes from 1→0 (AWS
   auto-populated forward when user specified `target_group_arn`).
2. Removes diff entries where `default_action.*.target_group_arn` goes from a
   value to empty/removed (AWS auto-populated target_group_arn when user
   specified `forward`).
3. Cleans up all sub-attributes of the removed block.

**Native equivalent:** In `Observe()`, when comparing listener default actions:
if the desired spec uses `target_group_arn` (no `forward` block), ignore the
`Forward` field returned by AWS. If the desired spec uses the `forward` block
(no `target_group_arn`), ignore the `TargetGroupArn` field returned by AWS.
Only compare the field the user actually set.

---

### aws_lb_target_group

**File:** `config/cluster/elbv2/config.go`

#### TerraformCustomDiff
**What it does:** Suppresses the diff for
`target_health_state.0.unhealthy_draining_interval` when the old value is empty
and the new value is `"0"`. This is a Terraform schema defaulting issue where
an unset integer field defaults to `0` in the diff but the AWS API returns
nothing (no health state config).

**Native equivalent:** In `Observe()`, when comparing `TargetHealthState.UnhealthyDrainingInterval`:
treat a nil value and a zero value as equivalent (no drift).

---

### aws_glue_catalog_table

**File:** `config/cluster/glue/config.go`

#### TerraformCustomDiff
**What it does:** Removes the `partition_index.#` diff attribute when both old
and new are empty strings and the value is NewComputed. This suppresses a
spurious computed diff on the partition index count when the user does not
specify any partition indexes.

**Native equivalent:** In `Observe()`, when comparing `PartitionKeys`: if the
desired spec has no `PartitionKeys` and the AWS API returns nil or empty, treat
as equal.

---

### aws_networkfirewall_firewall

**File:** `config/cluster/networkfirewall/config.go`

#### TerraformCustomDiff
**What it does:** Handles `subnet_mapping` set diffing. The `ip_address_type`
field in each subnet mapping is optional+computed. When the user only specifies
`subnet_id`, the AWS API fills in `ip_address_type`, causing a spurious
diff. The custom diff uses a custom hash function based only on `subnet_id` to
compare the old and new sets. If the sets of subnet IDs are the same, all
`subnet_mapping` diffs are removed. Also always removes `firewall_status.#`.

**Native equivalent:** In `Observe()`, when comparing `SubnetMappings`: build
a set of just subnet IDs from both desired and actual, and compare those sets.
If the subnet IDs match, do not treat differences in `IPAddressType` as drift
(since it's a computed field). Also ignore `FirewallStatus` entirely (status
field).

---

### aws_opensearch_domain

**File:** `config/cluster/opensearch/config.go`

#### TerraformCustomDiff
**What it does:** Removes the `advanced_security_options.#` diff attribute when
both old and new are empty strings and the value is NewComputed. This suppresses
a spurious computed diff on the advanced security options count when the user
doesn't specify them.

**Native equivalent:** In `Observe()`, when comparing `AdvancedSecurityOptions`:
if the desired spec has no `AdvancedSecurityOptions` configured, ignore the
value returned by AWS (it may return a default empty block).

---

### aws_rds_cluster

**File:** `config/cluster/rds/config.go`

#### TerraformCustomDiff
**What it does:** Suppresses `engine_version` diffs when the desired version is
lower than or equal to the actual version (downgrades not allowed by AWS RDS).
Uses `utils.CompareEngineVersions(new, old)` — if new ≤ old, removes the
engine_version diff so no update is triggered.

**Native equivalent:** In `Observe()`, when comparing `EngineVersion`: use a
semantic version comparison. If `spec.forProvider.engineVersion ≤
status.atProvider.engineVersion`, do not mark the resource as needing an
update. This prevents controller loops from attempting impossible downgrades.

---

### aws_db_instance

**File:** `config/cluster/rds/config.go`

#### TerraformCustomDiff
**What it does:** Same as `aws_rds_cluster` — suppresses `engine_version` diff
when the desired version is lower than or equal to the actual version.

**Native equivalent:** Same approach — in `Observe()`, semantic version
comparison on EngineVersion. If desired ≤ actual, do not trigger update.

---

### aws_route53_record

**File:** `config/cluster/route53/config.go`

#### TerraformCustomDiff
**What it does:** Suppresses `name` diff when the only difference is a trailing
dot. Route 53 always returns record names with a trailing dot (FQDN form), but
users typically specify them without. `TrimSuffix(new, ".") == TrimSuffix(old,
".")` → removes the name diff.

**Native equivalent:** In `Observe()`, when comparing record `Name`: trim
trailing dots from both desired (`spec.forProvider.name`) and actual (AWS
response `Name`) before comparing. This is the canonical FQDN normalization.

---

### aws_route53_resolver_endpoint

**File:** `config/cluster/route53resolver/config.go`

#### TerraformCustomDiff
**What it does:** Complex custom diff for the `ip_address` set field. The
Route 53 Resolver endpoint assigns IP addresses to subnets. When the user
specifies only a `subnet_id` (no explicit `ip`), AWS auto-assigns an IP. This
causes spurious diffs because TF's set hash includes the IP. The custom diff:
1. Builds a map of `subnet_id → {ip → hash}` from current state.
2. For desired ip_addresses with explicit IPs: removes the diff if already
   matched in state.
3. For desired ip_addresses with only `subnet_id`: matches them against
   remaining IPs in current state by count, removes creation diffs when counts
   match.
4. Adjusts `ip_address.#` count accordingly.

**Native equivalent:** In `Observe()`, when comparing `IpAddresses`: group by
`SubnetId`. For each subnet, if the desired spec has `n` entries with no
explicit IP, and the current state has `n` entries for that subnet, treat them
as equal regardless of which specific IPs are assigned. Only trigger an update
if the number of subnet entries changes or an explicit IP doesn't match.

---

### aws_s3_bucket_lifecycle_configuration

**File:** `config/cluster/s3/config.go`

#### TerraformCustomDiff
**What it does:** Removes the `expected_bucket_owner` diff attribute. The AWS
API may populate this field with the account ID even when the user does not
specify it, causing spurious diffs.

**Native equivalent:** In `Observe()`, when comparing `ExpectedBucketOwner`: if
the spec field is nil or empty, ignore the value returned by the AWS API. Do not
treat a missing desired value as a drift.

---

### aws_secretsmanager_secret

**File:** `config/cluster/secretsmanager/config.go`

#### TerraformCustomDiff
**What it does:** Complex custom diff for the `replica` set field. Secrets
Manager replicas have `region` and `kms_key_id`. When `kms_key_id` is not
specified, AWS auto-assigns one. The custom diff:
1. Builds a map of `region → {kms_key_id → hash}` from current state.
2. For desired replicas with explicit `kms_key_id`: removes the diff if already
   matched in state.
3. For desired replicas with only `region` (no explicit KMS key): matches them
   against remaining replicas in current state by count per region, removes
   creation diffs when counts match.
4. Adjusts `replica.#` count accordingly.

**Native equivalent:** In `Observe()`, when comparing `Replicas`: group by
`Region`. For each region, if the desired spec has `n` entries with no explicit
`KmsKeyId`, and the current state has `n` entries for that region, treat them as
equal regardless of which KMS key IDs are assigned. Only trigger an update if
the count per region changes or an explicit KMS key doesn't match.

---

### aws_sns_topic

**File:** `config/cluster/sns/config.go`

#### TerraformCustomDiff
**What it does:** Suppresses `policy` diff when the old and new JSON policies
are semantically equivalent. Uses `awspolicy.PoliciesAreEquivalent()` after
stripping the `Version` field from both sides (via `common.RemovePolicyVersion`).
This handles cases where AWS normalizes the policy JSON (e.g., reorders
Statement elements) differently than what the user wrote.

**Native equivalent:** In `Observe()`, when comparing the `Policy` field: use
`awspolicy.PoliciesAreEquivalent(desired, actual)` to compare semantically
rather than as strings. If equivalent, do not mark the resource as needing an
update.

---

### aws_sqs_queue

**File:** `config/cluster/sqs/config.go`

#### TerraformCustomDiff
**What it does:** Same as `aws_sns_topic` — suppresses `policy` diff when old
and new JSON policies are semantically equivalent, using
`awspolicy.PoliciesAreEquivalent()` after removing the `Version` field.

**Native equivalent:** Same as SNS — use `awspolicy.PoliciesAreEquivalent()` in
`Observe()` when comparing the SQS queue policy.

---

## UseAsync = true Resources

`UseAsync = true` tells the Terraform bridge that a resource's Create/Delete
operations are long-running (asynchronous). In native controllers, every
resource can be "async" via the standard reconciler loop — the controller
simply returns without marking the resource Ready until the AWS resource reaches
the desired state. The equivalent native pattern is to return `managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}` with an appropriate reason and rely on the reconciler loop.

The following table documents each resource, the long-running operation, and why
`UseAsync` was needed in the Terraform bridge context.

| Resource | Long-running Operation | Reason |
|----------|----------------------|--------|
| `aws_acm_certificate_validation` | DNS/email validation loop | ACM waits for domain validation which can take minutes |
| `aws_api_gateway_vpc_link` | VPC link provisioning | API GW VPC link creation can take minutes |
| `aws_apigatewayv2_stage` | Stage deployment | Stage create/update involves deploying the API |
| `aws_apigatewayv2_vpclink` | VPC link provisioning | Same as API GW v1 |
| `aws_apprunner_vpc_connector` | Connector provisioning | VPC connector initialization can take minutes |
| `aws_appstream_fleet` | Fleet provisioning | Fleet start/stop can take several minutes |
| `aws_appstream_image_builder` | Image builder provisioning | Image builder instance startup takes several minutes |
| `aws_autoscaling_group` | Instance launch/termination | ASG waits for instance health checks on scale events |
| `aws_backup_framework` | Framework activation | Backup framework requires compliance evaluation delay |
| `aws_backup_plan` | Plan creation | Backup plan creation involves policy propagation |
| `aws_cloudfront_distribution` | Distribution deploy | CloudFront propagates to edge locations (takes minutes) |
| `aws_cloudsearch_domain` | Domain indexing | CloudSearch domain initialization takes minutes |
| `aws_cloudsearch_domain_service_access_policy` | Policy update | Same domain initialization wait |
| `aws_cloudwatch_event_permission` | Permission propagation | EventBridge permission propagation |
| `aws_dax_cluster` | Cluster provisioning | DAX cluster node initialization takes minutes |
| `aws_dx_private_virtual_interface` | Virtual interface provisioning | DX VIF BGP session establishment takes time |
| `aws_dx_gateway_association` | Gateway association | DX gateway association propagation takes minutes |
| `aws_dx_hosted_transit_virtual_interface_accepter` | VIF acceptance | DX hosted VIF acceptance takes time |
| `aws_dx_connection` | Physical connection provisioning | DX physical connection setup takes time |
| `aws_dx_bgp_peer` | BGP session establishment | BGP session establishment takes time |
| `aws_dx_transit_virtual_interface` | VIF provisioning | Transit VIF provisioning takes time |
| `aws_docdb_cluster` | Cluster provisioning | DocumentDB cluster creation takes minutes |
| `aws_docdb_cluster_instance` | Instance provisioning | DocumentDB instance creation takes minutes |
| `aws_ecr_repository` | Repository deletion | ECR repository deletion waits for image cleanup |
| `aws_ecrpublic_repository` | Repository deletion | Same as ECR |
| `aws_ecs_cluster` | Cluster provisioning | ECS cluster activation/deactivation takes time |
| `aws_ecs_service` | Service stable state | ECS service waits for task health checks |
| `aws_efs_mount_target` | Mount target provisioning | EFS mount target NFS initialization takes minutes |
| `aws_efs_file_system` | Filesystem provisioning | EFS filesystem creation/deletion takes time |
| `aws_eip` | Address association/disassociation | EIP association with instance takes time |
| `aws_eks_cluster` | Cluster provisioning | EKS cluster creation/deletion takes 10-15 minutes |
| `aws_eks_node_group` | Node group provisioning | EC2 instance launch and k8s node registration takes time |
| `aws_eks_identity_provider_config` | OIDC config propagation | Identity provider config propagation takes time |
| `aws_eks_fargate_profile` | Fargate profile activation | Fargate profile takes time to activate |
| `aws_eks_addon` | Addon installation | EKS addon installation waits for k8s resources |
| `aws_elasticache_replication_group` | Cluster provisioning | ElastiCache replication group creation takes minutes |
| `aws_elasticache_serverless_cache` | Cache provisioning | Serverless cache initialization takes time |
| `aws_emrcontainers_virtual_cluster` | Cluster registration | EMR virtual cluster registration takes time |
| `aws_gamelift_fleet` | Fleet provisioning | GameLift fleet instance launch takes minutes |
| `aws_grafana_workspace` | Workspace provisioning | Grafana workspace initialization takes minutes |
| `aws_grafana_workspace_saml_configuration` | SAML config update | SAML config propagation takes time |
| `aws_grafana_license_association` | License association | License propagation takes time |
| `aws_instance` | Instance launch/termination | EC2 instance takes time to reach running/terminated state |
| `aws_kms_alias` | Alias creation | KMS alias propagation in multi-region |
| `aws_kms_ciphertext` | Encryption operation | KMS encryption can be async |
| `aws_lambda_event_source_mapping` | Mapping activation | Lambda ESM activation takes time |
| `aws_lambda_provisioned_concurrency_config` | Concurrency allocation | Lambda provisioned concurrency takes time to allocate |
| `aws_lb` | Load balancer provisioning | ALB/NLB provisioning takes minutes |
| `aws_lb_target_group_attachment` | Target attachment | ELBv2 target health check takes time |
| `aws_memorydb_cluster` | Cluster provisioning | MemoryDB cluster creation takes minutes |
| `aws_mq_broker` | Broker provisioning | MQ broker creation/deletion takes minutes |
| `aws_msk_single_scram_secret_association` | Secret association | MSK secret rotation propagation |
| `aws_msk_serverless_cluster` | Cluster provisioning | MSK serverless cluster creation takes time |
| `aws_neptune_cluster` | Cluster provisioning | Neptune cluster creation takes minutes |
| `aws_neptune_cluster_endpoint` | Endpoint provisioning | Neptune custom endpoint takes time |
| `aws_neptune_cluster_instance` | Instance provisioning | Neptune instance creation takes minutes |
| `aws_neptune_cluster_snapshot` | Snapshot creation | Neptune snapshot creation takes time |
| `aws_networkfirewall_firewall` | Firewall provisioning | Network Firewall creation/deletion takes minutes |
| `aws_opensearch_domain` | Domain provisioning | OpenSearch domain creation/update takes 10-30 minutes |
| `aws_rds_cluster` | Cluster provisioning | Aurora cluster creation takes minutes |
| `aws_rds_cluster_instance` | Instance provisioning | RDS instance creation takes minutes |
| `aws_db_instance` | Instance provisioning | RDS DB instance creation takes minutes |
| `aws_rds_global_cluster` | Global cluster setup | Cross-region replication setup takes time |
| `aws_db_proxy` | Proxy provisioning | RDS Proxy creation takes minutes |
| `aws_db_proxy_endpoint` | Endpoint provisioning | RDS Proxy endpoint takes time |
| `aws_rds_cluster_activity_stream` | Stream activation | Kinesis activity stream activation takes time |
| `aws_db_snapshot` | Snapshot creation | RDS snapshot takes time proportional to DB size |
| `aws_rds_cluster_endpoint` | Endpoint provisioning | Aurora custom endpoint takes time |
| `aws_rds_cluster_role_association` | IAM association | IAM role association propagation takes time |
| `aws_db_snapshot_copy` | Snapshot copy | Cross-region copy takes time |
| `aws_db_instance_automated_backups_replication` | Backup replication | Cross-region backup setup takes time |
| `aws_db_cluster_snapshot` | Cluster snapshot | Aurora cluster snapshot takes time |
| `aws_redshift_cluster` | Cluster provisioning | Redshift cluster creation takes minutes |
| `aws_route` | Route propagation | Route table propagation takes time |
| `aws_route53_zone` | Zone delegation | Route 53 zone delegation setup takes time |
| `aws_s3_bucket_metrics` | Metrics configuration | S3 metrics configuration propagation |
| `aws_sagemaker_device_fleet` | Fleet provisioning | SageMaker device fleet initialization |
| `aws_sagemaker_endpoint` | Endpoint provisioning | SageMaker endpoint creation/update takes minutes |
| `aws_security_group` | Security group deletion | AWS waits for all resources to detach before deletion |
| `aws_subnet` | Subnet provisioning | Subnet initialization takes time |
| `aws_vpc` | VPC provisioning | VPC creation/deletion can take time |

---

## Notes for Native Implementation

1. **UseAsync** is only relevant to the Terraform bridge. Native controllers
   already implement async via the reconciler loop. No special handling needed —
   just return early from `Observe()` when the resource is not yet in the
   desired state and set appropriate conditions.

2. **TerraformConfigurationInjector** logic (region copying, defaulting) should
   be replaced with:
   - Crossplane late-initialization (`lateInitialize()` in `Observe()`)
   - CRD default values (`// +kubebuilder:default=value`)
   - AWS SDK call construction that handles nil fields gracefully

3. **TerraformCustomDiff** logic (suppressing spurious diffs) should be
   replaced with careful comparison logic in `Observe()`:
   - Policy JSON: use `awspolicy.PoliciesAreEquivalent()`
   - Computed fields: skip comparison for fields not set in spec
   - Engine versions: semantic version comparison
   - Set fields with auto-assigned values: compare by count or key fields only

4. **Policy comparison helper:** The `awspolicy` package
   (`github.com/hashicorp/awspolicyequivalence`) is already a dependency and
   should be used for all IAM/resource policy comparisons in native controllers.

5. **JSON canonicalization:** For OpenSearch Serverless policies, use
   `json.Canonicalize()` or re-marshal through `encoding/json` (which sorts
   map keys) before comparing string policies.
