package archive_test

import (
	"runtime"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/buildpacks/lifecycle/archive"
	h "github.com/buildpacks/lifecycle/testhelpers"
)

func TestTarPath(t *testing.T) {
	spec.Run(t, "testTarPath", testTarPath, spec.Report(report.Terminal{}))
}

func testTarPath(t *testing.T, when spec.G, it spec.S) {
	when("#TarPath", func() {
		when("OS is Windows", func() {
			it.Before(func() {
				if runtime.GOOS != "windows" {
					t.Skip("Skipping for non-Windows")
				}
			})

			for path, expected := range map[string]string{
				`c:\`:                `/`,
				`C:\`:                `/`,
				`d:\`:                `/`,
				`c:\foo`:             `/foo//`,
				`c:\foo\bar`:         `/foo/bar`,
				`c:\foo\bar\baz.txt`: `/foo/bar/baz.txt`,
				`foo`:                `foo`,
			} {
				path := path
				expected := expected
				it("removes volume and converts slashes", func() {
					actual := archive.TarPath(path)
					h.AssertEq(t, actual, expected)
				})
			}
		})

		when("OS is POSIX", func() {
			it.Before(func() {
				if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
					t.Skip("Skipping for non-POSIX")
				}
			})

			for path, expected := range map[string]string{
				`/`:         `/`,
				`/foo/bar/`: `/foo/bar/`,
				`foo/bar/`:  `foo/bar/`,
			} {
				path := path
				expected := expected
				it("removes volume and converts slashes", func() {
					actual := archive.TarPath(path)
					h.AssertEq(t, actual, expected)
				})
			}
		})
	})
}
