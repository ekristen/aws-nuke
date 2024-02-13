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
"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const {{.Combined}}Resource = "{{.Combined}}"

func init() {
	registry.Register(&registry.Registration{
		Name:   {{.Combined}}Resource,
		Scope:  nuke.Account,
		Lister: &{{.Combined}}Lister{},
	})
}

type {{.Combined}}Lister struct{}

func (l *{{.Combined}}Lister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := {{.Service}}.New(opts.Session)
	var resources []resource.Resource

	// NOTE: you might have to modify the code below to actually work, this currently does not 
	// inspect the aws sdk instead is a jumping off point
	res, err := svc.List{{.ResourceTypeTitle}}s(&{{.Service}}.List{{.ResourceTypeTitle}}sInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.{{.ResourceTypeTitle}}s {
		resources = append(resources, &{{.Combined}}{
			svc:  svc,
			id:   p.Id,
			tags: p.Tags,
		})
	}

	return resources, nil
}

type {{.Combined}} struct {
	svc  *{{.Service}}.{{.ServiceTitle}}
	id   *string
	tags []*{{.Service}}.Tag
}

func (r *{{.Combined}}) Remove(_ context.Context) error {
	_, err := r.svc.Delete{{.ResourceTypeTitle}}(&{{.Service}}.Delete{{.ResourceTypeTitle}}Input{
		{{.ResourceTypeTitle}}Id: r.id, 
	})
	return err
}

func (r *{{.Combined}}) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	return properties
}

func (r *{{.Combined}}) String() string {
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
		Service           string
		ServiceTitle      string
		ResourceType      string
		ResourceTypeTitle string
		Combined          string
	}{
		Service:           strings.ToLower(service),
		ServiceTitle:      strings.Title(service),
		ResourceType:      resourceType,
		ResourceTypeTitle: strings.Title(resourceType),
		Combined:          fmt.Sprintf("%s%s", strings.Title(service), strings.Title(resourceType)),
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
