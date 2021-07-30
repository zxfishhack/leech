package main

import (
	"flag"
	"github.com/kataras/golog"
	gobuild "go/build"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
	"os"
	"path"
	"path/filepath"
)

var (
	fs = vfs.NewNameSpace()
)

var (
	flagSkipCodeGen bool
	flagModule      string
	flagVerbose     bool

	flagOutputFile string
	flagOutputPkg  string
)

func main() {
	flag.BoolVar(&flagSkipCodeGen, "no-codegen", true, "skip code gen")
	flag.StringVar(&flagModule, "m", "", "module to leech comments")
	flag.BoolVar(&flagVerbose, "v", false, "verbose")
	flag.StringVar(&flagOutputPkg, "pkg", "main", "Package name to use in the generated code. (default \"main\")")
	flag.StringVar(&flagOutputFile, "o", "./leech_gen.go", "Optional name of the output file to be generated. (default \"./leech_gen.go\")")
	flag.Parse()

	if flagModule == "" {
		flag.Usage()
		return
	}

	if flagVerbose {
		golog.SetLevel(golog.Levels[golog.DebugLevel].Name)
	}
	fs.Bind("/", vfs.OS(gobuild.Default.GOROOT), "/", vfs.BindReplace)
	for _, p := range filepath.SplitList(gobuild.Default.GOPATH) {
		mp := filepath.Join(p, "src", flagModule)
		if _, err := os.ReadDir(mp); err == nil {
			fs.Bind(path.Join("/src", flagModule), vfs.OS(mp), "/", vfs.BindReplace)
			fs.Bind(path.Join("/src", flagModule, "vendor"), mapfs.New(nil), "/", vfs.BindReplace)
			break
		}
	}

	leech, err := NewLeech(fs)

	if err != nil {
		golog.Fatal(err)
	}

	leech.Walk(flagModule)

	err = leech.save()
	if err != nil {
		golog.Fatal(err)
	}

	leech.PrintCommentRate()
}
