<h1 align="center">NOTRUST</h1>
<p align="center"><em>Nap Or Terminate: Resource Utilization Sleep Tool</em></p>

<p align="center">
  <a href="https://github.com/iammuuo/notrust/releases/latest"><img src="https://img.shields.io/github/v/release/iammuuo/notrust?style=flat-square" alt="Latest Release"></a>
  <a href="https://github.com/iammuuo/notrust/actions"><img src="https://img.shields.io/github/actions/workflow/status/iammuuo/notrust/release.yml?style=flat-square" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/iammuuo/notrust"><img src="https://goreportcard.com/badge/github.com/iammuuo/notrust?style=flat-square" alt="Go Report Card"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/github/license/iammuuo/notrust?style=flat-square" alt="License"></a>
</p>

<p align="center">Your Docker containers don't need to run while you're not using them.</p>

<p align="center">
  <!-- record with VHS (github.com/charmbracelet/vhs), then drop the gif here -->
  <img src=".github/demo.gif" alt="notrust pausing an idle container" width="640">
</p>

NOTRUST is a lightweight Linux daemon that watches your local Docker containers, notices when they've gone idle, and handles it for you: pause first, stop later if nobody comes back.

Less wasted RAM. Less CPU. Less noise.

```
    you stop working
           │
           ▼
      ┌────────┐    idle     ┌────────┐   still idle   ┌─────────┐
      │ ACTIVE │ ──────────▶ │ PAUSED │ ─────────────▶ │ STOPPED │
      └────────┘             └────────┘                └─────────┘
                             CPU → 0%                   RAM → freed
```

---

## Why

```sh
docker compose up -d
```

Postgres starts. Redis starts. RabbitMQ starts. Your API starts.

Then you go to lunch.

Three hours later, all four are still running. Tomorrow you start another project, and now there's a second Postgres, a second Redis, a second everything — quietly stacking up until your machine is spending more resources on things you *aren't* using than things you are.

NOTRUST exists so you don't have to remember to clean up your own development environment.

---

## Features

* **Pauses idle containers instantly** — a cgroup freeze, CPU drops to 0%, unpause takes ~50ms
* **Stops containers that stay idle** — hands the RAM all the way back
* **Actually checks idleness** — CPU delta, network delta, and open connections, not just "container exists"
* **Runs as a systemd user service** — no root daemon, no extra permissions
* **No dashboard. No account. No cloud.** — one daemon, one job

---

## Install

### One command

```sh
curl -fsSL https://raw.githubusercontent.com/iammuuo/notrust/main/install.sh | bash
```

Detects your architecture, downloads the latest release, installs the daemon, sets up the systemd unit, drops a default config, and starts it.

### From a release

Grab a `.tar.gz`, `.deb`, or `.rpm` from the [releases page](https://github.com/iammuuo/notrust/releases).

### From source

```sh
go install github.com/iammuuo/notrust/cmd/notrustd@latest
go install github.com/iammuuo/notrust/cmd/notrust@latest
```

**Requirements:** Linux · Docker · systemd · `curl` · `tar`
**Supported architectures:** `x86_64` · `arm64`

---

## After installing

```sh
systemctl --user status notrust     # is it running
journalctl --user -u notrust -f     # what it's doing
```

That's the whole interface for now. Leave it running and forget about it — that's the point.

---

## Configuration

Sensible defaults out of the box. Config lives at:

```
~/.config/notrust/config.yaml
```

A commented example ships with every release as `config.example.yaml`. Run with a specific file:

```sh
notrustd --config ./config.yaml
```

Every key can also be set as an environment variable prefixed `NOTRUST_`.

---

## What it isn't

not a container orchestrator · not a monitoring platform · not a Docker Compose replacement · not a Kubernetes tool · not a cloud service

It's a small utility for local Docker environments. One problem, solved quietly.

---

## Status

- [x] Idle detection — CPU, network, connections
- [x] Pause on idle
- [x] Escalate to stop after sustained idle
- [x] Runs as a systemd user service
- [ ] CLI tooling
- [ ] Desktop notifications
- [ ] Per container policy overrides
- [ ] Richer status output

Early days. The core loop works; the tooling around it is still being built.

---

## The name

People keep rewriting perfectly good tools in Rust because someone on the internet told them memory safety is a personality trait.

This one is staying in Go. Pausing a container is one syscall, not a lifestyle.

---

## Contributing

Still early. Broken things, ideas, and pull requests are all welcome — open an issue.

---

## License

[Apache License 2.0](./LICENSE)

<p align="center"><sub>not rust. just go, just build, just done.</sub></p>
