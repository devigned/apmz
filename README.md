# apmz: cmd line for sending events to Azure Application Insights in shell scripts

[![Go Report Card](https://goreportcard.com/badge/github.com/devigned/apmz)](https://goreportcard.com/report/github.com/devigned/apmz)
[![Actions Status](https://github.com/devigned/apmz/workflows/ci/badge.svg)](https://github.com/devigned/apmz/actions)
[![Documentation](https://godoc.org/github.com/devigned/apmz?status.svg)](https://godoc.org/github.com/devigned/apmz)
[![Coverage Status](https://coveralls.io/repos/github/devigned/apmz/badge.svg?branch=master)](https://coveralls.io/github/devigned/apmz?branch=master)

`apmz` enables shell script developers to instrument their scripts with [Application Insights](https://docs.microsoft.com/en-us/azure/azure-monitor/app/app-insights-overview), 
so you can trace events, measure script durations, duration of script functions, and catch script errors.
Best of all, this enables you to query, alert and build dashboards from this telemetry. 

## Install

Easiest way is to use golang, but there are also precompiled binaries on the [releases page](https://github.com/devigned/apmz/releases/).

```bash
$ go get github.com/devigned/apmz
...
$ apmz -h
...
```

## Usage

`apmz` offers a bunch of simple functionality to get you started quickly. First, you need to 
create an [Application Insights resource and fetch the instrumentation key](https://docs.microsoft.com/en-us/azure/azure-monitor/app/create-new-resource#copy-the-instrumentation-key).
Once you have created the resource and procured an instrumentation key, you can run the script below. 

__Be sure to add your key to the script or as an ENV var (INSTRUMENATION_KEY).__

```bash
#!/usr/bin/env bash

# put apmz in the path if it is not already
# this assumes ./bin directory holds the apmz binary
export PATH=./bin:$PATH

eval "$(apmz bash -n "myscript" -t "default=tag,something=cool" --api-keys "${INSTRUMENATION_KEYS}" )"
```

The script above evals a script generated by the `apmz bash` command. The script generated by 
`apmz bash` setups some env vars, functions and exit hooks to assist with instrumentation. Just by
executing the above script, there will be two events sent to Application Insights.
1) A trace event which contains the exit code of the script
2) A customMetric which contains the duration of the entire script

You should be able to view and query these traces and customMetrics via the [Log Query UI](https://docs.microsoft.com/en-us/azure/azure-monitor/log-query/log-query-overview).

### Tracing and Time Metrics
With the basics of setting up a script handled, here's a script where we setup some tracing and
metrics of our own.

```bash
#!/usr/bin/env bash

# put apmz in the path if it is not already
# this assumes ./bin directory holds the apmz binary
export PATH=./bin:$PATH

# inject tracing helpers into the script naming the script "myscript" with default tags of
# "default=tag,something=cool" applied to all traces and metrics logging to the Application
# Insights resource identified by the INSTRUMENATION_KEY
eval "$(apmz bash -n "myscript" -t "default=tag,something=cool" --api-keys "${INSTRUMENATION_KEYS}" )"

# simple function to measure
sleep_for_a_second() {
  sleep 1
}

# log info customTrace a name and tags
trace_info "name-of-info-event" "tag1=foo,tag2=bar"

# log an error customTrace with a name and tags
trace_error "oh-no-an-error" "bang=crash"

# log the duration of the function sleep_for_a_second and log it as a custom metric
time_metric "my-metric-name" sleep_for_a_second

# script ends here and the helper exit hook takes the local logged events and sends them to 
# Application Insights
```

If you are interested in seeing more of what the script does just run `apmz bash` and you can see.

### What can I use with out eval'ing `apmz bash`
Well, you can do all of the things that `apmz bash` does, but you have to write your own functions.

```
$ apmz
apmz provides a command line interface for the Azure Application Insights

Usage:
  apmz [command]

Available Commands:
  bash        prints a bash script to source which provides functionality for common tracing and metrics operations
  batch       upload a batch of telemetry to Application Insights
  help        Help about any command
  metadata    Azure instance metadata service related commands
  metric      send a metric (customMetrics) to Application Insights
  time        time related commands
  trace       send a trace event (traces) to Application Insights
  uuid        generate a new uuid
  version     Print the git ref

Flags:
      --api-keys strings   comma separated keys for the Application Insights accounts to send to; eg 'key1,key2,key3'
  -h, --help             help for apmz
  -o, --output           instead of sending directly to Application Insights, output event to stdout as json

Use "apmz [command] --help" for more information about a command.
```

#### Access the instance metadata endpoint on the VM
```bash
$ apmz metadata instance | jq
{
  "compute": {
    "azEnvironment": "AzurePublicCloud",
    "location": "westus2",
    "name": "test1",
    "offer": "UbuntuServer",
    "osType": "Linux",
    "plan": {},
    "platformFaultDomain": "0",
    "platformUpdateDomain": "0",
    "provider": "Microsoft.Compute",
    "publicKeys": [
      {
        "keyData": "key",
        "path": "/home/azureuser/.ssh/authorized_keys"
      }
    ],
    "publisher": "Canonical",
    "resourceGroupName": "test",
    "resourceId": "/subscriptions/something/resourceGroups/test/providers/Microsoft.Compute/virtualMachines/test1",
    "sku": "18.04-LTS",
    "subscriptionId": "foo",
    "version": "18.04.201912180",
    "vmId": "114433da-3f44-4bc1-9fa3-a714516c2abd",
    "vmSize": "Standard_DS1_v2"
  },
  "network": {
    "interface": [
      {
        "ipv4": {
          "ipAddress": [
            {
              "privateIpAddress": "10.0.0.4",
              "publicIpAddress": "52.228.7.11"
            }
          ],
          "subnet": [
            {
              "address": "10.0.0.0",
              "prefix": "24"
            }
          ]
        },
        "ipv6": {},
        "macAddress": "000D3AC2FB9A"
      }
    ]
  }
}
```

#### Get an auth token for the local VM identity
```bash
$ token=$(apmz metadata token -r "https://management.azure.com/" | jq -r ".access_token")
$ curl -H "Authorization: Bearer $token" https://management.azure.com/subscriptions?api-version=2019-11-01 | jq
{
  "value": [
    {
      "id": "/subscriptions/sub-id",
      "authorizationSource": "RoleBased",
      "managedByTenants": [],
      "subscriptionId": "sub-id",
      "tenantId": "tenant-id",
      "displayName": "SubscriptionName",
      "state": "Enabled",
      "subscriptionPolicies": {
        "locationPlacementId": "Public_2014-09-01",
        "quotaId": "MSDN_2014-09-01",
        "spendingLimit": "On"
      }
    }
  ],
  "count": {
    "type": "Total",
    "value": 1
  }
}
```

## Developing
With a standard Golang 1.13+ setup you should be able to pull down the repo and run `make` and `make test`.
All of the tools required should be installed for you upon first build.

If you run into a problem, please open an issue.

## Contributing
Contributions are always welcome. Please don't hesitate to open an issue or a pull request.
