package traefik_umami_plugin

import (
	"strconv"
	"strings"
)

type Encoding struct {
	name string
	q    float64
}

func ParseEncoding(encoding string) *Encoding {
	return &Encoding{
		name: encoding,
		q:    1.0,
	}
}

type Encodings struct {
	encodings []Encoding
}

func ParseEncodings(acceptEncoding string) *Encodings {
	encodingList := strings.Split(acceptEncoding, ",")
	result := make([]Encoding, 0, len(encodingList))

	for _, encoding := range encodingList {
		split := strings.Split(strings.TrimSpace(encoding), ";q=")
		q := 1.0
		if len(split) > 1 {
			q, _ = strconv.ParseFloat(split[1], 64)
		}
		result = append(result, Encoding{name: split[0], q: q})
	}

	return &Encodings{encodings: result}
}

func (ae *Encodings) String() string {
	result := make([]string, 0, len(ae.encodings))

	for _, encoding := range ae.encodings {
		result = append(result, encoding.name)
	}

	return strings.Join(result, ",")
}

func (ae *Encodings) FilterSupported() *Encodings {
	result := make([]Encoding, 0, len(ae.encodings))

	for _, encoding := range ae.encodings {
		switch encoding.name {
		case Gzip, Deflate, Identity:
			result = append(result, encoding)
		}
	}

	return &Encodings{encodings: result}
}

func (ae *Encodings) GetPreferred() *Encoding {
	maxQ := 0.0
	var preferred Encoding
	for _, encoding := range ae.encodings {
		if encoding.q > maxQ {
			preferred = encoding
			maxQ = encoding.q
		}
	}

	return &preferred
}
