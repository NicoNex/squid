/*
 * Squid
 * Copyright (C) 2019-2020  Nicolò Santamaria
 *
 * Squid is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Squid is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

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
				bfchroma.ChromaOptions(html.WithLineNumbers(false)),
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
	return fmt.Sprintf("<!DOCTYPE html>\n<style>\n%s\n</style>\n%s", style, in)
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
