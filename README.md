Noisy Neighbor Nozzle
[![slack.cloudfoundry.org][slack-badge]][loggregator-slack]
[![CI Badge][ci-badge]][ci-pipeline]
=====================

The Noisy Neighbor Nozzle is a Loggregator Firehose nozzle and
CLI Tool to help Operators identify applications producing a large
amount of logs - i.e. "noise".

## Getting Started & Authentication
In order to properly deploy the nozzle components you'll need to create
a UAA client with the apropriate permisions. This can be done by adding the
following client to your deoployment manifest and updating your deployment.


```
noisy-neighbor-nozzle:
  authorities: doppler.firehose,uaa.resource,cloud_controller.admin_read_only
  authorized-grant-types: client_credentials,refresh_token
  override: true
  scope: doppler.firehose
  secret: <secret>
```


## Deploying
The easiest way to deploy is to use the `deployer` binary for your local OS included in
our release package. If you add the flag `--interactive` you will be propted for all
the information required to deploy the nozzle components. If you are interested in
operationalizing the deployment experience you should use the [Noisy Neighbor Nozzle
Release][noisy-neighbor-nozzle-release].

## Using the CLI
The easisest way to quickly check your platform for top log producers is to use
the CLI tool. Download and install the binary for your local os using the command
`cf install-plugin`.

To see the top 10 log producers in the last minute run:

```
cf log-noise
```

## Integrating with the Noisy Neigbor Nozzle
The datadog-reporter is an optional component used for integrating with datadog.
When deployed, it will request rates from the accumulator every minute and
report the top 50 noisiest applications to [Datadog][datadog]


## How it works

The nozzle will read logs (excluding router logs by default) from the
Loggregator firehose keeping counts for the number of logs received for each
application. The nozzle stores the last 60 minutes worth of this data in an
in-memory cache.

The accumulator acts as a proxy for all the nozzles. When the accumulator
receives an HTTP request it will forward the same request to all the nozzles.
The accumulator then takes to rates from all the nozzles and sums them together,
responding with the total rates.


## Scaling

The nozzle can be scaled horizontally. We recommend having the same number of
nozzles as you have Loggregator Traffic Controllers.

The accumulator and datadog-reporter should only be deployed with a single
instance.

## Accumulator API

### **GET** `/rates/{timestamp}`

#### Headers

- `Authorization` - OAuth2 token, must have `doppler.firehose` scope.

#### Parameters

- `timestamp` - Unix timetamp truncated (to the minute) to the nozzles `POLLING_INTERVAL`.

### Query Parameter

- `truncate_timestamp` - Optional query parameter to truncate the given
  timestamp to the by the configured `RATE_INTERVAL` (Default is 1 minute). If
  `true` timestamp will be truncated, otherwise it will not be modified.

#### Example

```
curl -H "Authorization: $AUTH_TOKEN" https://nn-accumulator.<app-domain>/rates/$(python -c 'import time; n=time.time(); print(int(n-n%60)-(5*60))')
{
    "timestamp": 1514042640,
    "counts": {
        "06d83ae4-7632-46b9-af96-5f90f56ba0c5/0": 6456,
        "0dbb1e16-9da6-4a31-b8b3-fdff5258e20b/0": 129,
        "14213570-140d-41df-9a4e-481f7e010a08/0": 7,
        ...
    }
}
```


[bosh]:              https://bosh.io
[nn-releases]:        https://github.com/cloudfoundry/noisy-neighbor-nozzle/releases
[cf-cli]:            https://github.com/cloudfoundry/cli
[datadog]:           https://datadoghq.com
[ci-badge]:          https://loggregator.ci.cf-app.com/api/v1/pipelines/products/jobs/noisy-neighbor-nozzle-bump-submodule/badge
[ci-pipeline]:       https://loggregator.ci.cf-app.com/teams/main/pipelines/products/jobs/noisy-neighbor-nozzle-bump-submodule
[slack-badge]:       https://slack.cloudfoundry.org/badge.svg
[firehose-details]:  https://github.com/cloudfoundry/loggregator-release#consuming-the-firehose
[loggregator-slack]: https://cloudfoundry.slack.com/archives/loggregator
[noisy-neighbor-nozzle]:         https://code.cloudfoundry.org/noisy-neighbor-nozzle
[noisy-neighbor-nozzle-release]: https://code.cloudfoundry.org/noisy-neighbor-nozzle-release
[uaa-user-vs-client]: https://github.com/cloudfoundry/uaa/blob/master/docs/UAA-Tokens.md#users-and-clients-and-other-actors
