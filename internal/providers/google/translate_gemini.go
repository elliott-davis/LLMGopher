package google

import (
	"encoding/json"
	"fmt"
	"strings"

	gemini "github.com/google/generative-ai-go/genai"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// buildGeminiModel configures a GenerativeModel from the canonical request.
// It sets generation parameters, tools, and extracts system instructions from
// the message history, returning the model and the non-system content history.
func buildGeminiModel(client *gemini.Client, modelName string, req *llm.ChatCompletionRequest) (*gemini.GenerativeModel, []*gemini.Content) {
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
		var funcDecls []*gemini.FunctionDeclaration
		for _, t := range req.Tools {
			fd := &gemini.FunctionDeclaration{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  jsonSchemaToGeminiSchema(t.Function.Parameters),
			}
			funcDecls = append(funcDecls, fd)
		}
		model.Tools = []*gemini.Tool{{FunctionDeclarations: funcDecls}}

		if req.ToolChoice != nil {
			model.ToolConfig = buildGeminiToolConfig(req.ToolChoice)
		}
	}

	var systemParts []gemini.Part
	var history []*gemini.Content

	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			systemParts = append(systemParts, geminiContentPartsFromMessage(m)...)
		case "assistant":
			parts := textAndFuncCallParts(m)
			history = append(history, &gemini.Content{
				Role:  "model",
				Parts: parts,
			})
		case "tool":
			// Convert tool result to a FunctionResponse part.
			// Look back through history to find the function name by ToolCallID.
			name := findFunctionNameByCallID(history, m.ToolCallID)
			var responseMap map[string]any
			if err := json.Unmarshal(m.Content, &responseMap); err != nil {
				// Treat as a plain string result.
				responseMap = map[string]any{"result": m.ContentString()}
			}
			history = append(history, &gemini.Content{
				Role: "function",
				Parts: []gemini.Part{
					gemini.FunctionResponse{
						Name:     name,
						Response: responseMap,
					},
				},
			})
		default: // user
			parts := geminiContentPartsFromMessage(m)
			if len(parts) == 0 {
				parts = []gemini.Part{gemini.Text("")}
			}
			history = append(history, &gemini.Content{
				Role:  "user",
				Parts: parts,
			})
		}
	}

	if len(systemParts) > 0 {
		model.SystemInstruction = &gemini.Content{Parts: systemParts}
	}

	return model, history
}

// textAndFuncCallParts converts an assistant message to Gemini Parts,
// including FunctionCall parts for any tool calls.
func textAndFuncCallParts(m llm.Message) []gemini.Part {
	parts := geminiContentPartsFromMessage(m)
	for _, tc := range m.ToolCalls {
		var args map[string]any
		if tc.Function.Arguments != "" {
			json.Unmarshal([]byte(tc.Function.Arguments), &args)
		}
		parts = append(parts, gemini.FunctionCall{
			Name: tc.Function.Name,
			Args: args,
		})
	}
	if len(parts) == 0 {
		parts = []gemini.Part{gemini.Text("")}
	}
	return parts
}

func geminiContentPartsFromMessage(m llm.Message) []gemini.Part {
	if !rawMessageIsJSONArray(m.Content) {
		if len(m.Content) == 0 {
			return nil
		}
		return []gemini.Part{gemini.Text(m.ContentString())}
	}

	contentParts, err := m.ContentParts()
	if err != nil {
		return []gemini.Part{gemini.Text(m.ContentString())}
	}

	parts := make([]gemini.Part, 0, len(contentParts))
	for _, part := range contentParts {
		switch part.Type {
		case "text":
			parts = append(parts, gemini.Text(part.Text))
		case "image_url":
			if part.ImageURL == nil || part.ImageURL.URL == "" {
				continue
			}
			imagePart, ok := geminiImagePartFromURL(part.ImageURL.URL)
			if !ok {
				continue
			}
			parts = append(parts, imagePart)
		}
	}
	return parts
}

func geminiImagePartFromURL(rawURL string) (gemini.Part, bool) {
	if strings.HasPrefix(rawURL, "https://") {
		return gemini.FileData{
			URI:      rawURL,
			MIMEType: inferImageMIMEFromURL(rawURL),
		}, true
	}

	mediaType, data, ok := parseBase64ImageDataURI(rawURL)
	if !ok {
		return nil, false
	}
	return gemini.Blob{
		MIMEType: mediaType,
		Data:     data,
	}, true
}

// findFunctionNameByCallID scans backwards through the history to find the
// function name corresponding to a tool call ID. Returns the ID itself if
// the name cannot be determined.
func findFunctionNameByCallID(history []*gemini.Content, callID string) string {
	for i := len(history) - 1; i >= 0; i-- {
		for _, p := range history[i].Parts {
			if fc, ok := p.(gemini.FunctionCall); ok {
				// Gemini FunctionCall doesn't carry an ID; use name as a best-effort
				// match when there's only one pending call in the last model turn.
				_ = callID
				return fc.Name
			}
		}
	}
	return callID
}

// geminiResponseToOpenAI converts a Gemini GenerateContentResponse into
// an OpenAI-compatible ChatCompletionResponse.
func geminiResponseToOpenAI(resp *gemini.GenerateContentResponse, model string, id string, created int64) *llm.ChatCompletionResponse {
	out := &llm.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
	}

	for i, cand := range resp.Candidates {
		text := extractGeminiText(cand)
		toolCalls := extractGeminiToolCalls(cand)

		finishReason := mapGeminiFinishReason(cand.FinishReason)
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

// extractGeminiToolCalls extracts FunctionCall parts from a candidate.
func extractGeminiToolCalls(cand *gemini.Candidate) []llm.ToolCall {
	if cand.Content == nil {
		return nil
	}
	var toolCalls []llm.ToolCall
	for i, p := range cand.Content.Parts {
		fc, ok := p.(gemini.FunctionCall)
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

func extractGeminiText(cand *gemini.Candidate) string {
	if cand.Content == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range cand.Content.Parts {
		if t, ok := p.(gemini.Text); ok {
			b.WriteString(string(t))
		}
	}
	return b.String()
}

func mapGeminiFinishReason(r gemini.FinishReason) string {
	switch r {
	case gemini.FinishReasonStop:
		return "stop"
	case gemini.FinishReasonMaxTokens:
		return "length"
	case gemini.FinishReasonSafety:
		return "content_filter"
	case gemini.FinishReasonRecitation:
		return "content_filter"
	default:
		if r == gemini.FinishReasonUnspecified {
			return ""
		}
		return "stop"
	}
}

func parseStopSequences(raw json.RawMessage) []string {
	var stops []string
	if err := json.Unmarshal(raw, &stops); err != nil {
		var single string
		if err := json.Unmarshal(raw, &single); err == nil {
			stops = []string{single}
		}
	}
	return stops
}

// buildGeminiToolConfig converts the OpenAI tool_choice value to a Gemini ToolConfig.
func buildGeminiToolConfig(toolChoice json.RawMessage) *gemini.ToolConfig {
	// String forms: "none", "auto", "required"
	var s string
	if err := json.Unmarshal(toolChoice, &s); err == nil {
		switch s {
		case "none":
			return &gemini.ToolConfig{
				FunctionCallingConfig: &gemini.FunctionCallingConfig{Mode: gemini.FunctionCallingNone},
			}
		case "auto":
			return &gemini.ToolConfig{
				FunctionCallingConfig: &gemini.FunctionCallingConfig{Mode: gemini.FunctionCallingAuto},
			}
		case "required":
			return &gemini.ToolConfig{
				FunctionCallingConfig: &gemini.FunctionCallingConfig{Mode: gemini.FunctionCallingAny},
			}
		}
		return nil
	}

	// Object form: {"type": "function", "function": {"name": "..."}}
	var obj struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(toolChoice, &obj); err == nil && obj.Type == "function" {
		return &gemini.ToolConfig{
			FunctionCallingConfig: &gemini.FunctionCallingConfig{
				Mode:                 gemini.FunctionCallingAny,
				AllowedFunctionNames: []string{obj.Function.Name},
			},
		}
	}
	return nil
}

// jsonSchemaToGeminiSchema converts a JSON Schema object (json.RawMessage) to a *gemini.Schema.
// Handles the common subset: type, description, properties, required, items, enum.
func jsonSchemaToGeminiSchema(raw json.RawMessage) *gemini.Schema {
	if raw == nil {
		return nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}

	schema := &gemini.Schema{}

	if t, ok := m["type"]; ok {
		var typStr string
		json.Unmarshal(t, &typStr)
		switch typStr {
		case "string":
			schema.Type = gemini.TypeString
		case "number":
			schema.Type = gemini.TypeNumber
		case "integer":
			schema.Type = gemini.TypeInteger
		case "boolean":
			schema.Type = gemini.TypeBoolean
		case "array":
			schema.Type = gemini.TypeArray
		case "object":
			schema.Type = gemini.TypeObject
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
			schema.Properties = make(map[string]*gemini.Schema)
			for k, v := range propsMap {
				schema.Properties[k] = jsonSchemaToGeminiSchema(v)
			}
		}
	}

	if req, ok := m["required"]; ok {
		var required []string
		json.Unmarshal(req, &required)
		schema.Required = required
	}

	if items, ok := m["items"]; ok {
		schema.Items = jsonSchemaToGeminiSchema(items)
	}

	return schema
}
