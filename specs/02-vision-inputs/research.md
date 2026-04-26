# Research: Vision / Image Inputs

## Decision: Use OpenAI Content Parts as the Public Contract

**Rationale**: OpenAI-compatible clients already send multimodal messages as content arrays with `text` and `image_url` parts.

**Alternatives considered**: Provider-specific message formats were rejected because they would break SDK compatibility and force callers to know provider routing details.

## Decision: Support HTTPS Image URLs and Base64 Data URIs

**Rationale**: Hosted images and inline images cover the most common client workflows without adding file upload infrastructure.

**Alternatives considered**: Multipart uploads were rejected for this feature because OpenAI-compatible chat completions commonly use URL or data URI image references.

## Decision: Use Conservative Fixed Image Token Estimates

**Rationale**: Provider-specific image token accounting varies by model and image detail. A fixed conservative estimate avoids undercounting spend.

**Alternatives considered**: Exact image token counting was deferred because it requires provider-specific logic and image dimension handling that is not necessary for initial compatibility.

## Decision: Treat Existing Plan Status as Implementation Evidence Only

**Rationale**: The original plan marks all acceptance criteria complete, but functional validation against a running gateway or real vision-capable model is not documented.

**Alternatives considered**: Marking the Spec Kit feature as verified was rejected until a smoke test is recorded.
