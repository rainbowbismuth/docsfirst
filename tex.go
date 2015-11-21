package docsfirst

import "fmt"

type TexRewriter struct{}

func (_ *TexRewriter) Regex() string {
	return "% *DOCSFIRST *(.*)"
}

func (_ *TexRewriter) Match(blocks []*Block, out chan<- string) {
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
}
