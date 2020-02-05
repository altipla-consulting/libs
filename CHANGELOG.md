
This changelog will list only breaking changes in our APIs.


## v1.55.0

- **money:** `Format` now receives a `FormatConfig` object instead of a number. `Display` and `FormatHuman` are removed. If you want to keep the functionality use the equivalent `Format`. Money uses int32 for all operations now instead of int64.


## v1.54.0

- **connect:** Deleted without possible replacement. It is recommended to copy it externally.
- **services:** Deleted without possible replacement. It is recommended to copy it externally.


## v1.52.0

- **money:** Replace all usages of the removed `Markup` method with `AddTaxPercent`, which has the same functionality and parameters.


## v1.49.0

- **routing:** `NotFound`, `Unauthorized`, `BadRequest`, and `Internal` do not receive a format now. If you want to keep the functionality use the equivalent `NotFoundf`, `Unauthorizedf`, `BadRequestf` and `Internalf`.


## v1.46.0:

- `database.Database.Select`, `database.Collection.GetAll`, `database.Collection.Put`, ...: All the operations sending or receiving data receive now a context as a first argument. This allows these methods to be used inside a transaction.
- `database.Database.SelectContext`, `database.Collection.GetAllContext`, `database.Collection.PutContext`, ...: Removed and replaced with their normal counterparts now that they receive contexts.
- `pagination.Pager.Fetch` now receives a context to use the new transactional functions of the database package.


## v1.40.0:

- `redis.Database.DirectClient` is replaced by its equivalent `redis.Database.Cmdable`, but now it needs a context. When a normal context is passed no difference in behaviour is expected. If a transactional context is used a transactional connection will be returned instead.
- All write & read methods of every redis type now needs a context.


## v1.39.0:

- `services.Dial` is replaced by `connect.Internal` and it doesn't need the `grpc.WithInsecure()` argument, it is automatically added now. Any other option will remain the same.
- `services.Endpoint is now `connect.Endpoint` with the same semantics.
- `service.ConfigureBetaRouting` when configuring a new service is now removed. It was deprecated before. Use `ConfigureRouting(routing.WithBetaAuth(username, password))` instead.
