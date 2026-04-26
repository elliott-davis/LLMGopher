# Research: POST /v1/completions Endpoint

## Decision: Translate Legacy Completions to Chat Dispatch Internally

**Rationale**: The gateway already routes and proxies chat completions across supported providers. Translating a string prompt into a single user message avoids provider-specific completions implementations.

**Alternatives considered**: Calling provider-native completions endpoints was rejected because Anthropic and Gemini do not share a direct legacy completions contract and it would duplicate routing, audit, and cost behavior.

## Decision: Support Single String Prompts First

**Rationale**: A single string prompt covers the most common legacy completions workflow while keeping response mapping predictable.

**Alternatives considered**: Prompt arrays were rejected for the first version because they require batching semantics and response indexing that are distinct from the chat translation path.

## Decision: Accept but Ignore Compatibility-Only Parameters

**Rationale**: Accepting parameters such as `echo`, `logprobs`, `best_of`, and `suffix` avoids decode failures for existing clients, while the feature remains explicit that these behaviors are not implemented.

**Alternatives considered**: Rejecting unknown legacy fields was rejected because it would reduce SDK compatibility for no operational benefit.

## Decision: Treat Existing Plan Status as Implementation Evidence Only

**Rationale**: The original plan marks all acceptance criteria complete and includes implementation notes, but SDK smoke verification is only noted as optional.

**Alternatives considered**: Marking the Spec Kit feature as verified was rejected until the SDK smoke test is documented.
