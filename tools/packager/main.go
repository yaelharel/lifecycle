package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/buildpacks/lifecycle/archive"
)

var (
	archivePath    string
	descriptorPath string
	inputDir       string
	version        string
)

// Write contents of inputDir to archive at archivePath
func main() {
	flag.StringVar(&archivePath, "archivePath", "", "path to output")
	flag.StringVar(&descriptorPath, "descriptorPath", "", "path to lifecycle descriptor file")
	flag.StringVar(&inputDir, "inputDir", "", "dir to create package from")
	flag.StringVar(&version, "version", "", "lifecycle version")

	flag.Parse()
	if archivePath == "" || inputDir == "" || version == "" {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.OpenFile(archivePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		fmt.Printf("Failed to open -archivePath %s: %s", archivePath, err)
		os.Exit(2)
	}
	defer f.Close()
	zw := gzip.NewWriter(f)
	defer zw.Close()
	tw := archive.NewNormalizingTarWriter(tar.NewWriter(zw))
	tw.WithUID(0)
	tw.WithGID(0)
	defer tw.Close()

	templateContents, err := ioutil.ReadFile(descriptorPath)
	if err != nil {
		fmt.Printf("Failed read descriptor file at path %s: %s", descriptorPath, err)
		os.Exit(3)
	}

	descriptorContents, err := fillTemplate(templateContents, map[string]interface{}{"lifecycle_version": version})
	if err != nil {
		fmt.Printf("Failed fill template: %s", err)
		os.Exit(4)
	}

	descriptorInfo, err := os.Stat(descriptorPath)
	if err != nil {
		fmt.Printf("Failed stat descriptor file at path %s: %s", descriptorPath, err)
		os.Exit(5)
	}

	tempDir, err := ioutil.TempDir("", "lifecycle-descriptor")
	if err != nil {
		fmt.Printf("Failed to create a temp directory: %s", err)
		os.Exit(6)
	}

	tempFile, err := os.Create(filepath.Join(tempDir, "lifecycle.toml"))
	if err != nil {
		fmt.Printf("Failed create a temp file: %s", err)
		os.Exit(7)
	}

	if err := ioutil.WriteFile(tempFile.Name(), descriptorContents, descriptorInfo.Mode()); err != nil {
		fmt.Printf("Failed to write descriptor contents to tempFile %s: %s", tempFile.Name(), err)
		os.Exit(8)
	}

	if err := os.Chdir(tempDir); err != nil {
		fmt.Printf("Failed to switch directories to %s: %s", filepath.Dir(tempDir), err)
		os.Exit(9)
	}

	descriptorInfo, err = os.Stat(tempFile.Name())
	if err != nil {
		fmt.Printf("Failed stat descriptor file at path %s: %s", tempFile.Name(), err)
		os.Exit(10000)
	}

	if err := archive.AddFileToArchive(tw, "lifecycle.toml", descriptorInfo); err != nil {
		fmt.Printf("Failed to write descriptor to archive: %s\ntempDir: %s\ntempFile:%s\n", err, tempDir, tempFile.Name())
		os.Exit(10)
	}

	if err := os.Chdir(filepath.Dir(inputDir)); err != nil {
		fmt.Printf("Failed to switch directories to %s: %s", filepath.Dir(inputDir), err)
		os.Exit(11)
	}

	if err := archive.AddDirToArchive(tw, filepath.Base(inputDir)); err != nil {
		fmt.Printf("Failed to write dir to archive: %s", err)
		os.Exit(12)
	}
}

func fillTemplate(templateContents []byte, data map[string]interface{}) ([]byte, error) {
	tpl, err := template.New("").Parse(string(templateContents))
	if err != nil {
		return []byte{}, err
	}

	var templatedContent bytes.Buffer
	err = tpl.Execute(&templatedContent, data)
	if err != nil {
		return []byte{}, err
	}

	return templatedContent.Bytes(), nil
}
