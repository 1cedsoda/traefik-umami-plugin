package traefik_umami_plugin

import (
	"testing"
)

func TestParseEncodings(t *testing.T) {
	acceptEncodingStrings := []string{
		"gzip, deflate,br",
		"",
		"identity;q=0.5, *;q=0",
		"gzip, deflate, br;q=0.5",
	}
	expectedList := []Encodings{
		{encodings: []Encoding{{name: "gzip", q: 1.0}, {name: "deflate", q: 1.0}, {name: "br", q: 1.0}}},
		{encodings: []Encoding{{name: "identity", q: 1.0}}},
		{encodings: []Encoding{{name: "identity", q: 0.5}, {name: "identity", q: 0.0}, {name: "gzip", q: 0.0}, {name: "deflate", q: 0.0}, {name: "*", q: 0.0}}},
		{encodings: []Encoding{{name: "gzip", q: 1.0}, {name: "deflate", q: 1.0}, {name: "br", q: 0.5}}},
	}
	for i, acceptEncodingString := range acceptEncodingStrings {
		result := ParseEncodings(acceptEncodingString)
		expected := &expectedList[i]
		AssertEncodingsEquals(result, expected, t)
	}
}

func AssertEncodingEquals(a, b *Encoding, t *testing.T) {
	if a.name != b.name {
		t.Errorf("name does not match: %+v != %+v", a, b)
		return
	}
	if a.q != b.q {
		t.Errorf("q does not match: %+v != %+v", a, b)
		return
	}
}

func AssertEncodingsEquals(a, b *Encodings, t *testing.T) {
	if len(a.encodings) != len(b.encodings) {
		t.Errorf("len does not match: %+v != %+v", a, b)
		return
	}
	for i := range a.encodings {
		AssertEncodingEquals(&a.encodings[i], &b.encodings[i], t)
	}
}
