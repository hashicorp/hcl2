package gohcl_test

import (
	"fmt"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func ExampleEncodeIntoBody() {
	type Service struct {
		Name string   `hcl:"name,label"`
		Exe  []string `hcl:"executable"`
	}
	type Constraints struct {
		OS   string `hcl:"os"`
		Arch string `hcl:"arch"`
	}
	type Locals struct {
		Definitions hcl.Attributes `hcl:",remain"`
	}
	type App struct {
		Name        string       `hcl:"name"`
		Desc        string       `hcl:"description"`
		Constraints *Constraints `hcl:"constraints,block"`
		Services    []Service    `hcl:"service,block"`
		Locals      []*Locals    `hcl:"locals,block"`
	}

	app := App{
		Name: "awesome-app",
		Desc: "Such an awesome application",
		Constraints: &Constraints{
			OS:   "linux",
			Arch: "amd64",
		},
		Services: []Service{
			{
				Name: "web",
				Exe:  []string{"./web", "--listen=:8080"},
			},
			{
				Name: "worker",
				Exe:  []string{"./worker"},
			},
		},
		Locals: []*Locals{
			&Locals{
				Definitions: hcl.Attributes{
					"hello": &hcl.Attribute{
						Name: "hello",
						Expr: hcl.StaticExpr(cty.StringVal("world"), hcl.Range{}),
					},
					"foo": &hcl.Attribute{
						Name: "foo",
						Expr: &hclsyntax.ScopeTraversalExpr{
							Traversal: hcl.Traversal{
								hcl.TraverseRoot{
									Name: "bar",
								},
							},
						},
					},
				},
			},
		},
	}

	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&app, f.Body())
	fmt.Printf("%s", f.Bytes())

	// Output:
	// name        = "awesome-app"
	// description = "Such an awesome application"
	//
	// constraints {
	//   os   = "linux"
	//   arch = "amd64"
	// }
	//
	// service "web" {
	//   executable = ["./web", "--listen=:8080"]
	// }
	// service "worker" {
	//   executable = ["./worker"]
	// }
	//
	// locals {
	//   hello = "world"
	//   foo   = bar
	// }
}
