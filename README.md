Noisy Neighbor Nozzle
[![slack.cloudfoundry.org][slack-badge]][loggregator-slack]
[![CI Badge][ci-badge]][ci-pipeline]
=====================

This is a Loggregator Firehose nozzle. It keeps track of the log rates for
Cloud Foundry deployed applications.

This nozzle can be deployed via our `deployer` binary that is packaged in our
[releases][nn-releases] which will CF push the nozzle and accumulator
components, or via [BOSH][bosh] using the [Noisy Neighbor Nozzle
Release][noisy-neighbor-nozzle-release].

## How it works

The noisy neighbor nozzle consists of five components: nozzle, accumulator,
datadog-reporter, deployer, and cli-plugin.

The nozzle will read logs (excluding router logs by default) from the
Loggregator firehose keeping counts for the number of logs received for each
application. The nozzle stores the last 60 minutes worth of this data in an
in-memory cache.

The accumulator acts as a proxy for all the nozzles. When the accumulator
receives an HTTP request it will forward the same request to all the nozzles.
The accumulator then takes to rates from all the nozzles and sums them together,
responding with the total rates.

The datadog-reporter is an optional component. When deployed, it will request
rates from the accumulator every minute and report the top 50 noisiest
applications to [Datadog][datadog]

The deployer is a tool that CF pushes the nozzle and accumulator. It requires
the [Cloud Foundry CLI][cf-cli].

The cli-plugin is a [Cloud Foundry CLI][cf-cli] plugin that can be used to query
the accumulator for the top 10 log producers.

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

## Deploy to CF

Ensure your CF deployment has a [client configured][firehose-details] with the
`doppler.firehose` scope and authority as well as the `uaa.resource`
and `cloud_controller.admin_read_only` authorities.

Download the binaries from [releases][nn-releases], and use the `deployer` or
manually set the environment variables via the cf cli or an app manifest.

### Example UAA Client

```
noisy-neighbor-nozzle:
  authorities: oauth.login,doppler.firehose,uaa.resource,cloud_controller.admin_read_only
  authorized-grant-types: client_credentials,refresh_token
  override: true
  scope: doppler.firehose,oauth.approvals
  secret: <secret>
```

### Deployer

To use the deployer with interactive prompts and defaults, `./deployer-<my-os>
--interactive`. Otherwise, set all flags provided by `./deployer-<my-os>
--help` to provide all variables required for the deploy.

### Nozzle Properties

| Property | Description |
|----------|-------------|
| `UAA_ADDR`         | The address of the Cloud Foundry deployed UAA. Normally `https://uaa.<system-domain>`. NOTE: The schema (e.g., `https`) is required. |
| `CLIENT_ID`        | The [UAA client][uaa-user-vs-client] ID with the correct `authorities`. See [example](#example-uaa-client) UAA client. |
| `CLIENT_SECRET`    | The corresponding `secret` for the given `CLIENT_ID`. |
| `LOGGREGATOR_ADDR` | The address of the Cloud Foundry deployed Loggregator. This can be discovered by looking at the `doppler_logging_endpoint` field from `cf curl v2/info`. |
| `SUBSCRIPTION_ID`  | Any unique string that can identify the nozzle cluster for consuming from Loggregator. |
| `SKIP_CERT_VERIFY` | Set to true if the Cloud Foundry is using self signed certs. |

### Deploy the Nozzle

```
cf push nn-nozzle -b binary_buildpack -c ./<nozzle-binary> -i 4 --no-start
cf set-env nn-nozzle UAA_ADDR https://uaa.<system-domain>
cf set-env nn-nozzle CLIENT_ID <CLIENT_ID>
cf set-env nn-nozzle CLIENT_SECRET <CLIENT_SECRET>
cf set-env nn-nozzle LOGGREGATOR_ADDR wss://doppler.<system-domain>:<port>
cf set-env nn-nozzle SUBSCRIPTION_ID nn-nozzle-7691798872
cf start nn-nozzle
```

##### Example App Manifest

```
---
applications:
  - name: nn-nozzle
    buildpack: binary_buildpack
    command: ./<nozzle-binary>
    memory: 128M
    instances: 3
    env:
      UAA_ADDR: https://login.bosh-lite.com
      CLIENT_ID: noisy-neighbor-nozzle
      CLIENT_SECRET: <secret for the client>
      LOGGREGATOR_ADDR: "wss://doppler.bosh-lite.com:443"
      SUBSCRIPTION_ID: nozzle-test-subscription
      SKIP_CERT_VERIFY: false
```
### Accumulator Properties

| Property | Description |
|----------|-------------|
| `UAA_ADDR`         | The address of the Cloud Foundry deployed UAA. Normally `https://uaa.<system-domain>`. NOTE: The schema (e.g., `https`) is required. |
| `CLIENT_ID`        | The [UAA client][uaa-user-vs-client] ID with the correct `authorities`. See [example](#example-uaa-client) UAA client. |
| `CLIENT_SECRET`    | The corresponding `secret` for the given `CLIENT_ID`. |
| `NOZZLE_ADDRS` | The addresses of the nozzle (e.g. `https://<nozzle-app-name>.<app-domain>`). |
| `NOZZLE_COUNT`  | The number of noisy neighbor nozzles. Required if the nozzle is deployed via `cf push`. |
| `NOZZLE_APP_GUID`  | The guid of the nozzle app. Required if the nozzle is deployed via `cf push`. This can be obtained by `cf app <nozzle-app-name> --guid`.|
| `SKIP_CERT_VERIFY` | Set to true if the Cloud Foundry is using self signed certs. |

### Deploy the Accumulator

```
cf push nn-accumulator -b binary_buildpack -c ./<accumulator-binary> --no-start
cf set-env nn-accumulator UAA_ADDR https://uaa.<system-domain>
cf set-env nn-accumulator CLIENT_ID <CLIENT_ID>
cf set-env nn-accumulator CLIENT_SECRET <CLIENT_SECRET>
cf set-env nn-accumulator NOZZLE_ADDRS https://nn-nozzle.<app-domain>
cf set-env nn-accumulator NOZZLE_COUNT 4
cf set-env nn-accumulator NOZZLE_APP_GUID $(cf app nn-nozzle --guid)
cf start nn-accumulator
```

##### Example App Manifest

```
---
applications:
  - name: nn-accumulator
    buildpack: binary_buildpack
    command: ./<accumulator-binary>
    memory: 128M
    instances: 1
    env:
      UAA_ADDR: https://login.bosh-lite.com
      CLIENT_ID: noisy-neighbor-nozzle
      CLIENT_SECRET: <secret for the client>
      NOZZLE_ADDRS: http://nnn.bosh-lite.com
      NOZZLE_COUNT: 3
      NOZZLE_APP_GUID: <nozzle app guid>
      SKIP_CERT_VERIFY: false
```

### DataDog Reporter Properties

| Property | Description |
|----------|-------------|
| `UAA_ADDR`         | The address of the Cloud Foundry deployed UAA. Normally `https://uaa.<system-domain>`. NOTE: The schema (e.g., `https`) is required. |
| `CAPI_ADDR`         | The address of the Cloud Foundry API. Normally `https://api.<system-domain>`. NOTE: The schema (e.g., `https`) is required. |
| `ACCUMULATOR_ADDR` | The addresses of the accumulator (e.g. `https://<accumulator-app-name>.<app-domain>`). |
| `CLIENT_ID`        | The [UAA client][uaa-user-vs-client] ID with the correct `authorities`. See [example](#example-uaa-client) UAA client. |
| `CLIENT_SECRET`    | The corresponding `secret` for the given `CLIENT_ID`. |
| `DATADOG_API_KEY` | The API Key to identify the DataDog account to send to. |
| `SKIP_CERT_VERIFY` | Set to true if the Cloud Foundry is using self signed certs. |

### Deploy the Datadog Reporter (optional)

```
cf push nn-datadog-reporter -b binary_buildpack -c ./<reporter-binary> --no-start --health-check-type none
cf set-env nn-datadog-reporter UAA_ADDR https://uaa.<system-domain>
cf set-env nn-datadog-reporter CAPI_ADDR https://api.<system-domain>
cf set-env nn-datadog-reporter ACCUMULATOR_ADDR https://nn-accumulator.<app-domain>
cf set-env nn-datadog-reporter CLIENT_ID <CLIENT_ID>
cf set-env nn-datadog-reporter CLIENT_SECRET <CLIENT_SECRET>
cf set-env nn-datadog-reporter DATADOG_API_KEY <DATADOG_API_KEY>
cf start nn-datadog-reporter
```

##### Example App Manifest

```
---
applications:
- name: nn-datadog-reporter
  buildpack: binary_buildpack
  command: ./<reporter-binary>
  memory: 128M
  instances: 1
  health-check-type: none
  env:
    UAA_ADDR: https://login.bosh-lite.com
    CAPI_ADDR: https://api.bosh-lite.com
    ACCUMULATOR_ADDR: https://nna.bosh-lite.com
    CLIENT_ID: noisy-neighbor-nozzle
    CLIENT_SECRET: <secret for the client>
    DATADOG_API_KEY: <datadog API key>
    REPORTER_HOST: bosh-lite.com
    SKIP_CERT_VERIFY: false
```

### CLI Plugin

#### Install

To install the CLI plugin from source run:

```
go get -u code.cloudfoundry.org/noisy-neighbor-nozzle/cli-plugin
cf install-plugin $GOPATH/bin/cli-plugin
```

#### Usage

To see the top 10 log producers by instance run:

```
cf log-noise nn-accumulator
```

If you deployed the accumulator with a different app name replace `nn-accumulator`
with that name.

[bosh]:              https://bosh.io
[nn-releases]        https://github.com/cloudfoundry/noisy-neighbor-nozzle/releases
[cf-cli]:            https://github.com/cloudfoundry/cli
[datadog]:           https://datadoghq.com
[ci-badge]:          https://loggregator.ci.cf-app.com/api/v1/pipelines/loggregator/jobs/noisy-neighbor-nozzle-bump-submodule/badge
[ci-pipeline]:       https://loggregator.ci.cf-app.com/teams/main/pipelines/loggregator/jobs/noisy-neighbor-nozzle-bump-submodule
[slack-badge]:       https://slack.cloudfoundry.org/badge.svg
[firehose-details]:  https://github.com/cloudfoundry/loggregator-release#consuming-the-firehose
[loggregator-slack]: https://cloudfoundry.slack.com/archives/loggregator
[noisy-neighbor-nozzle]:         https://code.cloudfoundry.org/noisy-neighbor-nozzle
[noisy-neighbor-nozzle-release]: https://code.cloudfoundry.org/noisy-neighbor-nozzle-release
[uaa-user-vs-client]: https://github.com/cloudfoundry/uaa/blob/master/docs/UAA-Tokens.md#users-and-clients-and-other-actors
