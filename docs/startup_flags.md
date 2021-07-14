### Controller Startup Flags

Below is a list of command-line flags accepted by Morphling controller:

| Flag Name|  Type | Description    | Default |
|----------|---------|-------------| -----|
|metrics-addr|string|The address the metric endpoint binds to| 8088 
enable-leader-election |bool| Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager. | false
enable-grpc-probe-in-suggestion |bool|  Enable Pod readiness/liveness probes in samplings | true
