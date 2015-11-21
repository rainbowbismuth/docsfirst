package docsfirst

import "fmt"

type GithubMarkdownRewriter struct{}

func (_ *GithubMarkdownRewriter) Regex() string {
	return "% *DOCSFIRST *(.*)"
}

func (_ *GithubMarkdownRewriter) Match(blocks []*Block, out chan<- string) {
	firstBlock := blocks[0]

	out <- fmt.Sprint("| ", firstBlock.Description, " | ", firstBlock.FileName, ":", firstBlock.StartLine, " |")
	out <- "| ----- | ----- |"
	out <- "```" + firstBlock.Language.GithubMarkdownLanguage
	for _, bodyLine := range firstBlock.Body {
		out <- bodyLine
	}
	out <- "```"
	if len(blocks) > 1 {
		for i := 1; i < len(blocks); i++ {
			if blocks[i].Description == firstBlock.Description {
				out <- "```" + firstBlock.Language.GithubMarkdownLanguage
				for _, bodyLine := range blocks[i].Body {
					out <- bodyLine
				}
				out <- "```"
			} else {
				out <- fmt.Sprint(
					"```",
					firstBlock.Language.MintedLanguage,
					"\n",
					blocks[i].Indentation,
					firstBlock.Language.LineComment,
					blocks[i].Description,
					"\n```")
			}
		}
	}
}
