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
	"regexp"
)

type Language struct {
	FileEndingRegex string
	LineComment     string
	MintedLanguage  string
}

// BEGIN Define Block
type Block struct {
	Language    *Language
	FileName    string
	StartLine   int
	Indentation string
	Tag         string
	Description string
	Body        []string
}

// END

// BEGIN Regex Constants
const (
	SPACE     = `(\s*)`
	BEGIN     = ` *BEGIN(\([^\)]*\))? *(.*)`
	END       = " *END"
	DOCSFIRST = "% *DOCSFIRST *(.*)"
)

// BEGIN(ParseBlocks) Define ParseBlocks
func ParseBlocks(
	lang *Language, fileName string, in <-chan string) <-chan *Block {
	out := make(chan *Block, 64)
	go func() {
		defer close(out)
		// BEGIN(ParseBlocks) Initialize block parsing state
		lineNumber := 0
		beginRegex := regexp.MustCompile(SPACE + lang.LineComment + BEGIN)
		endRegex := regexp.MustCompile(SPACE + lang.LineComment + END)
		startedAtLine := 0
		var curIndentation string
		var curDescription string
		var curTag string
		var curBody []string

		// BEGIN(ParseBlocks) Start parsing blocks line by line
		for line := range in {
			lineNumber++
			if curDescription != "" {
				// BEGIN(ParseBlocks) Parsing a block and we find the start of another
				strings := beginRegex.FindStringSubmatch(line)
				if strings != nil {
					out <- &Block{
						Language:    lang,
						FileName:    fileName,
						StartLine:   startedAtLine,
						Indentation: curIndentation,
						Tag:         curTag,
						Description: curDescription,
						Body:        curBody}
					startedAtLine = lineNumber
					curIndentation = strings[1]
					curTag = strings[2]
					curDescription = strings[3]
					curBody = nil
					continue
				}
				// BEGIN(ParseBlocks) Parsing a block and we find an end block marker
				strings = endRegex.FindStringSubmatch(line)
				if strings != nil {
					out <- &Block{
						Language:    lang,
						FileName:    fileName,
						StartLine:   startedAtLine,
						Indentation: curIndentation,
						Tag:         curTag,
						Description: curDescription,
						Body:        curBody}
					curDescription = ""
					curBody = nil
					continue
				}
				// BEGIN(ParseBlocks) If no match, append line to body
				curBody = append(curBody, line)
			} else {
				// BEGIN(ParseBlocks) Start parsing a new block
				strings := beginRegex.FindStringSubmatch(line)
				if strings != nil {
					startedAtLine = lineNumber
					curIndentation = strings[1]
					curTag = strings[2]
					curDescription = strings[3]
					curBody = nil
					continue
				}
				// BEGIN(ParseBlocks) Handle dangling block ends
				strings = endRegex.FindStringSubmatch(line)
				if strings != nil {
					panic(fmt.Errorf(
						"Dangling end in %s at line %d", fileName, lineNumber))
				}
			}
		}
		// BEGIN(ParseBlocks) Check if the file ended while parsing a block
		if curDescription != "" {
			panic(fmt.Errorf("EOF in the middle of a block in %s", fileName))
		}
		// BEGIN(ParseBlocks) Define ParseBlocks
	}()
	return out
}

// END

func GatherBlockMap(in <-chan *Block) <-chan map[string][]*Block {
	out := make(chan map[string][]*Block, 1)
	go func() {
		defer close(out)
		blockMap := make(map[string][]*Block)
		for block := range in {
			if block.Tag != "" {
				blockMap[block.Tag] = append(blockMap[block.Tag], block)
			}
			blockMap[block.Description] = append(blockMap[block.Description], block)
		}
		out <- blockMap
	}()
	return out
}

func RewriteTex(blockMap map[string][]*Block, in <-chan string) (<-chan string, <-chan map[string]int) {
	out := make(chan string, 64)
	refcounts := make(chan map[string]int, 1)
	go func() {
		defer close(out)
		defer close(refcounts)
		counts := map[string]int{}
		docsFirstRegex := regexp.MustCompile(DOCSFIRST)
		for line := range in {
			strings := docsFirstRegex.FindStringSubmatch(line)
			if strings != nil {
				description := strings[1]
				counts[description]++
				blocks := blockMap[description]
				if blocks == nil {
					panic(fmt.Errorf("Missing block: %s", description))
				}
				firstBlock := blocks[0]
				out <- "\\phantomsection"
				out <- fmt.Sprintf(
					"\\label{lst:%s%d}",
					firstBlock.Description,
					firstBlock.StartLine)
				mintCommand := fmt.Sprint(
					"\\begin{minted}[tabsize=4]{",
					firstBlock.Language.MintedLanguage,
					"}")
				out <- "\\begin{defquote}"
				out <- fmt.Sprint(
					"\\noindent \\( \\ll \\) ",
					firstBlock.Description,
					" \\( \\gg \\enspace \\equiv \\) \\hfill ",
					firstBlock.FileName,
					":",
					firstBlock.StartLine)
				out <- mintCommand
				for _, bodyLine := range firstBlock.Body {
					out <- bodyLine
				}
				out <- "\\end{minted}"
				if len(blocks) > 1 {
					for i := 1; i < len(blocks); i++ {
						if blocks[i].Description == firstBlock.Description {
							out <- mintCommand
							for _, bodyLine := range blocks[i].Body {
								out <- bodyLine
							}
							out <- "\\end{minted}"
						} else {
							out <- fmt.Sprint(
								"\\mintinline[tabsize=4]{",
								firstBlock.Language.MintedLanguage,
								"}|",
								blocks[i].Indentation,
								firstBlock.Language.LineComment,
								"| \\( \\ll \\) ",
								blocks[i].Description,
								" \\( \\gg \\) \\hfill (\\ref{lst:",
								blocks[i].Description,
								blocks[i].StartLine,
								"}) \\\\")
						}
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

func CheckReferences(blockMap map[string][]*Block, refcounts map[string]int) {
	for desc, blocks := range blockMap {
		if refcounts[desc] == 0 && refcounts[blocks[0].Tag] == 0 {
			block := blockMap[desc][0]
			fn := block.FileName
			line := block.StartLine
			fmt.Printf(
				"WARNING: Block \"%s\" defined at %s:%d was never used\n",
				desc, fn, line)
		}
	}
}
