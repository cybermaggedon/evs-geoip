# `evs-geoip`

Eventstream analytic for Cyberprobe event streams.  Subscribes to Pulsar
for Cyberprobe events and annotates events with location information derived
from IP addresses using MaxMind GeoIP.

## Getting Started

The target deployment product is a container engine.  The analytic expects
a Pulsar service to be running.

You need a MaxMind licence/subscription, and to acquire GeoIP files to use
this analytic (available free).

```
  docker run -d \
      -e PULSAR_BROKER=pulsar://<PULSAR-HOST>:6650 \
      -e GEOIP_DB=/geoip/GeoLite2-City.mmdb \
      -e GEOIP_ASN_DB=/geoip/GeoLite2-ASN.mmdb \
      -v ./geoip:/geoip \
      -p 8088:8088 \
      docker.io/cybermaggedon/evs-geoip:<VERSION>
```

The above command mounts a directory under `/geoip` in the container,
and configures to use the GeoIP files contained therein. 

### Prerequisites

You need to have a container deployment system e.g. Podman, Docker, Moby.

You also need a Pulsar exchange, being fed by events from Cyberprobe.

### Installing

The easiest way is to use the containers we publish to Docker hub.
See https://hub.docker.com/r/cybermaggedon/evs-geoip

```
  docker pull docker.io/cybermaggedon/evs-geoip:<VERSION>
```

If you want to build this yourself, you can just clone the Github repo,
and type `make`.

## Deployment configuration

The following environment variables are used to configure:

| Variable | Purpose | Default |
|----------|---------|---------|
| `INPUT` | Specifies the Pulsar topic to subscribe to.  This is just the topic part of the URL e.g. `cyberprobe`. By default the input is `cyberprobe` which is the output  of `cybermon`. | `cyberprobe` |
| `OUTPUT` | Specifies a comma-separated list of Pulsar topics to publish annotated events to.  This is just the topic part of the URL e.g. `geo`. By default, the output is `withloc`. | `withloc` |
| `GEOIP_DB` | Specifies a filename of the GeoIP city database.  | |
| `GEOIP_ASN_DB` | Specifies a filename of the GeoIP ASN database.  | |
| `METRICS_PORT` | Specifies the port number to serve Prometheus metrics on.  If not set, metrics will not be served. The container has a default setting of 8088. | `8088` |

