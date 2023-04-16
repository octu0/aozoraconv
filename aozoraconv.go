package aozoraconv

import (
	"io"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var (
	aozoraCharMap = []string{
		"\u2014", "\u2015", // "—"
		"\u301C", "\uFF5E", // "〜"
		"\u2016", "\u2225", // "‖"
		"\u2212", "\uFF0D", // "−"
		"\u00A2", "\uFFE0", // "¢"
		"\u00A3", "\uFFE1", // "£"
		"\u00A5", "\uFFE5", // "¥"
		"\u00AC", "\uFFE2", // "¬"
	}
	aozoraUtf8CharReplacer  = strings.NewReplacer(aozoraCharMap...)
	aozoraUtf8CharReplacerR = strings.NewReplacer(reverse(aozoraCharMap)...)
)

// reverse reverses aozoraUtf8CharReplacer
func reverse(s []string) []string {
	r := make([]string, len(s))
	for i, v := range s {
		r[len(r)-i-1] = v
	}
	return r
}

// Conv replaces some characters in Unicode
func Conv(w io.Writer, r io.Reader, opts ...OptionFunc) error {
	option := newOption(opts...)
	esc := NewEscape(option)

	scan := NewAozoraTextScanner(r)
	for scan.Scan() {
		replaced := aozoraUtf8CharReplacer.Replace(scan.Text())
		newText, ok := esc.Escape(replaced)
		if ok != true {
			continue
		}
		if _, err := w.Write([]byte(newText)); err != nil {
			return errors.WithStack(err)
		}
	}
	if err := scan.Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ConvRev replaces some characters in Unicode
func ConvRev(w io.Writer, r io.Reader, opts ...OptionFunc) error {
	option := newOption(opts...)
	esc := NewEscape(option)

	scan := NewAozoraTextScanner(r)
	for scan.Scan() {
		text := scan.Text()
		newText, ok := esc.Escape(text)
		if ok != true {
			continue
		}
		replaced := aozoraUtf8CharReplacerR.Replace(newText)
		if _, err := w.Write([]byte(replaced)); err != nil {
			return errors.WithStack(err)
		}
	}
	if err := scan.Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Decode convert from UTF-8 into Aozora Bunko format (Shift_JIS)
func Decode(output io.Writer, input io.Reader, opts ...OptionFunc) (err error) {
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(input, decoder)
	if err := ConvRev(output, reader, opts...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Encode convert from Aozora Bunko format (Shift_JIS) into UTF-8
func Encode(output io.Writer, input io.Reader, opts ...OptionFunc) (err error) {
	encoder := japanese.ShiftJIS.NewEncoder()
	writer := transform.NewWriter(output, encoder)
	if err := Conv(writer, input, opts...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Jis2Uni returns a string from jis codepoint
func Jis2Uni(men, ku, ten int) (str string, err error) {
	if men < 1 || men > 2 || ku < 1 || ku > 94 || ten < 1 || ten > 94 {
		return "", errors.Errorf("error: args should be in 1..2, 1..94, 1..94")
	}
	chr := jis0213Decode[men-1][ku-1][ten-1]
	if chr == "" {
		return "", errors.Errorf("invalid access men: %v ku:%v ten:%v", men, ku, ten)
	}
	return chr, nil
}

// Uni2Jis returns a pointer of JisEntry
func Uni2Jis(str string) (jis JisEntry, err error) {
	var s1 uint16
	r := []rune(str)
	r1 := r[0]
	if len(r) == 1 {
		switch {
		case 0x20 <= r1 && r1 < 0x7f:
			return JisEntry{0, 0, 0}, errors.Errorf("ASCII character")
		case encode0Low <= r1 && r1 < encode0High:
			if s1 = encode0[r1-encode0Low]; (s1>>planeShift)&0x0003 > 0 {
				goto write2
			}
		case encode1Low <= r1 && r1 < encode1High:
			if s1 = encode1[r1-encode1Low]; (s1>>planeShift)&0x0003 > 0 {
				goto write2
			}
		case encode2Low <= r1 && r1 < encode2High:
			if s1 = encode2[r1-encode2Low]; (s1>>planeShift)&0x0003 > 0 {
				goto write2
			}
		case encode3Low <= r1 && r1 < encode3High:
			if s1 = encode3[r1-encode3Low]; (s1>>planeShift)&0x0003 > 0 {
				goto write2
			}
		case encode4Low <= r1 && r1 < encode4High:
			if s1 = encode4[r1-encode4Low]; (s1>>planeShift)&0x0003 > 0 {
				goto write2
			}
		}
		return JisEntry{0, 0, 0}, errors.Errorf("invalid character")
	write2:
		men := int8(s1 >> planeShift)
		ku := int8((s1 >> codeShift) & codeMask)
		ten := int8((s1) & codeMask)
		return JisEntry{men: men, ku: ku, ten: ten}, nil
	} else if len(r) == 2 {
		r2 := r[1]
		entry, ok := multichars[r1][r2]
		if !ok {
			return JisEntry{0, 0, 0}, err
		}
		return entry, nil
	}
	return JisEntry{0, 0, 0}, errors.Errorf("length of string should be 1 or 2")
}

// Is0208 checks triplet men-ku-ten is in JIS X 0208 or not
func Is0208(men, ku, ten int) bool {
	if men != 1 {
		return false
	}
	switch {
	case ku == 1:
		return 0x01 <= ten && ten <= 0x5e
	case ku == 2:
		return (0x01 <= ten && ten <= 0x0e) ||
			(0x1a <= ten && ten <= 0x1f) ||
			(0x20 <= ten && ten <= 0x21) ||
			(0x2a <= ten && ten <= 0x30) ||
			(0x3c <= ten && ten <= 0x4a) ||
			(0x52 <= ten && ten <= 0x59) ||
			(ten == 0x5e)
	case ku == 3:
		return (0x10 <= ten && ten <= 0x19) ||
			(0x21 <= ten && ten <= 0x3a) ||
			(0x41 <= ten && ten <= 0x5a)
	case ku == 4:
		return 0x01 <= ten && ten <= 0x53
	case ku == 5:
		return 0x01 <= ten && ten <= 0x56
	case ku == 6:
		return (0x01 <= ten && ten <= 0x18) ||
			(0x21 <= ten && ten <= 0x38)
	case ku == 7:
		return (0x01 <= ten && ten <= 0x21) ||
			(0x31 <= ten && ten <= 0x51)
	case ku == 8:
		return 0x01 <= ten && ten <= 0x20
	case 16 <= ku && ku <= 46:
		return 0x01 <= ten && ten <= 0x5e
	case ku == 47:
		return 0x01 <= ten && ten <= 0x33
	case 48 <= ku && ku <= 83:
		return 0x01 <= ten && ten <= 0x5e
	case ku == 84:
		return 0x01 <= ten && ten <= 0x06
	default:
		return false
	}
}

// Kuten2Sjis returns SJIS byte strings (2 byte) from ku-ten code
func Kuten2Sjis(ku, ten int) []byte {
	var seq, c1, c2, s1, s2 int
	seq = (ku-1)*94 + (ten - 1)
	c1 = seq / 188
	c2 = seq % 188
	if c1 < 31 {
		s1 = c1 + 129
	} else {
		s1 = c1 + 193
	}
	if c2 < 63 {
		s2 = c2 + 64
	} else {
		s2 = c2 + 65
	}
	return []byte{uint8(s1), uint8(s2)}
}
