// BEGIN Copyright
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
// END

package docsfirst

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Language struct {
	FileEndingRegex string
	LineComment     string
	Minted          string
}

// BEGIN Define Block
type Block struct {
	Language    *Language
	FileName    string
	StartLine   int
	Description string
	Body        []string
}

// END
const (
	BEGIN     = " BEGIN "
	END       = " END"
	DOCSFIRST = "%DOCSFIRST "
)

func ParseBlocks(lang *Language, fileName string, in <-chan string) <-chan *Block {
	out := make(chan *Block, 64)
	go func() {
		defer close(out)
		lineNumber := 0
		startedAtLine := 0
		var curDescription string
		var curBody []string
		specificBegin := lang.LineComment + BEGIN
		fmt.Println(specificBegin)
		specificEnd := lang.LineComment + END
		fmt.Println(specificEnd)
		for line := range in {
			lineNumber++
			if curDescription != "" {
				if strings.HasPrefix(line, specificBegin) {
					out <- &Block{
						Language:    lang,
						FileName:    fileName,
						StartLine:   startedAtLine,
						Description: curDescription,
						Body:        curBody}
					startedAtLine = lineNumber
					curDescription = strings.Replace(line, specificBegin, "", 1)
					curBody = nil
				} else if strings.HasPrefix(line, specificEnd) {
					out <- &Block{
						Language:    lang,
						FileName:    fileName,
						StartLine:   startedAtLine,
						Description: curDescription,
						Body:        curBody}
					startedAtLine = -1
					curDescription = ""
					curBody = nil
				} else {
					curBody = append(curBody, line)
				}
			} else {
				if strings.HasPrefix(line, specificBegin) {
					startedAtLine = lineNumber
					curDescription = strings.Replace(line, specificBegin, "", 1)
					curBody = nil
				} else if strings.HasPrefix(line, specificEnd) {
					panic(fmt.Errorf("Dangling end in %s at line %s", fileName, lineNumber))
				}
			}
		}
		if curBody != nil || curDescription != "" {
			panic(fmt.Errorf("EOF in the middle of a block in %s", fileName))
		}
	}()
	return out
}

func GatherBlockMap(in <-chan *Block) <-chan map[string][]*Block {
	out := make(chan map[string][]*Block)
	go func() {
		defer close(out)
		blockMap := make(map[string][]*Block)
		for block := range in {
			slice := blockMap[block.Description]
			blockMap[block.Description] = append(slice, block)
		}
		out <- blockMap
	}()
	return out
}

func RewriteTex(blockMap map[string][]*Block, in <-chan string) (<-chan string, <-chan map[string]int) {
	out := make(chan string, 64)
	refcounts := make(chan map[string]int)
	go func() {
		defer close(out)
		defer close(refcounts)
		counts := map[string]int{}
		for line := range in {
			if strings.HasPrefix(line, DOCSFIRST) {
				description := strings.Replace(line, DOCSFIRST, "", 1)
				counts[description]++
				blocks := blockMap[description]
				if blocks == nil {
					panic(fmt.Errorf("Missing block: %s", description))
				}
				firstBlock := blocks[0]
				out <- "\\begin{defquote}"
				out <- fmt.Sprint("\\noindent \\( \\ll \\) ",
					firstBlock.Description,
					" \\( \\gg \\enspace \\equiv \\) \\hfill ",
					firstBlock.FileName,
					":",
					firstBlock.StartLine)
				out <- firstBlock.Language.Minted
				for _, bodyLine := range firstBlock.Body {
					out <- bodyLine
				}
				out <- "\\end{minted}"
				for i := 1; i < len(blocks); i++ {
					if blocks[i].Description == firstBlock.Description {
						out <- firstBlock.Language.Minted
						for _, bodyLine := range blocks[i].Body {
							out <- bodyLine
						}
						out <- "\\end{minted}"
					} else {
						out <- fmt.Sprint("\\quad \\( \\ll \\) ",
							blocks[i].Description,
							" \\( \\gg \\)")
					}
				}
				out <- "\\end{defquote}"
			} else {
				out <- line
			}
		}
		refcounts <- counts
	}()
	return out, refcounts
}

func ReadLinesFromFile(fileName string) <-chan string {
	out := make(chan string, 64)
	go func() {
		defer close(out)
		inFile, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			out <- scanner.Text()
		}
		err = scanner.Err()
		if err != nil {
			panic(err)
		}
	}()
	return out
}

func WriteLinesToFile(fileName string, in <-chan string) {
	outFile, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	for line := range in {
		_, err := outFile.WriteString(line)
		if err != nil {
			panic(err)
		}
		_, err = outFile.WriteString("\n")
		if err != nil {
			panic(err)
		}
	}
}
