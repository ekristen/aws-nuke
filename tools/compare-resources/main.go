package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/fatih/color"
)

var OriginalRegisterRegex = regexp.MustCompile(
	`register\("(?P<resource>.*)",\s?(?P<function>\w+)(,)?(\s+mapCloudControl\("(?P<cc>.*)"\))?`)
var NewRegisterRegex = regexp.MustCompile(`registry.Registration{\s+Name:\s+(?P<name>.*),`)

var aliases = map[string]string{
	"NetpuneSnapshot":                    "NeptuneSnapshot",
	"EKSFargateProfiles":                 "EKSFargateProfile",
	"EKSNodegroups":                      "EKSNodegroup",
	"ComprehendPiiEntititesDetectionJob": "ComprehendPIIEntitiesDetectionJob",
}

func main() { //nolint:funlen,gocyclo
	args := os.Args[1:]

	if len(args) == 0 {
		panic("no arguments given")
	}

	upstreamDirectory := filepath.Clean(filepath.Join(args[0], "resources"))

	var upstreamResourceFiles []string
	var upstreamResourceTypes []string
	var upstreamTypeToFile = map[string]string{}

	//nolint:gosec // path from CLI args
	err := filepath.WalkDir(upstreamDirectory, func(path string, di fs.DirEntry, err error) error {
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		upstreamResourceFiles = append(upstreamResourceFiles, filepath.Base(path))
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, file := range upstreamResourceFiles {
		originalFileContents, err := os.ReadFile(filepath.Clean(filepath.Join(upstreamDirectory, file))) //nolint:gosec // path from CLI args
		if err != nil {
			panic(err)
		}

		matches := OriginalRegisterRegex.FindStringSubmatch(string(originalFileContents))

		if len(matches) < 3 {
			fmt.Printf("WARNING: ERROR no matches in %s\n", file)
			continue
		}
		resourceType := matches[1]
		funcName := matches[2]
		_ = funcName

		upstreamResourceTypes = append(upstreamResourceTypes, resourceType)
		upstreamTypeToFile[resourceType] = file
	}

	var localResourcesPath = "resources"
	var localResourceFiles []string
	var localResourceTypes []string

	if err := filepath.WalkDir(localResourcesPath, func(path string, di fs.DirEntry, err error) error {
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		localResourceFiles = append(localResourceFiles, filepath.Base(path))
		return nil
	}); err != nil {
		panic(err)
	}

	for _, file := range localResourceFiles {
		originalFileContents, err := os.ReadFile(filepath.Join(localResourcesPath, file))
		if err != nil {
			panic(err)
		}

		matches := NewRegisterRegex.FindStringSubmatch(string(originalFileContents))

		if len(matches) == 0 {
			continue
		}

		var NameRegex = regexp.MustCompile(fmt.Sprintf(`const %s = "(?P<name>.*)"`, matches[1]))

		nameMatches := NameRegex.FindStringSubmatch(string(originalFileContents))
		if len(nameMatches) == 0 {
			continue
		}

		resourceType := nameMatches[1]

		localResourceTypes = append(localResourceTypes, resourceType)
	}

	fmt.Println("\nrebuy-de/aws-nuke resource count:", len(upstreamResourceTypes))
	fmt.Println("ekristen/aws-nuke resource count:", len(localResourceTypes))

	fmt.Println("\nResources unique to ekristen/aws-nuke:")
	for _, resource := range localResourceTypes {
		found := false
		for _, v := range aliases {
			if v == resource {
				found = true
			}
		}

		if !slices.Contains(upstreamResourceTypes, resource) && !found {
			color.New(color.Bold, color.FgGreen).Printf("%-55s\n", resource)
		}
	}

	fmt.Println("\nResources not in ekristen/aws-nuke:")
	for _, resource := range upstreamResourceTypes {
		_, ok := aliases[resource]
		if !slices.Contains(localResourceTypes, resource) && !ok {
			color.New(color.Bold).Printf("%-55s", resource)
			color.New(color.FgYellow).Println(upstreamTypeToFile[resource])
		}
	}
}
