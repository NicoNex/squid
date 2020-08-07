package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	bf "github.com/russross/blackfriday/v2"
	mdfmt "github.com/shurcooL/markdownfmt/markdown"
)

// Defines the extensions that are used
const exts = bf.NoIntraEmphasis | bf.Tables | bf.FencedCode | bf.Autolink |
	bf.Strikethrough | bf.SpaceHeadings | bf.BackslashLineBreak |
	bf.DefinitionLists | bf.Footnotes

// Defines the HTML rendering flags that are used
const flags = bf.UseXHTML | bf.Smartypants | bf.SmartypantsFractions |
	bf.SmartypantsDashes | bf.SmartypantsLatexDashes

// Returns a formatted markdown file.
func format(input []byte) []byte {
	b, _ := mdfmt.Process("", input, nil)
	return b
}

// Returns the html rendered from a markdown bytes array.
func renderHtml(md []byte) string {
	return string(bf.Run(
		format(md),
		bf.WithExtensions(bf.CommonExtensions|bf.NoEmptyLineBeforeBlock),
		bf.WithRenderer(
			bfchroma.NewRenderer(
				bfchroma.WithoutAutodetect(),
				bfchroma.ChromaOptions(html.WithLineNumbers(true)),
				bfchroma.Extend(
					bf.NewHTMLRenderer(bf.HTMLRendererParameters{Flags: flags}),
				),
				bfchroma.Style("solarized-light"),
			),
		),
		bf.WithExtensions(exts),
	))
}

func addStyle(in string, style string) string {
	return fmt.Sprintf("<!DOCTYPE html>\n<style>\n\t%s</style>\n%s", style, in)
}

func loadCSS(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
