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

var outDir string
var wg sync.WaitGroup
var converter *Converter

func printErr(e error) {
    fmt.Println(aurora.BrightRed(e).Bold())
}

// func check(e error) {
//     if e != nil {
//         fmt.Println(e)
//         os.Exit(1)
//     }
// }

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

func copyFile(src string, dst string) {
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

func convertMarkdown(src string, dst string) {
    defer wg.Done()
    md, err := ioutil.ReadFile(src)
    check(err)
    html := converter.Convert(md)
    html = converter.AddStyle(html)

    ofile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        printErr(err)
        return
    }
    defer ofile.Close()
    ofile.WriteString(html)
    ofile.Close()
}

func walkDir(root string) {
	files, err := readDir(root)
	if err != nil {
		printErr(err)
		return
	}

	for _, finfo := range files {
        fname := finfo.Name()
        path := fmt.Sprintf("%s/%s", root, fname)
        buildDir := fmt.Sprintf("%s%s", outDir, root[1:])

        if finfo.IsDir() {
            buildPath := fmt.Sprintf("%s%s", outDir, path[1:])
            fmt.Printf("creating directory: %s\n", buildPath)
            err := os.Mkdir(buildPath, 0700)
            if err != nil {
                printErr(err)
            }
            walkDir(path)
		} else {
            if isMarkdown(fname) {
                htmlPath := fmt.Sprintf("%s%s/%s.html", outDir, root[1:], fname[:len(fname)-3])
                fmt.Printf("creating file: %s\n", htmlPath)
                wg.Add(1)
                go convertMarkdown(path, htmlPath)
            } else {
                destPath := fmt.Sprintf("%s/%s", buildDir, fname)
                fmt.Printf("copying %s to %s\n", path, destPath)
                wg.Add(1)
                go copyFile(path, destPath)
            }
		}
	}
}

func main() {
    var args []string
    var srcdir = "."

    flag.StringVar(&outDir, "o", "./build", "Output directory")
    flag.Parse()

    args = flag.Args()
    if len(args) > 0 {
        srcdir = args[0]
    }
    converter = NewConverter()
    fmt.Println(srcdir, outDir)

    // if _, err := os.Stat(outDir); os.IsNotExist(err) {
    //     os.Mkdir(outDir, 0700)
    // }
    //
    // walkDir(srcdir)
    // wg.Wait()
}
