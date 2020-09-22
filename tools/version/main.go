package main

import (
	"fmt"
	"os/exec"
	"regexp"
)

var version string

func main() {
	// TODO: all we're doing is trimming off the v - do we need this? We are also trimming off the 'g' in SCM commit...
	cmd := exec.Command("git", "describe", "--tags")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("0.0.0") // TODO: should we exit with error?
	}
	re := regexp.MustCompile("v(?P<version>.+)-(?P<commits>.+)-g(?P<sha>.+)")
	matches := re.FindStringSubmatch(string(output))
	if len(matches) != 4 {
		fmt.Println("0.0.0") // TODO: should we exit with error?
	}
	fmt.Println(matches[1] + "-" + matches[2] + "+" + matches[3])


}