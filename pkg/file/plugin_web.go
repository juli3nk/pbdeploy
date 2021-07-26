package file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"text/template"
)

func init() {
	RegisterDriver("js", NewJsFiler)
}

type JSFiler struct {
	config          *config
	packageJsonFile packageJsonFile
}

type packageJsonFile struct {
	Scope       string   `json:"scope,omitempty"`
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
	License     string   `json:"license,omitempty"`
	Author      string   `json:"author,omitempty"`
	Repository  string   `json:"repository,omitempty"`
	//DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Private       *bool             `json:"private,omitempty"`
	PublishConfig map[string]string `json:"publishConfig,omitempty"`
}

func NewJsFiler(conf map[string][]string) (Filer, error) {
	var srcPath string
	var dstPath string

	f := JSFiler{}

	pjf := packageJsonFile{}

	var name string
	var repoName string
	v, ok := conf["name"]
	if ok {
		name = v[0]
		repoName = v[0]
	}
	v, ok = conf["scope"]
	if ok {
		name = fmt.Sprintf("@%s/%s", v[0], name)
	}
	pjf.Name = name

	v, ok = conf["version"]
	if ok {
		if string(v[0][0]) == "v" {
			pjf.Version = string(v[0][1:])
		} else {
			pjf.Version = v[0]
		}
	}
	v, ok = conf["description"]
	if ok {
		pjf.Description = v[0]
	}
	v, ok = conf["keywords"]
	if ok {
		pjf.Keywords = v
	}
	v, ok = conf["homepage"]
	if ok {
		pjf.Homepage = v[0]
	}
	v, ok = conf["license"]
	if ok {
		pjf.License = v[0]
	}
	v, ok = conf["author"]
	if ok {
		pjf.Author = v[0]
	}
	v, ok = conf["repository"]
	if ok {
		pjf.Repository = v[0]
	}
	v, ok = conf["private"]
	if ok {
		val, err := strconv.ParseBool(v[0])
		if err != nil {
			return nil, err
		}
		pjf.Private = &val
	}
	v, ok = conf["publish_config"]
	if ok {
		r := make(map[string]string)
		for _, pc := range v {
			c := strings.Split(pc, ";")

			if len(c) > 0 {
				k := c[0]
				val := c[1]

				r[k] = val
			}
		}

		pjf.PublishConfig = r
	}
	if *pjf.Private && ok {
		pjf.Private = nil
	}

	f.packageJsonFile = pjf

	v, ok = conf["src_path"]
	if ok {
		srcPath = v[0]
	}
	v, ok = conf["dst_path"]
	if ok {
		dstPath = v[0]
	}

	dstPath = path.Join(dstPath, repoName)

	f.config = newFile(repoName, pjf.Description, srcPath, dstPath)

	return &f, nil
}

func (f *JSFiler) String() string {
	return "javascript"
}

func (f *JSFiler) DeleteFiles() error {
	return f.config.deleteFiles()
}

func (f *JSFiler) CopyFiles() error {
	return f.config.copyFiles()
}

func (f *JSFiler) CreateOrUpdateFiles() error {
	if err := f.createReadme(); err != nil {
		return err
	}

	if err := f.createPackageJson(); err != nil {
		return err
	}

	return nil
}

func (f *JSFiler) createReadme() error {
	type data struct {
		Name string
		Desc string
	}

	filename := path.Join(f.config.dstPath, "README.md")

	readmeMdTemplate := `# {{ .Name }}

{{ .Desc }}
	`

	t := template.Must(template.New("README.md").Parse(readmeMdTemplate))

	buf := new(bytes.Buffer)
	d := data{f.config.name, f.config.description}

	if err := t.Execute(buf, d); err != nil {
		return err
	}

	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func (f *JSFiler) createPackageJson() error {
	filename := path.Join(f.config.dstPath, "package.json")

	b, err := json.Marshal(f.packageJsonFile)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	json.Indent(buf, b, "", "\t")

	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}
