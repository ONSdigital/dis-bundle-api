package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/features/steps"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	MongoFeature *componenttest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(godogCtx *godog.ScenarioContext) {
	authorizationFeature := componenttest.NewAuthorizationFeature()
	bundleFeature, err := steps.NewBundleComponent(f.MongoFeature.Server.URI())
	if err != nil {
		panic(err)
	}

	apiFeature := componenttest.NewAPIFeature(bundleFeature.InitialiseService)

	godogCtx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		apiFeature.Reset()
		if err := bundleFeature.Reset(); err != nil {
			panic(err)
		}
		if err := f.MongoFeature.Reset(); err != nil {
			log.Error(context.Background(), "failed to reset mongo feature", err)
		}
		authorizationFeature.Reset()
		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		bundleFeature.Close()
		authorizationFeature.Close()
		return ctx, nil
	})

	bundleFeature.RegisterSteps(godogCtx)
	apiFeature.RegisterSteps(godogCtx)
	authorizationFeature.RegisterSteps(godogCtx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
			Strict: true,
		}

		f := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
