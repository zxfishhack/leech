package main

import (
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

func main() {
	module := os.Args[len(os.Args)-1]
	for _, p := range filepath.SplitList(gobuild.Default.GOPATH) {
		mp := filepath.Join(p, "src", module)
		if _, err := os.ReadDir(mp); err == nil {
			fs.Bind(path.Join("/src", module), vfs.OS(mp), "/", vfs.BindReplace)
			fs.Bind(path.Join("/src", module, "vendor"), mapfs.New(nil), "/", vfs.BindReplace)
			break
		}
	}

	leech, err := NewLeech(fs)

	if err != nil {
		golog.Fatal(err)
	}

	leech.Walk(module)

	leech.PrintCommentRate()

	leech.SaveDoc("doc.json")
	leech.SaveComment("comment.json")
}
