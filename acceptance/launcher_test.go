package acceptance

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	h "github.com/buildpacks/lifecycle/testhelpers"
)

var (
	launchDockerContext = filepath.Join("testdata", "launcher", )
	launcherBinaryDir = filepath.Join("acceptance", "testdata", "launcher", "container", "cnb", "lifecycle")
	launchImage            = "lifecycle/acceptance/launcher"
)

func TestLauncher(t *testing.T) {
	buildLauncher(t)
	buildLaunchImage(t)
	spec.Run(t, "acceptance", testLauncher, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testLauncher(t *testing.T, when spec.G, it spec.S) {
	when("no args and no default", func() {
		it("launches web", func() {
			cmd := exec.Command("docker", "run", launchImage)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
			}
			expected := "Executing web process-type"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute web:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})

	when("no args and CNB_PROCESS_TYPE", func() {
		it("launches given default process", func() {
			cmd := exec.Command("docker", "run", "--env", "CNB_PROCESS_TYPE=other-process", launchImage)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
			}
			expected := "Executing other-process process-type"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute other-process:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})

	when("process provided in CMD", func() {
		it("launches process", func() {
			cmd := exec.Command("docker", "run", launchImage, "other-process")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
			}
			expected := "Executing other-process process-type"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute other-process:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})

	when("CMD provided", func() {
		it.Focus("launches command with shell", func() {
			cmd := exec.Command("docker", "run", launchImage, "echo", "my-cmd")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
			}
			expected := "sourced app profile\nmy-cmd"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute provided CMD:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})

	when("CMD provided with --", func() {
		it("launches command directly", func() {
			cmd := exec.Command("docker", "run", launchImage, "echo", "my-cmd")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
			}
			expected := "my-cmd"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute provided CMD:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})

	when("CMD provided with --", func() {
		it("launches command directly", func() {
			cmd := exec.Command("docker", "run", launchImage, "echo", "my-cmd")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
			}
			expected := "my-cmd"
			if !strings.Contains(string(output), expected) {
				t.Fatalf("failed to execute provided CMD:\n\t got: %s\n\t want: %s", output, expected)
			}
		})
	})
}

func buildLaunchImage(t *testing.T) {
	cmd := exec.Command("docker", "build", "-t", launchImage, launchDockerContext)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, string(output), err)
	}
}

func buildLauncher(t *testing.T) {
	cmd := exec.Command("make", "clean", "build-linux-launcher")
	wd, err := os.Getwd()
	h.AssertNil(t, err)
	cmd.Dir = filepath.Join(wd, "..")
	cmd.Env = append(
		os.Environ(),
		"PWD="+cmd.Dir,
		"OUT_DIR="+launcherBinaryDir,
		"LIFECYCLE_VERSION=some-version",
		"SCM_COMMIT=asdf123",
	)

	t.Log("Building binaries: ", cmd.Args)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run %v\n OUTPUT: %s\n ERROR: %s\n", cmd.Args, output, err)
	}
	return
}
