# Early development
WARNING, this is in early development and many things dont work. But very very basic queries should.

parsing out values `by` currently is the big one.

* sum(kubernetes_state_container_memory_requested) by (SOMETHING)

# datadog-remote-adapter
Covert datadog metrics to prometheus metrics.


## Mappings
Sometimes its just not possible to remap to the right thing. To work around this just provide an entry in the map config

```
mappings:
  kubernetes_state_container_memory_requested: kubernetes_state.container.memory_requested
  some_metric_foo: some.metric_foo
```

## Kubernetes setup
Setup your k8s secret `datadog` prior to running
see documentation [here](docs/examples/kubernetes.yaml)