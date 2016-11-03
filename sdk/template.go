package sdk

import (
	"encoding/json"
	"fmt"

	"gopkg.in/gorp.v1"
)

// TemplateParam can be a String/Date/Script/URL...
type TemplateParam struct {
	ID          int64        `json:"id" yaml:"-"`
	Name        string       `json:"name"`
	Type        VariableType `json:"type"`
	Value       string       `json:"value"`
	Description string       `json:"description" yaml:"desc,omitempty"`
}

// Template definition to help users bootstrap their pipelines
type Template struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Params      []TemplateParam `json:"params"`
	Hook        bool            `json:"hook"`
}

// GetBuildTemplate Get the build template corresponding to the given name
func GetBuildTemplate(name string) (*Template, error) {
	tpls, err := GetBuildTemplates()
	if err != nil {
		return nil, err
	}

	for _, t := range tpls {
		if t.Name == name {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("%s: not found", err)
}

// GetBuildTemplates retrieves all existing build template from API
func GetBuildTemplates() ([]Template, error) {
	uri := "/template/build"

	data, code, err := Request("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	if code >= 300 {
		return nil, fmt.Errorf("HTTP %d", code)
	}

	var tmpls []Template
	err = json.Unmarshal(data, &tmpls)
	if err != nil {
		return nil, err
	}

	return tmpls, nil
}

// GetDeploymentTemplates retrieves all existing deployment template from API
func GetDeploymentTemplates() ([]Template, error) {
	uri := "/template/deploy"

	data, code, err := Request("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	if code >= 300 {
		return nil, fmt.Errorf("HTTP %d", code)
	}

	var tmpls []Template
	err = json.Unmarshal(data, &tmpls)
	if err != nil {
		return nil, err
	}

	return tmpls, nil
}

// ApplyApplicationTemplates creates given application and apply build and deployment templates
func ApplyApplicationTemplates(projectKey string, name, repo string, build, deploy Template) (*Application, error) {
	uri := fmt.Sprintf("/template/%s", projectKey)

	app := &Application{
		Name: name,
		//		BuildTemplate:  build,
		//		DeployTemplate: deploy,
		Variable: []Variable{
			Variable{
				Name:  "repo",
				Type:  StringVariable,
				Value: repo,
			},
		},
	}

	data, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}

	data, code, err := Request("POST", uri, data)
	if err != nil {
		return nil, err
	}
	if code >= 300 {
		return nil, fmt.Errorf("HTTP %d", code)
	}

	err = json.Unmarshal(data, app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

//TemplateExtention represents a template store as a binary extension
type TemplateExtention struct {
	ID          int64           `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Type        string          `json:"type" db:"type"`
	Author      string          `json:"author" db:"author"`
	Description string          `json:"description"`
	Identifier  string          `json:"identifier" db:"identifier"`
	Size        int64           `json:"-" db:"size"`
	Perm        uint32          `json:"-" db:"perm"`
	MD5Sum      string          `json:"md5sum" db:"md5sum"`
	ObjectPath  string          `json:"-" db:"object_path"`
	Filename    string          `json:"-" db:"-"`
	Path        string          `json:"-" db:"-"`
	Params      []TemplateParam `json:"-" db:"-"`
}

//PostInsert is a DB Hook on TemplateExtention to store params as JSON in DB
func (t *TemplateExtention) PostInsert(s gorp.SqlExecutor) error {
	btes, err := json.Marshal(t.Params)
	if err != nil {
		return err
	}

	query := "insert into template_params (template_id, params) values ($1, $2)"
	if _, err := s.Exec(query, t.ID, btes); err != nil {
		return err
	}
	return nil
}

//PostUpdate is a DB Hook on TemplateExtention to store params as JSON in DB
func (t *TemplateExtention) PostUpdate(s gorp.SqlExecutor) error {
	btes, err := json.Marshal(t.Params)
	if err != nil {
		return err
	}

	query := "update template_params set params = $2 where template_id = $1"
	if _, err := s.Exec(query, t.ID, btes); err != nil {
		return err
	}
	return nil
}

//PreDelete is a DB Hook on TemplateExtention to store params as JSON in DB
func (t *TemplateExtention) PreDelete(s gorp.SqlExecutor) error {
	query := "delete from template_params where template_id = $1"
	if _, err := s.Exec(query, t.ID); err != nil {
		return err
	}
	return nil

}

//GetName returns the name of the template extension
func (a *TemplateExtention) GetName() string {
	return a.Name
}

//GetPath returns the storage path of the template extension
func (a *TemplateExtention) GetPath() string {
	return fmt.Sprintf("templates")
}
