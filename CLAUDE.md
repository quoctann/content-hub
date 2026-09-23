# content-hub

Go/Gin backend for Content Hub. Frontend: sibling repo `content-hub-fe`.
Deploy manifests: sibling repo `content-gitops` (k3s; request path is
ingress-nginx -> frontend nginx `/api/*` -> this service).

## API contract with the frontend

The response/request DTOs in `internal/delivery/http/*_request.go` and
`*_response.go` are the contract. `content-hub-fe/src/types/*.ts` mirrors them.

- **Nullability is part of the contract**: a pointer field (`*string`, ...) is
  documented as nullable and the FE types it `T | null`. Changing a field to or
  from a pointer is a breaking change for the FE.
- **Partial updates use `NullableString`** (`internal/delivery/http/nullable.go`)
  so "omitted" (leave unchanged) and `null` (clear) are distinct. Plain `*string`
  on an update request cannot tell them apart; do not use it for nullable
  columns. The repository applies a `domain.ContentPatch` and only writes the
  columns that were set.
- **Changing a DTO is a two-repo change**: update the FE types/mapper in the same
  piece of work and regenerate swagger (`make swagger`). Mention the FE impact
  explicitly when handing off.
- Errors: repositories return `domain.ErrNotFound` (wrapped) for missing rows;
  handlers map it to 404 and everything else to a logged 500.

## Database

- Tables live in the schema from `DATABASE_SCHEMA` (prod: `meme`), passed as
  `search_path` by `database.URL`, which both the pgx pool and the migrator use.
- Migration rules: see `migrations/README.md` (unqualified table names, qualified
  `public.` functions, every down migration must run).
- Repository integration tests run when `TEST_DATABASE_URL` is set (CI does this).

## Checks

`go vet ./...` and `go test -race ./...` must pass.
