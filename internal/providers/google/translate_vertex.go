package google

import (
	"encoding/json"
	"fmt"
	"strings"

	vertex "cloud.google.com/go/vertexai/genai"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// buildVertexModel configures a Vertex AI GenerativeModel from the canonical
// request. It sets generation parameters, tools, and extracts system instructions
// from the message history, returning the model and the non-system content history.
func buildVertexModel(client *vertex.Client, modelName string, req *llm.ChatCompletionRequest) (*vertex.GenerativeModel, []*vertex.Content) {
	model := client.GenerativeModel(modelName)

	if req.Temperature != nil {
		model.SetTemperature(float32(*req.Temperature))
	}
	if req.TopP != nil {
		model.SetTopP(float32(*req.TopP))
	}
	if req.MaxTokens != nil {
		model.SetMaxOutputTokens(int32(*req.MaxTokens))
	}
	if req.N > 0 {
		model.SetCandidateCount(int32(req.N))
	}

	if req.Stop != nil {
		model.StopSequences = parseStopSequences(req.Stop)
	}

	// Translate tools.
	if len(req.Tools) > 0 {
		var funcDecls []*vertex.FunctionDeclaration
		for _, t := range req.Tools {
			fd := &vertex.FunctionDeclaration{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  jsonSchemaToVertexSchema(t.Function.Parameters),
			}
			funcDecls = append(funcDecls, fd)
		}
		model.Tools = []*vertex.Tool{{FunctionDeclarations: funcDecls}}

		if req.ToolChoice != nil {
			model.ToolConfig = buildVertexToolConfig(req.ToolChoice)
		}
	}

	var systemParts []vertex.Part
	var history []*vertex.Content

	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			systemParts = append(systemParts, vertexContentPartsFromMessage(m)...)
		case "assistant":
			parts := vertexTextAndFuncCallParts(m)
			history = append(history, &vertex.Content{
				Role:  "model",
				Parts: parts,
			})
		case "tool":
			name := findVertexFunctionNameByCallID(history, m.ToolCallID)
			var responseMap map[string]any
			if err := json.Unmarshal(m.Content, &responseMap); err != nil {
				responseMap = map[string]any{"result": m.ContentString()}
			}
			history = append(history, &vertex.Content{
				Role: "function",
				Parts: []vertex.Part{
					vertex.FunctionResponse{
						Name:     name,
						Response: responseMap,
					},
				},
			})
		default: // user
			parts := vertexContentPartsFromMessage(m)
			if len(parts) == 0 {
				parts = []vertex.Part{vertex.Text("")}
			}
			history = append(history, &vertex.Content{
				Role:  "user",
				Parts: parts,
			})
		}
	}

	if len(systemParts) > 0 {
		model.SystemInstruction = &vertex.Content{Parts: systemParts}
	}

	return model, history
}

func vertexTextAndFuncCallParts(m llm.Message) []vertex.Part {
	parts := vertexContentPartsFromMessage(m)
	for _, tc := range m.ToolCalls {
		var args map[string]any
		if tc.Function.Arguments != "" {
			json.Unmarshal([]byte(tc.Function.Arguments), &args)
		}
		parts = append(parts, vertex.FunctionCall{
			Name: tc.Function.Name,
			Args: args,
		})
	}
	if len(parts) == 0 {
		parts = []vertex.Part{vertex.Text("")}
	}
	return parts
}

func vertexContentPartsFromMessage(m llm.Message) []vertex.Part {
	if !rawMessageIsJSONArray(m.Content) {
		if len(m.Content) == 0 {
			return nil
		}
		return []vertex.Part{vertex.Text(m.ContentString())}
	}

	contentParts, err := m.ContentParts()
	if err != nil {
		return []vertex.Part{vertex.Text(m.ContentString())}
	}

	parts := make([]vertex.Part, 0, len(contentParts))
	for _, part := range contentParts {
		switch part.Type {
		case "text":
			parts = append(parts, vertex.Text(part.Text))
		case "image_url":
			if part.ImageURL == nil || part.ImageURL.URL == "" {
				continue
			}
			imagePart, ok := vertexImagePartFromURL(part.ImageURL.URL)
			if !ok {
				continue
			}
			parts = append(parts, imagePart)
		}
	}
	return parts
}

func vertexImagePartFromURL(rawURL string) (vertex.Part, bool) {
	if strings.HasPrefix(rawURL, "https://") {
		return vertex.FileData{
			FileURI:  rawURL,
			MIMEType: inferImageMIMEFromURL(rawURL),
		}, true
	}

	mediaType, data, ok := parseBase64ImageDataURI(rawURL)
	if !ok {
		return nil, false
	}
	return vertex.Blob{
		MIMEType: mediaType,
		Data:     data,
	}, true
}

func findVertexFunctionNameByCallID(history []*vertex.Content, callID string) string {
	for i := len(history) - 1; i >= 0; i-- {
		for _, p := range history[i].Parts {
			if fc, ok := p.(vertex.FunctionCall); ok {
				_ = callID
				return fc.Name
			}
		}
	}
	return callID
}

// vertexResponseToOpenAI converts a Vertex AI GenerateContentResponse into
// an OpenAI-compatible ChatCompletionResponse.
func vertexResponseToOpenAI(resp *vertex.GenerateContentResponse, model string, id string, created int64) *llm.ChatCompletionResponse {
	out := &llm.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
	}

	for i, cand := range resp.Candidates {
		text := extractVertexText(cand)
		toolCalls := extractVertexToolCalls(cand)

		finishReason := mapVertexFinishReason(cand.FinishReason)
		if len(toolCalls) > 0 && finishReason == "stop" {
			finishReason = "tool_calls"
		}

		msg := &llm.Message{
			Role:    "assistant",
			Content: llm.StringContent(text),
		}
		if len(toolCalls) > 0 {
			msg.ToolCalls = toolCalls
		}

		out.Choices = append(out.Choices, llm.Choice{
			Index:        i,
			Message:      msg,
			FinishReason: finishReason,
		})
	}

	if resp.UsageMetadata != nil {
		out.Usage = &llm.Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return out
}

func extractVertexToolCalls(cand *vertex.Candidate) []llm.ToolCall {
	if cand.Content == nil {
		return nil
	}
	var toolCalls []llm.ToolCall
	for i, p := range cand.Content.Parts {
		fc, ok := p.(vertex.FunctionCall)
		if !ok {
			continue
		}
		argsJSON, _ := json.Marshal(fc.Args)
		toolCalls = append(toolCalls, llm.ToolCall{
			ID:   fmt.Sprintf("call_%d", i),
			Type: "function",
			Function: llm.FunctionCall{
				Name:      fc.Name,
				Arguments: string(argsJSON),
			},
		})
	}
	return toolCalls
}

func extractVertexText(cand *vertex.Candidate) string {
	if cand.Content == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range cand.Content.Parts {
		if t, ok := p.(vertex.Text); ok {
			b.WriteString(string(t))
		}
	}
	return b.String()
}

func mapVertexFinishReason(r vertex.FinishReason) string {
	switch r {
	case vertex.FinishReasonStop:
		return "stop"
	case vertex.FinishReasonMaxTokens:
		return "length"
	case vertex.FinishReasonSafety:
		return "content_filter"
	case vertex.FinishReasonRecitation:
		return "content_filter"
	default:
		if r == vertex.FinishReasonUnspecified {
			return ""
		}
		return "stop"
	}
}

// buildVertexToolConfig converts the OpenAI tool_choice value to a Vertex ToolConfig.
func buildVertexToolConfig(toolChoice json.RawMessage) *vertex.ToolConfig {
	var s string
	if err := json.Unmarshal(toolChoice, &s); err == nil {
		switch s {
		case "none":
			return &vertex.ToolConfig{
				FunctionCallingConfig: &vertex.FunctionCallingConfig{Mode: vertex.FunctionCallingNone},
			}
		case "auto":
			return &vertex.ToolConfig{
				FunctionCallingConfig: &vertex.FunctionCallingConfig{Mode: vertex.FunctionCallingAuto},
			}
		case "required":
			return &vertex.ToolConfig{
				FunctionCallingConfig: &vertex.FunctionCallingConfig{Mode: vertex.FunctionCallingAny},
			}
		}
		return nil
	}

	var obj struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(toolChoice, &obj); err == nil && obj.Type == "function" {
		return &vertex.ToolConfig{
			FunctionCallingConfig: &vertex.FunctionCallingConfig{
				Mode:                 vertex.FunctionCallingAny,
				AllowedFunctionNames: []string{obj.Function.Name},
			},
		}
	}
	return nil
}

// jsonSchemaToVertexSchema converts a JSON Schema object to a *vertex.Schema.
func jsonSchemaToVertexSchema(raw json.RawMessage) *vertex.Schema {
	if raw == nil {
		return nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}

	schema := &vertex.Schema{}

	if t, ok := m["type"]; ok {
		var typStr string
		json.Unmarshal(t, &typStr)
		switch typStr {
		case "string":
			schema.Type = vertex.TypeString
		case "number":
			schema.Type = vertex.TypeNumber
		case "integer":
			schema.Type = vertex.TypeInteger
		case "boolean":
			schema.Type = vertex.TypeBoolean
		case "array":
			schema.Type = vertex.TypeArray
		case "object":
			schema.Type = vertex.TypeObject
		}
	}

	if d, ok := m["description"]; ok {
		json.Unmarshal(d, &schema.Description)
	}

	if e, ok := m["enum"]; ok {
		var enums []string
		json.Unmarshal(e, &enums)
		schema.Enum = enums
	}

	if props, ok := m["properties"]; ok {
		var propsMap map[string]json.RawMessage
		if err := json.Unmarshal(props, &propsMap); err == nil {
			schema.Properties = make(map[string]*vertex.Schema)
			for k, v := range propsMap {
				schema.Properties[k] = jsonSchemaToVertexSchema(v)
			}
		}
	}

	if req, ok := m["required"]; ok {
		var required []string
		json.Unmarshal(req, &required)
		schema.Required = required
	}

	if items, ok := m["items"]; ok {
		schema.Items = jsonSchemaToVertexSchema(items)
	}

	return schema
}
