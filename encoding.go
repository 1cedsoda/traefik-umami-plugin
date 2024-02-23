package traefik_umami_plugin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Encoding struct {
	name string
	q    float64
}

const encodingRegex = `(?:([a-z*]+)(?:\;?q=(\d(?:\.\d)?))?)`

var encodingRegexp = regexp.MustCompile(encodingRegex)

var IdentityEncoding = &Encoding{name: Identity, q: 1.0}

var IdentityEncodings = &Encodings{encodings: []Encoding{*IdentityEncoding}}

var SupportedEncodingNames = []string{Identity, Gzip, Deflate}

func GetSupportedEncodings(q float64) *Encodings {
	result := make([]Encoding, 0, len(SupportedEncodingNames))
	for _, name := range SupportedEncodingNames {
		result = append(result, Encoding{name: name, q: q})
	}

	return &Encodings{encodings: result}
}

func ParseEncoding(encoding string) (*Encoding, error) {
	matches := encodingRegexp.FindStringSubmatch(encoding)

	if len(matches) < 2 {
		return nil, fmt.Errorf("no matches")
	}

	// get q
	q := 1.0
	if matches[2] != "" {
		// if is float
		if strings.Contains(matches[2], ".") {
			q, _ = strconv.ParseFloat(matches[2], 64)
		} else {
			// if is int
			qInt, _ := strconv.ParseInt(matches[2], 10, 64)
			q = float64(qInt) / 100
		}
	}

	return &Encoding{
		name: matches[1],
		q:    q,
	}, nil
}

type Encodings struct {
	encodings []Encoding
}

func ParseEncodings(acceptEncoding string) *Encodings {
	// remove any spaces
	acceptEncoding = strings.ReplaceAll(acceptEncoding, " ", "")

	// split by separator
	encodingStringList := strings.Split(acceptEncoding, ",")

	// parse
	result := make([]Encoding, 0, len(encodingStringList))
	for _, encodingString := range encodingStringList {
		if encodingString == "" {
			continue
		}

		encoding, err := ParseEncoding(encodingString)
		if err != nil {
			continue
		}

		// if * save index
		if encoding.name == "*" {
			// append all supported with asterisk q
			for _, name := range SupportedEncodingNames {
				result = append(result, Encoding{name: name, q: encoding.q})
			}
		}
		result = append(result, *encoding)
	}

	if len(result) == 0 {
		return IdentityEncodings
	}

	return &Encodings{encodings: result}
}

func (ae *Encodings) String() string {
	result := make([]string, 0, len(ae.encodings))

	for _, encoding := range ae.encodings {
		result = append(result, encoding.name)
	}

	return strings.Join(result, ", ")
}

func (ae *Encodings) FilterSupported() *Encodings {
	result := make([]Encoding, 0, len(ae.encodings))

	for _, encoding := range ae.encodings {
		switch encoding.name {
		// * added all encodings in ParseEncodings, we still keep it here to prevent unexpected behavior
		case Gzip, Deflate, Identity, "*":
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
