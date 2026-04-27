# Test Plan: LiteLLM Integration

## Unit Tests

| Test | Expected |
|---|---|
| build request from review task | valid LiteLLM request |
| empty diff rejected | 400 |
| timeout mapped to 504 | pass |
| downstream 5xx mapped to 502 | pass |
| malformed downstream response mapped to 502 | pass |

## Integration Tests

| Test | Setup | Expected |
|---|---|---|
| call LiteLLM non-stream | LiteLLM running | result returned |
| fallback behavior | primary mock fails | fallback model used |
| usage returned | provider supports usage | usage captured |
| rate limit | low limit config | 429 observed |

## Notes

M1 only needs non-stream chat completion. Streaming, fallback, usage, and budget behavior are studied in M2 before custom replacements are designed.
