// Copyright (C) 2021 Red Hat, Inc.
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

package autodiscover_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config/autodiscover"
)

const (
	csvFile = "csv.json"
)

var (
	csvFilePath = path.Join(filePath, csvFile)
)

func loadCSVResource(filePath string) (csv autodiscover.CSVResource) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error (%s) loading CSVResource %s for testing", err, filePath)
	}
	err = json.Unmarshal(contents, &csv)
	if err != nil {
		log.Fatalf("error (%s) loading CSVResource %s for testing", err, filePath)
	}
	return
}

func TestCSVGetAnnotationValue(t *testing.T) {
	csv := loadCSVResource(csvFilePath)
	var val []string

	err := csv.GetAnnotationValue("notPresent", &val)
	assert.Equal(t, 0, len(val))
	assert.NotNil(t, err)

	err = csv.GetAnnotationValue("test-network-function.com/operator_tests", &val)
	assert.Equal(t, []string{"OPERATOR_STATUS", "ANOTHER_TEST"}, val)
	assert.Nil(t, err)
}
