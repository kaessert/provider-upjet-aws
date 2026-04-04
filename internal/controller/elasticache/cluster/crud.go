// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package cluster implements the shared CRUD logic for ClusterRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the ClusterCR interface.
package cluster

import (
	"context"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe = "cannot describe ElastiCache Cluster"
	errCreate   = "cannot create ElastiCache Cluster"
	errUpdate   = "cannot modify ElastiCache Cluster"
	errDelete   = "cannot delete ElastiCache Cluster"
	errListTags = "cannot list tags for ElastiCache Cluster"
	errAddTags  = "cannot add tags to ElastiCache Cluster"
	errDelTags  = "cannot remove tags from ElastiCache Cluster"
)

// ElastiCacheClusterClient is the interface for AWS ElastiCache operations
// required by the cluster controller. Defined as an interface to enable mocking
// in unit tests; *awselasticache.Client satisfies it.
type ElastiCacheClusterClient interface {
	DescribeCacheClusters(ctx context.Context, params *awselasticache.DescribeCacheClustersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error)
	CreateCacheCluster(ctx context.Context, params *awselasticache.CreateCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheClusterOutput, error)
	ModifyCacheCluster(ctx context.Context, params *awselasticache.ModifyCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheClusterOutput, error)
	DeleteCacheCluster(ctx context.Context, params *awselasticache.DeleteCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheClusterOutput, error)
	ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

// ClusterCR abstracts over cluster-scoped and namespaced ClusterRAW types.
type ClusterCR interface {
	resource.Managed
	GetForProvider() *clusternative.ClusterRAWParameters
	SetForProvider(clusternative.ClusterRAWParameters)
	GetInitProvider() *clusternative.ClusterRAWInitParameters
	GetAtProvider() clusternative.ClusterRAWObservation
	SetAtProvider(clusternative.ClusterRAWObservation)
}

// ExternalClient implements the shared CRUD logic for ClusterRAW resources.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCacheClusterClient
	// Kube is the Kubernetes client (reserved for future use).
	Kube client.Client
}

// Observe checks whether the external ClusterRAW resource exists and is up-to-date.
//
// Transitional state handling (spec §3):
//
//	"available"              → Available; proceed to isUpToDate
//	"creating"               → Unavailable; return UpToDate=true (prevent spurious Update)
//	"modifying"              → Unavailable; return UpToDate=true
//	"rebooting cluster nodes"→ Unavailable; return UpToDate=true
//	"snapshotting"           → Unavailable; return UpToDate=true
//	"deleting"               → Deleting; return UpToDate=true
//
// Without this, Observe calls isUpToDate on a "creating" cluster where AWS
// defaults aren't applied yet → returns false → triggers Update() → AWS returns
// InvalidCacheClusterState.
//
//nolint:gocyclo
func (e *ExternalClient) Observe(ctx context.Context, cr ClusterCR) (managed.ExternalObservation, error) {
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeCacheClusters(ctx, &awselasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(extName),
		ShowCacheNodeInfo: aws.Bool(true),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	if len(resp.CacheClusters) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cc := resp.CacheClusters[0]
	status := aws.ToString(cc.CacheClusterStatus)

	// Handle transitional states: skip isUpToDate to prevent spurious Updates.
	// spec §3: CacheClusterStatus transitional values.
	switch status {
	case "creating", "modifying", "rebooting cluster nodes", "snapshotting":
		cr.SetConditions(xpv1.Unavailable())
		setAtProviderFromCluster(cr, cc, nil)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
		setAtProviderFromCluster(cr, cc, nil)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	// Status is "available" (or unknown) — proceed to full observation.
	cr.SetConditions(xpv1.Available())

	// Fetch tags (non-fatal — tag errors should not fail Observe).
	var observedTags []ectypes.Tag
	if cc.ARN != nil {
		tagsResp, tagErr := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
			ResourceName: cc.ARN,
		})
		if tagErr == nil && tagsResp != nil {
			observedTags = tagsResp.TagList
		}
	}

	setAtProviderFromCluster(cr, cc, observedTags)

	// Late-initialize AWS-defaulted fields (spec §8).
	// The late-initialized values are AWS-assigned defaults already present in AWS.
	// We copy them into spec for future isUpToDate comparisons but do NOT need to
	// push them back to AWS. ResourceUpToDate=true prevents a spurious Update.
	spec := cr.GetForProvider()
	if lateInitializeCluster(spec, cc) {
		cr.SetForProvider(*spec)
		connDetails := buildConnectionDetails(cc)
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceUpToDate:        true,
			ResourceLateInitialized: true,
			ConnectionDetails:       connDetails,
		}, nil
	}

	// Build connection details from observed cluster (spec §7).
	connDetails := buildConnectionDetails(cc)

	upToDate := isUpToDate(cr.GetForProvider(), cc, observedTags)
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: connDetails,
	}, nil
}

// Create creates the external ClusterRAW resource.
func (e *ExternalClient) Create(ctx context.Context, cr ClusterCR) (managed.ExternalCreation, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	input := buildCreateInput(extName, spec)

	_, err := e.Client.CreateCacheCluster(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// ParameterAsIdentifier: external name is already set before Create is called.
	// No additional SetExternalName call is needed.

	// ClusterRAW publishes connection details from Observe (not Create).
	return managed.ExternalCreation{}, nil
}

// Update updates the external ClusterRAW resource.
func (e *ExternalClient) Update(ctx context.Context, cr ClusterCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	_, err := e.Client.ModifyCacheCluster(ctx, buildModifyInput(extName, spec))
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	// Sync tags using the ARN from atProvider.
	arn := aws.ToString(cr.GetAtProvider().Arn)
	if arn != "" {
		if err := e.syncTags(ctx, arn, spec.Tags); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external ClusterRAW resource.
// Idempotent: CacheClusterNotFoundFault is treated as success.
func (e *ExternalClient) Delete(ctx context.Context, cr ClusterCR) (managed.ExternalDelete, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	input := &awselasticache.DeleteCacheClusterInput{
		CacheClusterId: aws.String(extName),
	}
	// Include FinalSnapshotIdentifier if specified.
	if spec.FinalSnapshotIdentifier != nil {
		input.FinalSnapshotIdentifier = spec.FinalSnapshotIdentifier
	}

	_, err := e.Client.DeleteCacheCluster(ctx, input)
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}
	return managed.ExternalDelete{}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// buildCreateInput constructs a CreateCacheClusterInput from the spec.
//
//nolint:gocyclo
func buildCreateInput(clusterID string, spec *clusternative.ClusterRAWParameters) *awselasticache.CreateCacheClusterInput {
	input := &awselasticache.CreateCacheClusterInput{
		CacheClusterId:             aws.String(clusterID),
		Engine:                     spec.Engine,
		EngineVersion:              spec.EngineVersion,
		CacheNodeType:              spec.NodeType,
		CacheSubnetGroupName:       spec.SubnetGroupName,
		CacheParameterGroupName:    spec.ParameterGroupName,
		NotificationTopicArn:       spec.NotificationTopicArn,
		SnapshotName:               spec.SnapshotName,
		PreferredMaintenanceWindow: spec.MaintenanceWindow,
		SnapshotWindow:             spec.SnapshotWindow,
		ReplicationGroupId:         spec.ReplicationGroupID,
		Tags:                       mapToTags(spec.Tags),
	}

	// NumCacheNodes: *float64 → *int32.
	if spec.NumCacheNodes != nil {
		n := int32(*spec.NumCacheNodes)
		input.NumCacheNodes = &n
	}

	// Port: *float64 → *int32.
	if spec.Port != nil {
		p := int32(*spec.Port)
		input.Port = &p
	}

	// SnapshotRetentionLimit: *float64 → *int32.
	if spec.SnapshotRetentionLimit != nil {
		s := int32(*spec.SnapshotRetentionLimit)
		input.SnapshotRetentionLimit = &s
	}

	// AZMode: *string → ectypes.AZMode.
	if spec.AzMode != nil {
		input.AZMode = ectypes.AZMode(*spec.AzMode)
	}

	// IpDiscovery: *string → ectypes.IpDiscovery.
	if spec.IPDiscovery != nil {
		input.IpDiscovery = ectypes.IpDiscovery(*spec.IPDiscovery)
	}

	// NetworkType: *string → ectypes.NetworkType.
	if spec.NetworkType != nil {
		input.NetworkType = ectypes.NetworkType(*spec.NetworkType)
	}

	// OutpostMode: *string → ectypes.OutpostMode.
	if spec.OutpostMode != nil {
		input.OutpostMode = ectypes.OutpostMode(*spec.OutpostMode)
	}

	// AutoMinorVersionUpgrade: *string → *bool.
	if spec.AutoMinorVersionUpgrade != nil {
		v := *spec.AutoMinorVersionUpgrade == "true" || *spec.AutoMinorVersionUpgrade == "yes"
		input.AutoMinorVersionUpgrade = &v
	}

	// TransitEncryptionEnabled.
	if spec.TransitEncryptionEnabled != nil {
		input.TransitEncryptionEnabled = spec.TransitEncryptionEnabled
	}

	// SecurityGroupIds: []*string → []string.
	if len(spec.SecurityGroupIds) > 0 {
		input.SecurityGroupIds = derefStringSlice(spec.SecurityGroupIds)
	}

	// SnapshotArns: []*string → []string.
	if len(spec.SnapshotArns) > 0 {
		input.SnapshotArns = derefStringSlice(spec.SnapshotArns)
	}

	// PreferredAvailabilityZones: []*string → []string.
	if len(spec.PreferredAvailabilityZones) > 0 {
		input.PreferredAvailabilityZones = derefStringSlice(spec.PreferredAvailabilityZones)
	}

	// PreferredOutpostArn: *string (single).
	if spec.PreferredOutpostArn != nil {
		input.PreferredOutpostArn = spec.PreferredOutpostArn
	}

	// LogDeliveryConfigurations.
	if len(spec.LogDeliveryConfiguration) > 0 {
		input.LogDeliveryConfigurations = buildLogDeliveryConfigs(spec.LogDeliveryConfiguration)
	}

	return input
}

// buildModifyInput constructs a ModifyCacheClusterInput from the spec.
func buildModifyInput(clusterID string, spec *clusternative.ClusterRAWParameters) *awselasticache.ModifyCacheClusterInput {
	input := &awselasticache.ModifyCacheClusterInput{
		CacheClusterId:             aws.String(clusterID),
		ApplyImmediately:           spec.ApplyImmediately,
		CacheNodeType:              spec.NodeType,
		EngineVersion:              spec.EngineVersion,
		CacheParameterGroupName:    spec.ParameterGroupName,
		NotificationTopicArn:       spec.NotificationTopicArn,
		PreferredMaintenanceWindow: spec.MaintenanceWindow,
		SnapshotRetentionLimit:     nil,
		SnapshotWindow:             spec.SnapshotWindow,
	}

	// NumCacheNodes: *float64 → *int32.
	if spec.NumCacheNodes != nil {
		n := int32(*spec.NumCacheNodes)
		input.NumCacheNodes = &n
	}

	// SnapshotRetentionLimit: *float64 → *int32.
	if spec.SnapshotRetentionLimit != nil {
		s := int32(*spec.SnapshotRetentionLimit)
		input.SnapshotRetentionLimit = &s
	}

	// AZMode: *string → ectypes.AZMode.
	if spec.AzMode != nil {
		input.AZMode = ectypes.AZMode(*spec.AzMode)
	}

	// IpDiscovery: *string → ectypes.IpDiscovery.
	if spec.IPDiscovery != nil {
		input.IpDiscovery = ectypes.IpDiscovery(*spec.IPDiscovery)
	}

	// AutoMinorVersionUpgrade: *string → *bool.
	if spec.AutoMinorVersionUpgrade != nil {
		v := *spec.AutoMinorVersionUpgrade == "true" || *spec.AutoMinorVersionUpgrade == "yes"
		input.AutoMinorVersionUpgrade = &v
	}

	// SecurityGroupIds: []*string → []string.
	// security_group_names is intentionally omitted (EC2-Classic legacy, not supported in VPC).
	if len(spec.SecurityGroupIds) > 0 {
		input.SecurityGroupIds = derefStringSlice(spec.SecurityGroupIds)
	}

	// LogDeliveryConfigurations.
	if len(spec.LogDeliveryConfiguration) > 0 {
		input.LogDeliveryConfigurations = buildLogDeliveryConfigs(spec.LogDeliveryConfiguration)
	}

	return input
}

// buildLogDeliveryConfigs converts spec log delivery configurations to AWS types.
func buildLogDeliveryConfigs(cfgs []clusternative.ClusterLogDeliveryConfigurationRAWParameters) []ectypes.LogDeliveryConfigurationRequest {
	out := make([]ectypes.LogDeliveryConfigurationRequest, 0, len(cfgs))
	for _, c := range cfgs {
		c := c
		req := ectypes.LogDeliveryConfigurationRequest{
			DestinationDetails: &ectypes.DestinationDetails{},
		}
		if c.LogType != nil {
			req.LogType = ectypes.LogType(*c.LogType)
		}
		if c.LogFormat != nil {
			req.LogFormat = ectypes.LogFormat(*c.LogFormat)
		}
		if c.DestinationType != nil {
			req.DestinationType = ectypes.DestinationType(*c.DestinationType)
		}
		if c.Destination != nil {
			switch aws.ToString(c.DestinationType) {
			case "cloudwatch-logs":
				req.DestinationDetails.CloudWatchLogsDetails = &ectypes.CloudWatchLogsDestinationDetails{
					LogGroup: c.Destination,
				}
			case "kinesis-firehose":
				req.DestinationDetails.KinesisFirehoseDetails = &ectypes.KinesisFirehoseDestinationDetails{
					DeliveryStream: c.Destination,
				}
			}
		}
		out = append(out, req)
	}
	return out
}

// buildConnectionDetails builds connection details from the observed cache cluster.
//
// Connection details (spec §7):
//   - For Memcached: "cluster_address" (ConfigurationEndpoint.Address) + "port" (ConfigurationEndpoint.Port)
//   - For Redis/Valkey: "port" only (from CacheNodes[0].Endpoint.Port); ConfigurationEndpoint is nil
//
// Nil pointer guards: ConfigurationEndpoint is nil for Redis clusters, CacheNodes
// may be empty during creation.
func buildConnectionDetails(cc ectypes.CacheCluster) managed.ConnectionDetails {
	conn := managed.ConnectionDetails{}

	engine := aws.ToString(cc.Engine)

	if engine == "memcached" {
		// Memcached: use ConfigurationEndpoint for cluster_address + port.
		if cc.ConfigurationEndpoint != nil {
			if cc.ConfigurationEndpoint.Address != nil {
				conn["cluster_address"] = []byte(aws.ToString(cc.ConfigurationEndpoint.Address))
			}
			if cc.ConfigurationEndpoint.Port != nil {
				conn["port"] = []byte(strconv.FormatInt(int64(*cc.ConfigurationEndpoint.Port), 10))
			}
		}
	} else {
		// Redis / Valkey: port comes from CacheNodes[0].Endpoint.Port only.
		// ConfigurationEndpoint is nil for Redis; never publish cluster_address.
		if len(cc.CacheNodes) > 0 && cc.CacheNodes[0].Endpoint != nil {
			if cc.CacheNodes[0].Endpoint.Port != nil {
				conn["port"] = []byte(strconv.FormatInt(int64(*cc.CacheNodes[0].Endpoint.Port), 10))
			}
		}
	}

	return conn
}

// setAtProviderFromCluster populates the atProvider observation from the AWS
// CacheCluster state and the provided tag list.
//
//nolint:gocyclo
func setAtProviderFromCluster(cr ClusterCR, cc ectypes.CacheCluster, tags []ectypes.Tag) {
	o := clusternative.ClusterRAWObservation{
		Arn:                aws.String(aws.ToString(cc.ARN)),
		CacheClusterStatus: cc.CacheClusterStatus,
		Engine:             cc.Engine,
		EngineVersion:      cc.EngineVersion,
		NodeType:           cc.CacheNodeType,
		ID:                 cc.CacheClusterId,
		SubnetGroupName:    cc.CacheSubnetGroupName,
		ReplicationGroupID: cc.ReplicationGroupId,
		Tags:               tagsToMap(tags),
	}

	// NumCacheNodes: *int32 → *float64.
	if cc.NumCacheNodes != nil {
		f := float64(*cc.NumCacheNodes)
		o.NumCacheNodes = &f
	}

	// CacheNodes.
	if len(cc.CacheNodes) > 0 {
		cacheNodes := make([]clusternative.CacheNodeRAWObservation, 0, len(cc.CacheNodes))
		for _, n := range cc.CacheNodes {
			n := n
			nodeObs := clusternative.CacheNodeRAWObservation{
				ID:               n.CacheNodeId,
				AvailabilityZone: n.CustomerAvailabilityZone,
				OutpostArn:       n.CustomerOutpostArn,
			}
			if n.Endpoint != nil {
				nodeObs.Address = n.Endpoint.Address
				if n.Endpoint.Port != nil {
					f := float64(*n.Endpoint.Port)
					nodeObs.Port = &f
				}
			}
			cacheNodes = append(cacheNodes, nodeObs)
		}
		o.CacheNodes = cacheNodes
	}

	// SecurityGroupIds from SecurityGroups.
	if len(cc.SecurityGroups) > 0 {
		sgIDs := make([]*string, 0, len(cc.SecurityGroups))
		for _, sg := range cc.SecurityGroups {
			sg := sg
			sgIDs = append(sgIDs, sg.SecurityGroupId)
		}
		o.SecurityGroupIds = sgIDs
	}

	// LogDeliveryConfiguration.
	if len(cc.LogDeliveryConfigurations) > 0 {
		logConfigs := make([]clusternative.ClusterLogDeliveryConfigurationRAWObservation, 0, len(cc.LogDeliveryConfigurations))
		for _, ldc := range cc.LogDeliveryConfigurations {
			ldc := ldc
			logType := string(ldc.LogType)
			logFormat := string(ldc.LogFormat)
			destType := string(ldc.DestinationType)
			logObs := clusternative.ClusterLogDeliveryConfigurationRAWObservation{
				LogType:         &logType,
				LogFormat:       &logFormat,
				DestinationType: &destType,
			}
			if ldc.DestinationDetails != nil {
				if ldc.DestinationDetails.CloudWatchLogsDetails != nil {
					logObs.Destination = ldc.DestinationDetails.CloudWatchLogsDetails.LogGroup
				} else if ldc.DestinationDetails.KinesisFirehoseDetails != nil {
					logObs.Destination = ldc.DestinationDetails.KinesisFirehoseDetails.DeliveryStream
				}
			}
			logConfigs = append(logConfigs, logObs)
		}
		o.LogDeliveryConfiguration = logConfigs
	}

	// MaintenanceWindow from PreferredMaintenanceWindow.
	o.MaintenanceWindow = cc.PreferredMaintenanceWindow

	// SnapshotRetentionLimit: *int32 → *float64.
	if cc.SnapshotRetentionLimit != nil {
		f := float64(*cc.SnapshotRetentionLimit)
		o.SnapshotRetentionLimit = &f
	}

	// SnapshotWindow.
	o.SnapshotWindow = cc.SnapshotWindow

	// TransitEncryptionEnabled.
	o.TransitEncryptionEnabled = cc.TransitEncryptionEnabled

	// AvailabilityZone from PreferredAvailabilityZone.
	o.AvailabilityZone = cc.PreferredAvailabilityZone

	// AutoMinorVersionUpgrade: *bool → *string ("true"/"false").
	if cc.AutoMinorVersionUpgrade != nil {
		s := strconv.FormatBool(*cc.AutoMinorVersionUpgrade)
		o.AutoMinorVersionUpgrade = &s
	}

	// Port / ClusterAddress / ConfigurationEndpoint.
	// For Memcached, ConfigurationEndpoint is set at the cluster level and
	// contains the cluster's address and port.
	// For Redis/Valkey, ConfigurationEndpoint is nil; port comes from CacheNodes[0].
	if cc.ConfigurationEndpoint != nil {
		o.ClusterAddress = cc.ConfigurationEndpoint.Address
		if cc.ConfigurationEndpoint.Port != nil {
			f := float64(*cc.ConfigurationEndpoint.Port)
			o.Port = &f
		}
		// ConfigurationEndpoint atProvider field is "address:port" (TF parity).
		if cc.ConfigurationEndpoint.Address != nil && cc.ConfigurationEndpoint.Port != nil {
			ep := *cc.ConfigurationEndpoint.Address + ":" + strconv.Itoa(int(*cc.ConfigurationEndpoint.Port))
			o.ConfigurationEndpoint = &ep
		}
	} else if len(cc.CacheNodes) > 0 && cc.CacheNodes[0].Endpoint != nil && cc.CacheNodes[0].Endpoint.Port != nil {
		// Redis/Valkey: top-level Port from the first node's endpoint.
		f := float64(*cc.CacheNodes[0].Endpoint.Port)
		o.Port = &f
	}

	cr.SetAtProvider(o)
}

// lateInitializeCluster copies AWS-defaulted fields from the observed CacheCluster
// into spec when the spec fields are nil. Returns true if any field was changed.
//
// This prevents infinite reconciliation loops when optional fields (such as NodeType)
// are omitted from the user manifest and AWS has them set (e.g. a cluster linked to a
// replication group inherits NodeType from the group). Without late-init, isUpToDate
// would compare "" (nil spec) vs "cache.r7g.medium" (AWS) every cycle and always
// return false, triggering Update on every reconcile iteration (spec §8, §17).
//
//nolint:gocyclo
func lateInitializeCluster(spec *clusternative.ClusterRAWParameters, cc ectypes.CacheCluster) bool {
	changed := false
	changed = nativehelper.LateInitializeStringPtr(&spec.NodeType, cc.CacheNodeType) || changed
	changed = nativehelper.LateInitializeStringPtr(&spec.EngineVersion, cc.EngineVersion) || changed
	changed = nativehelper.LateInitializeStringPtr(&spec.MaintenanceWindow, cc.PreferredMaintenanceWindow) || changed
	changed = nativehelper.LateInitializeStringPtr(&spec.SnapshotWindow, cc.SnapshotWindow) || changed
	if spec.SnapshotRetentionLimit == nil && cc.SnapshotRetentionLimit != nil {
		v := float64(*cc.SnapshotRetentionLimit)
		spec.SnapshotRetentionLimit = &v
		changed = true
	}
	if cc.CacheParameterGroup != nil {
		changed = nativehelper.LateInitializeStringPtr(&spec.ParameterGroupName, cc.CacheParameterGroup.CacheParameterGroupName) || changed
	}
	if spec.NumCacheNodes == nil && cc.NumCacheNodes != nil {
		v := float64(*cc.NumCacheNodes)
		spec.NumCacheNodes = &v
		changed = true
	}
	return changed
}

// isUpToDate returns true when the spec is in sync with the observed AWS state.
//
//nolint:gocyclo
func isUpToDate(spec *clusternative.ClusterRAWParameters, cc ectypes.CacheCluster, observedTags []ectypes.Tag) bool {
	// NodeType: guard against nil — clusters linked to a replication group may not
	// have NodeType in spec (it is inherited from the RG). After late-initialization
	// the spec will be populated, but a nil check prevents a transient update call.
	if spec.NodeType != nil && aws.ToString(spec.NodeType) != aws.ToString(cc.CacheNodeType) {
		return false
	}

	// Engine/EngineVersion (only if specified in spec).
	if spec.EngineVersion != nil && aws.ToString(spec.EngineVersion) != aws.ToString(cc.EngineVersion) {
		return false
	}

	// NumCacheNodes.
	if spec.NumCacheNodes != nil && cc.NumCacheNodes != nil {
		if int32(*spec.NumCacheNodes) != *cc.NumCacheNodes {
			return false
		}
	}

	// SecurityGroupIds (order-independent set comparison).
	// security_group_names is intentionally skipped (EC2-Classic legacy field).
	if !securityGroupIDsUpToDate(spec.SecurityGroupIds, cc.SecurityGroups) {
		return false
	}

	// Tags.
	if !tagsUpToDate(spec.Tags, observedTags) {
		return false
	}

	// MaintenanceWindow.
	if spec.MaintenanceWindow != nil && aws.ToString(spec.MaintenanceWindow) != aws.ToString(cc.PreferredMaintenanceWindow) {
		return false
	}

	// SnapshotRetentionLimit: spec is *float64, AWS is *int32.
	if spec.SnapshotRetentionLimit != nil {
		specSRL := int32(*spec.SnapshotRetentionLimit)
		awsSRL := int32(0)
		if cc.SnapshotRetentionLimit != nil {
			awsSRL = *cc.SnapshotRetentionLimit
		}
		if specSRL != awsSRL {
			return false
		}
	}

	// SnapshotWindow.
	if spec.SnapshotWindow != nil && aws.ToString(spec.SnapshotWindow) != aws.ToString(cc.SnapshotWindow) {
		return false
	}

	// AutoMinorVersionUpgrade: spec is *string ("true"/"false"), AWS is *bool.
	if spec.AutoMinorVersionUpgrade != nil && cc.AutoMinorVersionUpgrade != nil {
		specAMVU := *spec.AutoMinorVersionUpgrade == "true" || *spec.AutoMinorVersionUpgrade == "yes"
		if specAMVU != *cc.AutoMinorVersionUpgrade {
			return false
		}
	}

	// NotificationTopicArn.
	if spec.NotificationTopicArn != nil && cc.NotificationConfiguration != nil {
		if aws.ToString(spec.NotificationTopicArn) != aws.ToString(cc.NotificationConfiguration.TopicArn) {
			return false
		}
	}

	// ParameterGroupName.
	if spec.ParameterGroupName != nil && cc.CacheParameterGroup != nil {
		if aws.ToString(spec.ParameterGroupName) != aws.ToString(cc.CacheParameterGroup.CacheParameterGroupName) {
			return false
		}
	}

	// LogDeliveryConfiguration.
	if !logDeliveryConfigUpToDate(spec.LogDeliveryConfiguration, cc.LogDeliveryConfigurations) {
		return false
	}

	return true
}

// logDeliveryConfigUpToDate returns true when the spec log delivery configurations
// match the observed AWS configurations.
//
// Comparison strategy:
//   - Count mismatch → not up-to-date
//   - For each spec entry, find a matching AWS entry by LogType (at most 2 entries)
//   - Compare LogFormat, DestinationType, and destination value
//
//nolint:gocyclo
func logDeliveryConfigUpToDate(spec []clusternative.ClusterLogDeliveryConfigurationRAWParameters, aws []ectypes.LogDeliveryConfiguration) bool {
	if len(spec) != len(aws) {
		return false
	}
	if len(spec) == 0 {
		return true
	}

	// Build a map keyed by LogType for the AWS observed configs.
	awsByLogType := make(map[ectypes.LogType]ectypes.LogDeliveryConfiguration, len(aws))
	for _, a := range aws {
		awsByLogType[a.LogType] = a
	}

	for _, s := range spec {
		if s.LogType == nil {
			continue
		}
		awsEntry, ok := awsByLogType[ectypes.LogType(*s.LogType)]
		if !ok {
			return false
		}
		// Compare LogFormat.
		if s.LogFormat != nil && string(awsEntry.LogFormat) != *s.LogFormat {
			return false
		}
		// Compare DestinationType.
		if s.DestinationType != nil && string(awsEntry.DestinationType) != *s.DestinationType {
			return false
		}
		// Compare destination value.
		if s.Destination != nil && awsEntry.DestinationDetails != nil {
			var awsDest string
			switch {
			case awsEntry.DestinationDetails.CloudWatchLogsDetails != nil:
				awsDest = aws2ToString(awsEntry.DestinationDetails.CloudWatchLogsDetails.LogGroup)
			case awsEntry.DestinationDetails.KinesisFirehoseDetails != nil:
				awsDest = aws2ToString(awsEntry.DestinationDetails.KinesisFirehoseDetails.DeliveryStream)
			}
			if *s.Destination != awsDest {
				return false
			}
		}
	}

	return true
}

// aws2ToString is a local alias to avoid shadowing the imported aws package
// inside logDeliveryConfigUpToDate which uses a parameter named "aws".
func aws2ToString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// securityGroupIDsUpToDate returns true when the spec security group IDs match
// the observed ones (order-independent).
func securityGroupIDsUpToDate(specIDs []*string, awsSGs []ectypes.SecurityGroupMembership) bool {
	awsIDs := make([]string, 0, len(awsSGs))
	for _, sg := range awsSGs {
		if sg.SecurityGroupId != nil {
			awsIDs = append(awsIDs, *sg.SecurityGroupId)
		}
	}
	specIDsStr := derefStringSlice(specIDs)
	return sortedStringSliceEqual(specIDsStr, awsIDs)
}

// syncTags reconciles desired tags on the cluster using its ARN.
func (e *ExternalClient) syncTags(ctx context.Context, arn string, desired map[string]*string) error {
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	observed := tagsResp.TagList
	toAdd, toRemove := diffTags(desired, observed)

	if len(toAdd) > 0 {
		if _, err := e.Client.AddTagsToResource(ctx, &awselasticache.AddTagsToResourceInput{
			ResourceName: aws.String(arn),
			Tags:         toAdd,
		}); err != nil {
			return nativehelper.Wrap(err, errAddTags)
		}
	}

	if len(toRemove) > 0 {
		keys := make([]string, 0, len(toRemove))
		for _, t := range toRemove {
			keys = append(keys, aws.ToString(t.Key))
		}
		if _, err := e.Client.RemoveTagsFromResource(ctx, &awselasticache.RemoveTagsFromResourceInput{
			ResourceName: aws.String(arn),
			TagKeys:      keys,
		}); err != nil {
			return nativehelper.Wrap(err, errDelTags)
		}
	}

	return nil
}

// tagsUpToDate returns true when the desired spec tags match the observed AWS tags.
func tagsUpToDate(specTags map[string]*string, observed []ectypes.Tag) bool {
	if len(specTags) != len(observed) {
		return false
	}
	observedMap := tagsToMap(observed)
	for k, v := range specTags {
		obsV, ok := observedMap[k]
		if !ok {
			return false
		}
		if aws.ToString(v) != aws.ToString(obsV) {
			return false
		}
	}
	return true
}

// diffTags computes tags to add and tags to remove.
func diffTags(desired map[string]*string, observed []ectypes.Tag) (toAdd []ectypes.Tag, toRemove []ectypes.Tag) {
	observedMap := tagsToMap(observed)

	// Tags to add or update.
	for k, v := range desired {
		tagKey := k
		obsVal, exists := observedMap[tagKey]
		if !exists || aws.ToString(v) != aws.ToString(obsVal) {
			toAdd = append(toAdd, ectypes.Tag{Key: aws.String(tagKey), Value: v})
		}
	}

	// Tags to remove (exist in AWS but not in spec).
	for _, t := range observed {
		if _, exists := desired[aws.ToString(t.Key)]; !exists {
			toRemove = append(toRemove, t)
		}
	}

	return toAdd, toRemove
}

// mapToTags converts a map[string]*string to a []ectypes.Tag slice.
func mapToTags(m map[string]*string) []ectypes.Tag {
	tags := make([]ectypes.Tag, 0, len(m))
	for k, v := range m {
		tagKey := k
		tags = append(tags, ectypes.Tag{Key: aws.String(tagKey), Value: v})
	}
	return tags
}

// tagsToMap converts a []ectypes.Tag slice to a map[string]*string.
func tagsToMap(tags []ectypes.Tag) map[string]*string {
	m := make(map[string]*string, len(tags))
	for _, t := range tags {
		if t.Key != nil {
			m[*t.Key] = t.Value
		}
	}
	return m
}

// derefStringSlice dereferences a []*string slice to []string, skipping nil pointers.
func derefStringSlice(s []*string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != nil {
			result = append(result, *v)
		}
	}
	return result
}

// sortedStringSliceEqual compares two string slices in a set-equal manner
// (order-independent).
func sortedStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sorted := func(ss []string) []string {
		cp := make([]string, len(ss))
		copy(cp, ss)
		sort.Strings(cp)
		return cp
	}
	a, b = sorted(a), sorted(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
