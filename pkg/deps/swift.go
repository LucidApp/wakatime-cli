package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers/s"
)

// nolint:noglobals
var swiftExcludeRegex = regexp.MustCompile(`(?i)^foundation$`)

// StateSwift is a token parsing state.
type StateSwift int

const (
	// StateSwiftUnknown represents unknown token parsing state.
	StateSwiftUnknown StateSwift = iota
	// StateSwiftImport means we are in hash section during token parsing.
	StateSwiftImport
)

// ParserSwift is a dependency parser for the swift programming language.
// It is not thread safe.
type ParserSwift struct {
	State  StateSwift
	Output []string
}

// Parse parses dependencies from Swift file content using the chroma Swift lexer.
func (p *ParserSwift) Parse(filepath string) ([]string, error) {
	reader, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer reader.Close()

	p.init()
	defer p.init()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := s.Swift.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserSwift) append(dep string) {
	dep = strings.TrimSpace(dep)

	if swiftExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserSwift) init() {
	p.State = StateSwiftUnknown
	p.Output = nil
}

func (p *ParserSwift) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordDeclaration:
		p.processKeywordDeclaration(token.Value)
	case chroma.NameClass:
		p.processNameClass(token.Value)
	}
}

func (p *ParserSwift) processKeywordDeclaration(value string) {
	switch value {
	case "import":
		p.State = StateSwiftImport
	default:
		p.State = StateSwiftUnknown
	}
}

func (p *ParserSwift) processNameClass(value string) {
	switch p.State {
	case StateSwiftImport:
		p.append(value)
	default:
		p.State = StateSwiftUnknown
	}
}
