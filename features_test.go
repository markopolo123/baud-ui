package baudui_test

import (
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs every .feature file under features/ via godog.
// Step definitions live in steps_test.go and assert on rendered templ
// output parsed as HTML — never string-contains.
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog suite failed")
	}
}
