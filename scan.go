package aozoraconv

import (
	"bufio"
	"bytes"
	"io"
)

func SplitCRLF(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) < 1 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\r'); 0 <= i {
		if i < (len(data)-1) && data[i+1] == '\n' {
			return i + 2, data[0 : i+2], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// NewTextScanner aozora text line-feed 'CRLF' scanner
func NewAozoraTextScanner(r io.Reader) *bufio.Scanner {
	scan := bufio.NewScanner(r)
	scan.Split(SplitCRLF)
	return scan
}

func NewDefaultTextScanner(r io.Reader) *bufio.Scanner {
	scan := bufio.NewScanner(r)
	return scan
}
