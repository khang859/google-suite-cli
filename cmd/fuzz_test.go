package cmd

import "testing"

func FuzzDecodeBase64URL(f *testing.F) {
	f.Add("aGVsbG8gd29ybGQ=")
	f.Add("dGVzdA")
	f.Add("")
	f.Add("!!!not-valid!!!")
	f.Add("SGVsbG8=")

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		decodeBase64URL(input)
	})
}

func FuzzInterpretEscapes(f *testing.F) {
	f.Add(`hello\nworld`)
	f.Add(`col1\tcol2`)
	f.Add(`path\\to\\file`)
	f.Add("plain text")
	f.Add(`hello\xworld`)
	f.Add(`trailing\`)
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		result := interpretEscapes(input)
		// Result should never be empty if input is non-empty
		// (unless input is only escape sequences that shrink)
		// At minimum: should not panic
		_ = result
	})
}

func FuzzTruncateSnippet(f *testing.F) {
	f.Add("hello world", 5)
	f.Add("short", 10)
	f.Add("", 5)
	f.Add("abcdef", 3)
	f.Add("exact", 5)

	f.Fuzz(func(t *testing.T, s string, maxLen int) {
		if maxLen < 0 {
			return // skip negative maxLen, not a valid input
		}
		result := truncateSnippet(s, maxLen)
		if len(result) > maxLen && maxLen >= 3 {
			t.Errorf("truncateSnippet(%q, %d) = %q (len %d), exceeds maxLen",
				s, maxLen, result, len(result))
		}
	})
}
