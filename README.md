
# Test Network Function ![build](https://github.com/test-network-function/test-network-function/actions/workflows/merge.yaml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/test-network-function/test-network-function)](https://goreportcard.com/report/github.com/test-network-function/test-network-function)


This repository contains a set of Cloud-Native Network Functions (CNFs) test cases and the framework to build more. The tests and framework are intended 
to test the interaction of CNFs with OpenShift Container Platform.  It also generates a report 
(claim.json) after running tests.

Please consult [CATALOG.md](./CATALOG.md) for a catalog of the included test cases and test case building blocks.

The suite is provided here in part so that CNF Developers can use the suite to test their CNFs readiness for
certification.  Please see "CNF Developers" below for more information.

## Overview 
 ![overview](docs/images/overview.svg)

In the diagram above:
- the `CNF under test` is the CNF to be certified. The certification suite identifies the resources (containers/pods/operators etc) belonging to the CNF via labels or static data entries in the config file
- the `Certification container/exec` is the certification test suite running on the platform or in a container. The executable verifies the CNF under test configuration and its interactions with openshift 
- the `Partner pod` can be any pod with the required tools in the same namespace as the `CNF under test`. For example, during connectivity tests, the partner pod will generate pings towards the `CNF under test` to verify connectivity. The partner pods/containers are auto deployed by the test suite prior a test run and can be auto discovered by the suite without any data entry in the config file.


## Test Configuration

The Test Network Function support auto-configuration using labels and annotations, but also a static configuration using a file. The following sections describe how to configure the TNF via labels/annotation and the corresponding settings in the config file. A sample config file can be found [here](test-network-function/tnf_config.yml).

### targetPodLabels
The goal of this section is to specify the label to be used to identify the CNF resources under test. It's highly recommended that the labels should be defined in pod definition rather than added after pod is created, as labels added later on will be lost in case the pod gets rescheduled. In case of pods defined as part a deployment, it's best to use the same label as the one defined in the `spec.selector.matchLabels` section of the deployment yaml. The namespace field is just a prefix for the label name, e.g. with the default configuration:
```shell script
targetPodLabels:
  - namespace: test-network-function.com
    name: generic
    value: target
```

The corresponding label used to match pods is: 
```shell script
test-network-function.com/generic: target 
```

Once the pods are found, all of their containers are also added to the target container list.

### testTarget
#### podsUnderTest / containersUnderTest
This section is usually not required if labels defined in the section above cover all resources that should be tested. If label based discovery is not sufficient, this section can be manually populated as shown in the commented part of the [sample config](test-network-function/tnf_config.yml). However, instrusive tests need to be skipped ([see here](#disable-intrusive-tests)) for a reliable test result. The pods and containers explicitly configured here are added to the target pod/container lists populated through label matching.

For both configured and discovered pods/containers, the autodiscovery mechanism will attempt to identify the default network device and all the IP addresses of the pods it
needs for network connectivity tests, though that information can be explicitly set using annotations if needed. For Pod IPs:

* The annotation `test-network-function.com/multusips` is the highest priority, and must contain a JSON-encoded list of
IP addresses to be tested for the pod. This must be explicitly set.
* If the above is not present, the `k8s.v1.cni.cncf.io/networks-status` annotation is checked and all IPs from it are
used. This annotation is automatically managed in OpenShift but may not be present in K8s.
* If neither of the above is present, then only known IPs associated with the pod are used (the pod `.status.ips` field).

For Network Interfaces:

* The annotation `test-network-function.com/defaultnetworkinterface` is the highest priority, and must contain a
JSON-encoded string of the primary network interface for the pod. This must be explicitly set if needed. Examples can
be seen in [cnf-certification-test-partner](https://github.com/test-network-function/cnf-certification-test-partner/local-test-infra/local-pod-under-test/local-partner-pod.yaml)
* If the above is not present, the `k8s.v1.cni.cncf.io/networks-status` annotation is checked and the `"interface"` from
the first entry found with `"default"=true` is used. This annotation is automatically managed in OpenShift but may not
be present in K8s.

If multus IP addresses are dicovered or configured, the partner pod needs to be deployed in the same namespace as the multus network interface for the connectivity test to pass. Refer to instruction [here](#specify-the-target-namespace-for-partner-pod-deployment).

If a pod is not suitable for network connectivity tests because it lacks binaries (e.g. `ping`), it should be
given the label `test-network-function.com/skip_connectivity_tests` to exclude it from those tests. The label value is
not important, only its presence. Equivalent to `excludeContainersFromConnectivityTests` in the config file.


#### operators

The section can be configured as well as auto discovered. For manual configuration, see the commented part of the [sample config](test-network-function/tnf_config.yml). For autodiscovery:

* CSVs to be tested by the `operator` spec are identified with the `test-network-function.com/operator=target`
label. Any value is permitted but `target` is used here for consistency with the other specs.
* Defining which tests are to be run on the operator is done using the `test-network-function.com/operator_tests`
annotation. This is equivalent to the `test-network-function.com/container_tests` and behaves the same.
* `test-network-function.com/subscription_name` is optional and should contain a JSON-encoded string that's the name of
the subscription for this CSV. If unset, the CSV name will be used.

### testPartner

This section can also be discovered automatically and should be left commented out unless the parter pods are modified from the original version in [cnf-certification-test-partner](https://github.com/test-network-function/cnf-certification-test-partner/local-test-infra/)

### certifiedcontainerinfo and certifiedoperatorinfo

The `certifiedcontainerinfo` and `certifiedoperatorinfo` sections contain information about CNFs and Operators that are
to be checked for certification status on Red Hat catalogs.

## Runtime environement variables
### Turn off openshift required tests
When test on CNFs that run on k8s only environment, execute shell command below before compile tool and run test shell script.

```shell script
export TNF_MINIKUBE_ONLY=true
```

### Disable intrusive tests
If you would like to skip intrusive tests which may disrupt cluster operations, issue the following:

```shell script
export TNF_NON_INTRUSIVE_ONLY=true
```

Likewise, to enable intrusive tests, set the following:

```shell script
export TNF_NON_INTRUSIVE_ONLY=false
```

### Specifiy the location of the partner repo
This env var is optional, but highly recommended if running the test suite from a clone of this github repo. It's not needed or used if running the tnf image.

To set it, clone the partner [repo](https://github.com/test-network-function/cnf-certification-test-partner) and set TNF_PARTNER_SRC_DIR to point to it.

```shell script
export TNF_PARTNER_SRC_DIR=/home/userid/code/cnf-certification-test-partner
```

When this variable is set, the run-cnf-suites.sh script will deploy/refresh the partner deployments/pods in the cluster before starting the test run.

### Specify the target namespace for partner pod deployment
Set this variable to deploy partner pods in a custom namespace instead of the default `tnf` namesapce.

```shell-script
export TNF_PARTNER_NAMESPACE="CNF-ns"
```


### Execute test suites from openshift-kni/cnf-feature-deploy
The test suites from openshift-kni/cnf-feature-deploy can be run prior to the actual CNF certification test execution and the results are incorporated in the same claim file if the following environment variable is set:


```shell script
export VERIFY_CNF_FEATURES=true
```

Currently, these suites are skipped:
* performance
* sriov
* ptp
* sctp
* xt_u32
* dpdk
* ovn

For more information on the test suites, refer to [the cnf-features-deploy repository](https://github.com/openshift-kni/cnf-features-deploy/tree/release-4.6)

## Running the tests with in a prebuild container

### Pulling test image
An image is built and is available at this repository: [quay.io](https://quay.io/repository/testnetworkfunction/test-network-function)
The image can be pulled using :
```shell script
docker pull quay.io/testnetworkfunction/test-network-function
```
### Cluster requirement
To run all the tests, OCP cluster should provide at least 3 worker nodes. If that's not the case, then ``platform-alteration`` test suite should be skipped.
### Check cluster resources
Some tests suites such as platform-alteration require node access to get node configuration like hugepage.
In order to get the required information, the test suite does not ssh into nodes, but instead rely on [oc debug tools ](https://docs.openshift.com/container-platform/3.7/cli_reference/basic_cli_operations.html#debug). This tool makes it easier to fetch information from nodes and also to debug running pods.

In short, oc debug tool will launch a new container ending with "-debug" suffix, the container will be destroyed once the debug session is done. To be able to create the debug pod,  the cluster should have enough resources, otherwise those tests would fail.

**Note:**
It's recommended to clean up disk space and make sure there's enough resources to deploy another container image in every node before starting the tests.
### Run the tests
``./run-tnf-container.sh`` script is used to launch the tests.  

There are several required arguments:

* `-t` gives the local directory that contains tnf config files set up for the test.
* `-o` gives the local directory that the test results will be available in once the container exits.
* `-f` gives the list of suites to be run, space separated.

Optional arguments are:

* `-i` gives a name to a custom TNF container image. Supports local images, as well as images from external registries.
* `-k` gives a path to one or more kubeconfig files to be used by the container to authenticate with the cluster. Paths must be separated by a colon.
* `-n` gives the network mode of the container. Defaults set to `host`, which requires selinux to be disabled. Alternatively, `bridge` mode can be used with selinux if TNF_CONTAINER_CLIENT is set to `docker` or running the test as root. See the [docker run --network parameter reference](https://docs.docker.com/engine/reference/run/#network-settings) for more information on how to configure network settings.
* `-s` gives the name of tests that should be skipped


If `-k` is not specified, autodiscovery is performed.
The autodiscovery first looks for paths in the `$KUBECONFIG` environment variable on the host system, and if the variable is not set or is empty, the default configuration stored in `$HOME/.kube/config` is checked.

```shell script
./run-tnf-container.sh -k ~/.kube/config -t ~/tnf/config -o ~/tnf/output -f diagnostic access-control -s access-control-host-resource-PRIVILEGED_POD
```

See [General tests](#general-tests) for a list of available keywords.

*Note*: The `run-tnf-container.sh` script performs autodiscovery of selected TNF environment variables.  
Currently supported environment variables include:
- `TNF_MINIKUBE_ONLY`

### Running using `docker` instead of `podman`

By default, `run-container.sh` utilizes `podman`.  However, you can configure an alternate container virtualization
client using `TNF_CONTAINER_CLIENT`.  This is particularly useful for operating systems that do not readily support
`podman`, such as macOS.  In order to configure the test harness to use `docker`, issue the following prior to
`run-tnf-container.sh`:

```shell script
export TNF_CONTAINER_CLIENT="docker"
```

### Building the container image locally

You can build an image locally by using the command below. Use the value of `TNF_VERSION` to set a branch, a tag, or a hash of a commit that will be installed into the image.

```shell script
docker build -t test-network-function:v1.0.5 --build-arg TNF_VERSION=v1.0.5 .
```

To build an image that installs TNF from an unofficial source (e.g. a fork of the TNF repository), use the `TNF_SRC_URL` build argument to override the URL to a source repository.

```shell script
docker build -t test-network-function:v1.0.5 \
  --build-arg TNF_VERSION=v1.0.5 \
  --build-arg TNF_SRC_URL=https://github.com/test-network-function/test-network-function .
```

To make `run-tnf-container.sh` use the newly built image, specify the custom TNF image using the `-i` parameter.

```shell script
./run-tnf-container.sh -i test-network-function:v1.0.5 -t ~/tnf/config -o ~/tnf/output -f diagnostic access-control
```
 Note: see [General tests](#general-tests) for a list of available keywords.


## Building and running the standalone test executable

Currently, all available tests are part of the "CNF Certification Test Suite" test suite, which serves as the entrypoint to run all test specs.
By default, `test-network-function` emits results to `test-network-function/cnf-certification-tests_junit.xml`.

The included default configuration is for running `diagnostic`  suite on the trivial example at
[cnf-certification-test-partner](https://github.com/test-network-function/cnf-certification-test-partner). To configure the tests for your own environment, 

please see the example [tnf_config.yml](https://github.com/test-network-function/test-network-function/blob/main/test-network-function/tnf_config.yml).

### Dependencies

At a minimum, the following dependencies must be installed *prior* to running `make install-tools`.

Dependency|Minimum Version
---|---
[GoLang](https://golang.org/dl/)|1.14
[golangci-lint](https://golangci-lint.run/usage/install/)|1.32.2
[jq](https://stedolan.github.io/jq/)|1.6
[OpenShift Client](https://docs.openshift.com/container-platform/4.4/welcome/index.html)|4.4

Other binary dependencies required to run tests can be installed using the following command:

```shell script
make install-tools
```

Finally the source dependencies can be installed with

```shell script
make update-deps
```

*Note*: You must also make sure that `$GOBIN` (default `$GOPATH/bin`) is on your `$PATH`.

*Note*:  Efforts to containerize this offering are considered a work in progress.



### Pulling The Code

In order to pull the code, issue the following command:

```shell script
mkdir ~/workspace
cd ~/workspace
git clone git@github.com:test-network-function/test-network-function.git
cd test-network-function
```

### Building the Tests

In order to build the test executable, first make sure you have satisfied the [dependencies](#dependencies).
```shell script
make build-cnf-tests
```

*Gotcha:* The `make build*` commands run unit tests where appropriate. They do NOT test the CNF.

### Testing a CNF

Once the executable is built, a CNF can be tested by specifying which suites to run using the `run-cnf-suites.sh` helper
script.

Run any combination of the suites keywords listed at in the [General tests](#general-tests) section, e.g.

```shell script
./run-cnf-suites.sh -f diagnostic
./run-cnf-suites.sh -f diagnostic lifecycle
./run-cnf-suites.sh -f diagnostic networking operator
./run-cnf-suites.sh -f diagnostic platform-alteration
./run-cnf-suites.sh -f diagnostic lifecycle affiliated-certification operator
```

By default the claim file will be output into the same location as the test executable. The `-o` argument for
`run-cnf-suites.sh` can be used to provide a new location that the output files will be saved to. For more detailed
control over the outputs, see the output of `test-network-function.test --help`.

```shell script
cd test-network-function && ./test-network-function.test --help
```

*Gotcha:* The generic test suite requires that the CNF has both `ping` and `ip` binaries installed.  Please add them
manually if the CNF under test does not include these.  Automated installation of missing dependencies is targeted
for a future version.
*Gotcha:* check that OCP cluster has resources to deploy [debug image](#check-cluster-resources)
## Available Test Specs

There are two categories for CNF tests;  'General' and 'CNF-specific' (TODO).

The 'General' tests are designed to test any commodity CNF running on OpenShift, and include specifications such as
'Default' network connectivity.

'CNF-specific' tests are designed to test some unique aspects of the CNF under test are behaving correctly.  This could
include specifications such as issuing a `GET` request to a web server, or passing traffic through an IPSEC tunnel.

'CNF-specific' test are yet to be defined.

### General tests

Test in the "general" category belong to multiple suites that can be run in any combination as is
appropriate for the CNF(s) under test. Test suites group tests by topic area:

Suite|Test Spec Description|Minimum OpenShift Version
---|---|---
`access-control`|The access-control test suite is used to test  service account, namespace and cluster/pod role binding for the pods under test. It also tests the pods/containers configuration.|4.4.3
`affiliated-certification`|The affiliated-certification test suite verifies that the containers in the pod under test and operator under test are certified by Redhat|4.4.3
`diagnostic`|The diagnostic test suite is used to gather node information from an OpenShift cluster.  The diagnostic test suite should be run whenever generating a claim.json file.|4.4.3
`lifecycle`| The lifecycle test suite verifies the pods deployment, creation, shutdown and  survivability. |4.4.3
`networking`|The networking test suite contains tests that check connectivity and networking config related best practices.|4.4.3
`operator`|The operator test suite is designed to test basic Kubernetes Operator functionality.|4.4.3
`platform-alteration`| verifies that key platform configuration is not modified by the CNF under test|4.4.3

Please consult [CATALOG.md](CATALOG.md) for a detailed description of tests in each suite.


### CNF-specific tests
TODO

## Test Output

### Claim File

The test suite generates a "claim" file, which describes the system(s) under test, the tests that were run, and the
outcome of all of the tests.  This claim file is the proof of the test run that is evaluated by Red Hat when
"certified" status is being considered.  For more information about the contents of the claim file please see the
[schema](https://github.com/test-network-function/test-network-function-claim/blob/main/schemas/claim.schema.json).  You can
read more about the purpose of the claim file and CNF Certification in the
[Guide](https://redhat-connect.gitbook.io/openshift-badges/badges/cloud-native-network-functions-cnf).

### Adding Test Results for the CNF Validation Test Suite to a Claim File 
e.g. Adding a cnf platform test results to your existing claim file.

You can use the claim cli tool to append other related test suite results to your existing claim.json file.
The output of the tool will be an updated claim file.
```
go run cmd/tools/cmd/main.go claim-add --claimfile=claim.json --reportdir=/home/$USER/reports
```
 Args:  
`
--claimfile is an existing claim.json file`
`
--repordir :path to test results that you want to include.
`

 The tests result files from the given report dir will be appended under the result section of the claim file using file name as the key/value pair.
 The tool will ignore the test result, if the key name is already present under result section of the claim file.
```
 "results": {
 "cnf-certification-tests_junit": {
 "testsuite": {
 "-errors": "0",
 "-failures": "2",
 "-name": "CNF Certification Test Suite",
 "-tests": "14",
```

### Command Line Output

When run the CNF test suite will output a report to the terminal that is primarily useful for Developers to evaluate and
address problems.  This output is similar to many testing tools.

#### Test successful output example
Here's an example of a Test pass.  It verifies that the CNF is using a replica set:

```shell
------------------------------
lifecycle when Testing owners of CNF pod 
  Should be only ReplicaSet
  /Users/$USER/cnf-cert/test-network-function/test-network-function/lifecycle/suite.go:339
2021/07/27 11:41:25 Sent: "oc -n tnf get pods test-697ff58f87-d55zx -o custom-columns=OWNERKIND:.metadata.ownerReferences\\[\\*\\].kind && echo END_OF_TEST_SENTINEL\n"
2021/07/27 11:41:26 Match for RE: "(?s)OWNERKIND\n.+((.|\n)*END_OF_TEST_SENTINEL\n)" found: ["OWNERKIND\nReplicaSet\nEND_OF_TEST_SENTINEL\n" "END_OF_TEST_SENTINEL\n" ""] Buffer: "OWNERKIND\nReplicaSet\nEND_OF_TEST_SENTINEL\n"
•

```

#### Test failed output examples

The following is the output from a Test failure.  In this case, the test is checking that a CSV (ClusterServiceVersion)
is installed correctly, but does not find it (the operator was not present on the cluster under test):

```shell
------------------------------
operator Runs test on operators when under test is: my-etcd/etcdoperator.v0.9.4  
  tests for: CSV_INSTALLED
  /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:122
2020/12/15 15:28:19 Sent: "oc get csv etcdoperator.v0.9.4 -n my-etcd -o json | jq -r '.status.phase'\n"

• Failure [10.002 seconds]
operator
/Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:58
  Runs test on operators
  /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:71
    when under test is: my-etcd/etcdoperator.v0.9.4 
    /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:121
      tests for: CSV_INSTALLED [It]
      /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:122

      Expected
          <int>: 0
      to equal
          <int>: 1

```

The following is the output from a Test failure.  In this case, the test is checking that a Subscription
is installed correctly, but does not find it (the operator was not present on the cluster under test):

```shell
------------------------------
operator Runs test on operators when under test is: my-etcd/etcd
  tests for: SUBSCRIPTION_INSTALLED
  /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:129
2021/04/09 12:37:10 Sent: "oc get subscription etcd -n my-etcd -ojson | jq -r '.spec.name'\n"

• Failure [10.000 seconds]
operator
/Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:55
  Runs test on operators
  /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:68
    when under test is: default/etcdoperator.v0.9.4 
    /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:128
      tests for: SUBSCRIPTION_INSTALLED [It]
      /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:129

      Expected
          <int>: 0
      to equal
          <int>: 1

```

The following is the output from a Test failure.  In this case, the test is checking clusterPermissions for
specific CSV, but does not find it (the operator was not present on the cluster under test):

```shell
------------------------------
operator Runs test on operators 
  should eventually be verified as certified (operator redhat-marketplace/etcd-operator)
  /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:146

• Failure [30.002 seconds]
operator
/Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:76
  Runs test on operators
  /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:89
    should eventually be verified as certified (operator redhat-marketplace/etcd-operator) [It]
    /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:146

    Timed out after 30.001s.
    Expected
        <bool>: false
    to be true

    /Users/$USER/cnf-cert/test-network-function/test-network-function/operator/suite.go:152
```
## Log level 
The optional LOG_LEVEL environment variable sets the log level. Defaults to "info" if not set. Valid values are: trace, debug, info, warn, error, fatal, panic.

## Grading Tool
### Overview
A tool for processing the claim file and producing a quality grade for the CNF.
The user supplies a policy conforming to [policy schema](schemas/gradetool-policy-schema.json).
A grade is considered `passed` if all its direct tests passed and its base grade passed.
In the output we use the field `propose` to indicate grade passed or failed.
See [policy example](pkg/gradetool/testdata/policy-good.json) for understanding the output of the grading tool.
### How to build and execute
```
make build
or
make build-gradetool
```
Executable name is `gradetool`.

## CNF Developers

Developers of CNFs, particularly those targeting 
[CNF Certification with Red Hat on OpenShift](https://redhat-connect.gitbook.io/openshift-badges/badges/cloud-native-network-functions-cnf),
can use this suite to test the interaction of their CNF with OpenShift.  If you are interested in CNF Certification
please contact [Red Hat](https://redhat-connect.gitbook.io/red-hat-partner-connect-general-guide/managing-your-account/getting-help/technology-partner-success-desk).

Refer to the rest of the documentation in this file to see how to install and run the tests as well as how to
interpret the results.

You will need an [OpenShift 4.4 installation](https://docs.openshift.com/container-platform/4.4/welcome/index.html)
running your CNF, and at least one other machine available to host the test suite.  The
[cnf-certification-test-partner](https://github.com/test-network-function/cnf-certification-test-partner) repository has a very
simple example of this you can model your setup on.

# Known Issues

## Issue #146:  Shell Output larger than 16KB requires specification of the TNF_DEFAULT_BUFFER_SIZE environment variable

When dealing with large output, you may occasionally overrun the default buffer size. The manifestation of this issue is
a `json.SyntaxError`, and may look similar to the following:

```shell script
    Expected
        <*json.SyntaxError | 0xc0002bc020>: {
            msg: "unexpected end of JSON input",
            Offset: 660,
        }
    to be nil
```

In such cases, you will need to set the TNF_DEFAULT_BUFFER_SIZE to a sufficient size (in bytes) to handle the expected
output.

For example:

```shell script
TNF_DEFAULT_BUFFER_SIZE=32768 ./run-cnf-suites.sh diagnostic generic
```

## Issue-161 Some containers under test do not contain `ping` or `ip` binary utilities

In some cases, containers do not provide ping or ip binary utilities. Since these binaries are required for the
connectivity tests, we must exclude such containers from the connectivity test suite.  In order to exclude these
containers, please add the following to `test-network-function/tnf_config.yml`:

```yaml
excludeContainersFromConnectivityTests:
  - namespace: <namespace>
    podName: <podName>
    containerName: <containerName>
```

Note:  Future work may involve installing missing binary dependencies.
