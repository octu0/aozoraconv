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
	repeatTwo       = regexp.MustCompile(`([^／]{2})(／＼)`)
)

var (
	headerPattern = regexp.MustCompile(`(?s)^(.*)\r?\n-------------------------------------------------------\r?\n【テキスト中に現れる記号について】\r?\n(.*)\r?\n-------------------------------------------------------\r?\n$`)
	footerPattern = regexp.MustCompile(`^底本：`)
)

type Escaper interface {
	Escape(string) (out string, continues bool)
}

var (
	_ Escaper = (*noopEscaper)(nil)
	_ Escaper = (*rubyEscaper)(nil)
	_ Escaper = (*annotationEscaper)(nil)
	_ Escaper = (*repeatTwoEscaper)(nil)
	_ Escaper = (*headerEscaper)(nil)
	_ Escaper = (*bufferEscaper)(nil)
	_ Escaper = (*chainEscaper)(nil)
)

type noopEscaper struct{}

func (*noopEscaper) Escape(s string) (string, bool) {
	return s, true
}

type rubyEscaper struct {
	reIndex *regexp.Regexp
	reQuote *regexp.Regexp
}

func (e *rubyEscaper) Escape(src string) (string, bool) {
	if e.reIndex.MatchString(src) {
		src = e.reIndex.ReplaceAllString(src, "$1$3")
	}
	if e.reQuote.MatchString(src) {
		src = e.reQuote.ReplaceAllString(src, "$1")
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
	if e.openclose.MatchString(src) {
		return e.openclose.ReplaceAllString(src, "$3"), true
	}
	if e.single.MatchString(src) {
		return e.single.ReplaceAllString(src, "$3"), true
	}
	return src, true
}

func newAnnotationEscaper() *annotationEscaper {
	return &annotationEscaper{
		openclose: annoteOpenClose,
		single:    annoteSingle,
	}
}

type repeatTwoEscaper struct {
	re *regexp.Regexp
}

func (e *repeatTwoEscaper) Escape(src string) (string, bool) {
	if e.re.MatchString(src) {
		return e.re.ReplaceAllString(src, "$1$1"), true
	}
	return src, true
}

func newRepeatTwoEscaper() *repeatTwoEscaper {
	return &repeatTwoEscaper{
		re: repeatTwo,
	}
}

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

type bufferEscaper struct {
	buf          *bytes.Buffer
	he           Escaper
	footer       *regexp.Regexp
	footerStart1 bool
	footerStart2 bool
	footerStart3 bool
	foundHeader  bool
	foundFooter  bool
	ce           Escaper
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

	if e.foundFooter {
		return "", false
	}
	if src == "\r\n" {
		if e.footerStart1 != true {
			e.footerStart1 = true
			return src, true
		}
		if e.footerStart2 != true {
			e.footerStart2 = true
			return src, true
		}
		if e.footerStart3 != true {
			e.footerStart3 = true
			return src, true
		}
	}

	if e.footerStart1 && e.footerStart2 && e.footerStart3 {
		if e.footer.MatchString(src) {
			e.foundFooter = true
			return "", false
		}
	} else {
		e.footerStart1 = false
		e.footerStart2 = false
		e.footerStart3 = false
	}
	return e.ce.Escape(src)
}

func newBufferEscaper(h Escaper, chain []Escaper) *bufferEscaper {
	return &bufferEscaper{
		buf:          bytes.NewBuffer(make([]byte, 0, 4*1024)),
		he:           h,
		footer:       footerPattern,
		footerStart1: false,
		footerStart2: false,
		footerStart3: false,
		foundHeader:  false,
		foundFooter:  false,
		ce:           &chainEscaper{chain},
	}
}

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
	if opt.RepeatTwo != nil {
		chain = append(chain, opt.RepeatTwo)
	}

	if opt.Header != nil {
		return newBufferEscaper(opt.Header, chain)
	}

	return &chainEscaper{chain}
}
