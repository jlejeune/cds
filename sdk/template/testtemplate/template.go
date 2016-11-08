package main

import (
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/template"
)

type TestTemplate struct {
	template.Common
}

func (t *TestTemplate) Name() string {
	return "testtemplate"
}

func (t *TestTemplate) Description() string {
	return "Description"
}

func (t *TestTemplate) Identifier() string {
	return "github.com/ovh/cds/sdk/template/TestTemplate"
}

func (t *TestTemplate) Author() string {
	return "François Samin <francois.samin@corp.ovh.com>"
}

func (t *TestTemplate) Type() string {
	return "BUILD"
}

func (t *TestTemplate) Parameters() []sdk.TemplateParam {
	return []sdk.TemplateParam{
		{
			Name:  "param1",
			Type:  sdk.StringVariable,
			Value: "value1",
		},
		{
			Name:  "param2",
			Type:  sdk.StringVariable,
			Value: "value2",
		},
	}
}

func (t *TestTemplate) Apply(opts template.IApplyOptions) (sdk.Application, error) {
	//Prepare action gitclone
	actionGitClone, err := sdk.NewActionFromRemoteScript("https://raw.githubusercontent.com/ovh/cds-contrib/actions/action-scripts/cds-git-clone.hcl", nil)
	if err != nil {
		return sdk.Application{}, err
	}

	//Return full application
	return sdk.Application{
		Name:       opts.ApplicationName(),
		ProjectKey: opts.ProjetKey(),
		Variable: []sdk.Variable{
			{
				Name:  "repo",
				Value: opts.Parameters().Get("repo"),
			},
			{
				Name:  "name",
				Value: opts.Parameters().Get("name"),
			},
		},
		Pipelines: []sdk.ApplicationPipeline{ //Pipelines
			{
				Pipeline: sdk.Pipeline{
					Name: "build",
					Type: sdk.BuildPipeline,
					Stages: []sdk.Stage{ //Stages
						{
							Name:       "Build",
							BuildOrder: 0,
							Enabled:    true,
							Actions: []sdk.Action{ //Jobs
								{
									Name: "Compile", //First job : compile
									Actions: []sdk.Action{
										*actionGitClone,
										sdk.NewActionScript("cd {{.cds.app.name}} && make", []sdk.Requirement{
											{
												Name:  "make",
												Type:  sdk.BinaryRequirement,
												Value: "make",
											},
										}),
										sdk.NewActionArtifactUpload("{{.cds.app.name}}", "{{.cds.version}}"),
									},
								},
								{
									Name: "Test", //Second job : test
									Actions: []sdk.Action{
										*actionGitClone,
										sdk.NewActionScript("cd {{.cds.app.name}} && make test", []sdk.Requirement{
											{
												Name:  "make",
												Type:  sdk.BinaryRequirement,
												Value: "make",
											},
										}),
										sdk.NewActionJUnit("*.xml"),
									},
								},
							},
						},
						{
							Name:       "Package",
							BuildOrder: 1,
							Enabled:    true,
							Actions: []sdk.Action{ //Jobs
								{
									Name: "Docker package",
									Actions: []sdk.Action{
										*actionGitClone,
										sdk.NewActionScript(`
cd {{.cds.app.name}} 
docker build -t cds/{{.cds.app.name}}-{{.cds.version}} . 
docker push cds/{{.cds.app.name}}-{{.cds.version}}`, []sdk.Requirement{
											{
												Name:  "docker",
												Type:  sdk.BinaryRequirement,
												Value: "docker",
											},
										}),
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func main() {
	p := TestTemplate{}
	template.Serve(&p)
}