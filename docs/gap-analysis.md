# Gap Analysis

## Known Gaps

| Area | Current state | Need |
|---|---|---|
| Go service | weak | HTTP server, context timeout, JSON, streaming |
| LiteLLM | unknown | proxy config, model alias, fallback, usage |
| Routing | conceptual | policy config, score-based decision |
| Context governance | conceptual | block model, keep/compress/drop |
| Usage logging | basic | task-level token/cost/latency |
| Docker | partial | compose for Go + LiteLLM |
| K8s | weak | later: deployment/service/config/secret |
| Observability | basic | logs first, metrics second, OTel later |

## Boundary Notes

- LiteLLM owns provider adapters, provider auth, fallback, and gateway behavior through M2.
- The Go service owns task API shape, context preparation, routing intent, request correlation, and task-level usage records as those modules are introduced.
- Kubernetes, OpenTelemetry, and multi-service splitting stay out of scope until the local loop is useful.
