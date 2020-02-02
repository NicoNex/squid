package main

import (
	"os"
	"io"
	"fmt"
	"sync"
	"flag"
	"regexp"
	"strings"
	"io/ioutil"

	"github.com/logrusorgru/aurora"
)

var buildRoot string
var stylefile string
var stylecont string
var wg sync.WaitGroup

func printErr(a interface{}) {
	fmt.Println(aurora.BrightRed(a).Bold())
}

func die(a interface{}) {
	fmt.Println(aurora.BrightRed(a).Bold())
	os.Exit(1)
}

// Takes as input an array of os.FileInfo and returns an array
// of non-hidden ones. (All files without '.' in front of the name)
func filterHidden(files []os.FileInfo) []os.FileInfo {
	var ret []os.FileInfo

	for _, f := range files {
		fname := f.Name()
		if fname[0] != '.' {
			ret = append(ret, f)
		}
	}
	return ret
}

// Returns an array of all the non-hidden files in a directory.
func readDir(filename string) ([]os.FileInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return []os.FileInfo{}, err
	}
	defer file.Close()
	files, err := file.Readdir(-1)
	if err != nil {
		return []os.FileInfo{}, err
	}

	return filterHidden(files), nil
}

func isMarkdown(fname string) bool {
	return fname[len(fname)-3:] == ".md"
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
	html := renderMarkdown(md)
	if stylefile != "" {
		html = addCustomStyle(html, stylecont)
	} else {
		html = addStyle(html)
	}

	err = ioutil.WriteFile(dst, []byte(html), 0666)
	if err != nil {
		printErr(err)
	}
}

func truncateFirstDir(path string) string {
	re := regexp.MustCompile(`(\w+)\/`)
	dirs := re.FindAllString(path, -1)
	return strings.Join(dirs[1:], "")
}

// Recursively walks in a directory tree.
func walkDir(root string) {
	var outdir string

	files, err := readDir(root)
	if err != nil {
		printErr(err)
		return
	}

	outdir = fmt.Sprintf("%s%s", buildRoot, truncateFirstDir(root))
	fmt.Printf("creating directory: %s\n", outdir)
	if err := os.MkdirAll(outdir, 0700); err != nil {
		die(err)
	}

	for _, finfo := range files {
		fname := finfo.Name()
		fpath := fmt.Sprintf("%s%s", root, fname)

		if finfo.IsDir() {
			walkDir(fpath+"/")
		} else {
			if isMarkdown(fname) {
				htmlPath := fmt.Sprintf("%s%s.html", outdir, fname[:len(fname)-3])
				fmt.Printf("creating file: %s\n", htmlPath)
				wg.Add(1)
				go render(fpath, htmlPath)
			} else {
				destPath := fmt.Sprintf("%s%s", outdir, fname)
				fmt.Printf("copying %s to %s\n", fpath, destPath)
				wg.Add(1)
				go copyFile(fpath, destPath)
			}
		}
	}
}

// Removes the './' from the beginning of file names and
// adds a '/' at the end if missing.
func sanitise(name string) string {
	if name != "." && name[:2] == "./" {
		name = name[2:]
	}
	if name[len(name)-1] != '/' {
		name += "/"
	}
	return name
}

func main() {
	var srcfile string

	flag.StringVar(&buildRoot, "o", "build/", "Output directory")
	flag.StringVar(&stylefile, "css", "", "CSS file")
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

	if flag.NArg() > 0 {
		srcfile = flag.Arg(0)
	} else {
		die("Please provide a valid source directory or file")
	}



	buildRoot = sanitise(buildRoot)

	srcinfo, err := os.Stat(srcfile)
	if err != nil {
		die(err)
	}
	if srcinfo.IsDir() {
		srcfile = sanitise(srcfile)
		walkDir(srcfile)
		wg.Wait()
	} else {
		if isMarkdown(srcinfo.Name()) {
			wg.Add(1)
			render(srcfile, srcfile+".html")
			wg.Wait()
		} else {
			die("Please provide a valid source directory or file")
		}
	}
}
