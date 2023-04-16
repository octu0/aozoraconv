package aozoraconv

import (
	"reflect"
	"strings"
	"testing"
)

func TestAozoraTextScanner(t *testing.T) {
	tests := []struct {
		in     string
		expect []string
	}{
		{
			in:     "\r\ntest1\r\ntest2",
			expect: []string{"\r\n", "test1\r\n", "test2"},
		},
		{
			in:     "test1\r\ntest2\r\n",
			expect: []string{"test1\r\n", "test2\r\n"},
		},
		{
			in:     "test1\ntest2\n",
			expect: []string{"test1\ntest2\n"},
		},
	}
	for _, tc := range tests {
		s := NewAozoraTextScanner(strings.NewReader(tc.in))
		result := []string{}
		for s.Scan() {
			result = append(result, s.Text())
		}
		if testing.Verbose() {
			for i, r := range result {
				t.Logf("[%d] '%v'(%d)", i, []byte(r), len(r))
			}
		}
		if reflect.DeepEqual(tc.expect, result) != true {
			t.Errorf("expect=%v actual=%v", tc.expect, result)
		}
	}
}
