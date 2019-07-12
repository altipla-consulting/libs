
This changelog will list breaking changes in our APIs.


## v1.40.0:

- `redis.Database.DirectClient` is replaced by its equivalent `redis.Database.Cmdable`, but now it needs a context. When a normal context is passed no difference in behaviour is expected. If a transactional context is used a transactional connection will be returned instead.
- All write & read methods of every redis type now needs a context.


## v1.39.0:

- `services.Dial` is replaced by `connect.Internal` and it doesn't need the `grpc.WithInsecure()` argument, it is automatically added now. Any other option will remain the same.
- `services.Endpoint is now `connect.Endpoint` with the same semantics.
- `service.ConfigureBetaRouting` when configuring a new service is now removed. It was deprecated before. Use `ConfigureRouting(routing.WithBetaAuth(username, password))` instead.
