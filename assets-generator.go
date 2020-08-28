// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	fp "path/filepath"

	"github.com/shurcooL/httpfs/filter"
	"github.com/shurcooL/vfsgen"
)

func main() {
	fs := filter.Skip(http.Dir("internal/view"), func(path string, fi os.FileInfo) bool {
		fileDir := fp.Dir(path)
		fileExt := fp.Ext(fi.Name())

		// Exclude JS file that not inside JS root dir
		if fileExt == ".js" && fileDir != "/js" {
			return true
		}

		// Exclude all LESS file
		if fileExt == ".less" {
			return true
		}

		return false
	})

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:     "internal/backend/assets-prod.go",
		PackageName:  "backend",
		BuildTags:    "prod",
		VariableName: "assets",
	})

	if err != nil {
		log.Fatalln(err)
	}
}
