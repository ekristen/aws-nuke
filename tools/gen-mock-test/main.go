package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iancoleman/strcase"

	_ "github.com/rebuy-de/aws-nuke/v2/resources"
)

const rawTmpl = `package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/{{ lower .Service }}"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_{{ lower .Service }}iface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_{{ .Service }}{{ .Resource }}_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock{{ .Service }} := mock_{{ lower .Service }}iface.NewMock{{ .Service }}API(ctrl)

	{{ lower .Service }}{{ .Resource }} := {{ .Service }}{{ .Resource }}{
		svc:  mock{{ .Service }},
		// populate fields
	}

	mockIAM.EXPECT().{{ .Action }}{{ .ModdedResource }}(gomock.Eq(&{{ lower .Service }}.{{ .Action }}{{ .ModdedResource }}Input{
		// need properties
	})).Return(&iam.{{ .Action }}{{ .ModdedResource }}Output{}, nil)

	err := {{ lower .Service }}{{ .Resource }}.Remove()
	a.Nil(err)
}
`

var service string
var resource string
var output string

func main() {
	flag.StringVar(&service, "service", "", "AWS Service Name (eg, IAM, EC2)")
	flag.StringVar(&resource, "resource", "", "AWS Resource Name for the Service (eg, Instance, Role, Policy")
	flag.StringVar(&output, "output", "stdout", "Where to write the generated code to (default: stdout)")

	flag.Parse()

	if service == "" {
		panic(errors.New("please provide a service"))
	}
	if resource == "" {
		panic(errors.New("please provide a resource"))
	}

	action := "Delete"
	moddedResource := resource
	if strings.HasSuffix(resource, "Attachment") {
		action = "Detach"
		moddedResource = strings.TrimSuffix(resource, "Attachment")
	}

	tmpl, err := template.New("test").Funcs(sprig.TxtFuncMap()).Parse(rawTmpl)
	if err != nil {
		panic(err)
	}

	data := struct {
		Service        string
		Resource       string
		ModdedResource string
		Action         string
	}{
		Service:        service,
		Resource:       resource,
		ModdedResource: moddedResource,
		Action:         action,
	}

	var err1 error
	out := os.Stdout
	if output == "file" {
		out, err1 = os.Create(fmt.Sprintf("resources/%s-%s_mock_test.go", strings.ToLower(service), strcase.ToKebab(resource)))
		if err != nil {
			panic(err1)
		}
		defer out.Close()
	}

	if err := tmpl.Execute(out, data); err != nil {
		panic(err)
	}
}
