package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var cmdInstall = &Command{
	Usage: "install",
	Short: "Install go dependencies",
	Long: `
Install makes sure that all the dependencies are
checked (with their right revisions) out locally
under $GOPATH/src folder.
`,
	Run: runInstall,
}

func runInstall(cmd *Command, args []string) {
	downloadDependencies()

}

// prepareGopath reads dependency information from the filesystem
// entry name, fetches any necessary code, and returns a gopath
// causing the specified dependencies to be used.
func downloadDependencies() {
	dir, isDir := findGodeps()
	if dir == "" {
		log.Fatalln("No Godeps found (or in any parent directory)")
	}
	if isDir {
		return
	}
	g, err := ReadGodeps(filepath.Join(dir, "Godeps"))
	if err != nil {
		log.Fatalln(err)
	}
	err = downloadAll(g.Deps)
	if err != nil {
		log.Fatalln(err)
	}
}

// findGodeps looks for a directory entry "Godeps" in the
// current directory or any parent, and returns the containing
// directory and whether the entry itself is a directory.
// If Godeps can't be found, findGodeps returns "".
// For any other error, it exits the program.
func findGodeps() (dir string, isDir bool) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	return findInParents(wd, "Godeps")
}

// findInParents returns the path to the directory containing name
// in dir or any ancestor, and whether name itself is a directory.
// If name cannot be found, findInParents returns the empty string.
func findInParents(dir, name string) (container string, isDir bool) {
	for {
		fi, err := os.Stat(filepath.Join(dir, name))
		if os.IsNotExist(err) && dir == "/" {
			return "", false
		}
		if os.IsNotExist(err) {
			dir = filepath.Dir(dir)
			continue
		}
		if err != nil {
			log.Fatalln(err)
		}
		return dir, fi.IsDir()
	}
}

func envNoGopath() (a []string) {
	for _, s := range os.Environ() {
		if !strings.HasPrefix(s, "GOPATH=") {
			a = append(a, s)
		}
	}
	return a
}

// sandboxAll ensures that the commits in deps are available
// on disk, and returns a GOPATH string that will cause them
// to be used.
func downloadAll(a []Dependency) (err error) {
	for _, dep := range a {
		err := download(dep)
		if err != nil {
			return err
		}
	}
    return
}

// sandbox ensures that commit d is available on disk,
// and returns a GOPATH string that will cause it to be used.
func download(d Dependency) (err error) {
	if !exists(d.RepoPath()) {
		if err = d.CreateRepo("main"); err != nil {
			return fmt.Errorf("create repo: %s", err)
		}
        d.fetchAndCheckout("main")
	}
    err = d.checkout()
	if err != nil {
		err = d.fetchAndCheckout("main")
	}
    return
}
