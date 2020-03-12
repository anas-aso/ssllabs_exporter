# SSLLabs exporter
[![Release](https://img.shields.io/github/release/anas-aso/ssllabs_exporter.svg?style=flat)](https://github.com/anas-aso/ssllabs_exporter/releases/latest)
[![Build Status](https://github.com/anas-aso/ssllabs_exporter/workflows/test/badge.svg)](https://github.com/anas-aso/ssllabs_exporter/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/anas-aso/ssllabs_exporter)](https://goreportcard.com/report/github.com/anas-aso/ssllabs_exporter)

Getting deep analysis of the configuration of any SSL web server on the public Internet à la blackbox_exporter style.

This exporter relays the target server hostname to [SSLLabs API](https://www.ssllabs.com/ssltest), parses the result and export it as Prometheus metrics. It covers retries in case of failures and simplifies the assessment result.

## SSLLabs
> SSL Labs is a non-commercial research effort, run by [Qualys](https://www.qualys.com/), to better understand how SSL, TLS, and PKI technologies are used in practice.

source: https://www.ssllabs.com/about/assessment.html

This exporter implements SSLLabs API client that would get you the same results as if you use the [web interface](https://www.ssllabs.com/ssltest/).

## Configuration
ssllabs_exporter doesn't require any configuration file and the available flags can be found as below :
```bash
$ ssllabs_exporter --help
usage: ssllabs_exporter [<flags>]

Flags:
  --help                     Show context-sensitive help (also try --help-long and --help-man).
  --listen-address=":19115"  The address to listen on for HTTP requests.
  --timeout=300              Assessment timeout in seconds (including retries).
  --log-level=debug          Printed logs level.
  --version                  Show application version.
```

## Docker
The Prometheus exporter is available as a [docker image](https://hub.docker.com/repository/docker/anasaso/ssllabs_exporter) :
```
docker run --rm -it anasaso/ssllabs_exporter:latest --help
```

## How To Use it
Deploy the exporter to your infrastructure. Kubernetes deployment and service Yaml file are provided [here](examples/kubernetes) as an example.

Then adjust Prometheus config to add a new scrape configuration. Examples of how this look like can be found [here](examples/prometheus) (it includes both static config and Kubernetes service discovery to auto check all the cluster ingresses).

Once deployed, Prometheus Targets view page should look like this : 
![prometheus-targets-view](https://i.imgur.com/fJCun72.png "Prometheus Targets View")

The Grafana dashboard below is available [here](examples/grafana_dashboard.json).
![grafana-dashboard](https://i.imgur.com/q71BpOa.png "Grafana Dashboard")

## Available metrics
| Metric Name | Description |
|----|-----------|
| ssllabs_probe_duration_seconds | how long the assessment took in seconds |
| ssllabs_probe_success | whether we were able to fetch an assessment result from SSLLabs API (value of 1) or not (value of 0) regardless of the result content |
| ssllabs_grade | the grade of the target host |
| ssllabs_grade_age_seconds | when the result was generated in Unix time |

#### `ssllabs_grade` possible values:
  - `1` : Assessment was successful and the grade is exposed in the `grade` label of the metric.
  - `0` : Target host doesn't have any endpoint (list of returned [endpoints](https://github.com/ssllabs/ssllabs-scan/blob/master/ssllabs-api-docs-v3.md#host) is empty).
  - `-1` : Error while processing the assessment (e.g rate limiting from SSLLabs API side).
