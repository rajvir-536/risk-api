# Risk API Service

A Go-based mock backend service for a risk/fraud case management system. This service is designed to be registered with and called by an MCP server over HTTP as a separate service, rather than being queried directly by an LLM client.

## Core Features
- **Embedded Storage**: Uses Go's `//go:embed` package to compile seed JSON files directly into the executable binary.
- **In-Memory Store**: No database of any kind is required at runtime. On startup, datasets are loaded into thread-safe in-memory maps.
- **Dependency-Free**: Built purely using the Go Standard Library (requires Go 1.22+ for ServeMux routing).
- **Process Lifetime State**: Creating a risk flag via `POST /flags` modifies state only in memory; flags are lost on process exit. This is intentional for a lightweight, self-contained demo.

---

## Technical Architecture Note (Swapping Stores)

The codebase implements a strict separation of concerns:
- **`internal/store/store.go`** defines a database-agnostic interface (`Store`) and sentinel errors (like `ErrNotFound`).
- **`internal/handlers/`** interacts exclusively with the `Store` interface and knows nothing about databases, filesystems, or JSON parsing.

If you want to swap this mock store with a persistent database like SQLite, Postgres, or MongoDB later, you only need to write a new implementation of the `Store` interface under `internal/store/` and switch the wiring inside `cmd/api/main.go`. The HTTP handlers will remain completely untouched.

---

## Getting Started

### Running the App
Ensure you have Go 1.22+ installed and run:
```bash
go run ./cmd/api
```
The server will boot up instantly and start listening on port `8080`.

You can configure the listener port via the `PORT` environment variable:
```bash
PORT=9000 go run ./cmd/api
```

---

## API Documentation & Example Curl Commands

### 1. Health Check
Checks if the API is active.
- **URL**: `GET /healthz`
- **Response**: `{"status": "ok"}`
- **Curl**:
  ```bash
  curl -i http://localhost:8080/healthz
  ```

### 2. Retrieve Case
Retrieves a fraud/risk case by ID.
- **URL**: `GET /cases/{id}`
- **Response**: `{"id": "CASE-0001", "account_id": "ACC-0001", "status": "open", "opened_at": "...", "summary": "..."}`
- **Curl**:
  ```bash
  curl -i http://localhost:8080/cases/CASE-0001
  ```

### 3. Retrieve KYC Record
Retrieves the KYC profile details for an account ID.
- **URL**: `GET /accounts/{id}/kyc`
- **Response**: `{"account_id": "ACC-0001", "verification_level": "Level 1", "prior_flags": 3, "jurisdiction": "US", "watchlist_matches": ["PEP Match - Tier 2", "OFAC SDN Watchlist"]}`
- **Curl**:
  ```bash
  curl -i http://localhost:8080/accounts/ACC-0001/kyc
  ```

### 4. Retrieve Account Transactions
Retrieves all transactions for an account ID, sorted newest to oldest. Includes an optional `since_days` filter.
- **URL**: `GET /accounts/{id}/transactions`
- **Query Params**: `since_days=N` (optional, filter to transactions in the last N days)
- **Response**: List of transaction objects.
- **Curl**:
  - Get all:
    ```bash
    curl -i http://localhost:8080/accounts/ACC-0001/transactions
    ```
  - Filter to last 30 days:
    ```bash
    curl -i "http://localhost:8080/accounts/ACC-0001/transactions?since_days=30"
    ```

### 5. Raise Risk Flag
Creates a new risk flag for an account in memory.
- **URL**: `POST /flags`
- **Payload**:
  ```json
  {
    "account_id": "ACC-0001",
    "reason": "Suspicious login location",
    "raised_by": "mcp-agent"
  }
  ```
- **Response**: The created Flag object (returns HTTP 201).
- **Curl**:
  ```bash
  curl -i -X POST -H "Content-Type: application/json" \
    -d '{"account_id": "ACC-0001", "reason": "Suspicious login location", "raised_by": "mcp-agent"}' \
    http://localhost:8080/flags
  ```
