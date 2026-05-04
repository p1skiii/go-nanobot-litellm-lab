# Test Plan: LiteLLM Integration

## Unit Tests

| Test | Expected |
|---|---|
| build request from review task | valid LiteLLM request |
| empty diff rejected | 400 |
| timeout mapped to 504 | pass |
| downstream 5xx mapped to 502 | pass |
| malformed downstream response mapped to 502 | pass |
| request-level model alias used | pass |
| final context used as prompt | pass |

## Integration Tests

| Test | Setup | Expected |
|---|---|---|
| call LiteLLM non-stream | LiteLLM running | result returned |
| fallback behavior | primary mock fails | fallback model used |
| usage returned | provider supports usage | usage captured |
| rate limit | low limit config | 429 observed |
| real provider non-stream chain | Xiaomi MiMo API key | result returned |

## Notes

M1 only needs non-stream chat completion. Streaming remains out of API scope until a separate contract exists.
M2 studied fallback, usage, streaming, and budget behavior with real provider-backed LiteLLM.
M5 owns usage persistence.
