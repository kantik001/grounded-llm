# ИНСТРУКЦИЯ ДЛЯ AI-АГЕНТОВ CURSOR
## Проект: Verifiable AI Infrastructure Platform
## Версия: 1.0 | Дата: 28 июля 2026

> **Superseded for execution priority** by [`CAREER_OPTIMIZED_PLAN_v1.1.md`](./CAREER_OPTIMIZED_PLAN_v1.1.md).  
> Keep this file as historical backlog only.  
> **Port correction:** `grounded-guardrails` listens on **`:50052`** (not `:50051` — Retriever owns `:50051`).

---

## 0. МЕТАДАННЫЕ И ПРАВИЛА

### 0.1. Роль агента

Ты — senior AI infrastructure engineer. Твоя задача — пошагово реализовать платформу verifiable AI infrastructure. Ты пишешь production-код, а не прототипы. Каждый файл должен компилироваться, каждый модуль — иметь тесты.

### 0.2. Стек и версии

```yaml
languages:
  go: "1.25+"
  rust: "1.85+ (edition 2024)"
  python: "3.12+"
  cpp: "C++20"
  cuda: "12.4+"

tools:
  build:
    go: "go build"
    rust: "cargo build --release"
    python: "uv" или "pip"
    cpp: "cmake 3.28+, nvcc"
  test:
    go: "go test -race -cover"
    rust: "cargo test"
    python: "pytest"
    cpp: "ctest + gtest"
  lint:
    go: "golangci-lint"
    rust: "clippy + rustfmt"
    python: "ruff"
    cpp: "clang-tidy"
  ci: "GitHub Actions"
  container: "Docker 26+"
```

### 0.3. Правила кода

```yaml
commits:
  format: "type(scope): description"
  types: [feat, fix, test, docs, refactor, perf, ci]
  example: "feat(guardrails): add token ring buffer with zero-copy"

code_style:
  go:
    - golangci-lint run
    - все экспортируемые функции с godoc
    - ошибки оборачивать: fmt.Errorf("context: %w", err)
  rust:
    - cargo clippy -- -D warnings
    - cargo fmt
    - никаких unwrap() в production-коде
    - thiserror для ошибок
  python:
    - ruff check --fix
    - type hints везде
    - pydantic для моделей данных
  cpp:
    - clang-tidy
    - никаких raw new/delete
    - smart pointers только

testing:
  minimum_coverage: 80%
  required:
    - unit tests для каждого модуля
    - integration tests для API
    - benchmark tests для performance-критичных путей
  naming: "TestFunctionName_Scenario_Expected"

documentation:
  - README.md на английском для каждого репо
  - ARCHITECTURE.md с диаграммой
  - BENCHMARKS.md с цифрами
  - все публичные API с примерами
```

### 0.4. Структура монорепозитория

```
verifiable-ai-platform/
├── README.md
├── ARCHITECTURE.md
├── BENCHMARKS.md
├── Makefile
├── docker-compose.yml
├── .github/workflows/ci.yml
│
├── grounded-llm/              # Существующий (Go + Python)
│   ├── cmd/server/main.go
│   ├── internal/
│   ├── api/proto/
│   └── python/retrieval/
│
├── grounded-agent/            # Существующий (Go)
│   ├── cmd/agent/main.go
│   └── internal/
│
├── mcp-gateway/               # Существующий (Go)
│   ├── cmd/gateway/main.go
│   └── internal/
│
├── grounded-guardrails/       # НОВЫЙ (Rust + Python + Go)
│   ├── rust/
│   │   ├── Cargo.toml
│   │   ├── src/
│   │   └── tests/
│   ├── python/
│   │   ├── pyproject.toml
│   │   └── grounded_guardrails/
│   ├── go/
│   │   ├── go.mod
│   │   └── cmd/server/main.go
│   └── proto/guardrails.proto
│
├── grounded-wasm/             # НОВЫЙ (Rust + Go)
│   ├── host/                  # Go (wazero)
│   │   ├── go.mod
│   │   └── cmd/wasm-gateway/main.go
│   ├── guests/                # Rust → WASM
│   │   ├── calculator/
│   │   ├── sql-sandbox/
│   │   └── http-fetch/
│   └── proto/wasm.proto
│
├── grounded-eval/             # НОВЫЙ (Go + Rust + Python)
│   ├── coordinator/           # Go
│   ├── worker/                # Rust
│   ├── metrics/               # Python
│   └── specs/                 # YAML eval specs
│
├── grounded-bench/            # НОВЫЙ (Python + Go)
│   ├── dataset/               # Python: генерация датасета
│   ├── runner/                # Go: запуск бенчмарков
│   └── leaderboard/           # Static HTML
│
└── grounded-kernels/          # НОВЫЙ (C++ + CUDA + Python)
    ├── csrc/
    │   ├── numeric_verify.cu
    │   ├── pii_detect.cu
    │   └── bindings.cpp
    ├── python/
    │   ├── pyproject.toml
    │   └── grounded_kernels/
    ├── CMakeLists.txt
    └── tests/
```

---

## ФАЗА 1: grounded-guardrails (Недели 1–8)

### Спринт 1.1: Rust token buffer (Неделя 1–2)

#### Задача 1.1.1: Инициализация Rust-проекта

```bash
cd grounded-guardrails/rust
cargo init --name grounded-guardrails
```

**Файл: `Cargo.toml`**
```toml
[package]
name = "grounded-guardrails"
version = "0.1.0"
edition = "2024"

[dependencies]
tokio = { version = "1", features = ["full"] }
thiserror = "2"
tracing = "0.1"
tracing-subscriber = "0.3"
regex = "1"
serde = { version = "1", features = ["derive"] }
serde_json = "1"

[dev-dependencies]
criterion = "0.5"
proptest = "1"

[[bench]]
name = "token_buffer"
harness = false
```

**Критерий приёмки:** `cargo build` проходит без ошибок. `cargo clippy -- -D warnings` чист.

---

#### Задача 1.1.2: Token Ring Buffer

**Файл: `src/buffer.rs`**

Создать структуру `TokenRingBuffer`:

```rust
/// Zero-copy ring buffer для streaming токенов.
/// Вместимость: 4096 токенов (фиксированная).
/// При переполнении: перезапись самых старых.
pub struct TokenRingBuffer {
    tokens: Vec<u32>,      // token IDs
    positions: Vec<usize>, // позиции в оригинальном потоке
    head: usize,
    len: usize,
    capacity: usize,
}

impl TokenRingBuffer {
    pub fn new(capacity: usize) -> Self;
    
    /// Добавить токен. O(1).
    pub fn push(&mut self, token_id: u32, position: usize);
    
    /// Получить последние N токенов без аллокации.
    /// Возвращает итератор.
    pub fn last_n(&self, n: usize) -> impl Iterator<Item = (u32, usize)> + '_;
    
    /// Текущая длина.
    pub fn len(&self) -> usize;
    
    /// Очистить.
    pub fn clear(&mut self);
}
```

**Требования:**
- Никаких аллокаций при `push()` после инициализации
- `last_n()` возвращает итератор, не Vec
- Thread-safe: `Send + Sync`

**Тесты:**
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_push_and_last_n() {
        // push 10 токенов, last_n(3) возвращает последние 3
    }
    
    #[test]
    fn test_overflow() {
        // push capacity+10, len == capacity, старые перезаписаны
    }
    
    #[test]
    fn test_no_allocation_on_push() {
        // использовать dhat или proptest для проверки
    }
}
```

**Бенчмарк: `benches/token_buffer.rs`**
```rust
use criterion::{criterion_group, criterion_main, Criterion};

fn bench_push(c: &mut Criterion) {
    // 1M push операций < 10ms
}

fn bench_last_n(c: &mut Criterion) {
    // last_n(128) на полном буфере < 100ns
}
```

**Критерий приёмки:**
- `cargo test` — все тесты проходят
- `cargo bench` — push: < 10ns/операция, last_n: < 100ns
- `cargo clippy` — 0 warnings

---

#### Задача 1.1.3: Regex PII Detection

**Файл: `src/pii.rs`**

```rust
use regex::Regex;
use std::sync::LazyLock;

pub enum PiiType {
    Email,
    Phone,
    Ssn,
    CreditCard,
}

pub struct PiiMatch {
    pub pii_type: PiiType,
    pub start: usize,
    pub end: usize,
    pub masked: String, // "j***@example.com"
}

/// Детектор PII на основе regex.
/// Все regex компилируются один раз (LazyLock).
pub struct PiiDetector {
    email: Regex,
    phone: Regex,
    ssn: Regex,
    credit_card: Regex,
}

impl PiiDetector {
    pub fn new() -> Self;
    
    /// Проверить текст. Возвращает все найденные PII.
    /// Производительность: < 100μs на 1KB текста.
    pub fn detect(&self, text: &str) -> Vec<PiiMatch>;
    
    /// Быстрый путь: только проверка наличия (без позиций).
    pub fn contains_pii(&self, text: &str) -> bool;
}
```

**Тесты:**
- Email: `user@example.com`, `first.last+tag@sub.domain.co.uk`
- Phone: `+1 (555) 123-4567`, `555.123.4567`, `+7 999 123-45-67`
- SSN: `123-45-6789`
- Credit Card: `4111 1111 1111 1111`, `4111111111111111`
- False positives: `version 1.2.3`, `order #123-456`

**Критерий приёмки:**
- Все тесты проходят
- `contains_pii()` на 1KB: < 50μs (criterion bench)
- 0 false positives на тестовом наборе

---

#### Задача 1.1.4: Numeric Extraction

**Файл: `src/numeric.rs`**

```rust
pub struct NumericValue {
    pub value: f64,
    pub start: usize,
    pub end: usize,
    pub raw: String, // "14,000,000" или "14.5%"
}

/// Извлечь все числа из текста.
/// Поддержка: integers, floats, percentages, comma-separated.
pub fn extract_numerics(text: &str) -> Vec<NumericValue>;

/// Сверить число с контекстом.
/// tolerance: абсолютное отклонение (по умолчанию 0.01).
pub fn verify_numeric(value: f64, context: &[NumericValue], tolerance: f64) -> bool;
```

**Тесты:**
- `"Выручка составила 14,000,000 рублей"` → `14000000.0`
- `"Ставка НДС 20%"` → `20.0`
- `"Курс 92.5 ₽/$"` → `92.5`
- `"±0.01"` → `0.01`
- Verify: `14000000.0` в контексте `[14000000.0]` → `true`
- Verify: `14000000.0` в контексте `[15000000.0]` → `false`

**Критерий приёмки:**
- Все тесты проходят
- Extraction на 1KB: < 100μs
- Verify на 100 числах: < 10μs

---

### Спринт 1.2: Python ML-слой (Неделя 3–4)

#### Задача 1.2.1: Инициализация Python-проекта

```bash
cd grounded-guardrails/python
uv init grounded-guardrails
```

**Файл: `pyproject.toml`**
```toml
[project]
name = "grounded-guardrails"
version = "0.1.0"
requires-python = ">=3.12"
dependencies = [
    "onnxruntime>=1.19",
    "numpy>=2.0",
    "transformers>=4.45",
    "pydantic>=2.9",
    "grpcio>=1.66",
    "grpcio-tools>=1.66",
]

[project.optional-dependencies]
dev = ["pytest>=8", "pytest-benchmark", "ruff"]
```

**Критерий приёмки:** `uv sync` проходит. `python -c "import grounded_guardrails"` работает.

---

#### Задача 1.2.2: Toxicity Classifier (ONNX)

**Файл: `grounded_guardrails/toxicity.py`**

```python
from dataclasses import dataclass
import numpy as np
import onnxruntime as ort

@dataclass
class ToxicityVerdict:
    score: float          # 0.0 – 1.0
    is_toxic: bool        # score > threshold
    label: str            # "toxic" | "neutral"

class ToxicityClassifier:
    """ONNX-based toxicity classifier.
    Модель: distilbert-toxicity (экспортировать в ONNX заранее).
    Batch inference: до 32 текстов за раз.
    """
    
    def __init__(self, model_path: str, threshold: float = 0.7):
        self.session = ort.InferenceSession(
            model_path,
            providers=["CUDAExecutionProvider", "CPUExecutionProvider"],
        )
        self.threshold = threshold
    
    def classify(self, text: str) -> ToxicityVerdict:
        """Классифицировать один текст."""
        ...
    
    def classify_batch(self, texts: list[str]) -> list[ToxicityVerdict]:
        """Batch классификация. Максимум 32 текста."""
        ...
```

**Тесты:**
- `"You are great"` → score < 0.3, is_toxic = False
- `"I will kill you"` → score > 0.7, is_toxic = True
- Batch из 10 текстов → 10 вердиктов

**Критерий приёмки:**
- `pytest` проходит
- Latency: < 5ms на текст (CPU), < 1ms (GPU)
- Batch 32: < 20ms (GPU)

---

#### Задача 1.2.3: Numeric Verify (Python-обёртка)

**Файл: `grounded_guardrails/numeric_verify.py`**

```python
from pydantic import BaseModel

class NumericVerifyRequest(BaseModel):
    answer: str
    context: str
    tolerance: float = 0.01

class NumericVerifyResult(BaseModel):
    passed: bool
    answer_numbers: list[float]
    context_numbers: list[float]
    unmatched: list[float]  # числа в ответе, не найденные в контексте

def verify_numerics(req: NumericVerifyRequest) -> NumericVerifyResult:
    """Извлечь числа из answer и context, сверить с tolerance."""
    ...
```

**Критерий приёмки:**
- Тесты на русской и английской локали
- `"Выручка 14 млн"` + context `"14,000,000"` → passed = True
- `"Выручка 15 млн"` + context `"14,000,000"` → passed = False

---

### Спринт 1.3: Go gRPC API (Неделя 5–6)

#### Задача 1.3.1: Proto-контракт

**Файл: `proto/guardrails.proto`**
```protobuf
syntax = "proto3";
package guardrails.v1;

option go_package = "github.com/kantik001/verifiable-ai-platform/grounded-guardrails/go/proto";

service GuardrailsService {
  // Streaming verification: клиент шлёт токены, сервер верифицирует.
  rpc VerifyStream(stream TokenBatch) returns (stream Verdict);
  
  // Unary: проверить готовый текст.
  rpc VerifyText(TextRequest) returns (TextVerdict);
}

message TokenBatch {
  repeated uint32 token_ids = 1;
  string tenant_id = 2;
  string session_id = 3;
}

message Verdict {
  enum Action {
    PASS = 0;
    BLOCK = 1;
    FLAG = 2;
  }
  Action action = 1;
  string reason = 2;
  repeated string matched_rules = 3;
  float latency_ms = 4;
}

message TextRequest {
  string text = 1;
  string context = 2;
  string tenant_id = 3;
  repeated string rules = 4; // ["numeric_verify", "pii_block", "toxicity"]
}

message TextVerdict {
  bool passed = 1;
  repeated string violations = 2;
  float latency_ms = 3;
}
```

**Команда:**
```bash
cd grounded-guardrails
protoc --go_out=. --go-grpc_out=. proto/guardrails.proto
```

**Критерий приёмки:** Сгенерированные Go-файлы компилируются.

---

#### Задача 1.3.2: Go gRPC сервер

**Файл: `go/cmd/server/main.go`**

```go
package main

import (
    "log"
    "net"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    healthpb "google.golang.org/grpc/health/grpc_health_v1"
    
    pb "github.com/kantik001/verifiable-ai-platform/grounded-guardrails/go/proto"
)

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    
    grpcServer := grpc.NewServer()
    
    // Guardrails service
    guardrailsSvc := NewGuardrailsService()
    pb.RegisterGuardrailsServiceServer(grpcServer, guardrailsSvc)
    
    // Health check
    healthServer := health.NewServer()
    healthpb.RegisterHealthServer(grpcServer, healthServer)
    healthServer.SetServingStatus("guardrails.v1.GuardrailsService", healthpb.HealthCheckResponse_SERVING)
    
    log.Println("grounded-guardrails listening on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

**Критерий приёмки:**
- `go build ./...` проходит
- `go test -race ./...` проходит
- `grpcurl -plaintext localhost:50051 list` показывает сервис

---

### Спринт 1.4: Интеграция и бенчмарки (Неделя 7–8)

#### Задача 1.4.1: Docker Compose

**Файл: `docker-compose.yml` (корневой)**
```yaml
services:
  guardrails:
    build: ./grounded-guardrails
    ports:
      - "50051:50051"
    environment:
      - TOXICITY_MODEL_PATH=/models/distilbert-toxicity.onnx
    volumes:
      - ./models:/models:ro
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

#### Задача 1.4.2: End-to-end бенчмарк

**Файл: `grounded-guardrails/BENCHMARKS.md`**

Запустить и зафиксировать:

```markdown
## Token-level verification (Rust, CPU)
- 1M токенов, ring buffer push: X ms (цель: < 10ms)
- PII detection на 1KB: X μs (цель: < 50μs)
- Numeric extraction на 1KB: X μs (цель: < 100μs)

## Toxicity classification (Python, GPU)
- 1 текст: X ms (цель: < 1ms)
- Batch 32: X ms (цель: < 20ms)

## gRPC streaming (Go)
- 10K токенов/сек: p99 latency X ms (цель: < 5ms)
```

**Критерий приёмки:**
- Все бенчмарки запущены, цифры зафиксированы
- `docker compose up` поднимает сервис
- gRPC health check отвечает `SERVING`

---

## ФАЗА 2: grounded-wasm (Недели 9–16)

### Спринт 2.1: Go host на wazero (Неделя 9–10)

#### Задача 2.1.1: Инициализация

```bash
cd grounded-wasm/host
go mod init github.com/kantik001/verifiable-ai-platform/grounded-wasm/host
go get github.com/tetratelabs/wazero@latest
```

#### Задача 2.1.2: WASM Runtime с capabilities

**Файл: `host/internal/runtime/runtime.go`**

```go
package runtime

import (
    "context"
    
    "github.com/tetratelabs/wazero"
    "github.com/tetratelabs/wazero/api"
)

// Capability определяет, что разрешено WASM-модулю.
type Capability struct {
    FSReadPaths  []string // разрешённые пути для чтения
    FSWritePaths []string // разрешённые пути для записи
    NetWhitelist []string // разрешённые домены
    MaxMemoryMB  uint32   // лимит памяти
    MaxCPUms     uint64   // лимит CPU времени
}

// Runtime управляет WASM-модулями с capability-based security.
type Runtime struct {
    r wazero.Runtime
}

func New(ctx context.Context) (*Runtime, error);

// LoadModule компилирует WASM-модуль с ограничениями.
func (rt *Runtime) LoadModule(ctx context.Context, wasmBytes []byte, cap Capability) (api.Module, error);

// Invoke вызывает экспортную функцию модуля.
func (rt *Runtime) Invoke(ctx context.Context, mod api.Module, fn string, args ...uint64) ([]uint64, error);
```

**Критерий приёмки:**
- `go test -race ./...` проходит
- Тест: загрузить trivial.wasm, вызвать `add(2,3)` → `5`
- Тест: модуль пытается открыть `/etc/passwd` → ошибка `permission denied`

---

### Спринт 2.2: Rust WASM guests (Неделя 11–12)

#### Задача 2.2.1: Calculator guest

```bash
cd grounded-wasm/guests/calculator
cargo init --lib
```

**Файл: `Cargo.toml`**
```toml
[lib]
crate-type = ["cdylib"]

[dependencies]
wasm-bindgen = "0.2"
```

**Файл: `src/lib.rs`**
```rust
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub extern "C" fn calculate(op: u32, a: f64, b: f64) -> f64 {
    match op {
        0 => a + b,
        1 => a - b,
        2 => a * b,
        3 => if b != 0.0 { a / b } else { f64::NAN },
        _ => f64::NAN,
    }
}
```

**Сборка:**
```bash
cargo build --target wasm32-unknown-unknown --release
wasm-opt -Oz target/wasm32-unknown-unknown/release/calculator.wasm -o calculator.wasm
```

**Критерий приёмки:**
- `calculator.wasm` < 50KB
- Go host загружает и вызывает `calculate(0, 2.0, 3.0)` → `5.0`

---

#### Задача 2.2.2: SQL Sandbox guest

Аналогично calculator, но:
- Вход: SQL query (string)
- Разрешено: только `SELECT`
- Запрещено: `INSERT`, `UPDATE`, `DELETE`, `DROP`, `;` (multi-statement)
- Выход: JSON result или error

**Критерий приёмки:**
- `SELECT * FROM users` → работает
- `DROP TABLE users` → ошибка `forbidden operation`

---

### Спринт 2.3: HTTP Gateway (Неделя 13–14)

#### Задача 2.3.1: Совместимость с mcp-gateway

**Файл: `host/cmd/wasm-gateway/main.go`**

HTTP API, идентичный mcp-gateway:

```
GET  /v1/tools/schema       → OpenAI function JSON
POST /v1/tools/{name}/call  → WASM invoke
GET  /healthz               → health check
GET  /metrics               → Prometheus
```

**Критерий приёмки:**
- `curl localhost:8080/v1/tools/schema` возвращает JSON с `calculator`, `sql-sandbox`
- `curl -X POST localhost:8080/v1/tools/calculator/call -d '{"op":0,"a":2,"b":3}'` → `{"result":5}`
- Drop-in замена: тот же YAML конфиг, что mcp-gateway

---

### Спринт 2.4: Security audit и бенчмарки (Неделя 15–16)

#### Задача 2.4.1: Тесты изоляции

**Файл: `host/tests/isolation_test.go`**

```go
func TestModuleCannotAccessHostFS(t *testing.T) {
    // WASM модуль пытается открыть /etc/passwd → denied
}

func TestModuleCannotExceedMemory(t *testing.T) {
    // Модуль аллоцирует > MaxMemoryMB → trap
}

func TestModuleCannotMakeNetworkCalls(t *testing.T) {
    // Модуль пытается HTTP запрос → denied (если не в whitelist)
}

func TestModuleDeterminism(t *testing.T) {
    // Один вход → идентичный выход при 100 запусках
}
```

#### Задача 2.4.2: Бенчмарки

**Файл: `grounded-wasm/BENCHMARKS.md`**

```markdown
## Cold start
- subprocess (python): ~50ms
- WASM (wazero): X ms (цель: < 5ms)

## Invocation overhead
- native: X ns
- WASM: X ns (цель: < 10% overhead)

## Memory isolation
- Max memory: 64MB enforced
- OOM trap: verified
```

**Критерий приёмки:**
- Все isolation-тесты проходят
- Cold start < 5ms
- Overhead < 10%

---

## ФАЗА 3: grounded-eval (Недели 17–24)

### Спринт 3.1: Go coordinator (Неделя 17–18)

#### Задача 3.1.1: Eval Spec Parser

**Файл: `coordinator/internal/spec/spec.go`**

```go
package spec

type EvalSpec struct {
    Dataset    string   `yaml:"dataset"`
    Model      Model    `yaml:"model"`
    Rules      []Rule   `yaml:"rules"`
    Parallelism int     `yaml:"parallelism"`
    Seed       int64    `yaml:"seed"`
}

type Model struct {
    Endpoint string `yaml:"endpoint"`
    Provider string `yaml:"provider"` // vllm | ollama | openai
}

type Rule struct {
    Name     string  `yaml:"name"` // numeric_verify | citation_check | toxicity
    Params   map[string]any `yaml:"params"`
}

func Load(path string) (*EvalSpec, error);
func (s *EvalSpec) Validate() error;
```

**Критерий приёмки:**
- YAML парсится
- Валидация: parallelism > 0, endpoint не пустой

---

#### Задача 3.1.2: Distributed Scheduler

**Файл: `coordinator/internal/scheduler/scheduler.go`**

```go
package scheduler

// Scheduler распределяет eval-кейсы по воркерам.
type Scheduler struct {
    workers []WorkerClient
    results chan CaseResult
}

// Run запускает eval на N воркерах.
// Детерминизм: при одном seed результат идентичен.
func (s *Scheduler) Run(ctx context.Context, spec *spec.EvalSpec, cases []EvalCase) (*Report, error);
```

**Критерий приёмки:**
- 100 кейсов, 10 воркеров → все завершаются
- Seed 42 → идентичный report при 2 запусках

---

### Спринт 3.2: Rust workers (Неделя 19–20)

#### Задача 3.2.1: Worker binary

```bash
cd grounded-eval/worker
cargo init --name grounded-eval-worker
```

**Файл: `src/main.rs`**

```rust
/// Worker получает gRPC stream кейсов, прогоняет inference,
/// вычисляет метрики, возвращает результаты.
/// Детерминизм: seed фиксируется в HTTP headers к модели.
```

**Критерий приёмки:**
- `cargo build --release` проходит
- Worker обрабатывает 1000 кейсов/минуту

---

### Спринт 3.3: Python metrics (Неделя 21–22)

#### Задача 3.3.1: Statistical significance

**Файл: `metrics/significance.py`**

```python
from dataclasses import dataclass
import numpy as np
from scipy import stats

@dataclass
class SignificanceResult:
    p_value: float
    confidence_interval: tuple[float, float]
    significant: bool  # p < 0.05

def paired_bootstrap(
    scores_a: list[float],
    scores_b: list[float],
    n_bootstrap: int = 10000,
    seed: int = 42,
) -> SignificanceResult:
    """Paired bootstrap test для сравнения двух моделей."""
    ...
```

**Критерий приёмки:**
- Тест: одинаковые scores → p > 0.05
- Тест: разные scores → p < 0.05

---

### Спринт 3.4: CLI и отчёты (Неделя 23–24)

#### Задача 3.4.1: CLI

**Файл: `coordinator/cmd/grounded-eval/main.go`**

```bash
grounded-eval run --spec eval.yaml --workers 50 --report html
grounded-eval compare --results run1.json run2.json
```

**Критерий приёмки:**
- `grounded-eval run` на 100 кейсах завершается
- `report.html` открывается в браузере
- `results.jsonl` содержит вердикт для каждого кейса

---

## ФАЗА 4: grounded-bench (Недели 25–32)

### Спринт 4.1: Датасет (Неделя 25–28)

#### Задача 4.1.1: Генерация датасета

**Файл: `dataset/generate.py`**

```python
"""Генерация 1000 question-answer пар для verifiable AI benchmark.

Домены:
- finance: 250 кейсов (выручка, НДС, курсы)
- medical: 250 кейсов (дозировки, противопоказания)
- legal: 250 кейсов (статьи, сроки, штрафы)
- technical: 250 кейсов (допуски, спецификации)

Формат:
{
  "id": "fin-001",
  "domain": "finance",
  "question": "Какая выручка компании за Q3 2025?",
  "ground_truth": "14,000,000 рублей",
  "source_document": "annual_report_2025.pdf",
  "source_location": "стр. 42, табл. 3",
  "numeric_values": [14000000.0],
  "citation_required": true
}
"""
```

**Критерий приёмки:**
- 1000 кейсов в `dataset/grounded-bench-1000.jsonl`
- Валидация: все JSON валидны, все numeric_values извлекаются

---

### Спринт 4.2: Runner (Неделя 29–30)

#### Задача 4.2.1: Go runner

**Файл: `runner/cmd/grounded-bench/main.go`**

```bash
grounded-bench run --model vllm:llama3 --dataset finance-250
grounded-bench compare --results run1.json run2.json
grounded-bench publish --leaderboard
```

**Метрики:**
```go
type Metrics struct {
    NumericVerifyRate float64 // % чисел, подтверждённых контекстом
    CitationPrecision float64 // % цитат, ведущих на реальный источник
    HallucinationRate float64 // % утверждений без подтверждения
    RefusalRate       float64 // % отказов при неподтверждённых данных
}
```

**Критерий приёмки:**
- Прогон 250 кейсов завершается
- `results.json` содержит все 4 метрики

---

### Спринт 4.3: Leaderboard (Неделя 31–32)

#### Задача 4.3.1: Static HTML

**Файл: `leaderboard/index.html`**

- Таблица: Model | NVR | CP | HR | RR | Date
- Сортировка по NVR
- Reproducibility: seed, hardware, version
- Submit form: JSON upload

**Критерий приёмки:**
- Открывается локально
- Данные из `results.json` отображаются

---

## ФАЗА 5: grounded-kernels (Недели 33–40)

### Спринт 5.1: C++ / CUDA основы (Неделя 33–34)

#### Задача 5.1.1: CMake проект

**Файл: `grounded-kernels/CMakeLists.txt`**
```cmake
cmake_minimum_required(VERSION 3.28)
project(grounded_kernels LANGUAGES CXX CUDA)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CUDA_STANDARD 17)

find_package(CUDAToolkit REQUIRED)
find_package(Python REQUIRED COMPONENTS Interpreter Development)

add_library(grounded_kernels SHARED
    csrc/numeric_verify.cu
    csrc/pii_detect.cu
    csrc/bindings.cpp
)

target_link_libraries(grounded_kernels PRIVATE
    CUDA::cudart
    Python::Python
)
```

**Критерий приёмки:** `cmake -B build && cmake --build build` проходит.

---

### Спринт 5.2: CUDA kernels (Неделя 35–36)

#### Задача 5.2.1: numeric_verify.cu

**Файл: `csrc/numeric_verify.cu`**

```cuda
/// CUDA kernel: parallel numeric verification.
/// Input: token_ids (GPU, batch x seq_len), context_numbers (GPU)
/// Output: verdict_mask (GPU, batch)
/// Каждый thread обрабатывает один токен.
/// Warp-level reduction для агрегации.

__global__ void numeric_verify_kernel(
    const int* __restrict__ token_ids,
    const float* __restrict__ context_numbers,
    int* __restrict__ verdict_mask,
    int batch_size,
    int seq_len,
    int num_context_numbers,
    float tolerance
);
```

**Критерий приёмки:**
- Kernel компилируется
- Тест: batch 1024, seq_len 128 → verdict_mask корректен
- nsight profile: occupancy > 50%

---

#### Задача 5.2.2: pii_detect.cu

Аналогично, но pattern matching (email, phone) на GPU.

**Критерий приёмки:**
- Batch 1024 → pii_mask корректен
- Latency: < 0.1ms на batch

---

### Спринт 5.3: Python bindings (Неделя 37–38)

#### Задача 5.3.1: PyBind11

**Файл: `csrc/bindings.cpp`**

```cpp
#include <pybind11/pybind11.h>
#include <torch/extension.h>

torch::Tensor numeric_verify_cuda(
    torch::Tensor token_ids,
    torch::Tensor context_numbers,
    float tolerance
);

PYBIND11_MODULE(TORCH_EXTENSION_NAME, m) {
    m.def("numeric_verify", &numeric_verify_cuda, "Numeric verify (CUDA)");
}
```

**Файл: `python/grounded_kernels/__init__.py`**

```python
from ._C import numeric_verify

class NumericVerifyKernel:
    def __init__(self, tolerance: float = 0.01):
        self.tolerance = tolerance
    
    def verify(self, token_ids: torch.Tensor, context: torch.Tensor) -> torch.Tensor:
        return numeric_verify(token_ids, context, self.tolerance)
```

**Критерий приёмки:**
- `pip install -e .` проходит
- `python -c "import grounded_kernels; print('ok')"` работает
- Тест: GPU tensor → verdict tensor

---

### Спринт 5.4: vLLM integration (Неделя 39–40)

#### Задача 5.4.1: vLLM plugin

**Файл: `python/grounded_kernels/vllm_plugin.py`**

```python
"""vLLM custom callback: grounded verification на каждый batch."""

from vllm import LLMEngine
from grounded_kernels import NumericVerifyKernel

class GroundedVerificationCallback:
    def __init__(self, tolerance: float = 0.01):
        self.kernel = NumericVerifyKernel(tolerance)
    
    def on_batch_generated(self, token_ids, context):
        verdict = self.kernel.verify(token_ids, context)
        if not verdict.all():
            # Остановить генерацию для неподтверждённых
            ...
```

**Критерий приёмки:**
- vLLM запускается с plugin
- Генерация с verify: overhead < 5%
- Без verify: baseline latency

---

## ФАЗА 6: Продвижение (Недели 41–48)

### Спринт 6.1: Документация и README

#### Задача 6.1.1: README для каждого репо

Шаблон:

```markdown
# grounded-guardrails

Token-level verification for LLM inference. Blocks hallucinated numbers at token 5, not token 500.

## Benchmarks

| Metric | Value |
|--------|-------|
| Overhead | < 1ms p99 per token |
| Throughput | > 10K tokens/sec |
| Numeric Verify F1 | 0.95 |

## Quick Start

```bash
docker compose up
grpcurl -plaintext localhost:50051 list
```

## Architecture

[диаграмма]

## Integration

- vLLM: callback
- Ollama: streaming wrapper
- grounded-llm: gRPC client
```

**Критерий приёмки:** Все 8 репо имеют README на английском с бенчмарками.

---

### Спринт 6.2: Публикации

#### Задача 6.2.1: Hacker News

Для каждого продукта:

```
Show HN: Token-level guardrails for vLLM with <1ms overhead (Rust + Go)
Show HN: MCP tools in WASM sandboxes – provable isolation
Show HN: Distributed eval engine for verifiable AI – 10K cases in 3 min
Show HN: Grounded-Bench – first benchmark for verifiable inference
Show HN: CUDA kernels for numeric verification in vLLM – 16x speedup
```

**Критерий приёмки:** 2+ поста на HN front page.

---

### Спринт 6.3: Контрибуции

#### Задача 6.3.1: vLLM PR

- Документация: добавить раздел "Custom Verification Callbacks"
- Баг-фикс или мелкая фича
- Цель: 2+ merged PR

**Критерий приёмки:** PR merged. Профиль GitHub показывает контрибуции.

---

## КРИТЕРИИ УСПЕХА ЧЕРЕЗ 12 МЕСЯЦЕВ

```yaml
github:
  stars_total: 1000+
  forks: 100+
  contributors: 5+

publications:
  hacker_news_frontpage: 2+
  medium_articles: 4+
  conference_talks: 2+

contributions:
  vllm_prs: 2+
  ollama_prs: 1+

network:
  linkedin_connections_target_companies: 15+
  inbound_recruiter_messages: 5+/month

offers:
  tier1_interviews: 3+
  tier1_offers: 1+
```

---

## ПРАВИЛА ДЛЯ АГЕНТОВ CURSOR

1. **Один шаг за раз.** Не переходи к следующему шагу, пока текущий не прошёл все критерии приёмки.
2. **Коммит после каждого шага.** Формат: `feat(scope): description`.
3. **Тесты обязательны.** Код без тестов не принимается.
4. **Бенчмарки обязательны.** Performance-критичный код без цифр не принимается.
5. **README обновляется сразу.** Не «потом напишу», а в том же коммите.
6. **Ошибки обрабатываются.** Никаких `unwrap()`, `panic()`, bare `except:`.
7. **Английский язык.** Код, комментарии, коммиты, README — только английский.
8. **Если шаг не проходит:** остановись, выведи ошибку, предложи фикс. Не пропускай.

---

## НАЧАЛО РАБОТЫ

```bash
# 1. Создать монорепозиторий
mkdir verifiable-ai-platform && cd verifiable-ai-platform
git init

# 2. Скопировать существующие проекты
cp -r ~/grounded-llm ./grounded-llm
cp -r ~/grounded-agent ./grounded-agent
cp -r ~/mcp-gateway ./mcp-gateway

# 3. Создать структуру новых проектов
mkdir -p grounded-guardrails/{rust/src,python/grounded_guardrails,go/cmd/server,proto}
mkdir -p grounded-wasm/{host/cmd/wasm-gateway,guests/calculator/src}
mkdir -p grounded-eval/{coordinator/cmd/grounded-eval,worker/src,metrics}
mkdir -p grounded-bench/{dataset,runner/cmd/grounded-bench,leaderboard}
mkdir -p grounded-kernels/{csrc,python/grounded_kernels}

# 4. Первый коммит
git add .
git commit -m "feat: initialize verifiable-ai-platform monorepo"

# 5. Начать с Фазы 1, Спринт 1.1, Задача 1.1.1
```

Инструкция завершена
