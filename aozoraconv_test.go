package aozoraconv

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// toSjis convert string into Shift_JIS ([]byte)
func toSjis(s string) []byte {
	outBuf := new(bytes.Buffer)
	reader := strings.NewReader(s)
	writer := transform.NewWriter(outBuf, japanese.ShiftJIS.NewEncoder())

	//Copy
	_, err := io.Copy(writer, reader)
	if err != nil {
		return nil
	}
	ret := outBuf.Bytes()
	return ret
}

func TestConv(t *testing.T) {
	var convertedStrings = []struct {
		in  string
		out string
	}{
		{"", ""},
		{"あ", "あ"},
		{"〜", "～"},
		{"‖", "∥"},
		{"∥", "∥"},
		{"¢", "￠"},
	}
	for _, tt := range convertedStrings {
		if got, want := Conv(tt.in), tt.out; got != want {
			t.Errorf("Conv: got %v want %v", got, want)
		}
	}
}

func TestConvR(t *testing.T) {
	var convertedStrings = []struct {
		in  string
		out string
	}{
		{"", ""},
		{"あ", "あ"},
		{"～", "〜"},
		{"\u301C", "\u301C"},
		{"\uFF5E", "\u301C"},
		{"∥", "‖"},
		{"‖", "‖"},
		{"￠", "¢"},
	}
	for _, tt := range convertedStrings {
		if got, want := ConvRev(tt.in), tt.out; got != want {
			t.Errorf("ConvRev: got %v want %v", got, want)
		}
	}
}

func TestSjisConv(t *testing.T) {
	encoder := japanese.ShiftJIS.NewEncoder()

	input := "\uFF5E∥￠\u2015"
	reader := transform.NewReader(strings.NewReader(input), encoder)
	ret, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("sjis: error %v", err)
	}
	s := ""
	for _, c := range ret {
		s += fmt.Sprintf("%x", c)
	}
	want := "8160" + "8161" + "8191" + "815c"
	if s != want {
		t.Errorf("sjis: got %v want %v", s, want)
	}
}

func TestSjisConv2(t *testing.T) {
	input := "\uFF5E∥￠\u2015"
	out := new(bytes.Buffer)

	writer := transform.NewWriter(out, japanese.ShiftJIS.NewEncoder())
	n, err := writer.Write([]byte(input))
	if err != nil {
		t.Errorf("sjis: in column %v, val: %v, 'U+%x'", n, err, []rune(input[n:])[0])
	}
	got := ""
	for _, c := range out.Bytes() {
		got += fmt.Sprintf("%x", c)
	}
	want := "8160" + "8161" + "8191" + "815c"
	if got != want {
		t.Errorf("sjis: got %v want %v", got, want)
	}
}

func TestSjisConv3(t *testing.T) {
	input := "\uFF5E∥￠\u2015 \u2014123"
	out := new(bytes.Buffer)

	writer := transform.NewWriter(out, japanese.ShiftJIS.NewEncoder())
	n, err := writer.Write([]byte(input))
	if n != 13 || err == nil {
		t.Errorf("should be fail: %v, val: %v, 'U+%x'", n, err, []rune(input[n:])[0])
	}
}

func TestEncode(t *testing.T) {
	var convertedPairs = []struct {
		in  string
		out []byte
	}{
		{"あいうえお", toSjis("あいうえお")},
		{"\u301C", toSjis("\uFF5E")},
		{"\uFF5E", toSjis("\uFF5E")},
		{"¢", toSjis("￠")},
	}
	for _, tt := range convertedPairs {
		input := strings.NewReader(tt.in)
		output := new(bytes.Buffer)

		Encode(input, output)
		if got, want := output.Bytes(), tt.out; bytes.Compare(got, want) != 0 {
			t.Errorf("Encode got: %v, want: %v", got, want)
		}
	}
}

func TestDecode(t *testing.T) {
	var convertedPairs = []struct {
		out string
		in  []byte
	}{
		{"あいうえお", toSjis("あいうえお")},
		{"\u301C", toSjis("\uFF5E")},
		{"¢", toSjis("￠")},
	}
	for _, tt := range convertedPairs {
		input := bytes.NewBuffer(tt.in)
		output := new(bytes.Buffer)

		Decode(input, output)
		if got, want := output.String(), tt.out; got != want {
			t.Errorf("Encode got: %v, want: %v", got, want)
		}
	}
}

func TestUni2Jis(t *testing.T) {
	var convertedPairs = []struct {
		in        string
		out       JisEntry
		isSuccess bool
	}{
		{"あ", JisEntry{men: 1, ku: 4, ten: 2}, true},
		{"。", JisEntry{men: 1, ku: 1, ten: 3}, true},
		{"◆", JisEntry{men: 1, ku: 2, ten: 1}, true},
		{"A", JisEntry{0, 0, 0}, false},
		{"☺", JisEntry{0, 0, 0}, false},
	}
	for _, tt := range convertedPairs {
		got, err := Uni2Jis(tt.in)
		if want := tt.out; got != want {
			t.Errorf("Uni2Jis got: %v, want: %v", got, tt.out)
		}
		if err != nil && tt.isSuccess {
			t.Errorf("Uni2Jis got: %v want: %v; should be error but %v", got, tt.out, err)
		} else if err == nil && !tt.isSuccess {
			t.Errorf("Uni2Jis got: %v want: %v; should be error but %v", got, tt.out, err)
		}
	}
}

func TestJis2Uni(t *testing.T) {
	var convertedPairs = []struct {
		men, ku, ten int
		out          string
		isSuccess    bool
	}{
		{1, 4, 2, "あ", true},
		{1, 1, 3, "。", true},
		{1, 2, 1, "◆", true},
		{2, 3, 17, "𠗖", true},
		{2, 2, 80, "", false},
		{3, 2, 10, "", false},
		{1, 2, 100, "", false},
	}
	for _, tt := range convertedPairs {
		got, err := Jis2Uni(tt.men, tt.ku, tt.ten)
		if err != nil && tt.isSuccess {
			t.Errorf("Jis2Uni got: %v want: %v; should not be error but %v", got, tt.out, err)
		} else if err == nil {
			if !tt.isSuccess {
				t.Errorf("Jis2Uni got: %v want: %v; should be error but %v", got, tt.out, err)
			}
			if want := tt.out; got != want {
				t.Errorf("Jis2Uni got: %v, want: %v", got, tt.out)
			}
		}
	}
}

func TestIs0208(t *testing.T) {
	var convertedPairs = []struct {
		men, ku, ten int
		isSuccess    bool
	}{
		{1, 4, 2, true},
		{1, 1, 3, true},
		{1, 2, 1, true},
		{2, 3, 17, false},
		{2, 2, 80, false},
		{3, 2, 10, false},
		{1, 2, 100, false},
		{0, 1, 1, false},

		{1, 0, 0, false},
		{1, 1, 0, false},
		{1, 1, 1, true},
		{1, 1, 93, true}, {1, 1, 94, true}, {1, 1, 95, false},
		{1, 2, 0, false}, {1, 2, 1, true},
		{1, 2, 14, true}, {1, 2, 15, false}, {1, 2, 16, false},
		{1, 2, 25, false}, {1, 2, 26, true},
		{1, 2, 33, true}, {1, 2, 34, false},
		{1, 2, 41, false}, {1, 2, 42, true},
		{1, 2, 48, true}, {1, 2, 49, false},
		{1, 2, 59, false}, {1, 2, 60, true},
		{1, 2, 74, true}, {1, 2, 75, false},
		{1, 2, 81, false}, {1, 2, 82, true},
		{1, 2, 89, true}, {1, 2, 90, false},
		{1, 2, 93, false}, {1, 2, 94, true}, {1, 2, 95, false},
		{1, 3, 1, false},
		{1, 3, 15, false}, {1, 3, 16, true},
		{1, 3, 25, true}, {1, 3, 26, false},
		{1, 3, 32, false}, {1, 3, 33, true},
		{1, 3, 58, true}, {1, 3, 59, false},
		{1, 3, 64, false}, {1, 3, 65, true},
		{1, 3, 90, true}, {1, 3, 91, false},
		{1, 3, 94, false},
		{1, 4, 1, true}, {1, 4, 83, true}, {1, 4, 84, false}, {1, 4, 94, false},
		{1, 5, 1, true},
		{1, 5, 86, true}, {1, 5, 87, false},
		{1, 5, 94, false},
		{1, 6, 1, true},
		{1, 6, 24, true}, {1, 6, 25, false},
		{1, 6, 32, false}, {1, 6, 33, true},
		{1, 6, 56, true}, {1, 6, 57, false},
		{1, 7, 1, true},
		{1, 7, 33, true}, {1, 7, 34, false},
		{1, 7, 48, false}, {1, 7, 49, true},
		{1, 7, 81, true}, {1, 7, 82, false},
		{1, 8, 1, true},
		{1, 8, 32, true},
		{1, 8, 33, false},
		{1, 9, 1, false}, {1, 15, 1, false},
		{1, 16, 1, true},
		{1, 16, 94, true}, {1, 16, 95, false},
		{1, 17, 1, true},
		{1, 47, 1, true},
		{1, 47, 51, true}, {1, 47, 52, false},
		{1, 48, 1, true},
		{1, 84, 1, true},
		{1, 84, 6, true}, {1, 84, 7, false},
	}
	for _, tt := range convertedPairs {
		got := Is0208(tt.men, tt.ku, tt.ten)
		if got != tt.isSuccess {
			t.Errorf("Is0208 %v,%v,%v, got: %v want: %v", tt.men, tt.ku, tt.ten, got, tt.isSuccess)
		}
	}
}

func TestKuten2Sjis(t *testing.T) {
	var convertedPairs = []struct {
		ku, ten int
		sjis    []byte
	}{
		{1, 1, []byte{0x81, 0x40}},   // "　"
		{2, 1, []byte{0x81, 0x9F}},   // "◆"
		{4, 2, []byte{0x82, 0xA0}},   // "あ"
		{16, 1, []byte{0x88, 0x9F}},  // "亜"
		{47, 52, []byte{0x98, 0x73}}, // only in JIS X 0213, not 0208
		{84, 6, []byte{0xEA, 0xA4}},
	}
	for _, tt := range convertedPairs {
		got, want := Kuten2Sjis(tt.ku, tt.ten), tt.sjis
		if bytes.Compare(got, want) != 0 {
			t.Errorf("Kuten2Sjis got: %v want: %v", got, want)
		}
	}

}

func TestAllKuten2SjisChars(t *testing.T) {
	buf := make([]byte, 0)
	for ku := 1; ku < 96; ku++ {
		for ten := 0; ten < 96; ten++ {
			if Is0208(1, ku, ten) {
				chrs := Kuten2Sjis(ku, ten)
				buf = append(buf, chrs[0], chrs[1])
			}
		}
	}
	got, want := len(buf), 13758
	if len(buf) != want {
		t.Errorf("kuten2sjis got: %v want: %v", got, want)
	}
}

func TestReverse(t *testing.T) {
	v := reverse([]string{"a", "b", "c", "d"})
	if (v[0] == "d" && v[1] == "c" && v[2] == "b" && v[3] == "a") != true {
		t.Errorf("actual %v", v)
	}
	v = reverse([]string{"a", "b", "c"})
	if (v[0] == "c" && v[1] == "b" && v[2] == "a") != true {
		t.Errorf("actual %v", v)
	}
}
