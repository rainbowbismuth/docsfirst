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

package docsfirst

import (
	"testing"
	"time"
)

func TestParseBlocks(t *testing.T) {
	testLang := &Language{
		FileEndingRegex: "*.hs",
		LineComment:     "--",
		MintedLanguage:  "haskell",
	}
	in := make(chan string, 8)
	body := "main = putStrLn \"Hello, world!\""
	go func() {
		defer close(in)
		in <- "-- BEGIN HelloWorld"
		in <- body
		in <- "-- END"
	}()
	blocks := ParseBlocks(testLang, "main.hs", in)
	select {
	case block := <-blocks:
		if block == nil ||
			block.FileName != "main.hs" ||
			block.Body[0] != body ||
			block.StartLine != 1 ||
			block.Description != "HelloWorld" ||
			block.Language != testLang {
			t.Fatal(block)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out")
	}
}

func TestRefCounts(t *testing.T) {
	testLang := &Language{
		FileEndingRegex: "*.py",
		LineComment:     "#",
		MintedLanguage:  "python",
	}
	inSrc := make(chan string, 8)
	inTex := make(chan string, 8)
	go func() {
		defer close(inSrc)
		inSrc <- "# BEGIN Say hello world"
		inSrc <- "print(\"Hello, world!\")"
		inSrc <- "# BEGIN Say goodbye world"
		inSrc <- "print(\"Goodbye, world!\")"
		inSrc <- "# END"
	}()
	go func() {
		defer close(inTex)
		inTex <- "%DOCSFIRST Say hello world"
		inTex <- "%DOCSFIRST Say goodbye world"
		inTex <- "%DOCSFIRST Say hello world"
	}()
	blocks := ParseBlocks(testLang, "main.py", inSrc)
	blockMap := <-GatherBlockMap(blocks)
	_, refcount_future := RewriteTex(blockMap, inTex)
	refcounts := <-refcount_future
	if refcounts["Say hello world"] != 2 ||
		refcounts["Say goodbye world"] != 1 {
		t.Fatal(refcounts)
	}
}
