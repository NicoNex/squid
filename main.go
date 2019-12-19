package main

import (
    "os"
    "io"
    "fmt"
    "sync"
    "flag"
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

    ofile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        printErr(err)
        return
    }
    defer ofile.Close()
    ofile.WriteString(html)
}

// Recursively walks in a directory tree.
func walkDir(root string) {
	files, err := readDir(root)
	if err != nil {
		printErr(err)
		return
	}

	for _, finfo := range files {
        fname := finfo.Name()
        fpath := fmt.Sprintf("%s%s", root, fname)
        buildDir := fmt.Sprintf("%s%s", buildRoot, root)

        if finfo.IsDir() {
            fpath += "/"
            tmp := fmt.Sprintf("%s%s", buildRoot, fpath)
            fmt.Printf("creating directory: %s\n", tmp)
            if err := os.MkdirAll(tmp, 0700); err != nil {
                printErr(err)
            }
            walkDir(fpath)
        } else {
            if isMarkdown(fname) {
                htmlPath := fmt.Sprintf("%s%s.html", buildDir, fname[:len(fname)-3])
                fmt.Printf("creating file: %s\n", htmlPath)
                wg.Add(1)
                go render(fpath, htmlPath)
            } else {
                destPath := fmt.Sprintf("%s%s", buildDir, fname)
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
    var ret = name
    if ret[:2] == "./" {
        ret = ret[2:]
    }
    if ret[len(ret)-1] != '/' {
        ret += "/"
    }
    return ret
}

func main() {
    var args []string
    var srcdir string

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

    args = flag.Args()
    if len(args) > 0 {
        srcdir = args[0]
    } else {
        die("Please provide a source directory")
    }

    buildRoot = sanitise(buildRoot)
    srcdir = sanitise(srcdir)
    walkDir(srcdir)
    wg.Wait()
}
