# Car Pooling Service

## Goal
Implement the challenge service while preserving the original HTTP contract and improving code quality, robustness, and testability.

## Scope of this solution
- A **single implementation** is kept (no extra modes and no out-of-scope features).
- In-memory state (self-contained).
- Challenge API contract is respected.
- Fairness rule implemented: **FIFO when possible**.

## API Contract

### `GET /status`
- `200 OK`

### `PUT /cars`
- Content-Type: `application/json`
- Body:
```json
[
  { "id": 1, "seats": 4 },
  { "id": 2, "seats": 6 }
]
```
- Responses: `200`, `400`
- Effect: resets cars, journeys, and waiting queue.

### `POST /journey`
- Content-Type: `application/json`
- Body:
```json
{ "id": 1, "people": 4 }
```
- Responses: `200` (assigned), `202` (waiting), `400`

### `POST /dropoff`
- Content-Type: `application/x-www-form-urlencoded`
- Body: `ID=X`
- Responses: `204`, `404`, `400`

### `POST /locate`
- Content-Type: `application/x-www-form-urlencoded`
- Body: `ID=X`
- Responses:
  - `200` + payload `{ "id": <car_id>, "seats": <car_seats> }` when riding
  - `204` when waiting
  - `404`, `400`

## Technical decisions

### 1) Simple domain model
- In-memory `CarPool` with:
  - `cars` by `id`
  - `journeys` by `id`
  - `pending` as a FIFO waiting queue

### 2) Deterministic assignment
- Car selection uses **best-fit**:
  - choose the car with minimum `availableSeats` that can still fit the group
  - tie-break by lower `car.id`
- Fairness:
  - process `pending` in arrival order
  - a later group is served first only when earlier groups cannot fit in any car at that moment

### 3) Cascade reassignment
- After `dropoff`, the service keeps assigning waiting groups until no valid assignment remains.

### 4) Validation and HTTP codes
- Invalid payload or invalid content type -> `400`
- Invalid method -> `405`
- Unknown journey ID in journey operations -> `404`

### 5) Maintainability
- Business rules are concentrated in `service/carpool.go`.
- HTTP layer (`api/controller.go`) translates requests/responses and domain errors.
- Comments are limited to non-obvious rules.

## What was fixed to pass acceptance tests
- **Payload contract alignment**:
  - Cars use `seats` (not `maxSeats`).
  - Journeys use `people` (not `passengers`).
  - `locate` returns only `{id,seats}` for assigned groups.
- **HTTP behavior alignment**:
  - Invalid body/content-type returns `400`.
  - Invalid method returns `405`.
  - `journey` returns `200` when assigned and `202` when waiting.
  - `dropoff` returns `204` on success and `404` when ID is unknown.
  - `locate` returns `200` when riding, `204` when waiting, and `404` when unknown.
- **State handling fixes**:
  - `PUT /cars` fully resets service state.
  - `dropoff` correctly frees seats and triggers reassignment.
- **Assignment logic fixes**:
  - FIFO fairness is respected when possible.
  - Best-fit car selection is deterministic (tie-break by car ID).

Implementation approach:
- First, fix API contract and status codes in `api/controller.go`.
- Then, rewrite allocation/state transitions in `service/carpool.go`.
- Finally, validate with unit tests plus local harness run (`17/17` acceptance assertions passed).

## Relevant structure
- `cmd/carpool/main.go`: server bootstrap
- `api/controller.go`: HTTP handlers
- `service/carpool.go`: business rules
- `service/model/*.go`: models
- `service/carpool_test.go`: business tests
- `api/controller_test.go`: basic API tests

## Run

### Unit tests
```bash
go test ./...
```

### Local acceptance (harness)
```bash
make test.acceptance
```

## Validation against `INSTRUCTIONS.md`
- No out-of-scope features were added.
- Code quality and testing strategy were improved.
- Acceptance flow is preserved (`.gitlab-ci.yml` / harness).
- Solution is self-contained (no external database).

## Possible future improvements (without changing functionality)
- Add more HTTP edge-case tests.
- Refine internal package structure while preserving behavior.
- Add simple benchmarks for large waiting queues.
