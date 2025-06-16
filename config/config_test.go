package config

import (
	"os"
	"testing"
	"time"

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
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.ClusterEndpoint, ShouldEqual, "localhost:27017")
				So(cfg.Database, ShouldEqual, "bundles")
				So(cfg.Collections, ShouldResemble, map[string]string{"BundlesCollection": "bundles", "BundleEventsCollection": "bundle_events", "BundleContentsCollection": "bundle_contents"})
				So(cfg.Username, ShouldEqual, "")
				So(cfg.Password, ShouldEqual, "")
				So(cfg.IsSSL, ShouldEqual, false)
				So(cfg.QueryTimeout, ShouldEqual, 15*time.Second)
				So(cfg.ConnectTimeout, ShouldEqual, 5*time.Second)
				So(cfg.IsStrongReadConcernEnabled, ShouldEqual, false)
				So(cfg.IsWriteConcernMajorityEnabled, ShouldEqual, true)
			})
		})
	})
}
