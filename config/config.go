package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

type MongoConfig struct {
	mongodriver.MongoDriverConfig
}

// Config represents service configuration for dis-bundle-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	OTBatchTimeout             time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	MongoConfig
}

var cfg *Config

const (
	BundlesCollection = "BundlesCollection"
)

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   ":29800",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		OTBatchTimeout:             5 * time.Second,
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dis-bundle-api",
		OtelEnabled:                false,
		MongoConfig: MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint:               "localhost:27017",
				Username:                      "",
				Password:                      "",
				Database:                      "bundles",
				Collections:                   map[string]string{BundlesCollection: "bundles"},
				ReplicaSet:                    "",
				IsStrongReadConcernEnabled:    false,
				IsWriteConcernMajorityEnabled: true,
				ConnectTimeout:                5 * time.Second,
				QueryTimeout:                  15 * time.Second,
				TLSConnectionConfig: mongodriver.TLSConnectionConfig{
					IsSSL: false,
				},
			},
		},
	}

	return cfg, envconfig.Process("", cfg)
}
