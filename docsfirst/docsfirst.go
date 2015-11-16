// docsfirst
// Copyright (C) 2015  Emily A. Bellows
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import "github.com/rainbowbismuth/docsfirst"

func main() {
	language := &docsfirst.Language{
		FileEndingRegex: "*.go",
		LineComment:     "// ",
		Minted:          "\\begin{minted}{go}",
	}
	codeSrc := docsfirst.ReadLinesFromFile("docsfirst.go")
	texSrc := docsfirst.ReadLinesFromFile("book.tex")
	blocks := docsfirst.ParseBlocks(language, "docsfirst.go", codeSrc)
	blockMap := <-docsfirst.GatherBlockMap(blocks)
	linesOut := docsfirst.RewriteTex(blockMap, texSrc)
	docsfirst.WriteLinesToFile("out.tex", linesOut)
}
