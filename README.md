# Test ODCDS

A dummy ODCDS server for Envoy.

## Running

```
go run .
```

In a new shell:

```
# Configure the path to the Envoy binary.
export ENVOY_BIN=/tmp/envoy-bin/envoy

# Run Envoy.
$ENVOY_BIN --config-path examples/envoy.yaml
```

Send a request through Envoy:

```
curl -H "Cluster-Name: foo" localhost:8080
```
