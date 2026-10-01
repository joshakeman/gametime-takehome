# Order Payment Flow

## Description

A small Go service that takes an order through payment authorization and completion, voiding the payment if completion fails and flagging the order as `needs_attention` if the void fails too.

## How to Run It

```sh
go run .                     # serves on :8080
go test -race ./...
```

```sh
# Create an order (201)
curl -s -X POST localhost:8080/orders -d '{"user_id":"u1","amount":100}'

# Run payment + completion (200, returns the final order and any error)
curl -s -X POST localhost:8080/orders/1/process

# Get the current state and history
curl -s localhost:8080/orders/1
```

| Route                       | Responses                                                        |
| --------------------------- | ---------------------------------------------------------------- |
| `POST /orders`              | 201 created; 400 bad body                                        |
| `POST /orders/{id}/process` | 200 with `{order, error}`; 404 unknown id; 409 already processed |
| `GET /orders/{id}`          | 200; 404 unknown id                                              |

A decline or a failed completion still returns **200** from `/process`. The
request succeeded and the outcome was recorded; the response carries the final
state and the error text.

For ease of testing, simple interfaces are established that allow you to create fakes that set expected behaviors by either the `PaymentServicer` or `Completer` (`ShouldDecline`, `ShouldFailComplete`,`ShouldFailVoid`). Failure states are tested in `service_test.go`

## Layout

- `data/` uses a map for recording state. Obviously a mature application would have a full database, but a simple in-memory store worked here for testing the core logic. The map is protected with a mutex to control data races.
- `service/` holds the payment flow (`Create`, `Process`, `Get`), the `PaymentServicer` and `Completer` interfaces, and the fakes. The `Completer` interface is a minimal stand-in for whatever "other" actions might be necessary in a real application to complete an order. It's here just to test how our payment flow is affected by a failure in that order completion phase.
- `api/` holds the HTTP handlers.
- `main.go` wires the pieces together.

## Tradeoffs & With More Time

- Persist orders and history in a real database. In this case, you would consider the best practices for controlling data races and enforcing atomicity where necessary.
- The "payment service" in this case is well controlled, as this is a simplified model intended to test a state machine. In reality, you would be worried about network failures, retries and similar issues which are best prevented with idempotency controls. This was excluded from this exercise but would be necessary for a real application. You would pass an idempotency key to the payment servicer (ie Stripe) to ensure retries don't result in multiple charges.
- Retries to services like Stripe would require a backoff strategy and more general approach to handling distributed failures.
- Ambiguous states like `needs_attention` would need to be handled, probably by an intermittent clean up job of some kind.
- Tests were only written for the core state machine logic. You would want to also flesh out unit tests for all other aspects of the application, including handlers and data layer.
- The HTTP codes would bear further consideration, many of the states are returned as 200 plus a status, but this arguably should be adapted to a more fine grained error system.

## AI Usage

The core functionality of the application was developed first without AI assistance, including package structure and basic state machine flow. I then used Claude to review what I'd written and suggest feedback or improvements, specifying that I wanted particular focus on potential race conditions.
Claude suggested the `ClaimPending` pattern to ensure multiple concurrent requests wouldn't simultaneously modify the same pending order (thus potentially causing a double charge). Claude also helped with the in memory `History` representation and wrote the api handlers.
Claude was also tasked with writing the tests for the core logic.
I reviewed every change, ran `go vet` and `go test -race` after each round, and confirmed the tests fail when the code is deliberately broken.
