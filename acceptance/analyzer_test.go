// +build acceptance

package acceptance

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	h "github.com/buildpacks/lifecycle/testhelpers"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	analyzeDockerContext = filepath.Join("testdata", "analyzer")
	analyzerBinaryDir    = filepath.Join("testdata", "analyzer", "container", "cnb", "lifecycle")
	analyzeImage         = "lifecycle/acceptance/analyzer"
)

func TestAnalyzer(t *testing.T) {
	//outDir, err := filepath.Abs(analyzerBinaryDir)
	//h.AssertNil(t, err)
	//buildLifecycle(t, outDir)
	buildTestImage(t, analyzeImage, analyzeDockerContext)
	defer removeTestImage(t, analyzeImage)
	spec.Run(t, "acceptance", testAnalyzer, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testAnalyzer(t *testing.T, when spec.G, it spec.S) {
	when("called without an image", func() {
		it("errors", func() {
			cmd := exec.Command("docker", "run", "--rm", analyzeImage)
			output, err := cmd.CombinedOutput()
			expected := "failed to parse arguments: received 0 arguments, but expected 1"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute provided CMD:\n\t got: %s\n\t want: %s", output, expected)
			}
			h.AssertNotNil(t, err)
		})
	})
	when("cache image tag and cache directory are both blank", func() {
		it("warns", func() {
			cmd := exec.Command("docker", "run", "--rm", analyzeImage, "some-image")
			output, _ := cmd.CombinedOutput() // TODO: provide the correct setup so that we can assert err is nil
			expected := "Not restoring cached layer metadata, no cache flag specified." // TODO: why does this fail if we add "Warning"... color?
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute provided CMD:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})
	when("CNB_USER_ID is not provided", func() {
		it("errors", func() {

		})
	})
	when("CNB_GROUP_ID is not provided", func() {
		it("FOO", func() {

		})
	})
	when("the provided layers directory isn't writeable", func() {
		it("recursively chowns the directory", func() {

		})
	})
	when("the provided cache directory isn't writeable", func() {
		it("recursively chowns the directory", func() {

		})
	})
	it("runs as the provided user", func() {
		// test setup should set CNB_USER_ID and CNB_GROUP_ID in the environment
	})
	when("group path is provided", func() {
		it("uses the provided group path", func() {

		})
	})
	when("cache is provided", func() {
		when("cache image case", func() {
			it("uses the provided cache", func() {

			})
		})
		when("cache directory case", func() {
			it("uses the provided cache", func() {

			})
		})
	})
	when("daemon case", func() {
		it("FOO", func() {

		})
	})
	when("registry case", func() {
		when("auth is required", func() {
			it("FOO", func() {

			})
		})
	})
	it("writes analyzed.toml", func() { // TODO: should this assertion be repeated in different contexts?
	})
	when("analyzed path is provided", func() {
		it("writes analyzed.toml at the provided path", func() {

		})
	})
	when("layers path is provided", func() {
		it("uses the provided layers path", func() {

		})
	})
	when("skip layers is provided", func() {
		it("does not write buildpack layer metadata", func() {

		})
	})
}

// TODO: move buildLifecycle into more central location, or adapt buildBinaries. Also, we should only have to build the binaries once when running acceptance.
func buildLifecycle(t *testing.T, dir string) {
	cmd := exec.Command("make", "build-linux")
	wd, err := os.Getwd()
	h.AssertNil(t, err)
	cmd.Dir = filepath.Join(wd, "..")
	cmd.Env = append(
		os.Environ(),
		"PWD="+cmd.Dir,
		"OUT_DIR="+dir,
		"PLATFORM_API=0.9",
		"LIFECYCLE_VERSION=some-version",
		"SCM_COMMIT=asdf123",
	)

	t.Log("Building binaries: ", cmd.Args)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
	}
}
