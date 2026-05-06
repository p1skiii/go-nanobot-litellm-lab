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

## `smoke-failure-cases.sh`

Runs expected failure and recovery checks against a running local stack:

```bash
BASE_URL=http://127.0.0.1:8080 \
LITELLM_BASE_URL=http://127.0.0.1:4000 \
LITELLM_API_KEY=sk-local-dev \
scripts/smoke-failure-cases.sh
```

It verifies:

- empty diff maps to `400`
- streaming request maps to `400`
- tiny LiteLLM timeout maps to `504` and writes failed usage
- missing LiteLLM model maps to `502` and writes failed usage
- normal request still succeeds after failure cases

Keep scripts small, reviewable, and easy to disable.

Do not add automation that rewrites formal protocol files without creating a proposal first.
