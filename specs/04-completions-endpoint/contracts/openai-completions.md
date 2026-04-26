# Contract: OpenAI-Compatible Legacy Completions

## Endpoint

`POST /v1/completions`

## Supported Request Shape

```json
{
  "model": "gpt-3.5-turbo-instruct",
  "prompt": "Say hello",
  "max_tokens": 100,
  "temperature": 0.7,
  "top_p": 1,
  "stream": false,
  "stop": null,
  "presence_penalty": 0,
  "frequency_penalty": 0,
  "user": "optional-user-id"
}
```

## Successful Response Shape

```json
{
  "id": "cmpl-...",
  "object": "text_completion",
  "created": 1234567890,
  "model": "gpt-3.5-turbo-instruct",
  "choices": [
    {
      "text": "Hello!",
      "index": 0,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 3,
    "completion_tokens": 5,
    "total_tokens": 8
  }
}
```

## Streaming Response Shape

Streaming responses use server-sent events with `object: "text_completion"` and incremental `choices[].text` values.

## Unsupported Behavior

- Prompt arrays are rejected with HTTP 400.
- `echo`, `logprobs`, `best_of`, and `suffix` may be accepted for decode compatibility but are not implemented.
- Provider-native completions APIs are not required for this feature.

## Compatibility Notes

- Authentication, rate limiting, guardrails, audit logging, and cost tracking follow the chat completions path.
- Errors must use the gateway's OpenAI-compatible error envelope.
