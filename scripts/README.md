# Scripts

## `smoke-real-provider.sh`

Runs a local smoke flow against a running Go service:

```bash
BASE_URL=http://127.0.0.1:8080 scripts/smoke-real-provider.sh
```

It verifies:

- `/health`
- `POST /tasks/review-diff`
- `GET /tasks/{id}`
- `GET /usage/tasks/{id}`
- `GET /usage/recent`

Keep scripts small, reviewable, and easy to disable.

Do not add automation that rewrites formal protocol files without creating a proposal first.
