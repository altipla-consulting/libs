
## v1.39.0:

- `services.Dial` is replaced by `connect.Internal` and it doesn't need the `grpc.WithInsecure()` argument, it is automatically added now. Any other option will remain the same.
- `services.Endpoint is now `connect.Endpoint` with the same semantics.
- `service.ConfigureBetaRouting` when configuring a new service is now removed. It was deprecated before. Use `ConfigureRouting(routing.WithBetaAuth(username, password))` instead.
