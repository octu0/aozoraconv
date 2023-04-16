package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/octu0/aozoraconv"
)

func getOuput(path string) (output io.Writer, err error) {
	if path == "" {
		return os.Stdout, nil
	}
	output, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func getInput(path string, stdin bool) (input io.Reader, err error) {
	if stdin {
		return os.Stdin, nil
	}
	if path == "" {
		return os.Stdin, nil
	}
	input, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func main() {
	var (
		useSjis, useUtf8 bool
		useStdin         bool
		path, outpath    string
		encoding         string
	)

	flag.StringVar(&encoding, "e", "sjis", "set output encoding (sjis or utf8)")
	flag.BoolVar(&useSjis, "s", false, "convert from UTF-8 into Shift_JIS")
	flag.BoolVar(&useUtf8, "u", false, "convert from Shift_JIS into UTF-8")
	flag.StringVar(&outpath, "o", "", "output filename")
	flag.BoolVar(&useStdin, "stdin", false, "use standard input")
	flag.Parse()

	if useSjis && useUtf8 {
		log.Fatalf("only -s or -u can be enabled")
	}

	path = flag.Arg(0)

	input, err := getInput(path, useStdin)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	output, err := getOuput(outpath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if useUtf8 {
		encoding = "utf8"
	}
	if useSjis {
		encoding = "sjis"
	}

	options := []aozoraconv.OptionFunc{}
	options = append(options, aozoraconv.WithoutHeader())
	options = append(options, aozoraconv.WithoutRuby())
	options = append(options, aozoraconv.WithoutAnnotation())
	options = append(options, aozoraconv.WithoutRepeatTwo())

	switch strings.ToLower(encoding) {
	case "utf8", "utf-8":
		if err := aozoraconv.Encode(output, input, options...); err != nil {
			log.Fatalf("error: %+v", err)
		}
	case "sjis", "shift_jis":
		if err := aozoraconv.Decode(output, input, options...); err != nil {
			log.Fatalf("error: %+v", err)
		}
	default:
		log.Fatalf("require encoding args: -s (Shift_JIS) or -u (UTF-8) or -e sting")
	}
}
