package aozoraconv

type OptionFunc func(*option)

type option struct {
	Header     Escaper
	Ruby       Escaper
	Annotation Escaper
	RepeatTwo  Escaper
}

func WithoutHeader() OptionFunc {
	return func(opt *option) {
		opt.Header = newHeaderEscaper()
	}
}

func WithoutRuby() OptionFunc {
	return func(opt *option) {
		opt.Ruby = newRubyEscaper()
	}
}

func WithoutAnnotation() OptionFunc {
	return func(opt *option) {
		opt.Annotation = newAnnotationEscaper()
	}
}

func WithoutRepeatTwo() OptionFunc {
	return func(opt *option) {
		opt.RepeatTwo = newRepeatTwoEscaper()
	}
}

func defaultOption() *option {
	return &option{
		Header:     nil,
		Ruby:       nil,
		Annotation: nil,
		RepeatTwo:  nil,
	}
}

func newOption(funcs ...OptionFunc) *option {
	opt := defaultOption()
	for _, fn := range funcs {
		fn(opt)
	}
	return opt
}
