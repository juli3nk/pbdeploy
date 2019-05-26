package file

import (
	"bytes"
	"io/ioutil"
	"path"
	"text/template"
)

func init() {
	RegisterDriver("go", NewGoFiler)
}

type GoFiler struct {
	config *config
}

func NewGoFiler(conf map[string][]string) (Filer, error) {
	var name string
	var description string
	var srcPath string
	var dstPath string

	f := GoFiler{}

	v, ok := conf["name"]
	if ok {
		name = v[0]
	}
	v, ok = conf["description"]
	if ok {
		description = v[0]
	}
	v, ok = conf["src_path"]
	if ok {
		srcPath = v[0]
	}
	v, ok = conf["dst_path"]
	if ok {
		dstPath = v[0]
	}

	dstPath = path.Join(dstPath, name)

	f.config = newFile(name, description, srcPath, dstPath)

	return &f, nil
}

func (f *GoFiler) String() string {
	return "Golang"
}

func (f *GoFiler) DeleteFiles() error {
	return f.config.deleteFiles()
}

func (f *GoFiler) CopyFiles() error {
	return f.config.copyFiles()
}

func (f *GoFiler) CreateOrUpdateFiles() error {
	if err := f.createReadme(); err != nil {
		return err
	}

	return nil
}

func (f *GoFiler) createReadme() error {
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
