package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var resourceTemplate = `const {{.ResourceType}}Resource = "{{.ResourceType}}"

func init() {
	registry.Register(&registry.Registration{
		Name:   {{.ResourceType}}Resource,
		Scope:  nuke.Account,
		Lister: &{{.ResourceType}}Lister{},
	})
}

type {{.ResourceType}}Lister struct{}`

var funcTemplate = `func (l *{{.ResourceType}}Lister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
`

var imports = `import (
	"context"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		panic("no arguments given")
	}

	if len(args) != 2 {
		fmt.Println("usage: migrate-resource <source-aws-nuke> <resource-type>")
		os.Exit(1)
	}

	originalSourceDir := filepath.Join(args[0], "resources")

	repl := regexp.MustCompile(`func init\(\) {\s+.*[\s+].*\s}`)
	match := regexp.MustCompile(`register\("(?P<resource>.*)",\s?(?P<function>\w+)(,)?(\s+mapCloudControl\("(?P<cc>.*)"\))?`)
	funcMatch := regexp.MustCompile(`func List.*{`)

	filename := filepath.Clean(filepath.Join(originalSourceDir, args[1]+".go"))

	originalFileContents, err := os.ReadFile(filename) //nolint:gosec // path from CLI args is intentional
	if err != nil {
		panic(err)
	}

	matches := match.FindStringSubmatch(string(originalFileContents))

	if len(matches) < 3 {
		panic("no matches")
	}
	resourceType := matches[1]

	data := struct {
		ResourceType string
	}{
		ResourceType: resourceType,
	}

	tmpl, err := template.New("resource").Parse(resourceTemplate)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		panic(err)
	}

	funcTmpl, err := template.New("function").Parse(funcTemplate)
	if err != nil {
		panic(err)
	}

	var funcTpl bytes.Buffer
	if err := funcTmpl.Execute(&funcTpl, data); err != nil {
		panic(err)
	}

	newContents := repl.ReplaceAllString(string(originalFileContents), tpl.String())

	newContents = strings.ReplaceAll(newContents, "github.com/rebuy-de/aws-nuke/v2/pkg/types", "")

	newContents = funcMatch.ReplaceAllString(newContents, funcTpl.String())
	newContents = strings.ReplaceAll(newContents, "[]Resource", "[]resource.Resource")
	newContents = strings.ReplaceAll(newContents, "(sess)", "(opts.Session)")

	newContents = strings.ReplaceAll(newContents,
		"resources := []resource.Resource{}", "resources := make([]resource.Resource, 0)")

	newContents = strings.ReplaceAll(newContents, "import (", imports)
	newContents = strings.ReplaceAll(newContents, "\"github.com/aws/aws-sdk-go/aws/session\"", "")

	newContents = strings.ReplaceAll(newContents,
		"\"github.com/rebuy-de/aws-nuke/v2/pkg/config\"", "\"github.com/ekristen/libnuke/pkg/featureflag\"")

	newContents = strings.ReplaceAll(newContents, "config.FeatureFlags", "*featureflag.FeatureFlags")
	newContents = strings.ReplaceAll(newContents, ") Remove() error {", ") Remove(_ context.Context) error {")

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile( //nolint:gosec
		filepath.Join(cwd, "resources", args[1]+".go"), []byte(newContents), 0644); err != nil {
		panic(err)
	}
}
