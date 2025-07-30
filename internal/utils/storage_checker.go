package cobrautil

import (
	"context"
	"fmt"
	"github.com/ceph/go-ceph/rados"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"net"
	"strings"
	"time"
)

type StorageChecker interface {
	Check(ctx context.Context) (bool, error)
	Name() string
}

type S3Checker struct {
	client     *minio.Client
	accessKey  string
	secretKey  string
	bucketName string
	endpoint   string
	timeout    time.Duration
}

func NewS3Checker(endpoint, accessKey, secretKey, bucket string, timeout time.Duration) (StorageChecker, error) {
	useSSL := false
	if IsHTTPS(endpoint) {
		useSSL = true
	}
	endpoint = RemoveHTTPPrefix(endpoint)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client failed, error : %w", err)
	}

	return &S3Checker{
		client:     client,
		accessKey:  accessKey,
		secretKey:  secretKey,
		bucketName: bucket,
		endpoint:   endpoint,
		timeout:    timeout,
	}, nil
}

func (s *S3Checker) Check(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return false, fmt.Errorf("check bucket %s failed, error : %w", s.bucketName, err)
	}
	if !exists {
		return false, fmt.Errorf("bucket %s is not exist", s.bucketName)
	}

	return true, nil
}

func (s *S3Checker) Name() string {
	return fmt.Sprintf("S3(endpoint: %s/%s, ak: %s, sk: %s)", s.endpoint, s.bucketName, s.accessKey, s.secretKey)
}

type RadosChecker struct {
	conn     *rados.Conn
	monHost  string
	user     string
	key      string
	poolName string
	timeout  time.Duration
}

func NewRadosChecker(monHost, user, key, poolName, cluster string, timeout time.Duration) (StorageChecker, error) {
	conn, err := rados.NewConnWithClusterAndUser(cluster, user)
	if err != nil {
		return nil, fmt.Errorf("create rados connection failed: %w", err)
	}

	// set mon host
	if err := conn.SetConfigOption("mon_host", monHost); err != nil {
		return nil, fmt.Errorf("set mon_host failed: %w", err)
	}
	// set client key
	if err := conn.SetConfigOption("key", key); err != nil {
		return nil, fmt.Errorf("set key failed: %w", err)
	}

	return &RadosChecker{
		conn:     conn,
		monHost:  monHost,
		user:     user,
		key:      key,
		poolName: poolName,
		timeout:  timeout,
	}, nil
}

func (r *RadosChecker) HostIsConnected(host string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

func (r *RadosChecker) AllHostsUnavailable() (bool, error) {
	hosts := strings.Split(r.monHost, ",")
	var failedHosts []string

	for _, host := range hosts {
		if !r.HostIsConnected(host, r.timeout) {
			failedHosts = append(failedHosts, host)
		}
	}

	if len(failedHosts) == len(hosts) {
		return true, fmt.Errorf("all hosts are unavailable: %v", failedHosts)
	}

	return false, nil
}

func (r *RadosChecker) Check(ctx context.Context) (bool, error) {
	if ok, err := r.AllHostsUnavailable(); ok {
		return false, err
	}

	if err := r.conn.Connect(); err != nil {
		return false, fmt.Errorf("connect to cluster failed: %w", err)
	}

	// check cluster stats
	_, err := r.conn.GetClusterStats()
	if err != nil {
		return false, fmt.Errorf("get cluster stats failed: %w", err)
	}

	// check pool is available
	ioCtx, err := r.conn.OpenIOContext(r.poolName)
	if err != nil {
		return false, fmt.Errorf("open io context failed (pool %s may not exist): %w", r.poolName, err)
	}
	ioCtx.Destroy()

	return true, nil
}

func (r *RadosChecker) Name() string {
	return fmt.Sprintf("RADOS(pool:%s, user:%s, key:%s)", r.poolName, r.user, r.key)
}

func (r *RadosChecker) Close() error {
	if r.conn != nil {
		r.conn.Shutdown()
	}
	return nil
}
