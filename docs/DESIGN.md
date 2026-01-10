# High Level Design

Sinkplot is meant to be a simple to use runtime-configurable reverse-proxy.

The `App` schema is a simple configuration object that allows you to choose your listeners, routes and sinks.

### Terminology

- `App` is the high-level container that hosts your various routes that are associated in some way. A good example is like `product-service`.
- `Listeners` is which ports sinkplot should listen on for requests for a given `App`.
- `Routes` is basically a multiplexer with different routing rules to dictate which route to send a request to.
  - `path` - the URL path to match against
  - `methods` - (optional) list of HTTP methods to match
  - `match` - (optional) matching strategy: `exact`, `prefix`, or `regex` (defaults to `exact`)
  - `sink` - the name of the sink to forward matching requests to
- `Sink` is one or more _upstream_ IP/hosts. These are the actual services you want to forward your request to.
  - `name` - identifier used by routes to reference this sink
  - `strategy` - (optional) load balancing strategy (e.g., `random`)
  - `upstreams` - list of upstream servers
- `Upstream` represents a single backend server.
  - `address` - IP or hostname of the upstream
  - `port` - port number
  - `weight` - (optional) weight for load balancing

### Usage

1. Create a YAML file with the following schema (certain fields are optional and can be omitted, while others have different options).

```yaml
app:
  name: product-service
  listeners:
  - 8080
  - 8081
  routes:
  - path: /backend/pay
    methods: ['GET', 'POST']
    match: exact
    sink: payments
  - path: /backend/v2
    methods: ['GET', 'POST', 'DELETE', 'PATCH', 'PUT', 'OPTIONS']
    match: prefix
    sink: v2
  - path: /backend
    methods: ['GET', 'POST', 'DELETE', 'PATCH', 'PUT', 'OPTIONS']
    match: prefix
    sink: v1
  sinks:
  - name: v1
    strategy: random
    upstreams:
    - address: '1.0.0.1'
      port: 80
      weight: 10
  - name: v2
    strategy: random
    upstreams:
    - address: '1.0.0.2'
      port: 80
      weight: 10
  - name: payments
    upstreams:
    - address: '1.0.0.3'
      port: 443
      weight: 10
```

2. Start sinkplot locally on a port of your choosing.

```bash
sinkctl start --port 8443
```

3. Send your configuration to the `/v1/config` endpoint.

```bash
curl -XPOST localhost:8443/v1/config \
  -H "Content-Type: application/yaml" \
  --data-binary @config.yaml
```

4. Now that the configuration was accepted, you can try to use sinkplot to hit one of your registered routes (assuming you have that backend actually running somewhere).

```bash
curl -XGET localhost:8080/backend/pay
```

