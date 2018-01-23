Noisy Neighbor Nozzle
[![slack.cloudfoundry.org][slack-badge]][loggregator-slack]
[![CI Badge][ci-badge]][ci-pipeline]
=====================

This is a Loggregator Firehose nozzle. It keeps track of the log rates for
Cloud Foundry deployed applications.

This nozzle can be deployed via [BOSH][bosh] using the
[Noisy Neighbor Nozzle Release][noisy-neighbor-nozzle-release] or via CF push.

## How it works.

The noisy neighbor nozzle consists of three components, nozzle, accumulator and
datadog-reporter.

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

- `timestamp` - Unix timetamp truncated to the nozzles `POLLING_INTERVAL`
  (Default is 1 minute).

#### Example

```
curl -H "Authorization: $AUTH_TOKEN" https://nn-accumulator.<app-domain>/rates/$(date --date="-5 minutes" +%s)
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

Download the binaries from [releases](https://github.com/cloudfoundry/noisy-neighbor-nozzle/releases), set the environment variables via the cf cli or an app manifest.

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

[bosh]:              https://bosh.io
[datadog]:           https://datadoghq.com
[ci-badge]:          https://loggregator.ci.cf-app.com/api/v1/pipelines/loggregator/jobs/noisy-neighbor-nozzle-bump-submodule/badge
[ci-pipeline]:       https://loggregator.ci.cf-app.com/teams/main/pipelines/loggregator/jobs/noisy-neighbor-nozzle-bump-submodule
[slack-badge]:       https://slack.cloudfoundry.org/badge.svg
[firehose-details]:  https://github.com/cloudfoundry/loggregator-release#consuming-the-firehose
[loggregator-slack]: https://cloudfoundry.slack.com/archives/loggregator
[noisy-neighbor-nozzle]:         https://code.cloudfoundry.org/noisy-neighbor-nozzle
[noisy-neighbor-nozzle-release]: https://code.cloudfoundry.org/noisy-neighbor-nozzle-release
