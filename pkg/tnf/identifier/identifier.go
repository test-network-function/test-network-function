// Copyright (C) 2020-2022 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package identifier

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const (
	// semanticVersionKey is the JSON Key representing a SemanticVersion payload.
	semanticVersionKey = "version"
	// urlKey is the JSON key representing a URL payload.
	urlKey = "url"

	urlTests = "http://test-network-function.com/tests"
)

// Identifier is a per tnf.Test unique identifier.
type Identifier struct {
	// URL stores the unique identifier for a test.
	URL string `json:"url" yaml:"url"`
	// SemanticVersion stores the version of the test.
	SemanticVersion string `json:"version" yaml:"version"`
}

// unmarshalURL deciphers a URL and ensures that the URL parses correctly.
func (i *Identifier) unmarshalURL(objMap map[string]*json.RawMessage) error {
	if u, ok := objMap[urlKey]; ok {
		var strURL string
		if err := json.Unmarshal(*u, &strURL); err != nil {
			return err
		}
		if _, err := url.Parse(strURL); err != nil {
			return err
		}
		i.URL = strURL
		return nil
	}
	return fmt.Errorf("missing required field: \"%s\"", urlKey)
}

// unmarshalSemanticVersion deciphers a SemanticVersion and ensures the SemanticVersion parses correctly.
func (i *Identifier) unmarshalSemanticVersion(objMap map[string]*json.RawMessage) error {
	if s, ok := objMap[semanticVersionKey]; ok {
		var strSemanticVersion string
		if err := json.Unmarshal(*s, &strSemanticVersion); err != nil {
			return err
		}
		if _, err := semver.NewVersion(strSemanticVersion); err != nil {
			return err
		}
		i.SemanticVersion = strSemanticVersion
		return nil
	}
	return fmt.Errorf("missing required field: \"%s\"", semanticVersionKey)
}

// UnmarshalJSON provides a custom JSON Unmarshal function which performs URL and SemanticVersion validation.
func (i *Identifier) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	if err = i.unmarshalURL(objMap); err != nil {
		return err
	}

	return i.unmarshalSemanticVersion(objMap)
}

// XformToGinkgoItIdentifier transform the Identifier into a test Id that can be used to skip
// specific tests
func XformToGinkgoItIdentifier(identifier Identifier) string {
	return XformToGinkgoItIdentifierExtended(identifier, "")
}

// XformToGinkgoItIdentifierExtended transform the claim.Identifier into a test Id that can be used to skip
// specific tests
// returns empty string if identifierURL was not correct about the base domain
// for the url of tests
func XformToGinkgoItIdentifierExtended(identifier Identifier, extra string) string {
	if !strings.HasPrefix(identifier.URL, urlTests+"/") {
		return ""
	}
	itID := strings.ReplaceAll(strings.TrimPrefix(identifier.URL, urlTests+"/"), "/", "-")
	var key string
	if extra != "" {
		key = itID + "-" + extra
	} else {
		key = itID
	}
	return key
}

func GetIdentifierURLBaseDomain() string {
	return urlTests
}
