package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const resourceTemplate = `package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/{{.Service}}"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const {{.ResourceType}}Resource = "{{.ResourceType}}"

func init() {
	resource.Register(&resource.Registration{
		Name:   {{.ResourceType}}Resource,
		Scope:  nuke.Account,
		Lister: &{{.ResourceType}}Lister{},
	})
}

type {{.ResourceType}}Lister struct{}

func (l *{{.ResourceType}}Lister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := {{.Service}}.New(opts.Session)
	var resources []resource.Resource

	// INSERT CODE HERE TO ITERATE AND ADD RESOURCES

	return resources, nil
}

type {{.ResourceType}} struct {
	svc  *{{.Service}}.{{.ResourceType}}
	id   *string
	tags []*{{.Service}}.Tag
}

func (r *{{.ResourceType}}) Remove(_ context.Context) error {
	return nil
}

func (r *{{.ResourceType}}) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	return properties
}

func (r *{{.ResourceType}}) String() string {
	return *r.id
}
`

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		fmt.Println("usage: create-resource <service> <resource>")
		os.Exit(1)
	}

	service := args[0]
	resourceType := args[1]

	data := struct {
		Service      string
		ResourceType string
	}{
		Service:      strings.ToLower(service),
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

	fmt.Println(tpl.String())
}
