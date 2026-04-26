# Contract: OpenAI-Compatible Chat Tool Use

## Request Fields

- `tools`: Optional array of tool definitions. Initial supported tool type is `function`.
- `tool_choice`: Optional string or object controlling tool call behavior.
- `functions`: Optional legacy compatibility field accepted for OpenAI pass-through.
- `function_call`: Optional legacy compatibility field accepted for OpenAI pass-through.
- `messages[].tool_calls`: Optional assistant tool call list.
- `messages[].tool_call_id`: Required when `role` is `tool`.

## Response Fields

- `choices[].message.tool_calls`: Tool calls returned by the model in non-streaming responses.
- `choices[].finish_reason`: Uses `tool_calls` when the model stops to request tool execution.

## Streaming Fields

- `choices[].delta.tool_calls[].index`: Position of the streamed tool call.
- `choices[].delta.tool_calls[].id`: Tool call identifier when first known.
- `choices[].delta.tool_calls[].type`: Tool call type, usually `function`.
- `choices[].delta.tool_calls[].function.name`: Function name when first known.
- `choices[].delta.tool_calls[].function.arguments`: Partial JSON argument fragments.

## Compatibility Notes

- Requests without any tool fields keep existing chat completion behavior.
- Legacy `functions` and `function_call` translation is not required for Anthropic or Gemini in this feature.
- Errors must use the gateway's OpenAI-compatible error envelope.
