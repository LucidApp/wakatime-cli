package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/chroma"
)

// ParserGo is a dependency parser for the go programming language.
// It is not thread safe.
type ParserGo struct {
	Parenthesis int
	State       State
	Output      []string
}

// Parse parses dependencies from golang file content via ReadCloser using the chroma golang lexer.
func (p *ParserGo) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
	defer reader.Close()

	p.init()
	defer p.init()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := lexer.Tokenise(&chroma.TokeniseOptions{
		State:    "root",
		EnsureLF: true,
	}, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserGo) init() {
	p.Output = []string{}
	p.Parenthesis = 0
}

func (p *ParserGo) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordNamespace:
		p.processNamespace(token.Value)
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	case chroma.LiteralString:
		p.processString(token.Value)
	case chroma.Text:
		p.processText(token.Value)
	}
}

func (p *ParserGo) processNamespace(value string) {
	p.Parenthesis = 0

	switch value {
	case "import":
		p.State = StateImport
	default:
		p.State = StateUnknown
	}
}

func (p *ParserGo) processPunctuation(value string) {
	switch value {
	case "(":
		p.Parenthesis++
	case ")":
		p.Parenthesis--
	}
}

func (p *ParserGo) processText(value string) {
	if p.State == StateImport {
		if value == "\n" && p.Parenthesis <= 0 {
			p.State = StateUnknown
			p.Parenthesis = 0
		}
	} else {
		p.State = StateUnknown
	}
}

func (p *ParserGo) processString(value string) {
	if p.State == StateImport {
		p.Output = append(p.Output, strings.TrimSpace(value))
	}
}
