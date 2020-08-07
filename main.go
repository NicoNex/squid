package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/logrusorgru/aurora"
)

var (
	buildRoot string
	stylefile string
	stylecont string
	wg        sync.WaitGroup
)

func printErr(a interface{}) {
	fmt.Println(aurora.BrightRed(a).Bold())
}

func die(a interface{}) {
	fmt.Println(aurora.BrightRed(a).Bold())
	os.Exit(1)
}

func isHidden(fname string) bool {
	return fname[0] == '.'
}

// Returns true if the filename ends with '.md'.
func isMarkdown(fname string) bool {
	ok, err := filepath.Match("*.md", fname)
	if err != nil {
		printErr(err)
		return false
	}
	return ok
}

// Copies a file from src to dst.
func copyFile(src string, dst string) {
	defer wg.Done()
	in, err := os.Open(src)
	if err != nil {
		printErr(err)
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		printErr(err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		printErr(err)
		return
	}
}

// Converts the markdown src into the html dst.
func render(src string, dst string) {
	defer wg.Done()
	md, err := ioutil.ReadFile(src)
	if err != nil {
		printErr(err)
		return
	}
	html := renderHtml(md)
	if stylefile != "" {
		html = addStyle(html, stylecont)
	} else {
		html = addStyle(html, CSS)
	}

	err = ioutil.WriteFile(dst, []byte(html), 0644)
	if err != nil {
		printErr(err)
	}
}

func truncateFirstDir(path string) string {
	toks := strings.Split(path, string(os.PathSeparator))
	return filepath.Join(toks[1:]...)
}

func removeExt(fname string) string {
	return fname[:len(fname)-3]
}

// Replaces the '.md' in a filepath with '.html'
func toHtml(fname string) string {
	return strings.Replace(fname, ".md", ".html", -1)
}

func evaluate(path string, info os.FileInfo, err error) error {
	var dir, file = filepath.Split(path)

	if err != nil {
		return err
	}

	outdir := filepath.Join(buildRoot, truncateFirstDir(dir))
	fmt.Printf("Creating directory: %s\n", outdir)
	if err := os.MkdirAll(outdir, 0755); err != nil {
		return err
	}

	if isMarkdown(file) {
		htmlPath := filepath.Join(outdir, toHtml(file))
		fmt.Printf("Creating file: %s\n", htmlPath)
		wg.Add(1)
		go render(path, htmlPath)
	} else if !info.IsDir() {
		destPath := filepath.Join(outdir, file)
		fmt.Printf("Copying %s to %s\n", path, destPath)
		wg.Add(1)
		go copyFile(path, destPath)
	}
	return nil
}

func main() {
	var srcfile string

	flag.StringVar(&stylefile, "css", "", "CSS file")
	flag.Usage = usage
	flag.Parse()

	if stylefile != "" {
		if cont, err := loadCSS(stylefile); err == nil {
			stylecont = cont
		} else {
			printErr(err)
			printErr("Using fallback theme...")
			stylefile = ""
		}
	}

	switch argc := flag.NArg(); {
	case argc == 1:
		srcfile = filepath.Clean(flag.Arg(0))
		buildRoot = "build/"
	case argc > 1:
		srcfile = filepath.Clean(flag.Arg(0))
		buildRoot = filepath.Clean(flag.Arg(1))
	default:
		flag.Usage()
		os.Exit(1)
	}

	srcinfo, err := os.Stat(srcfile)
	if err != nil {
		die(err)
	}
	if srcinfo.IsDir() {
		filepath.Walk(srcfile, evaluate)
		wg.Wait()
	} else {
		if isMarkdown(srcinfo.Name()) {
			wg.Add(1)
			render(srcfile, srcfile+".html")
			wg.Wait()
		} else {
			die("Please provide a valid source directory or file!")
		}
	}
}

func usage() {
	fmt.Printf(`squid - A fast markdown to HTML converter.
Squid convert to HTML a single markdown file or an entire project.
Additionally it copies eventual non-markdown files or assets found in the
project tree, preserving their relative position to the markdown files.

If a build destination is not specified, squid builds in 'build/' by default.

Usage:
    %s [OPTIONS] SOURCE [DESTINATION]

Options:
    -css string
        Specify a CSS file to use for styling.
    -h, --help
        Prints this help message and exits.
`, os.Args[0])
}
