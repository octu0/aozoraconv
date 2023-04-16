package aozoraconv

import (
	"bytes"
	"regexp"
)

var (
	rubyIndex       = regexp.MustCompile(`([^｜]+?)([｜])([^《]+)《([^》]+)》`)
	rubyQuote       = regexp.MustCompile(`([^《]+)《([^》]+)》`)
	annoteOpenClose = regexp.MustCompile(`(［＃([^］]+)］)(.*)(［＃([^］]+)］)`)
	annoteSingle    = regexp.MustCompile(`(［＃([^］]+)］)(.*)`)
)

var (
	headerPattern = regexp.MustCompile(`(?s)^(.*)\r?\n-------------------------------------------------------\r?\n【テキスト中に現れる記号について】\r?\n(.*)\r?\n-------------------------------------------------------\r?\n$`)
	footerPattern = regexp.MustCompile(`(?s)^\r?\n\r?\n底本：`)
)

type Escaper interface {
	Escape(string) (out string, continues bool)
}

var (
	_ Escaper = (*noopEscaper)(nil)
)

type noopEscaper struct{}

func (*noopEscaper) Escape(s string) (string, bool) {
	return s, true
}

var (
	_ Escaper = (*rubyEscaper)(nil)
)

type rubyEscaper struct {
	reIndex *regexp.Regexp
	reQuote *regexp.Regexp
}

func (e *rubyEscaper) Escape(src string) (string, bool) {
	idxMatches := e.reIndex.FindStringSubmatch(src)
	if 5 <= len(idxMatches) {
		return idxMatches[1] + idxMatches[3], true
	}
	quoteMatches := e.reQuote.FindStringSubmatch(src)
	if 3 <= len(quoteMatches) {
		return quoteMatches[1], true
	}
	return src, true
}

func newRubyEscaper() *rubyEscaper {
	return &rubyEscaper{
		reIndex: rubyIndex,
		reQuote: rubyQuote,
	}
}

type annotationEscaper struct {
	openclose *regexp.Regexp
	single    *regexp.Regexp
}

func (e *annotationEscaper) Escape(src string) (string, bool) {
	ocMatches := e.openclose.FindStringSubmatch(src)
	if 6 <= len(ocMatches) {
		return ocMatches[3], true
	}
	siMatches := e.single.FindStringSubmatch(src)
	if 4 <= len(siMatches) {
		return siMatches[3], true
	}
	return src, true
}

func newAnnotationEscaper() *annotationEscaper {
	return &annotationEscaper{
		openclose: annoteOpenClose,
		single:    annoteSingle,
	}
}

var (
	_ Escaper = (*headerEscaper)(nil)
)

type headerEscaper struct {
	re *regexp.Regexp
}

func (e *headerEscaper) Escape(src string) (string, bool) {
	m := e.re.FindStringSubmatch(src)
	if 3 <= len(m) {
		return m[1], true
	}
	return "", false
}

func newHeaderEscaper() *headerEscaper {
	return &headerEscaper{
		re: headerPattern,
	}
}

var (
	_ Escaper = (*bufferEscaper)(nil)
)

type bufferEscaper struct {
	buf         *bytes.Buffer
	he          Escaper
	footer      *regexp.Regexp
	foundHeader bool
	foundFooter bool
	ce          Escaper
}

func (e *bufferEscaper) Escape(src string) (string, bool) {
	if e.foundHeader != true {
		e.buf.WriteString(src)
		out, ok := e.he.Escape(e.buf.String())
		if ok != true {
			return "", false
		}
		e.foundHeader = true
		e.buf.Reset()
		return out, true
	}

	return e.ce.Escape(src)
}

func newBufferEscaper(h Escaper, chain []Escaper) *bufferEscaper {
	return &bufferEscaper{
		buf:         bytes.NewBuffer(make([]byte, 0, 4*1024)),
		he:          h,
		footer:      footerPattern,
		foundHeader: false,
		foundFooter: false,
		ce:          &chainEscaper{chain},
	}
}

var (
	_ Escaper = (*chainEscaper)(nil)
)

type chainEscaper struct {
	chain []Escaper
}

func (e *chainEscaper) Escape(src string) (string, bool) {
	out := src
	for _, c := range e.chain {
		esc, ok := c.Escape(out)
		if ok != true {
			return out, false
		}
		out = esc
	}
	return out, true
}

func NewEscape(opt *option) Escaper {
	if opt.Header == nil && opt.Ruby == nil && opt.Annotation == nil {
		return new(noopEscaper)
	}

	chain := make([]Escaper, 0, 2)
	if opt.Ruby != nil {
		chain = append(chain, opt.Ruby)
	}
	if opt.Annotation != nil {
		chain = append(chain, opt.Annotation)
	}

	if opt.Header != nil {
		return newBufferEscaper(opt.Header, chain)
	}

	return &chainEscaper{chain}
}
