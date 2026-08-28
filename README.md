# NOTRUST

### Nap Or Terminate: Resource Utilization Sleep Tool

You forget to stop your containers. Everyone does. NOTRUST notices before your fans do.

Postgres, Redis, RabbitMQ, whatever you spun up an hour ago and abandoned for a meeting, still running, still burning cycles, still holding onto RAM you will want back the second Chrome opens forty more tabs. NOTRUST watches. When you stop using something, it pauses it. When you keep ignoring it, it stops it and gives the memory back. When you come back, it just works.

## The name

People keep rewriting perfectly fine tools in Rust because someone on the internet told them memory safety is a personality trait. This one is staying in Go. Pausing a container is one syscall, not a lifestyle.

## What it actually does

```
$ notrust status
postgres_dev   paused    idle 12m
redis_dev      active
api_dev        stopped   idle 47m
```

No dashboard to check. No config you have to think about. It just does the thing you keep forgetting to do.

## Status

Early. Small. Does one thing.

<p align="center"><sub>btw it's not rust</sub></p>
