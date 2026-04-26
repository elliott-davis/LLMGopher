# Contract: OpenAI-Compatible Chat Content Parts

## Message Content

`messages[].content` accepts either:

- A JSON string for plain text messages.
- An array of content parts for multimodal messages.

## Text Content Part

```json
{
  "type": "text",
  "text": "Describe this image"
}
```

## Image URL Content Part

```json
{
  "type": "image_url",
  "image_url": {
    "url": "https://example.com/image.jpg",
    "detail": "auto"
  }
}
```

## Base64 Image Data URI

```json
{
  "type": "image_url",
  "image_url": {
    "url": "data:image/jpeg;base64,..."
  }
}
```

## Compatibility Notes

- The `detail` field is accepted and preserved where possible, but this feature does not change token estimates based on detail.
- Video inputs and multipart file uploads are outside this contract.
- Errors must use the gateway's OpenAI-compatible error envelope.
