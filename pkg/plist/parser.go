package plist

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

func ModifyXML(r io.Reader, key string, newValue string) (io.ReadWriter, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.New("error reading xml contents to buffer")
	}
	input := bytes.NewBuffer(buf)
	output, err := parseAndUpdate(input, key, newValue)
	if err != nil {
		return nil, errors.New("error parsing or updating xml file - " + err.Error())
	}
	return output, nil
}

func parseAndUpdate(buf *bytes.Buffer, key string, newValue string) (*bytes.Buffer, error) {
	output := bytes.NewBuffer([]byte{})
	encode := xml.NewEncoder(output)
	decoder := xml.NewDecoder(buf)
	found := false

Outer:
	for {
		token, _ := decoder.Token()
		if token == nil {
			if !found {
				return nil, errors.New("required key not found in xml contents")
			}
			break
		}
		switch node := token.(type) {
		case xml.StartElement:
			switch node.Name.Local {
			case "key":
				var k string
				if err := decodeElement(decoder, &k, &node); err != nil {
					return nil, err
				}
				if k == key {
					found = true
				}
				if err := encodeElement(encode, &k, node); err != nil {
					return nil, err
				}
			case "string":
				if found {
					var s string
					if err := decodeElement(decoder, &s, &node); err != nil {
						return nil, err
					}
					s = newValue
					if err := encodeElement(encode, &s, node); err != nil {
						return nil, err

					}
					break Outer
				}
				if err := encodeToken(encode, token); err != nil {
					return nil, err

				}
			default:
				if err := encodeToken(encode, token); err != nil {
					return nil, err

				}
			}
		default:
			if err := encodeToken(encode, token); err != nil {
				return nil, err

			}

		}

	}
	output.Write(buf.Bytes())
	return output, nil
}

func decodeElement(d *xml.Decoder, into *string, se *xml.StartElement) error {
	if err := d.DecodeElement(into, se); err != nil {
		fmt.Println("failed to decoding element " + se.Name.Local)
		return errors.New("error parsing xml element")
	}
	return nil
}

func encodeElement(e *xml.Encoder, value interface{}, se xml.StartElement) error {
	if err := e.EncodeElement(value, se); err != nil {
		fmt.Println("failed to encode element " + se.Name.Local)
		return errors.New("error writing parsed xml node")
	}
	return nil
}

func encodeToken(e *xml.Encoder, t xml.Token) error {
	if err := e.EncodeToken(t); err != nil {
		fmt.Println("failed to encode token")
		return errors.New("error writing parsed xml node")
	}
	return nil
}
