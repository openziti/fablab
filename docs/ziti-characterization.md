# Ziti Characterization

The `zitilab/characterization` package contains a set of scalable "characterization" models, used to continually evaluate and refine the performance of Ziti network overlay implementations.

## Geolocation

Define a model for a Ziti configuration, which includes a _local region_ (`us-east-1`) containing an `iperf3` (server) service.

The model will facilitate access to the _local region_ services from the following _remote regions:_

* _Short distance region:_ `us-west-1`, N. California

* _Medium distance region:_ `ap-south-1`, Mumbai

* _Long distance region:_ `ap-southeast-2`, Sydney

The characterization model will expose this configuration, allowing the implementation to be targeted at any (and more or less) geographic regions.

## Scenario

The following scenarios will be executed from each of the _remote regions_, accessing the necessary services (`iperf3`) in the _local region_:

* TCP maximum bandwidth testing, using `iperf3` to determine the maximum sustained throughput available to the client

* UDP jitter testing, using `iperf3` configured with a fixed `1mbit` sender/receiver rate. Capture both jitter and percentage of lost datagrams

* UDP jitter testing, using `iperf3` configured with a fixed `100mbit` sender/receiver rate. Capture both jitter and percentage of lost datagrams

## Ambient Metric Cloud

The characterization implementation will include telemetry probes, which continually monitor the following metrics across the entire model:

* _CPU and memory utilization:_ periodic snapshots of host performance for all hosts

* _Ziti overlay metrics:_ periodic snapshots capturing the configuration (and "shape") of the overlay mesh, along with the observed performance through the overlay

## Variations and Configurability

A `fablab` model will be created that supports the following kinds of variations:

* _Comparison model(s):_ By default, the characterization implementation will compare the performance of the Ziti deployment to one or more comparison models. In this way, the comparison implementation(s) can also be captured as separate `fablab` models.

* _EC2 instance type:_ The prior work done in network characterization included deployments on `t2.micro`, `t2.small`, `t2.medium`, `t2.large`, `m5.xlarge`, `m5.2xlarge`, and `m5.4xlarge` instance types. Ziti characterization should target at least those instance types, but will be parameterized to allow for testing any permutations of instance types.

* _Scenario run time:_ How long should each scenario be executed?

* _Session count:_ Scenario will have a configurable number of concurrent sessions, and the data from those sessions will be aggregated in reporting.

* _Ziti deployment architecture:_ The characterization model will prescribe a default architecture, but will be capable of executing on multiple different network topologies, which illustrate different kinds of benefits and tradeoffs of underlying Ziti deployments.

## Hands Off

`fablab` will provide a mechanism to allow an entire "manifest" of model configurations to be characterized sequentially (or possibly in parallel, eventually), creating a detailed data set that accurately describes the performance of specific Ziti network configuration.

## NetFoundry Reference Characterizations

Previous efforts to characterize NetFoundry network implementations, comparing them against various other technologies:

https://netfoundry.atlassian.net/wiki/spaces/NE/pages/255819826/VPN+compare+Network+Characterization

https://netfoundry.atlassian.net/wiki/spaces/NE/pages/533496058/IPERF3+AWS+Testing+For+DVN+3.6.6

_(NetFoundry internal wiki)_