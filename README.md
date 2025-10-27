# 🧭 Chronum

**Chronum** is a local-first workflow runtime built for developers — a modern, lightweight orchestrator that runs your automation DAGs directly from the command line, with optional sync to the cloud.

> Think of it as “GitHub Actions, but local and hackable.”

---

## 🚀 Overview

Chronum lets you define task flows (DAGs) in YAML or code, execute them locally with full visibility, and optionally sync state or logs to a shared workspace.  
It’s designed for **developer workflows** — CI/CD, data pipelines, build orchestration, and test automation — without the friction of remote runners.

---

## ✨ Core Features

| Capability | Description |
|-------------|--------------|
| **🧩 Pluggable Parsers** | Supports YAML and future typed DSLs (TypeScript/Python SDKs). |
| **⚙️ DAG Execution Engine** | Executes steps in correct dependency order, parallelized where possible. |
| **🔀 Concurrency Control** | `maxParallel` config controls how many tasks can run simultaneously. |
| **🛑 Failure Policies** | `stopOnFail`, `retry` logic, and dependency-aware halting ensure deterministic runs. |
| **⏱️ Exponential Backoff** | Built-in retry backoff with jitter for network or transient failures. |
| **🧠 Executor Dispatch** | Each step can use a different executor (e.g. `shell`, `python`, `docker`, custom). |
| **🖥️ Local-First Design** | SQLite-backed state store; no cloud dependency required. |
| **📡 Event Bus (upcoming)** | Publish structured `start/success/fail` events for the local dashboard or WebSocket observers. |

---

## 📦 Project Structure

```bash
chronum/
├── cmd/
│   └── test/                 # Example entrypoint for testing flows
├── engine/                   # Execution engine, runner, state, executors
├── parser/                   # DAG + parser + YAML handler
│   ├── dag.go
│   ├── yaml/
│   └── types/
└── examples/
    └── chronum.yaml          # Example workflow definition
```

---

## 🧱 Example Workflow

```yaml
name: deploy-api
maxParallel: 3
stopOnFail: true
defaultRetry: 2

steps:
  - name: build
    run: echo "Building project"

  - name: test
    run: echo "Running tests"
    needs: [build]

  - name: lint
    run: echo "Linting code"
    needs: [build]

  - name: deploy
    run: echo "Deploying..."
    needs: [test, lint]

  - name: preprocess
    executor: python
    run: scripts/preprocess.py

  - name: deployed
    run: ./deploy.sh
```

---

## 🧰 Running a Flow

```bash
go run ./cmd/test
```
---

## 🧩 Example Output

```
✅ Parsed Chronum: deploy-api
🚀 Running deploy-api (maxParallel=3, stopOnFail=true, retry=2)
[build] Running...
→ [shell] executing step: build
"Building project"
[build] complete
[lint] Running...
→ [shell] executing step: lint
"Linting code"
[lint] complete
[test] Running...
→ [shell] executing step: test
"Running tests"
[test] complete
[deploy] Running...
→ [shell] executing step: deploy
"Deploying..."
[deploy] complete

✅ Flow complete.
```

---

## ⚙️ Configuration Reference

| Field | Description | Default |
|--------|--------------|----------|
| `name` | Name of the flow | — |
| `maxParallel` | Maximum concurrent tasks | `1` |
| `stopOnFail` | Halt the DAG on first failure | `false` |
| `defaultRetry` | Number of retry attempts per step | `0` |
| `executor` | Step executor (`shell`, `python`, `docker`, custom) | `shell` |

---

## 🔌 Extending Chronum

Chronum is fully extensible:

- **Executors** — implement the `Executor` interface (`Execute(node *dag.Node) error`)
- **Parsers** — register new formats via `parser.Register("format", parserImpl)`
- **Event Bus (WIP)** — connect your own event subscriber to stream real-time state changes
- **Exporters (future)** — send run data to external systems (e.g., Prometheus, S3, Datadog)

---

## 🧩 Roadmap

| Milestone     | Status     | Description                                     |
| ------------- | ---------- | ----------------------------------------------- |
| **ENG-004.2** | ✅          | Max concurrency limit                           |
| **ENG-004.3** | ✅          | Failure policies (`stopOnFail`, `retry`)        |
| **ENG-004.4** | ✅          | Executor dispatch (`shell`, `python`, `docker`) |
| **ENG-005**   | ✅          | Exponential backoff with jitter                 |
| **OBS-001**   | 🚧          | EventBus for UI + structured logs               |
| **OBS-002**   | 🧠 Planned  | Local dashboard for DAG visualization           |
| **NET-001**   | 🧠 Planned  | Chronum Cloud sync & team mode                  |

---
## 🧑‍💻 Philosophy

Chronum’s guiding principles:

1. **Local-first** — your automation should work offline.
2. **Composable** — everything is a plugin or module.
3. **Observable** — runs should be inspectable and replayable.
4. **Deterministic** — same DAG, same result, every time.

---

## 🧪 Status

Chronum is in early development.  
Core engine features are implemented and tested.  
Next milestone: event bus + local dashboard.

Contributions, discussions, and crazy ideas are welcome.

---