package utf8_test

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
	"github.com/hitzhangjie/codemaster/encoding/utf8/pb/hello"
	"github.com/stretchr/testify/require"
)

func Test_UTF8_Invalid(t *testing.T) {
	// supposing this data is received from external system
	//
	// In Go, when you convert a byte slice containing invalid UTF-8 sequences to a string,
	// the language does not raise an error because the conversion is essentially a byte-for-byte copy.
	// This means that the underlying byte representation is preserved, regardless of whether it
	// forms valid UTF-8 characters or not.
	b := []byte{'\xff'}
	bb := []byte(string(b))
	require.Equal(t, b, bb)

	// If the byte slice contains invalid UTF-8 sequences, those bytes will still be included in the
	// resulting string. However, when you attempt to process or manipulate this string (for example,
	// when iterating over it or decoding it), you may encounter issues because the invalid bytes do
	// not correspond to valid Unicode code points.
	s := "Hello \xFF World"
	fmt.Println(s)

	// Go provides mechanisms to handle invalid UTF-8 sequences. For instance, functions in the unicode/utf8
	// package can be used to check if a string is valid UTF-8 or to replace invalid sequences with a
	// replacement character (typically U+FFFD, known as the "replacement character")
	valid := utf8.Valid(b)
	require.False(t, valid)

	valid = utf8.ValidString(s)
	require.False(t, valid)

	// how to detect and replace invalid UTF-8 code point
	ss := "Hello \xFF World \xC0 \x80"
	ss = replaceInvalidUTF8([]byte(ss))
	fmt.Println(ss)
}

func replaceInvalidUTF8(input []byte) string {
	result := make([]rune, 0, len(input))

	for len(input) > 0 {
		r, size := utf8.DecodeRune(input)
		if r == utf8.RuneError && size == 1 {
			// If it's an invalid UTF-8 sequence, append U+FFFD
			result = append(result, []rune("\\U+FFFD")...) // U+FFFD
			input = input[1:]                              // Move to the next byte
		} else {
			// Valid UTF-8 sequence
			result = append(result, r)
			input = input[size:] // Move past the valid rune
		}
	}

	return string(result)
}

func Test_UTF8_GoProtobuf_String_Check_UTF8_OR_NOT(t *testing.T) {
	t.Run("marshal report error: msg.field contains invalid UTF-8 codepoint", func(t *testing.T) {
		req := &hello.HelloReq{
			Msg: "Hello \xFF World",
		}
		_, err := proto.Marshal(req)
		require.NotNil(t, err) // string field contains invalid UTF-8
	})

	t.Run("unmarshal report error: msg.field contains invalid UTF-8 codepoint", func(t *testing.T) {
		reqx := &hello.HelloReqX{
			Msg: []byte("Hello \xFF World"),
		}
		b, err := proto.Marshal(reqx)
		require.Nil(t, err)
		require.NotEmpty(t, b)

		req := &hello.HelloReq{}
		err = proto.Unmarshal(b, req)
		require.NotNil(t, err) // string field contains invalid UTF-8
	})
}
