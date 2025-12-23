package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

type MongoConfig struct {
	mongodriver.MongoDriverConfig
}

type AuthConfig = authorisation.Config

// Config represents service configuration for dis-bundle-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	OTBatchTimeout             time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	DefaultMaxLimit            int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultLimit               int           `envconfig:"DEFAULT_LIMIT"`
	DefaultOffset              int           `envconfig:"DEFAULT_OFFSET"`
	EnablePermissionsAuth      bool          `envconfig:"ENABLE_PERMISSIONS_AUTH"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	ZebedeeClientTimeout       time.Duration `envconfig:"ZEBEDEE_CLIENT_TIMEOUT"`
	MongoConfig
	AuthConfig                                *authorisation.Config
	DataBundlePublicationServiceSlackAPIToken string `envconfig:"DATA_BUNDLE_PUBLICATION_SERVICE_SLACK_API_TOKEN"`
	SlackConfig                               *slack.SlackConfig
}

var cfg *Config

const (
	BundlesCollection        = "BundlesCollection"
	BundleEventsCollection   = "BundleEventsCollection"
	BundleContentsCollection = "BundleContentsCollection"
)

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   ":29800",
		DatasetAPIURL:              "http://localhost:22000",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		OTBatchTimeout:             5 * time.Second,
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dis-bundle-api",
		OtelEnabled:                false,
		DefaultMaxLimit:            1000,
		DefaultLimit:               20,
		DefaultOffset:              0,
		ZebedeeURL:                 "http://localhost:8082",
		ZebedeeClientTimeout:       30 * time.Second,
		MongoConfig: MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint:               "localhost:27017",
				Username:                      "",
				Password:                      "",
				Database:                      "bundles",
				Collections:                   map[string]string{BundlesCollection: "bundles", BundleEventsCollection: "bundle_events", BundleContentsCollection: "bundle_contents"},
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
		AuthConfig: authorisation.NewDefaultConfig(),
		DataBundlePublicationServiceSlackAPIToken: "test-data-bundle-publication-service-slack-api-token",
		SlackConfig: &slack.SlackConfig{},
	}

	return cfg, envconfig.Process("", cfg)
}
