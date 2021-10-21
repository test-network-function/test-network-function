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

package nodenames

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	utils "github.com/test-network-function/test-network-function/pkg/utils"
)

const (
	nnRegex = "(?s).+"
)

// NodeNames holds information derived from running "oc get nodes" on the command line.
type NodeNames struct {
	nodeNames []string
	result    int
	timeout   time.Duration
	args      []string
}

// NewNodeNames creates a new NodeNames tnf.Test.
// labels is a map of label names and values for filtering nodes
// nil label value is like a wildcard - any value as long as the label exists
func NewNodeNames(timeout time.Duration, labels map[string]*string) *NodeNames {
	args := []string{"oc", "get", "nodes"}
	var labelsString string
	var findSchedulable bool
	for labelName, labelValue := range labels {
		if labelName == configsections.Schedulable {
			findSchedulable = true
			break
		}
		labelsString += labelName
		if labelValue != nil {
			labelsString += "=" + *labelValue
		}
		labelsString += ","
	}
	if findSchedulable {
		args = append(args, "-o", "'go-template={{range .items}}{{$taints:=\"\"}}{{range .spec.taints}}{{if eq .effect \"NoSchedule\"}}{{$taints = print $taints .key \",\"}}{{end}}{{end}}{{if not $taints}}{{.metadata.name}}{{ \"\\n\"}}{{end}}{{end}}'")
	} else if labelsString != "" {
		labelsString = labelsString[:len(labelsString)-1]
		args = append(args, "-o", "custom-columns=NAME:.metadata.name", "-l", labelsString, "| tail -n +2")
	}

	return &NodeNames{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    args,
	}
}

// GetNodeNames returns node names extracted from running the NodeNames tnf.Test.
func (nn *NodeNames) GetNodeNames() []string {
	return nn.nodeNames
}

// Args returns the command line args for the test.
func (nn *NodeNames) Args() []string {
	return nn.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (nn *NodeNames) GetIdentifier() identifier.Identifier {
	return identifier.NodeNamesIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (nn *NodeNames) Timeout() time.Duration {
	return nn.timeout
}

// Result returns the test result.
func (nn *NodeNames) Result() int {
	return nn.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (nn *NodeNames) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{nnRegex},
		Timeout: nn.timeout,
	}
}

// ReelMatch ensures that list of nodes is not empty and stores the names as []string
func (nn *NodeNames) ReelMatch(_, _, match string) *reel.Step {
	trimmedMatch := strings.Trim(match, "\n")
	nn.nodeNames = strings.Split(trimmedMatch, "\n") // First line is the headers/titles line

	nn.nodeNames = utils.DeleteEmpty(nn.nodeNames)

	if len(nn.nodeNames) == 0 {
		nn.result = tnf.FAILURE
	} else {
		nn.result = tnf.SUCCESS
	}

	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (nn *NodeNames) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (nn *NodeNames) ReelEOF() {
}
