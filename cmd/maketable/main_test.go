package main

import (
	"testing"
)

func TestParseLine(test *testing.T) {
	var err error
	var s string
	m, k, t, uni, uni2 := 0, 0, 0, int32(0), int32(0)
	s = "3-2121\tU+3000\t# IDEOGRAPHIC SPACE"
	err = ParseLine(s, &m, &k, &t, &uni, &uni2)
	if m != 3 || k != 0x21 || t != 0x21 || uni != 0x3000 || uni2 != 0 {
		test.Errorf("AozoraConv: m,k,t,uni,uni2: %v,%X,%X,%X,%X\nerr:%q\ns:%q",
			m, k, t, uni, uni2, err, s)
	}

	s = "3-2477	U+304B+309A	# 	[2000]"
	err = ParseLine(s, &m, &k, &t, &uni, &uni2)
	if m != 3 || k != 0x24 || t != 0x77 || uni != 0x304B || uni2 != 0x309A {
		test.Errorf("AozoraConv: m,k,t,uni,uni2: %v,%X,%X,%X,%X\nerr:%q\ns:%q",
			m, k, t, uni, uni2, err, s)
	}
}
