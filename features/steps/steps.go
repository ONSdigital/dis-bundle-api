package steps

import (
	"context"

	"github.com/cucumber/godog"
)

func (c *BundleComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^there are no bundles$`, c.thereAreNoBundles)
}

func (c *BundleComponent) thereAreNoBundles() error {
	return c.MongoClient.Connection.DropDatabase(context.Background())
}
