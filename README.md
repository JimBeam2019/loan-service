# Loan Service

A small Go REST API implementing the Loan Service assignment with the **State pattern**, PostgreSQL persistence, and an in-memory repository for tests.

## Loan lifecycle

```text
proposed -> approved -> invested -> disbursed
```

- A new loan starts as `proposed`.
- A proposed loan can be approved only with validator ID, visit-proof URL, and approval date.
- An approved loan can receive multiple investments.
- Total investment may not exceed the principal.
- The loan becomes `invested` only when total investment equals the principal.
- An invested loan can be disbursed only with signed agreement URL, field officer ID, and disbursement date.
- No backward state transitions are allowed.

## Duplicate loans

An active duplicate is defined as the same:

- borrower ID
- principal amount
- rate
- return on investment

A loan is considered active until it reaches `disbursed`.

The PostgreSQL database enforces this rule with a partial unique index. The in-memory repository applies the same rule atomically.

## Architecture

```text
HTTP -> Use Case -> Domain (State Pattern)
                  |
                  +-> Repository -> PostgreSQL
                                  -> In-memory (tests)
```

- `internal/domain` — loan aggregate, states, validation, and business rules.
- `internal/usecase` — application orchestration.
- `internal/repository` — persistence interface.
- `internal/infrastructure/postgres` — PostgreSQL implementation and transactions.
- `internal/infrastructure/memory` — in-memory implementation used by tests.
- `internal/interface/http` — Gin handlers and request parsing.
- `pkg/postgres` — PostgreSQL connection/query helpers.

Approval, investment, and disbursement persistence use database transactions. Investment operations lock the loan row with `FOR UPDATE` before calculating the current total, preventing concurrent investments from exceeding the principal.

## API

### Create loan

```http
POST /loans
Content-Type: application/json

{
  "borrowerId": "<uuid>",
  "principalAmount": "5000000",
  "rate": "0.05",
  "returnOnInvestment": "0.03"
}
```

CLI:

```bash
curl -X POST http://localhost:8080/loans -H "Content-Type: application/json" -d '{"borrowerId": "<uuid>", "principalAmount": "5000000", "rate": "0.05", "returnOnInvestment": "0.03"}'
```

### Get loan

```http
GET /loans/<loan-id>
```

CLI:

```bash
curl http://localhost:8080/loans/<loan-id>
```

### Approve loan

```http
POST /loans/<loan-id>/approval
Content-Type: application/json

{
  "fieldValidatorEmployeeId": "<uuid>",
  "visitProofUrl": "https://example.com/proof.pdf",
  "approvalDate": "2026-08-30T10:00:00Z"
}
```

CLI:

```bash
curl -X POST http://localhost:8080/loans/<loan-id>/approval -H "Content-Type: application/json" -d '{"fieldValidatorEmployeeId": "<uuid>", "visitProofUrl": "https://example.com/proof.pdf", "approvalDate": "2026-08-30T10:00:00Z"}'
```

### Add investment

```http
POST /loans/<loan-id>/investments
Content-Type: application/json

{
  "investorId": "<uuid>",
  "amount": "2000000"
}
```

CLI:

```bash
curl -X POST http://localhost:8080/loans/<loan-id>/investments -H "Content-Type: application/json" -d '{"investorId": "<uuid>", "amount": "2000000"}'
```

### Disburse loan

```http
POST /loans/<loan-id>/disbursement
Content-Type: application/json

{
  "signedAgreementUrl": "https://example.com/agreement.pdf",
  "fieldOfficerEmployeeId": "<uuid>",
  "disbursementDate": "2026-08-30T10:00:00Z"
}
```

CLI:

```bash
curl -X POST http://localhost:8080/loans/<loan-id>/disbursement -H "Content-Type: application/json" -d '{"signedAgreementUrl": "https://example.com/agreement.pdf","fieldOfficerEmployeeId": "<uuid>", "disbursementDate": "2026-08-30T10:00:00Z"}'
```

Amounts are accepted as strings to avoid JSON floating-point representation issues at the API boundary.

## Run locally

The application defaults to the in-memory repository. To use PostgreSQL, set `USE_POSTGRES=true` and the PostgreSQL connection variables.

```bash
go run ./cmd/server
```

The API listens on `http://localhost:8080`.

## Run with Docker

```bash
docker compose up --build
```

PostgreSQL is initialized from `db/init.sql` and is available to the application as the Docker service `postgres`.

If the database volume was created with an incompatible PostgreSQL image/version, recreate it:

```bash
docker compose down -v
docker compose up --build
```

## Tests

```bash
go test ./...
```

The test suite covers:

- initial `proposed` state
- approval validation and duplicate approval
- all forward-only state transitions
- invalid investment amounts
- multiple investors
- exact funding and over-funding
- disbursement validation
- HTTP end-to-end lifecycle
- missing-loan HTTP response
- duplicate active loans
- in-memory repository duplicate protection
- numeric helper functions

## Scope note

The assignment describes sending each investor an agreement-letter PDF link after a loan is fully invested. This project models and persists an agreement-letter URL on investments, but it does not implement a PDF-generation or notification service. A real document/notification service can populate that URL when the loan becomes fully invested.
