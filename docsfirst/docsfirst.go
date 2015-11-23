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

import (
	"github.com/rainbowbismuth/docsfirst"
	"flag"
	"fmt"
	"os"
	"sync"
	"regexp"
)

func findLanguage(filename string) *docsfirst.Language {
	languages := []*docsfirst.Language{&docsfirst.Language{
		FileEndingRegex:        ".*.go",
		LineComment:            "//",
		MintedLanguage:         "go",
		GithubMarkdownLanguage: "go",
	}}

	for _, language := range languages {
		found, err := regexp.MatchString(language.FileEndingRegex, filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Regex error in %s\n %s", language.FileEndingRegex, err)
			os.Exit(1)
		}
		if found {
			return language
		}
	}

	fmt.Fprintf(os.Stderr, "No file ending match on %s\n", filename)
	os.Exit(1)
	return nil
}


func main() {
	warnOnUnused := flag.Bool("warn", true, "warn on unused entries")
	inputDoc := flag.String("input", "", "input document")
	outputDoc := flag.String("output", "", "output document")
	flag.Parse()

	src := flag.Args()

	if *inputDoc == "" || *outputDoc == "" || src == nil {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	allBlocks := make(chan *docsfirst.Block, 64)
	var doneReadingSrc sync.WaitGroup
	doneReadingSrc.Add(len(src))

	go func() {
		doneReadingSrc.Wait()
		close(allBlocks)
	}()

	for _, filename := range src {
		language := findLanguage(filename)
		codeSrc := docsfirst.ReadLinesFromFile(filename)
		blocks := docsfirst.ParseBlocks(language, filename, codeSrc)
		go func() {
			for block := range blocks {
				allBlocks <- block
			}
			doneReadingSrc.Done()
		}()
	}

	texSrc := docsfirst.ReadLinesFromFile(*inputDoc)
	blockMap := <-docsfirst.GatherBlockMap(allBlocks)
	linesOut, refCounts := docsfirst.Rewrite(blockMap, texSrc, &docsfirst.GithubMarkdownRewriter{})
	docsfirst.WriteLinesToFile(*outputDoc, linesOut)

	if *warnOnUnused {
		docsfirst.CheckReferences(blockMap, <-refCounts)
	}
}
