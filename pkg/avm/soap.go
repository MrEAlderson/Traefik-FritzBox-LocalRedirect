package avm

import (
	"bytes"
	"errors"
	"net"
	"encoding/xml"
	"io"
)

func parseGetExternalIPAddressResponse(data []byte) (net.IP, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))

	for {
		tok, err := dec.Token()

		if err == io.EOF || err != nil {
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		if start.Name.Local == "NewExternalIPAddress" {
			var v string
			if err := dec.DecodeElement(&v, &start); err != nil {
				return nil, err
			}

			ip := net.ParseIP(v)

			if ip == nil {
				return nil, errors.New("failed to parse soap response into IPv4")
			}

			return ip, nil
		}
	}
}

func parseGetExternalIPv6Address(data []byte) (net.IP, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var address string = ""

	for {
		tok, err := dec.Token()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		var v string

		if start.Name.Local == "NewValidLifetime" {
			if err := dec.DecodeElement(&v, &start); err != nil {
				return nil, err
			}

			if v == "0" {
				return nil, nil
			}
		} else if start.Name.Local == "NewExternalIPv6Address" {
			if err := dec.DecodeElement(&v, &start); err != nil {
				return nil, err
			}

			address = v
		}
	}

	if address == "" {
		return nil, errors.New("missing address")
	}

	ip := net.ParseIP(address)

	if ip == nil {
		return nil, errors.New("failed to parse soap response into IPv6")
	}

	return ip, nil
}

func parseGetIPv6Prefix(data []byte) (*net.IPNet, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var address string = ""
	var prefixLength string = ""

	for {
		tok, err := dec.Token()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		var v string

		if start.Name.Local == "NewValidLifetime" {
			if err := dec.DecodeElement(&v, &start); err != nil {
				return nil, err
			}

			if v == "0" {
				return nil, nil
			}
		} else if start.Name.Local == "NewIPv6Prefix" {
			if err := dec.DecodeElement(&v, &start); err != nil {
				return nil, err
			}

			address = v
		} else if start.Name.Local == "NewIPv6Prefix" {
			if err := dec.DecodeElement(&v, &start); err != nil {
				return nil, err
			}

			prefixLength = v
		}
	}

	if address == "" || prefixLength == "" {
		return nil, errors.New("missing address or prefixLength")
	}

	ip := net.ParseIP(address)

	if ip == nil {
		return nil, errors.New("failed to parse soap response into IPv6")
	}

	_, ipNet, err := net.ParseCIDR(address + "/" + prefixLength)

	if err != nil {
		return nil, err
	}

	return ipNet, nil
}
