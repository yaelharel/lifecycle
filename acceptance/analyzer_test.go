// +build acceptance

package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	analyzerPath         = "/cnb/lifecycle/analyzer"
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
			cmd := exec.Command("docker", "run", "--rm", analyzeImage, analyzerPath)
			output, err := cmd.CombinedOutput()
			expected := "failed to parse arguments: received 0 arguments, but expected 1"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to receive expected output:\n\t got: %s\n\t want: %s", output, expected)
			}
			h.AssertNotNil(t, err)
		})
	})
	when("cache image tag and cache directory are both blank", func() {
		it("warns", func() {
			cmd := exec.Command("docker", "run", "--rm", analyzeImage, analyzerPath, "some-image")
			output, err := cmd.CombinedOutput()
			h.AssertNil(t, err)
			expected := "Not restoring cached layer metadata, no cache flag specified." // TODO: why does this fail if we add "Warning"... color?
			if !strings.Contains(string(output), expected) { // TODO: make test helper to reduce duplication.
				t.Fatalf("failed to receive expected output:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})
	when("CNB_USER_ID is not provided", func() {
		it("defaults to 0", func() {
			// TODO: not sure how to demonstrate this. Maybe it's not necessary.
		})
	})
	when("CNB_GROUP_ID is not provided", func() {
		it("defaults to 0", func() {

		})
	})
	when("the provided layers directory isn't writeable", func() {
		it("recursively chowns the directory", func() { // TODO: there is some subtlety around canWrite() that is likely not covered here...
			cmd := exec.Command(
				"docker",
				"run",
				"--rm",
				analyzeImage,
				"/bin/bash",
				"-c",
				// TODO: if CNB_USER_ID and CNB_GROUP_ID are provided, CNB_REGISTRY_AUTH must also be provided or we will fail to stat /root/.docker/config.json. This seems brittle...
				fmt.Sprintf("CNB_USER_ID=%s CNB_GROUP_ID=%s CNB_REGISTRY_AUTH={} %s some-image; ls -al /layers", "2222", "3333", analyzerPath),
			)
			output, err := cmd.CombinedOutput()
			h.AssertNil(t, err)
			fileChowned, err := regexp.MatchString("2222 3333 .+ analyzed.toml", string(output))
			h.AssertNil(t, err)
			h.AssertEq(t, fileChowned, true)
			dirChowned, err := regexp.MatchString("2222 3333 .+ \\.", string(output))
			h.AssertNil(t, err)
			h.AssertEq(t, dirChowned, true)
		})
	})
	when("the provided cache directory isn't writeable", func() {
		it("recursively chowns the directory", func() {

		})
	})
	it("runs as the provided user", func() {
		// TODO: not sure how to demonstrate this... it seems important though.
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
