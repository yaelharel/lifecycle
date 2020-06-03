// +build acceptance

package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	outDir, err := filepath.Abs(analyzerBinaryDir)
	h.AssertNil(t, err)
	buildLifecycle(t, outDir)
	buildTestImage(t, analyzeImage, analyzeDockerContext)
	defer removeTestImage(t, analyzeImage)
	spec.Run(t, "acceptance", testAnalyzer, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testAnalyzer(t *testing.T, when spec.G, it spec.S) {
	it("works", func() {
		// TODO: set args on the cmd
		cmd := exec.Command("docker", "run", "--rm", analyzeImage)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
		}
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
