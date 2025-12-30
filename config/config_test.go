package config

import (
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {
			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":29800")
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.OTBatchTimeout, ShouldEqual, 5*time.Second)
				So(cfg.OTExporterOTLPEndpoint, ShouldEqual, "localhost:4317")
				So(cfg.OTServiceName, ShouldEqual, "dis-bundle-api")
				So(cfg.OtelEnabled, ShouldBeFalse)
				So(cfg.DefaultMaxLimit, ShouldEqual, 1000)
				So(cfg.DefaultLimit, ShouldEqual, 20)
				So(cfg.DefaultOffset, ShouldEqual, 0)
				So(cfg.EnablePermissionsAuth, ShouldBeFalse)
				So(cfg.ZebedeeURL, ShouldEqual, "http://localhost:8082")
				So(cfg.ZebedeeClientTimeout, ShouldEqual, 30*time.Second)

				So(cfg.ClusterEndpoint, ShouldEqual, "localhost:27017")
				So(cfg.Username, ShouldEqual, "")
				So(cfg.Password, ShouldEqual, "")
				So(cfg.Database, ShouldEqual, "bundles")
				So(cfg.Collections, ShouldResemble, map[string]string{
					BundlesCollection:        "bundles",
					BundleEventsCollection:   "bundle_events",
					BundleContentsCollection: "bundle_contents",
				})
				So(cfg.ReplicaSet, ShouldEqual, "")
				So(cfg.IsStrongReadConcernEnabled, ShouldBeFalse)
				So(cfg.IsWriteConcernMajorityEnabled, ShouldBeTrue)
				So(cfg.ConnectTimeout, ShouldEqual, 5*time.Second)
				So(cfg.QueryTimeout, ShouldEqual, 15*time.Second)
				So(cfg.IsSSL, ShouldBeFalse)

				So(cfg.AuthConfig, ShouldResemble, authorisation.NewDefaultConfig())

				So(cfg.DataBundlePublicationServiceSlackEnabled, ShouldBeFalse)
				So(cfg.DataBundlePublicationServiceSlackAPIToken, ShouldEqual, "test-data-bundle-publication-service-slack-api-token")
				So(cfg.SlackConfig, ShouldResemble, &slack.SlackConfig{})
			})
		})
	})
}
