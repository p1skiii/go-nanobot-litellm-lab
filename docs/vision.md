# Vision

## Goal

Build a Go nanobot backend that calls LiteLLM as a downstream LLM gateway.

The project is used to learn:

- how upper-layer agent services consume an LLM gateway
- how LiteLLM handles provider abstraction, routing, fallback, usage, cost, and rate limits
- how to gradually replace selected parts with custom ContextManager, PolicyRouter, and UsageLogger

## Non-goals

- Do not build a full custom LLM Gateway in M1
- Do not clone Claude Code
- Do not start with Kubernetes
- Do not train ML-based routing models
- Do not split into microservices before the local loop works
