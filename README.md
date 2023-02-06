# datadog-remote-adapter
Covert datadog metrics to prometheus metrics.


## Mappings
Sometimes its just not possible to remap to the right thing. To work around this just provide an entry in the map config

```
mappings:
  kubernetes_state_container_memory_requested: kubernetes_state.container.memory_requested
  some_metric_foo: some.metric_foo
```