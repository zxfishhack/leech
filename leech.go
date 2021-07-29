package main

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/golog"
	"go/ast"
	"go/doc"
	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/vfs"
	"io/ioutil"
	"path"
	"strings"
)

type counter struct {
	Type, Func, Value int
}

type Leech struct {
	comments, docs map[string]string

	pres *godoc.Presentation

	total, commented counter
}

func NewLeech(fs vfs.NameSpace) (res *Leech, err error) {
	corpus := godoc.NewCorpus(fs)
	err = corpus.Init()
	if err != nil {
		return
	}
	res = &Leech{
		comments: make(map[string]string),
		docs:     make(map[string]string),
		pres:     godoc.NewPresentation(corpus),
	}

	return
}

func (l *Leech) PrintCommentRate() {
	fmt.Printf("------Comment Rate------\n")
	fmt.Printf("       Type: %6.2f%%\n", float32(l.commented.Type)/float32(l.total.Type)*100)
	fmt.Printf("       Func: %6.2f%%\n", float32(l.commented.Func)/float32(l.total.Func)*100)
	fmt.Printf("      Value: %6.2f%%\n", float32(l.commented.Value)/float32(l.total.Value)*100)
	fmt.Printf("------------------------\n")
}

func (l *Leech) SaveDoc(fn string) error {
	b, _ := json.Marshal(l.docs)
	return ioutil.WriteFile(fn, b, 0666)
}

func (l *Leech) SaveComment(fn string) error {
	b, _ := json.Marshal(l.comments)
	return ioutil.WriteFile(fn, b, 0666)
}

func (l *Leech) Walk(module string) {
	absPath := path.Join("/src", module)
	info := l.pres.GetPkgPageInfo(absPath, ".", godoc.NoFiltering)

	l.packageDoc(module, info.PDoc)
	if info.Dirs != nil {
		for _, dir := range info.Dirs.List {
			dirInfo := l.pres.GetPkgPageInfo(path.Join(absPath, dir.Path), ".", godoc.NoFiltering)
			l.packageDoc(path.Join(module, dir.Path), dirInfo.PDoc)
		}
	}
	golog.Debug(l.docs)
	golog.Debug(l.comments)
}

func (l *Leech) packageDoc(name string, pkg *doc.Package) {
	if pkg == nil {
		return
	}
	golog.Debug(name)
	if pkg.Doc != "" {
		l.docs[name] = strings.TrimSpace(pkg.Doc)
	}
	for _, v := range pkg.Consts {
		l.valueDoc(name, v)
	}

	for _, v := range pkg.Vars {
		l.valueDoc(name, v)
	}

	for _, v := range pkg.Types {
		l.typeDoc(name, v)
	}

	for _, v := range pkg.Funcs {
		l.funcDoc(name, v)
	}
}

func (l *Leech) typeDoc(prefix string, v *doc.Type) {
	if v == nil {
		return
	}
	l.total.Type++
	hasDoc := false
	if v.Doc != "" {
		l.docs[prefix+"."+v.Name] = strings.TrimSpace(v.Doc)
		hasDoc = true
	}

	for _, f := range v.Funcs {
		l.funcDoc(prefix+"."+v.Name, f)
	}

	for _, f := range v.Methods {
		l.funcDoc(prefix+"."+v.Name, f)
	}

	if v.Decl != nil {
		for _, spec := range v.Decl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				if interfaceSpec, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					if interfaceSpec.Methods != nil {
						for _, field := range interfaceSpec.Methods.List {
							l.fieldDoc(prefix+"."+v.Name, field)
						}
					}
				} else if structSpec, ok := typeSpec.Type.(*ast.StructType); ok {
					if structSpec.Fields != nil {
						for _, field := range structSpec.Fields.List {
							l.fieldDoc(prefix+"."+v.Name, field)
						}
					}
				}
			}
		}
	}

	if hasDoc {
		l.commented.Type++
	}
}

func (l *Leech) funcDoc(prefix string, v *doc.Func) {
	if v == nil {
		return
	}
	hasDoc := false
	l.total.Func++
	funcName := prefix + "." + v.Name
	if v.Name == "TypeToSchema" {
		golog.Debug(v.Name)
	}
	if v.Doc != "" {
		l.docs[funcName] = strings.TrimSpace(v.Doc)
		hasDoc = true
	}

	if hasDoc {
		l.commented.Func++
	}
}

func (l *Leech) valueDoc(prefix string, v *doc.Value) {
	if v == nil {
		return
	}
	hasDoc := false
	l.total.Value++

	if v.Doc != "" {
		for _, name := range v.Names {
			fieldName := prefix + "." + name
			l.docs[fieldName] = strings.TrimSpace(v.Doc)
			hasDoc = true
		}
	}
	if v.Decl != nil {
		for _, spec := range v.Decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					fieldName := prefix + "." + name.Name
					txt := strings.TrimSpace(valueSpec.Doc.Text())
					if txt != "" {
						l.docs[fieldName] = txt
						hasDoc = true
					}
					txt = strings.TrimSpace(valueSpec.Comment.Text())
					if txt != "" {
						l.comments[fieldName] = txt
						hasDoc = true
					}
				}
			}
		}
	}

	if hasDoc {
		l.commented.Value++
	}
}

func (l *Leech) fieldDoc(prefix string, v *ast.Field) {
	if v == nil {
		return
	}
	for _, name := range v.Names {
		fieldName := prefix + "." + name.Name
		txt := strings.TrimSpace(v.Doc.Text())
		if txt != "" {
			l.docs[fieldName] = txt
		}
		txt = strings.TrimSpace(v.Comment.Text())
		if txt != "" {
			l.comments[fieldName] = txt
		}
	}
}
