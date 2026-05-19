# Transaction Intake API Implementation Notes

Branch: `feat/transaction-intake-api`

This slice turns the core routing domain into a bank-callable API surface.

## What It Implements

- OpenAPI contract at `api/openapi.yaml`
- `POST /v1/transactions`
- `GET /v1/transactions/{transaction_id}`
- request validation for required bank/channel fields
- required idempotency key
- idempotent replay with no duplicate lifecycle event emission
- transaction record persistence through an interface
- in-memory transaction store for pilot and tests
- lifecycle event sink interface
- in-memory event sink for pilot and tests
- route decision persistence through the core route decision store
- `net/http` handlers using Go's standard library router
- runnable demo command at `cmd/imsi-api`

## What It Intentionally Does Not Implement Yet

- authentication
- RBAC/maker-checker approval
- PostgreSQL transaction persistence
- NATS/JetStream event publication
- provider submission after route selection
- real bank/provider callback handling
- production deployment wiring

Those belong in later slices so intake remains small and testable.

## Verification

Run:

```powershell
go test ./...
go vet ./...
go run ./cmd/imsi-api
```
