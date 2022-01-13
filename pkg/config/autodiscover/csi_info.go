package autodiscover

import (
	"fmt"
	"strings"
)

// PodSetList holds the data from an `oc get deployment/statefulset -o json` command

// PodSetResource defines deployment/statefulset resources
var (
	csicommand = "oc get csidriver -o go-template='{{ range .items}}{{.metadata.name}} {{end}}'"
)

func getpackageandorg(csi string) (packag, organization string) {
	command := fmt.Sprintf("oc get pods -A -o go-template='{{ range .items}}{{ $alllabels := .metadata.labels}}{{ $namespace := .metadata.namespace}}{{ range .spec.containers }}{{ range .args }}{{if eq . \"--driver-name=%s\"}}{{ range $label,$value := $alllabels}}{{if eq $label \"app.kubernetes.io/managed-by\"}}{{$value}} {{$namespace}}{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}'", csi)
	out := execCommandOutput(command)
	operatorname := ""
	namespace := ""
	subscription := ""
	if out != "" {
		out := strings.Split(out, " ")
		operatorname = out[0]
		namespace = out[1]
	}
	command = fmt.Sprintf("oc get deployment %s -n %s -o go-template='{{ range $label,$value := .metadata.labels}}{{$label}}{{print \"\n\"}}{{end}}' |grep \"operators.coreos.com\"| sed \"s#operators.coreos.com/##g\"", operatorname, namespace)
	out = execCommandOutput(command)
	if out != "" {
		operatorname = out
	}
	command = fmt.Sprintf("oc get operator %s -o go-template='{{ range .status.components.refs }}{{if eq .kind \"Subscription\"}}{{.name}}{{end}}{{end}}'", operatorname)
	out = execCommandOutput(command)
	if out != "" {
		subscription = out
	}
	command = fmt.Sprintf("oc get subscription -n %s %s -o go-template='{{.spec.source}} {{.spec.name}}'", namespace, subscription)
	out = execCommandOutput(command)
	if out != "" {
		out := strings.Split(out, " ")
		organization = out[0]
		packag = out[1]
	}
	return packag, organization
}

// GetTargetCsi will return all podsets(deployments/statefulset )that have pods with a given label.
func GetTargetCsi() ([]string, error) {
	ocCmd := csicommand

	out := execCommandOutput(ocCmd)

	csiList := strings.Split(out, " ")
	return csiList, nil
}