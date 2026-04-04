// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/native"
)

// ── Mock client ────────────────────────────────────────────────────────────────

type mockClusterClient struct {
	describeFn   func(ctx context.Context, params *awselasticache.DescribeCacheClustersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error)
	createFn     func(ctx context.Context, params *awselasticache.CreateCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheClusterOutput, error)
	modifyFn     func(ctx context.Context, params *awselasticache.ModifyCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheClusterOutput, error)
	deleteFn     func(ctx context.Context, params *awselasticache.DeleteCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheClusterOutput, error)
	listTagsFn   func(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	addTagsFn    func(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	removeTagsFn func(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

func (m *mockClusterClient) DescribeCacheClusters(ctx context.Context, params *awselasticache.DescribeCacheClustersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
	return m.describeFn(ctx, params, optFns...)
}
func (m *mockClusterClient) CreateCacheCluster(ctx context.Context, params *awselasticache.CreateCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheClusterOutput, error) {
	return m.createFn(ctx, params, optFns...)
}
func (m *mockClusterClient) ModifyCacheCluster(ctx context.Context, params *awselasticache.ModifyCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheClusterOutput, error) {
	return m.modifyFn(ctx, params, optFns...)
}
func (m *mockClusterClient) DeleteCacheCluster(ctx context.Context, params *awselasticache.DeleteCacheClusterInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheClusterOutput, error) {
	return m.deleteFn(ctx, params, optFns...)
}
func (m *mockClusterClient) ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	if m.listTagsFn != nil {
		return m.listTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}
func (m *mockClusterClient) AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
	if m.addTagsFn != nil {
		return m.addTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.AddTagsToResourceOutput{}, nil
}
func (m *mockClusterClient) RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error) {
	if m.removeTagsFn != nil {
		return m.removeTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.RemoveTagsFromResourceOutput{}, nil
}

// ── Constants ──────────────────────────────────────────────────────────────────

const (
	testClusterID  = "my-cluster"
	testClusterARN = "arn:aws:elasticache:us-east-1:609897127049:cluster:my-cluster"
)

// ── Test Helpers ───────────────────────────────────────────────────────────────

// newTestCR builds a minimal cluster-scoped ClusterRAW for tests.
func newTestCR(name, extName string) *clusternative.ClusterRAW {
	cr := &clusternative.ClusterRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: clusternative.ClusterRAWSpec{
			ForProvider: clusternative.ClusterRAWParameters{
				Region:   aws.String("us-east-1"),
				Engine:   aws.String("redis"),
				NodeType: aws.String("cache.t3.micro"),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// noopListTags returns an empty tag list without error.
func noopListTags(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}

// redisClusterResponse returns a DescribeCacheClustersOutput for a Redis cluster.
func redisClusterResponse(id, arn, status string, nodes []ectypes.CacheNode) *awselasticache.DescribeCacheClustersOutput {
	return &awselasticache.DescribeCacheClustersOutput{
		CacheClusters: []ectypes.CacheCluster{
			{
				CacheClusterId:     aws.String(id),
				ARN:                aws.String(arn),
				CacheClusterStatus: aws.String(status),
				Engine:             aws.String("redis"),
				CacheNodes:         nodes,
			},
		},
	}
}

// memcachedClusterResponse returns a DescribeCacheClustersOutput for a Memcached cluster.
func memcachedClusterResponse(id, arn, status string, cfgEndpoint *ectypes.Endpoint, nodes []ectypes.CacheNode) *awselasticache.DescribeCacheClustersOutput {
	return &awselasticache.DescribeCacheClustersOutput{
		CacheClusters: []ectypes.CacheCluster{
			{
				CacheClusterId:        aws.String(id),
				ARN:                   aws.String(arn),
				CacheClusterStatus:    aws.String(status),
				Engine:                aws.String("memcached"),
				ConfigurationEndpoint: cfgEndpoint,
				CacheNodes:            nodes,
			},
		},
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-cluster", "")
	e := &ExternalClient{Client: &mockClusterClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_ClusterNotFound(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return nil, &ectypes.CacheClusterNotFoundFault{Message: aws.String("not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for NotFound, got: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when cluster not found")
	}
}

func TestObserve_EmptyList(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return &awselasticache.DescribeCacheClustersOutput{CacheClusters: []ectypes.CacheCluster{}}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty CacheClusters list")
	}
}

// TestObserve_Available_Redis_ConnectionDetails verifies that for a Redis cluster in
// "available" state, connection details contain only "port" (from CacheNodes[0].Endpoint)
// and NOT "cluster_address" (ConfigurationEndpoint is nil for Redis).
func TestObserve_Available_Redis_ConnectionDetails(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("redis")

	// Redis cluster — ConfigurationEndpoint is nil, port comes from CacheNodes[0].Endpoint.
	nodes := []ectypes.CacheNode{
		{
			Endpoint: &ectypes.Endpoint{
				Address: aws.String("my-cluster.abc.0001.use1.cache.amazonaws.com"),
				Port:    aws.Int32(6379),
			},
		},
	}

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return redisClusterResponse(testClusterID, testClusterARN, "available", nodes), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}

	// port must be published.
	portVal, ok := obs.ConnectionDetails["port"]
	if !ok {
		t.Error("expected 'port' in connection details for Redis cluster")
	} else if string(portVal) != "6379" {
		t.Errorf("expected port='6379', got '%s'", string(portVal))
	}

	// cluster_address must NOT be published for Redis.
	if _, ok := obs.ConnectionDetails["cluster_address"]; ok {
		t.Error("'cluster_address' must NOT be in connection details for Redis cluster")
	}
}

// TestObserve_Available_Memcached_ConnectionDetails verifies that for a Memcached cluster
// in "available" state, connection details contain "cluster_address" AND "port"
// from ConfigurationEndpoint.
func TestObserve_Available_Memcached_ConnectionDetails(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("memcached")

	cfgEndpoint := &ectypes.Endpoint{
		Address: aws.String("my-cluster.cfg.usw2.cache.amazonaws.com"),
		Port:    aws.Int32(11211),
	}

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return memcachedClusterResponse(testClusterID, testClusterARN, "available", cfgEndpoint, nil), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}

	// cluster_address must be published for Memcached.
	addrVal, ok := obs.ConnectionDetails["cluster_address"]
	if !ok {
		t.Error("expected 'cluster_address' in connection details for Memcached cluster")
	} else if string(addrVal) != "my-cluster.cfg.usw2.cache.amazonaws.com" {
		t.Errorf("unexpected cluster_address: %s", string(addrVal))
	}

	// port must be published.
	portVal, ok := obs.ConnectionDetails["port"]
	if !ok {
		t.Error("expected 'port' in connection details for Memcached cluster")
	} else if string(portVal) != "11211" {
		t.Errorf("expected port='11211', got '%s'", string(portVal))
	}
}

// TestObserve_NilGuard_CacheNodesEmpty verifies that when CacheNodes is empty
// (during creation), Observe does not panic and returns UpToDate=true for Redis.
func TestObserve_NilGuard_CacheNodesEmpty(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("redis")

	// Redis cluster with no nodes yet (creating).
	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return redisClusterResponse(testClusterID, testClusterARN, "creating", nil), nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("should not panic or error with empty CacheNodes: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for creating state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for creating state")
	}
}

// TestObserve_Creating_ReturnsUpToDate verifies that "creating" state returns
// UpToDate=true without calling isUpToDate (prevents spurious Update).
func TestObserve_Creating_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	var listTagsCalled bool
	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return redisClusterResponse(testClusterID, testClusterARN, "creating", nil), nil
		},
		listTagsFn: func(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
			listTagsCalled = true
			return &awselasticache.ListTagsForResourceOutput{}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for creating state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for creating state (prevent spurious Update)")
	}
	if listTagsCalled {
		t.Error("ListTagsForResource should not be called during 'creating' state")
	}
}

// TestObserve_Modifying_ReturnsUpToDate verifies that "modifying" state returns
// UpToDate=true without calling isUpToDate.
func TestObserve_Modifying_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	var listTagsCalled bool
	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return redisClusterResponse(testClusterID, testClusterARN, "modifying", nil), nil
		},
		listTagsFn: func(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
			listTagsCalled = true
			return &awselasticache.ListTagsForResourceOutput{}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for modifying state")
	}
	if listTagsCalled {
		t.Error("ListTagsForResource should not be called during 'modifying' state")
	}
}

// TestObserve_Snapshotting_ReturnsUpToDate verifies "snapshotting" state.
func TestObserve_Snapshotting_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return redisClusterResponse(testClusterID, testClusterARN, "snapshotting", nil), nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for snapshotting state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for snapshotting state")
	}
}

// TestObserve_RebootingClusterNodes_ReturnsUpToDate verifies "rebooting cluster nodes" state.
func TestObserve_RebootingClusterNodes_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return redisClusterResponse(testClusterID, testClusterARN, "rebooting cluster nodes", nil), nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for 'rebooting cluster nodes' state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for 'rebooting cluster nodes' state")
	}
}

// TestObserve_Deleting_ReturnsUpToDate verifies "deleting" state.
func TestObserve_Deleting_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return redisClusterResponse(testClusterID, testClusterARN, "deleting", nil), nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for deleting state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for deleting state")
	}
}

func TestObserve_DescribeError(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return nil, errors.New("unexpected AWS error")
		},
	}}

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Observe when describe fails")
	}
}

// TestObserve_Available_UpToDate verifies that an available cluster with matching
// spec is reported as up-to-date and atProvider is populated.
func TestObserve_Available_UpToDate(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("redis")
	cr.Spec.ForProvider.NodeType = aws.String("cache.t3.micro")

	nodes := []ectypes.CacheNode{
		{
			Endpoint: &ectypes.Endpoint{
				Address: aws.String("my-cluster.abc.0001.use1.cache.amazonaws.com"),
				Port:    aws.Int32(6379),
			},
		},
	}

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return &awselasticache.DescribeCacheClustersOutput{
				CacheClusters: []ectypes.CacheCluster{
					{
						CacheClusterId:     aws.String(testClusterID),
						ARN:                aws.String(testClusterARN),
						CacheClusterStatus: aws.String("available"),
						Engine:             aws.String("redis"),
						CacheNodeType:      aws.String("cache.t3.micro"),
						CacheNodes:         nodes,
					},
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}

	// atProvider should be populated.
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testClusterARN {
		t.Errorf("expected atProvider.Arn=%s, got %v", testClusterARN, cr.Status.AtProvider.Arn)
	}
	if cr.Status.AtProvider.CacheClusterStatus == nil || *cr.Status.AtProvider.CacheClusterStatus != "available" {
		t.Errorf("expected atProvider.CacheClusterStatus='available', got %v", cr.Status.AtProvider.CacheClusterStatus)
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_Success_Redis(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("redis")
	cr.Spec.ForProvider.NodeType = aws.String("cache.t3.micro")
	numNodes := float64(1)
	cr.Spec.ForProvider.NumCacheNodes = &numNodes

	var gotClusterID string
	var gotEngine string
	var gotNodeType string

	e := &ExternalClient{Client: &mockClusterClient{
		createFn: func(_ context.Context, params *awselasticache.CreateCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheClusterOutput, error) {
			gotClusterID = aws.ToString(params.CacheClusterId)
			gotEngine = aws.ToString(params.Engine)
			gotNodeType = aws.ToString(params.CacheNodeType)
			return &awselasticache.CreateCacheClusterOutput{
				CacheCluster: &ectypes.CacheCluster{
					CacheClusterId: params.CacheClusterId,
					Engine:         params.Engine,
				},
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotClusterID != testClusterID {
		t.Errorf("expected CacheClusterId=%s, got %s", testClusterID, gotClusterID)
	}
	if gotEngine != "redis" {
		t.Errorf("expected Engine=redis, got %s", gotEngine)
	}
	if gotNodeType != "cache.t3.micro" {
		t.Errorf("expected CacheNodeType=cache.t3.micro, got %s", gotNodeType)
	}
}

func TestCreate_Success_Memcached(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("memcached")
	cr.Spec.ForProvider.NodeType = aws.String("cache.t3.micro")
	numNodes := float64(2)
	cr.Spec.ForProvider.NumCacheNodes = &numNodes

	var gotClusterID string
	var gotNumNodes int32

	e := &ExternalClient{Client: &mockClusterClient{
		createFn: func(_ context.Context, params *awselasticache.CreateCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheClusterOutput, error) {
			gotClusterID = aws.ToString(params.CacheClusterId)
			if params.NumCacheNodes != nil {
				gotNumNodes = *params.NumCacheNodes
			}
			return &awselasticache.CreateCacheClusterOutput{
				CacheCluster: &ectypes.CacheCluster{
					CacheClusterId: params.CacheClusterId,
				},
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotClusterID != testClusterID {
		t.Errorf("expected CacheClusterId=%s, got %s", testClusterID, gotClusterID)
	}
	if gotNumNodes != 2 {
		t.Errorf("expected NumCacheNodes=2, got %d", gotNumNodes)
	}
}

func TestCreate_Error(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		createFn: func(_ context.Context, _ *awselasticache.CreateCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheClusterOutput, error) {
			return nil, errors.New("create failed")
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_FieldChange(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.NodeType = aws.String("cache.m5.large")
	cr.Status.AtProvider.Arn = aws.String(testClusterARN)

	var gotNodeType string

	e := &ExternalClient{Client: &mockClusterClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheClusterOutput, error) {
			gotNodeType = aws.ToString(params.CacheNodeType)
			return &awselasticache.ModifyCacheClusterOutput{
				CacheCluster: &ectypes.CacheCluster{
					CacheClusterId: aws.String(testClusterID),
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotNodeType != "cache.m5.large" {
		t.Errorf("expected CacheNodeType=cache.m5.large, got %s", gotNodeType)
	}
}

func TestUpdate_Error(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheClusterOutput, error) {
			return nil, errors.New("modify failed")
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Update")
	}
}

func TestUpdate_TagSync(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Tags = map[string]*string{"env": aws.String("prod")}
	cr.Status.AtProvider.Arn = aws.String(testClusterARN)

	var addTagsCalled bool
	e := &ExternalClient{Client: &mockClusterClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheClusterOutput, error) {
			return &awselasticache.ModifyCacheClusterOutput{CacheCluster: &ectypes.CacheCluster{
				CacheClusterId: aws.String(testClusterID),
			}}, nil
		},
		listTagsFn: noopListTags, // no existing tags
		addTagsFn: func(_ context.Context, _ *awselasticache.AddTagsToResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
			addTagsCalled = true
			return &awselasticache.AddTagsToResourceOutput{}, nil
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !addTagsCalled {
		t.Error("expected AddTagsToResource to be called when tags differ")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		deleteFn: func(_ context.Context, params *awselasticache.DeleteCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheClusterOutput, error) {
			return &awselasticache.DeleteCacheClusterOutput{}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDelete_NotFound_Idempotent verifies that CacheClusterNotFoundFault is treated
// as success (idempotent delete).
func TestDelete_NotFound_Idempotent(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheClusterOutput, error) {
			return nil, &ectypes.CacheClusterNotFoundFault{Message: aws.String("not found")}
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for CacheClusterNotFoundFault (idempotent delete), got: %v", err)
	}
}

func TestDelete_OtherError(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	e := &ExternalClient{Client: &mockClusterClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteCacheClusterInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheClusterOutput, error) {
			return nil, errors.New("some AWS error")
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Delete when AWS returns unexpected error")
	}
}

// ── Engine-based connection detail discrimination ─────────────────────────────

// TestObserve_Valkey_Engine_ConnectionDetails verifies that a "valkey" engine
// cluster behaves like Redis (port only, no cluster_address).
func TestObserve_Valkey_Engine_ConnectionDetails(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("valkey")

	nodes := []ectypes.CacheNode{
		{
			Endpoint: &ectypes.Endpoint{
				Address: aws.String("my-cluster.abc.0001.use1.cache.amazonaws.com"),
				Port:    aws.Int32(6379),
			},
		},
	}

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return &awselasticache.DescribeCacheClustersOutput{
				CacheClusters: []ectypes.CacheCluster{
					{
						CacheClusterId:     aws.String(testClusterID),
						ARN:                aws.String(testClusterARN),
						CacheClusterStatus: aws.String("available"),
						Engine:             aws.String("valkey"),
						CacheNodes:         nodes,
					},
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// port must be published.
	if _, ok := obs.ConnectionDetails["port"]; !ok {
		t.Error("expected 'port' in connection details for Valkey cluster")
	}

	// cluster_address must NOT be published for Valkey.
	if _, ok := obs.ConnectionDetails["cluster_address"]; ok {
		t.Error("'cluster_address' must NOT be in connection details for Valkey cluster")
	}
}

// ── Late-initialization tests ──────────────────────────────────────────────────

// TestObserve_LateInit_NilNodeType verifies that when spec.NodeType is nil and
// AWS returns a CacheNodeType, Observe returns ResourceLateInitialized=true and
// populates the spec field — preventing the infinite reconciliation loop that
// would occur when a cluster is linked to a replication group (inheriting the node type).
func TestObserve_LateInit_NilNodeType(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	// Simulate a cluster linked to a replication group: no NodeType in spec.
	cr.Spec.ForProvider.NodeType = nil

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return &awselasticache.DescribeCacheClustersOutput{
				CacheClusters: []ectypes.CacheCluster{
					{
						CacheClusterId:     aws.String(testClusterID),
						ARN:                aws.String(testClusterARN),
						CacheClusterStatus: aws.String("available"),
						Engine:             aws.String("redis"),
						CacheNodeType:      aws.String("cache.r7g.medium"),
						NumCacheNodes:      aws.Int32(1),
					},
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must signal late initialization so the controller saves the spec.
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when NodeType is nil and AWS returns a value")
	}

	// spec.NodeType must be populated from the AWS response.
	if cr.Spec.ForProvider.NodeType == nil {
		t.Fatal("expected spec.NodeType to be populated after late initialization")
	}
	if got, want := *cr.Spec.ForProvider.NodeType, "cache.r7g.medium"; got != want {
		t.Errorf("spec.NodeType: got %q, want %q", got, want)
	}
}

// TestObserve_NoLateInit_NodeTypeAlreadySet verifies that when all AWS-defaulted
// spec fields are already populated, Observe does NOT return ResourceLateInitialized=true.
func TestObserve_NoLateInit_NodeTypeAlreadySet(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	// Fully populate all fields that lateInitializeCluster would fill in.
	cr.Spec.ForProvider.NodeType = aws.String("cache.t3.micro")
	cr.Spec.ForProvider.EngineVersion = aws.String("7.0.7")
	cr.Spec.ForProvider.MaintenanceWindow = aws.String("sun:05:00-sun:06:00")
	cr.Spec.ForProvider.SnapshotWindow = aws.String("03:00-04:00")
	snapshotLimit := float64(1)
	cr.Spec.ForProvider.SnapshotRetentionLimit = &snapshotLimit
	numNodes := float64(1)
	cr.Spec.ForProvider.NumCacheNodes = &numNodes

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return &awselasticache.DescribeCacheClustersOutput{
				CacheClusters: []ectypes.CacheCluster{
					{
						CacheClusterId:             aws.String(testClusterID),
						ARN:                        aws.String(testClusterARN),
						CacheClusterStatus:         aws.String("available"),
						Engine:                     aws.String("redis"),
						CacheNodeType:              aws.String("cache.t3.micro"),
						EngineVersion:              aws.String("7.0.7"),
						PreferredMaintenanceWindow: aws.String("sun:05:00-sun:06:00"),
						SnapshotWindow:             aws.String("03:00-04:00"),
						SnapshotRetentionLimit:     aws.Int32(1),
						NumCacheNodes:              aws.Int32(1),
					},
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must NOT trigger late initialization when all fields are already set.
	if obs.ResourceLateInitialized {
		t.Error("unexpected ResourceLateInitialized=true when all spec fields are already populated")
	}
}

// TestIsUpToDate_NilNodeType_DoesNotTriggerUpdate verifies that when
// spec.NodeType is nil (cluster linked to a replication group), isUpToDate
// does not return false for the NodeType field — preventing the infinite
// reconciliation loop described in the ticket.
func TestIsUpToDate_NilNodeType_DoesNotTriggerUpdate(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:   aws.String("us-east-1"),
		Engine:   aws.String("redis"),
		NodeType: nil, // intentionally nil
	}
	cc := ectypes.CacheCluster{
		CacheNodeType: aws.String("cache.r7g.medium"),
		NumCacheNodes: aws.Int32(1),
	}

	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when spec.NodeType is nil — no spurious update")
	}
}

// ── isUpToDate mutable-field tests ────────────────────────────────────────────
// Each test verifies that isUpToDate returns false when a single mutable field
// differs between spec and AWS state, and true when they match.

// TestIsUpToDate_MaintenanceWindow_Changed verifies that isUpToDate detects drift
// in MaintenanceWindow and returns false.
func TestIsUpToDate_MaintenanceWindow_Changed(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:            aws.String("us-east-1"),
		MaintenanceWindow: aws.String("mon:05:00-mon:06:00"),
	}
	cc := ectypes.CacheCluster{
		PreferredMaintenanceWindow: aws.String("sun:05:00-sun:06:00"), // different
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when MaintenanceWindow differs")
	}
}

// TestIsUpToDate_MaintenanceWindow_Same verifies that isUpToDate returns true
// when MaintenanceWindow matches.
func TestIsUpToDate_MaintenanceWindow_Same(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:            aws.String("us-east-1"),
		MaintenanceWindow: aws.String("sun:05:00-sun:06:00"),
	}
	cc := ectypes.CacheCluster{
		PreferredMaintenanceWindow: aws.String("sun:05:00-sun:06:00"), // same
	}
	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when MaintenanceWindow matches")
	}
}

// TestIsUpToDate_SnapshotRetentionLimit_Changed verifies drift detection for
// SnapshotRetentionLimit.
func TestIsUpToDate_SnapshotRetentionLimit_Changed(t *testing.T) {
	srl := float64(7)
	spec := &clusternative.ClusterRAWParameters{
		Region:                 aws.String("us-east-1"),
		SnapshotRetentionLimit: &srl,
	}
	cc := ectypes.CacheCluster{
		SnapshotRetentionLimit: aws.Int32(3), // different
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when SnapshotRetentionLimit differs")
	}
}

// TestIsUpToDate_SnapshotRetentionLimit_Same verifies isUpToDate returns true
// when SnapshotRetentionLimit matches.
func TestIsUpToDate_SnapshotRetentionLimit_Same(t *testing.T) {
	srl := float64(7)
	spec := &clusternative.ClusterRAWParameters{
		Region:                 aws.String("us-east-1"),
		SnapshotRetentionLimit: &srl,
	}
	cc := ectypes.CacheCluster{
		SnapshotRetentionLimit: aws.Int32(7), // same
	}
	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when SnapshotRetentionLimit matches")
	}
}

// TestIsUpToDate_SnapshotWindow_Changed verifies drift detection for SnapshotWindow.
func TestIsUpToDate_SnapshotWindow_Changed(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:         aws.String("us-east-1"),
		SnapshotWindow: aws.String("03:00-04:00"),
	}
	cc := ectypes.CacheCluster{
		SnapshotWindow: aws.String("05:00-06:00"), // different
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when SnapshotWindow differs")
	}
}

// TestIsUpToDate_SnapshotWindow_Same verifies isUpToDate returns true
// when SnapshotWindow matches.
func TestIsUpToDate_SnapshotWindow_Same(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:         aws.String("us-east-1"),
		SnapshotWindow: aws.String("03:00-04:00"),
	}
	cc := ectypes.CacheCluster{
		SnapshotWindow: aws.String("03:00-04:00"), // same
	}
	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when SnapshotWindow matches")
	}
}

// TestIsUpToDate_AutoMinorVersionUpgrade_Changed verifies drift detection for
// AutoMinorVersionUpgrade (*string "true"/"false" in spec, *bool in AWS).
func TestIsUpToDate_AutoMinorVersionUpgrade_Changed(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:                  aws.String("us-east-1"),
		AutoMinorVersionUpgrade: aws.String("true"),
	}
	cc := ectypes.CacheCluster{
		AutoMinorVersionUpgrade: aws.Bool(false), // different
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when AutoMinorVersionUpgrade differs")
	}
}

// TestIsUpToDate_AutoMinorVersionUpgrade_Same verifies isUpToDate returns true
// when AutoMinorVersionUpgrade matches.
func TestIsUpToDate_AutoMinorVersionUpgrade_Same(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:                  aws.String("us-east-1"),
		AutoMinorVersionUpgrade: aws.String("true"),
	}
	cc := ectypes.CacheCluster{
		AutoMinorVersionUpgrade: aws.Bool(true), // same
	}
	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when AutoMinorVersionUpgrade matches")
	}
}

// TestIsUpToDate_NotificationTopicArn_Changed verifies drift detection for
// NotificationTopicArn.
func TestIsUpToDate_NotificationTopicArn_Changed(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:               aws.String("us-east-1"),
		NotificationTopicArn: aws.String("arn:aws:sns:us-east-1:123456789012:new-topic"),
	}
	cc := ectypes.CacheCluster{
		NotificationConfiguration: &ectypes.NotificationConfiguration{
			TopicArn: aws.String("arn:aws:sns:us-east-1:123456789012:old-topic"), // different
		},
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when NotificationTopicArn differs")
	}
}

// TestIsUpToDate_NotificationTopicArn_Same verifies isUpToDate returns true
// when NotificationTopicArn matches.
func TestIsUpToDate_NotificationTopicArn_Same(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:               aws.String("us-east-1"),
		NotificationTopicArn: aws.String("arn:aws:sns:us-east-1:123456789012:my-topic"),
	}
	cc := ectypes.CacheCluster{
		NotificationConfiguration: &ectypes.NotificationConfiguration{
			TopicArn: aws.String("arn:aws:sns:us-east-1:123456789012:my-topic"), // same
		},
	}
	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when NotificationTopicArn matches")
	}
}

// TestIsUpToDate_ParameterGroupName_Changed verifies drift detection for
// ParameterGroupName.
func TestIsUpToDate_ParameterGroupName_Changed(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:             aws.String("us-east-1"),
		ParameterGroupName: aws.String("my-param-group-v2"),
	}
	cc := ectypes.CacheCluster{
		CacheParameterGroup: &ectypes.CacheParameterGroupStatus{
			CacheParameterGroupName: aws.String("my-param-group-v1"), // different
		},
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when ParameterGroupName differs")
	}
}

// TestIsUpToDate_ParameterGroupName_Same verifies isUpToDate returns true
// when ParameterGroupName matches.
func TestIsUpToDate_ParameterGroupName_Same(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region:             aws.String("us-east-1"),
		ParameterGroupName: aws.String("my-param-group"),
	}
	cc := ectypes.CacheCluster{
		CacheParameterGroup: &ectypes.CacheParameterGroupStatus{
			CacheParameterGroupName: aws.String("my-param-group"), // same
		},
	}
	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when ParameterGroupName matches")
	}
}

// TestIsUpToDate_LogDeliveryConfiguration_Changed verifies drift detection for
// LogDeliveryConfiguration when log type count changes.
func TestIsUpToDate_LogDeliveryConfiguration_Changed(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region: aws.String("us-east-1"),
		LogDeliveryConfiguration: []clusternative.ClusterLogDeliveryConfigurationRAWParameters{
			{
				LogType:         aws.String("slow-log"),
				LogFormat:       aws.String("json"),
				DestinationType: aws.String("cloudwatch-logs"),
				Destination:     aws.String("/aws/elasticache/cluster/slow-log"),
			},
		},
	}
	// AWS has no log delivery configurations.
	cc := ectypes.CacheCluster{
		LogDeliveryConfigurations: []ectypes.LogDeliveryConfiguration{}, // different (empty)
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when LogDeliveryConfiguration count differs")
	}
}

// TestIsUpToDate_LogDeliveryConfiguration_DestinationChanged verifies drift
// detection when the log destination changes.
func TestIsUpToDate_LogDeliveryConfiguration_DestinationChanged(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region: aws.String("us-east-1"),
		LogDeliveryConfiguration: []clusternative.ClusterLogDeliveryConfigurationRAWParameters{
			{
				LogType:         aws.String("slow-log"),
				LogFormat:       aws.String("json"),
				DestinationType: aws.String("cloudwatch-logs"),
				Destination:     aws.String("/aws/elasticache/cluster/slow-log-v2"),
			},
		},
	}
	cc := ectypes.CacheCluster{
		LogDeliveryConfigurations: []ectypes.LogDeliveryConfiguration{
			{
				LogType:         ectypes.LogTypeSlowLog,
				LogFormat:       ectypes.LogFormatJson,
				DestinationType: ectypes.DestinationTypeCloudWatchLogs,
				DestinationDetails: &ectypes.DestinationDetails{
					CloudWatchLogsDetails: &ectypes.CloudWatchLogsDestinationDetails{
						LogGroup: aws.String("/aws/elasticache/cluster/slow-log-v1"), // different
					},
				},
			},
		},
	}
	if isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return false when log destination differs")
	}
}

// TestIsUpToDate_LogDeliveryConfiguration_Same verifies isUpToDate returns true
// when log delivery configuration matches.
func TestIsUpToDate_LogDeliveryConfiguration_Same(t *testing.T) {
	spec := &clusternative.ClusterRAWParameters{
		Region: aws.String("us-east-1"),
		LogDeliveryConfiguration: []clusternative.ClusterLogDeliveryConfigurationRAWParameters{
			{
				LogType:         aws.String("slow-log"),
				LogFormat:       aws.String("json"),
				DestinationType: aws.String("cloudwatch-logs"),
				Destination:     aws.String("/aws/elasticache/cluster/slow-log"),
			},
		},
	}
	cc := ectypes.CacheCluster{
		LogDeliveryConfigurations: []ectypes.LogDeliveryConfiguration{
			{
				LogType:         ectypes.LogTypeSlowLog,
				LogFormat:       ectypes.LogFormatJson,
				DestinationType: ectypes.DestinationTypeCloudWatchLogs,
				DestinationDetails: &ectypes.DestinationDetails{
					CloudWatchLogsDetails: &ectypes.CloudWatchLogsDestinationDetails{
						LogGroup: aws.String("/aws/elasticache/cluster/slow-log"), // same
					},
				},
			},
		},
	}
	if !isUpToDate(spec, cc, nil) {
		t.Error("isUpToDate should return true when LogDeliveryConfiguration matches")
	}
}

// ── setAtProviderFromCluster observation field tests ──────────────────────────

// TestSetAtProvider_ObservationFields_Redis verifies that all observation fields
// (Port, MaintenanceWindow, SnapshotRetentionLimit, SnapshotWindow,
// TransitEncryptionEnabled, AvailabilityZone, AutoMinorVersionUpgrade) are
// populated correctly for a Redis cluster where ConfigurationEndpoint is nil
// and Port comes from CacheNodes[0].Endpoint.Port.
func TestSetAtProvider_ObservationFields_Redis(t *testing.T) {
	cr := newTestCR("my-redis", testClusterID)

	cc := ectypes.CacheCluster{
		CacheClusterId:             aws.String(testClusterID),
		ARN:                        aws.String(testClusterARN),
		CacheClusterStatus:         aws.String("available"),
		Engine:                     aws.String("redis"),
		PreferredMaintenanceWindow: aws.String("sun:05:00-sun:06:00"),
		SnapshotRetentionLimit:     aws.Int32(7),
		SnapshotWindow:             aws.String("03:00-04:00"),
		TransitEncryptionEnabled:   aws.Bool(true),
		PreferredAvailabilityZone:  aws.String("us-east-1a"),
		AutoMinorVersionUpgrade:    aws.Bool(true),
		CacheNodes: []ectypes.CacheNode{
			{
				Endpoint: &ectypes.Endpoint{
					Address: aws.String("my-cluster.abc.0001.use1.cache.amazonaws.com"),
					Port:    aws.Int32(6379),
				},
			},
		},
	}

	setAtProviderFromCluster(cr, cc, nil)
	obs := cr.GetAtProvider()

	// Port from CacheNodes[0].Endpoint.Port for Redis.
	if obs.Port == nil || *obs.Port != 6379 {
		t.Errorf("expected Port=6379, got %v", obs.Port)
	}
	// MaintenanceWindow.
	if obs.MaintenanceWindow == nil || *obs.MaintenanceWindow != "sun:05:00-sun:06:00" {
		t.Errorf("expected MaintenanceWindow='sun:05:00-sun:06:00', got %v", obs.MaintenanceWindow)
	}
	// SnapshotRetentionLimit: *int32(7) → *float64(7).
	if obs.SnapshotRetentionLimit == nil || *obs.SnapshotRetentionLimit != 7.0 {
		t.Errorf("expected SnapshotRetentionLimit=7, got %v", obs.SnapshotRetentionLimit)
	}
	// SnapshotWindow.
	if obs.SnapshotWindow == nil || *obs.SnapshotWindow != "03:00-04:00" {
		t.Errorf("expected SnapshotWindow='03:00-04:00', got %v", obs.SnapshotWindow)
	}
	// TransitEncryptionEnabled.
	if obs.TransitEncryptionEnabled == nil || !*obs.TransitEncryptionEnabled {
		t.Errorf("expected TransitEncryptionEnabled=true, got %v", obs.TransitEncryptionEnabled)
	}
	// AvailabilityZone from PreferredAvailabilityZone.
	if obs.AvailabilityZone == nil || *obs.AvailabilityZone != "us-east-1a" {
		t.Errorf("expected AvailabilityZone='us-east-1a', got %v", obs.AvailabilityZone)
	}
	// AutoMinorVersionUpgrade: *bool(true) → *string("true").
	if obs.AutoMinorVersionUpgrade == nil || *obs.AutoMinorVersionUpgrade != "true" {
		t.Errorf("expected AutoMinorVersionUpgrade='true', got %v", obs.AutoMinorVersionUpgrade)
	}
	// ClusterAddress must NOT be set for Redis (ConfigurationEndpoint is nil).
	if obs.ClusterAddress != nil {
		t.Errorf("expected ClusterAddress=nil for Redis, got %v", obs.ClusterAddress)
	}
	// ConfigurationEndpoint must NOT be set for Redis.
	if obs.ConfigurationEndpoint != nil {
		t.Errorf("expected ConfigurationEndpoint=nil for Redis, got %v", obs.ConfigurationEndpoint)
	}
}

// TestSetAtProvider_ObservationFields_Memcached verifies that Port, ClusterAddress,
// and ConfigurationEndpoint are set correctly from ConfigurationEndpoint for a
// Memcached cluster, plus AutoMinorVersionUpgrade=false → "false".
func TestSetAtProvider_ObservationFields_Memcached(t *testing.T) {
	cr := newTestCR("my-memcached", "my-memcached")
	cr.Spec.ForProvider.Engine = aws.String("memcached")

	cc := ectypes.CacheCluster{
		CacheClusterId:     aws.String("my-memcached"),
		ARN:                aws.String(testClusterARN),
		CacheClusterStatus: aws.String("available"),
		Engine:             aws.String("memcached"),
		ConfigurationEndpoint: &ectypes.Endpoint{
			Address: aws.String("my-cluster.cfg.usw2.cache.amazonaws.com"),
			Port:    aws.Int32(11211),
		},
		PreferredMaintenanceWindow: aws.String("mon:02:00-mon:03:00"),
		AutoMinorVersionUpgrade:    aws.Bool(false),
	}

	setAtProviderFromCluster(cr, cc, nil)
	obs := cr.GetAtProvider()

	// Port from ConfigurationEndpoint.Port for Memcached.
	if obs.Port == nil || *obs.Port != 11211 {
		t.Errorf("expected Port=11211, got %v", obs.Port)
	}
	// ClusterAddress from ConfigurationEndpoint.Address.
	if obs.ClusterAddress == nil || *obs.ClusterAddress != "my-cluster.cfg.usw2.cache.amazonaws.com" {
		t.Errorf("expected ClusterAddress='my-cluster.cfg.usw2.cache.amazonaws.com', got %v", obs.ClusterAddress)
	}
	// ConfigurationEndpoint in "address:port" format.
	want := "my-cluster.cfg.usw2.cache.amazonaws.com:11211"
	if obs.ConfigurationEndpoint == nil || *obs.ConfigurationEndpoint != want {
		t.Errorf("expected ConfigurationEndpoint=%q, got %v", want, obs.ConfigurationEndpoint)
	}
	// MaintenanceWindow.
	if obs.MaintenanceWindow == nil || *obs.MaintenanceWindow != "mon:02:00-mon:03:00" {
		t.Errorf("expected MaintenanceWindow='mon:02:00-mon:03:00', got %v", obs.MaintenanceWindow)
	}
	// AutoMinorVersionUpgrade: *bool(false) → *string("false").
	if obs.AutoMinorVersionUpgrade == nil || *obs.AutoMinorVersionUpgrade != "false" {
		t.Errorf("expected AutoMinorVersionUpgrade='false', got %v", obs.AutoMinorVersionUpgrade)
	}
}

// TestSetAtProvider_NilOptionalFields verifies that nil optional fields in the
// AWS response (AutoMinorVersionUpgrade, TransitEncryptionEnabled, etc.) result
// in nil observation fields — no nil-pointer panics.
func TestSetAtProvider_NilOptionalFields(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)

	cc := ectypes.CacheCluster{
		CacheClusterId:     aws.String(testClusterID),
		ARN:                aws.String(testClusterARN),
		CacheClusterStatus: aws.String("available"),
		// All optional fields nil.
	}

	// Must not panic.
	setAtProviderFromCluster(cr, cc, nil)
	obs := cr.GetAtProvider()

	if obs.Port != nil {
		t.Errorf("expected nil Port when CacheNodes is empty and ConfigurationEndpoint is nil")
	}
	if obs.AutoMinorVersionUpgrade != nil {
		t.Errorf("expected nil AutoMinorVersionUpgrade when AWS field is nil")
	}
	if obs.TransitEncryptionEnabled != nil {
		t.Errorf("expected nil TransitEncryptionEnabled when AWS field is nil")
	}
	if obs.SnapshotRetentionLimit != nil {
		t.Errorf("expected nil SnapshotRetentionLimit when AWS field is nil")
	}
	if obs.AvailabilityZone != nil {
		t.Errorf("expected nil AvailabilityZone when AWS field is nil")
	}
}

// TestObserve_NilGuard_ConfigEndpoint verifies nil-safety when ConfigurationEndpoint
// is nil (as it is for Redis/Valkey).
func TestObserve_NilGuard_ConfigEndpoint(t *testing.T) {
	cr := newTestCR("my-cluster", testClusterID)
	cr.Spec.ForProvider.Engine = aws.String("redis")

	nodes := []ectypes.CacheNode{
		{Endpoint: &ectypes.Endpoint{Port: aws.Int32(6379)}},
	}

	e := &ExternalClient{Client: &mockClusterClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheClustersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheClustersOutput, error) {
			return &awselasticache.DescribeCacheClustersOutput{
				CacheClusters: []ectypes.CacheCluster{
					{
						CacheClusterId:        aws.String(testClusterID),
						ARN:                   aws.String(testClusterARN),
						CacheClusterStatus:    aws.String("available"),
						Engine:                aws.String("redis"),
						ConfigurationEndpoint: nil, // explicitly nil
						CacheNodes:            nodes,
					},
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	// Should not panic.
	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error with nil ConfigurationEndpoint: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
}
