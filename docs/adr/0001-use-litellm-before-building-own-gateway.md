# ADR-0001: Use LiteLLM before building our own gateway

## Status

Accepted

## Context

The project goal is to understand real LLM gateway behavior before implementing a custom infra layer.
Building provider adapters, fallback, budget, rate limits, and usage tracking from scratch too early would create ungrounded abstractions.

## Decision

M1-M2 will use LiteLLM Proxy as the downstream LLM gateway.
The Go nanobot backend acts as an upstream client.

Custom components will first be built around:

- ContextManager
- PolicyRouter
- UsageLogger

## Consequences

### Pros

- Learn real gateway behavior first
- Avoid premature abstraction
- Keep provider integration simple
- Create a concrete baseline before custom implementation

### Cons

- Some routing behavior remains delegated to LiteLLM
- Custom gateway implementation is delayed
