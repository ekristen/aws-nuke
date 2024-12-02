package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ekristen/libnuke/pkg/docs"
	"github.com/ekristen/libnuke/pkg/registry"

	"github.com/ekristen/aws-nuke/v3/pkg/common"

	_ "github.com/ekristen/aws-nuke/v3/resources"
)

//go:embed files/*
var ResourceTemplates embed.FS

type TemplateData struct {
	Name                string
	Description         string
	Properties          map[string]string
	Settings            []string
	DependsOn           []string
	DeprecatedAliases   []string
	AlternativeResource string
}

func execute(c *cli.Context) error { //nolint:funlen,gocyclo
	var regs registry.Registrations

	if c.String("resource") == "all" {
		regs = registry.GetRegistrations()
	} else if c.String("resource") != "" {
		regs = registry.Registrations{
			c.String("resource"): registry.GetRegistration(c.String("resource")),
		}
	}

	for _, reg := range regs {
		if reg.Resource == nil {
			continue
		}

		data := TemplateData{
			Name:                reg.Name,
			Properties:          docs.GeneratePropertiesMap(reg.Resource),
			Settings:            reg.Settings,
			DependsOn:           reg.DependsOn,
			DeprecatedAliases:   reg.DeprecatedAliases,
			AlternativeResource: reg.AlternativeResource,
		}

		rawTmpl, err := ResourceTemplates.ReadFile("files/resource.gomd")
		if err != nil {
			return err
		}

		funcMap := template.FuncMap{
			"KebabCase":      KebabCase,
			"SplitCamelCase": SplitCamelCase,
			"ToLower":        toLower,
		}

		tmpl, err := template.New("example").Funcs(funcMap).Parse(string(rawTmpl))
		if err != nil {
			return err
		}

		var buf bytes.Buffer

		err = tmpl.Execute(&buf, data)
		if err != nil {
			return err
		}

		if c.Bool("write-to-disk") {
			err := os.WriteFile(
				fmt.Sprintf("docs/resources/%s.md",
					KebabCase(strings.ToLower(SplitCamelCase(reg.Name)))), buf.Bytes(), 0600)
			if err != nil {
				return err
			}

			fmt.Printf("Wrote docs/resources/%s.md\n", reg.Name)

			continue
		}

		fmt.Println(buf.String())
	}

	mkdocs, err := os.ReadFile("mkdocs.yml")
	if err != nil {
		return err
	}

	resources, err := filepath.Glob("docs/resources/*.md")
	if err != nil {
		return err
	}

	var newResources []string

	for _, resource := range resources {
		if strings.Contains(resource, "overview") {
			continue
		}

		newResource := strings.Replace(resource, "docs/", "", 1)
		filename := filepath.Base(resource)
		title := strings.Join(strings.Split(filename, "-"), " ")
		title = strings.Replace(title, ".md", "", 1)
		title = cases.Title(language.English).String(title)
		title = strings.Replace(title, "Ec2", "EC2", 1)
		title = strings.Replace(title, "Iam", "IAM", 1)
		title = strings.Replace(title, "Sqs", "SQS", 1)
		title = strings.Replace(title, "Ssm", "SSM", 1)
		title = strings.Replace(title, "Kms", "KMS", 1)
		title = strings.Replace(title, "Ebs", "EBS", 1)
		title = strings.Replace(title, "Efs", "EFS", 1)
		title = strings.Replace(title, "Rds", "RDS", 1)
		title = strings.Replace(title, "Vpc", "VPC", 1)
		title = strings.Replace(title, "Acm", "ACM", 1)
		title = strings.Replace(title, "Sns", "SNS", 1)

		newResources = append(newResources, fmt.Sprintf("%s: %s", title, newResource))
	}

	slices.Sort(newResources)
	newResources = append([]string{"Overview: resources/overview.md"}, newResources...)
	newMkdocs := updateResources(string(mkdocs), newResources)

	if c.Bool("write-to-disk") {
		err := os.WriteFile("mkdocs.yml", []byte(newMkdocs), 0600)
		if err != nil {
			return err
		}

		fmt.Println("Wrote mkdocs.yml")

		return nil
	}

	fmt.Println(newMkdocs)

	return nil
}

func main() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "resource",
			Value: "all",
		},
		&cli.BoolFlag{
			Name:    "write-to-disk",
			Aliases: []string{"write"},
		},
	}

	defer func() {
		if r := recover(); r != nil {
			// log panics forces exit
			if _, ok := r.(*logrus.Entry); ok {
				os.Exit(1)
			}
			panic(r)
		}
	}()

	app := cli.NewApp()
	app.Name = "generate-docs"
	app.Usage = "generate resource docs from code"
	app.Version = common.AppVersion.Summary
	app.Authors = []*cli.Author{
		{
			Name:  "Erik Kristensen",
			Email: "erik@erikkristensen.com",
		},
	}
	app.Flags = flags
	app.Action = execute

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

var (
	spaces      = regexp.MustCompile(`\s+`)
	nonAlphaNum = regexp.MustCompile(`[^\pL\pN]+`)
)

// KebabCase -
func KebabCase(in string) string {
	s := casePrepare(in)
	return spaces.ReplaceAllString(s, "-")
}

func casePrepare(in string) string {
	in = strings.TrimSpace(in)
	s := strings.ToLower(in)
	// make sure the first letter remains lower- or upper-cased
	s = strings.Replace(s, string(s[0]), string(in[0]), 1)
	s = nonAlphaNum.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func toLower(in string) string {
	return strings.ToLower(in)
}

func SplitCamelCase(input string) string {
	// Regular expression to find boundaries between lowercase and uppercase letters,
	// and between sequences of uppercase letters followed by lowercase letters.
	re := regexp.MustCompile(`([a-z])([A-Z0-9])|([A-Z]+)([A-Z][a-z])|(\d)([A-Z])`)
	// Replace boundaries with a space followed by the uppercase letter.
	return re.ReplaceAllString(input, "${1}${3}${5} ${2}${4}${6}")
}

// Function to update the 'Resources' section with new list values
func updateResources(markdown string, newResources []string) string {
	// Define the regex to match the 'Resources:' section and the following list
	re := regexp.MustCompile(`(?ms)(^\s*- Resources:(?:\n\s*-\s.+)+)`)

	// Join new resource list as markdown format
	newList := "  - Resources:\n"
	for _, resource := range newResources {
		newList += fmt.Sprintf("    - %s\n", resource)
	}

	// Replace the matched 'Resources' list with the new list
	updatedMarkdown := re.ReplaceAllString(markdown, newList)
	return updatedMarkdown
}
