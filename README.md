# loan-svc

service to handle loaning system.

## Initialization & Requirements

as usual in many go project we run `go mod tidy` to fetch all required dependencies,
after that the project is ready to be started by `go run ./cmd/loan-svc/main.go`.

note that our service is context-aware that will help with all modern technology to support it,
we're missing telemetry (preferably using open-telemetry) but this is fairly easy to implement,
aside from telemetry we're also able to implement graceful shutdown in case of autoscaling event,
many issues arise when we're not properly killing the service leading to states of irregularity,
mainly in transactional data, thus properly closing all db connection & transaction is a must have
on microservice architecture.

on database, we choose for simplicity, we're using sqlite3 with dependency to `github.com/mattn/go-sqlite3`,
upon starting up for the first time will initialize the migration process, reflected by code on
`./internal/repository/datastore/datastore.go:105-109`, creating tables using script on
`./internal/repository/datastore/queries/loan-svc.sqlite3.migration.000.sql`. Also in this repo, it exists
`./local.db` and prepopulated with some data, user can try the HTTP REST API via `./rest.http` file.

## Quick Review

we're using YAML files to manage our `config` & `secret`, `config` consists of mainly feature flags &
configuration defined from other packages, while `secret` is used to manage secret & sensitive like DB dsn, API key, etc.

we're using a variation of clean code & SOLID, named FoReST (Feature-oriented, Repository, Service, & Test),
as we can see `./internal` splitted into `./internal/feature`, `./internal/repository`, `./internal/service`.
`service` is layer from our endpoint to `feature` mostly just doing mapping request & response,
`feature` hold our business logic and depends on `repository` to fetch from stateful datasource,
`repository` is mapping our datasource like DB, API, localfile, etc.

## Improvement

definitive error type for `fmt.Errorf` & `errors.New` so that we could use `errors.As` & `errors.Is`
for more detailed error handling, the usecase is this returned error is very much needed to return
detailed information that useful in our telemetry like logging, but the same error should be brief
and doesn't contains any sensitive data when return to the client, e.g. unsuccesful login attempt,
system will need to log the userID, IP, etc that maybe useful for defense mechanism, but user/client
don't need all of that, they only need to know what's wrong like simple "username / password mismatch"

telemetry is a requirement for microservices handling multitude of services, so when any issues arise,
we could be responding quickly, by employing logs, metrics, & traces, we could build our dashboard &
alert manager that could be notified on-call engineer to take a quick glance and respond accordingly.
It could be false positive and we should update the alert manager to response better in the future,
or if it's a real issues, on-call engineer can planned for the follow-up, usualy in form of root-cause analysis.
other than that, this telemetry could help tech lead/ manager planning the services to be more attractive &
efficient, e.g. looking at trafic trends, planning new features, etc.
