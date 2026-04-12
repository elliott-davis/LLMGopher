package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockdocument "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	bedrocktypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/ed007183/llmgopher/pkg/llm"
)

const bedrockDefaultMaxTokens = int32(4096)

// BedrockModelAliases maps common shorthand names to Bedrock model IDs.
var BedrockModelAliases = map[string]string{
	"claude-3-5-sonnet":  "anthropic.claude-3-5-sonnet-20241022-v2:0",
	"claude-3-5-haiku":   "anthropic.claude-3-5-haiku-20241022-v1:0",
	"claude-3-opus":      "anthropic.claude-3-opus-20240229-v1:0",
	"llama-3-70b":        "meta.llama3-70b-instruct-v1:0",
	"mistral-large":      "mistral.mistral-large-2402-v1:0",
	"command-r-plus":     "cohere.command-r-plus-v1:0",
	"titan-text-express": "amazon.titan-text-express-v1",
}

type bedrockConverseStream interface {
	Events() <-chan bedrocktypes.ConverseStreamOutput
	Err() error
	Close() error
}

type bedrockConverseClient interface {
	Converse(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error)
	ConverseStream(ctx context.Context, input *bedrockruntime.ConverseStreamInput) (bedrockConverseStream, error)
}

type awsBedrockClient struct {
	client *bedrockruntime.Client
}

func (c *awsBedrockClient) Converse(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
	return c.client.Converse(ctx, input)
}

func (c *awsBedrockClient) ConverseStream(ctx context.Context, input *bedrockruntime.ConverseStreamInput) (bedrockConverseStream, error) {
	out, err := c.client.ConverseStream(ctx, input)
	if err != nil {
		return nil, err
	}
	return out.GetStream(), nil
}

// BedrockProvider translates OpenAI-format requests into Bedrock Converse API
// requests and translates responses back to OpenAI-compatible payloads.
type BedrockProvider struct {
	client bedrockConverseClient
	region string
}

func NewBedrockProvider(region, accessKeyID, secretAccessKey, sessionToken string) (*BedrockProvider, error) {
	region = strings.TrimSpace(region)
	if region == "" {
		return nil, fmt.Errorf("bedrock region is required")
	}

	loadOptions := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	accessKeyID = strings.TrimSpace(accessKeyID)
	secretAccessKey = strings.TrimSpace(secretAccessKey)
	sessionToken = strings.TrimSpace(sessionToken)
	if accessKeyID != "" {
		if secretAccessKey == "" {
			return nil, fmt.Errorf("secret access key is required when access key ID is provided")
		}
		loadOptions = append(loadOptions, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken),
		))
	}

	awsConfig, err := config.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &BedrockProvider{
		client: &awsBedrockClient{
			client: bedrockruntime.NewFromConfig(awsConfig),
		},
		region: region,
	}, nil
}

func (p *BedrockProvider) Name() string { return "bedrock" }

func (p *BedrockProvider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	input, err := p.toConverseInput(req)
	if err != nil {
		return nil, err
	}

	out, err := p.client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("bedrock converse request: %w", err)
	}

	return p.fromConverseOutput(out, req.Model), nil
}

func (p *BedrockProvider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (io.ReadCloser, error) {
	input, err := p.toConverseInput(req)
	if err != nil {
		return nil, err
	}

	streamInput := &bedrockruntime.ConverseStreamInput{
		ModelId:         input.ModelId,
		InferenceConfig: input.InferenceConfig,
		Messages:        input.Messages,
		System:          input.System,
		ToolConfig:      input.ToolConfig,
	}

	stream, err := p.client.ConverseStream(ctx, streamInput)
	if err != nil {
		return nil, fmt.Errorf("bedrock converse stream request: %w", err)
	}

	pr, pw := io.Pipe()
	go p.translateStream(stream, pw, req.Model)
	return pr, nil
}

func (p *BedrockProvider) toConverseInput(req *llm.ChatCompletionRequest) (*bedrockruntime.ConverseInput, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	maxTokens := bedrockDefaultMaxTokens
	if req.MaxTokens != nil {
		maxTokens = int32(*req.MaxTokens)
	}

	inferenceConfig := &bedrocktypes.InferenceConfiguration{
		MaxTokens: &maxTokens,
	}
	if req.Temperature != nil {
		t := float32(*req.Temperature)
		inferenceConfig.Temperature = &t
	}
	if req.TopP != nil {
		topP := float32(*req.TopP)
		inferenceConfig.TopP = &topP
	}

	stopSequences, err := parseStopSequences(req.Stop)
	if err != nil {
		return nil, err
	}
	if len(stopSequences) > 0 {
		inferenceConfig.StopSequences = stopSequences
	}

	messages := make([]bedrocktypes.Message, 0, len(req.Messages))
	system := make([]bedrocktypes.SystemContentBlock, 0)
	for _, message := range req.Messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "system":
			if text := strings.TrimSpace(message.ContentString()); text != "" {
				system = append(system, &bedrocktypes.SystemContentBlockMemberText{Value: text})
			}
		case "assistant":
			bedrockMsg, err := buildBedrockAssistantMessage(message)
			if err != nil {
				return nil, err
			}
			messages = append(messages, bedrockMsg)
		case "tool":
			bedrockMsg, err := buildBedrockToolResultMessage(message)
			if err != nil {
				return nil, err
			}
			messages = append(messages, bedrockMsg)
		default:
			messages = append(messages, buildBedrockTextMessage(message, bedrocktypes.ConversationRoleUser))
		}
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:         bedrockPtr(resolveBedrockModelID(req.Model)),
		InferenceConfig: inferenceConfig,
		Messages:        messages,
	}
	if len(system) > 0 {
		input.System = system
	}

	if len(req.Tools) > 0 {
		toolConfig, err := toBedrockToolConfig(req.Tools, req.ToolChoice)
		if err != nil {
			return nil, err
		}
		input.ToolConfig = toolConfig
	}

	return input, nil
}

func buildBedrockTextMessage(message llm.Message, role bedrocktypes.ConversationRole) bedrocktypes.Message {
	content := contentBlocksFromMessage(message)
	if len(content) == 0 {
		content = append(content, &bedrocktypes.ContentBlockMemberText{Value: ""})
	}

	return bedrocktypes.Message{
		Role:    role,
		Content: content,
	}
}

func buildBedrockAssistantMessage(message llm.Message) (bedrocktypes.Message, error) {
	content := contentBlocksFromMessage(message)
	for idx, toolCall := range message.ToolCalls {
		name := strings.TrimSpace(toolCall.Function.Name)
		if name == "" {
			continue
		}

		toolUseID := strings.TrimSpace(toolCall.ID)
		if toolUseID == "" {
			toolUseID = fmt.Sprintf("tool_call_%d", idx)
		}

		toolInput, err := parseToolArguments(toolCall.Function.Arguments)
		if err != nil {
			return bedrocktypes.Message{}, err
		}

		content = append(content, &bedrocktypes.ContentBlockMemberToolUse{
			Value: bedrocktypes.ToolUseBlock{
				ToolUseId: bedrockPtr(toolUseID),
				Name:      bedrockPtr(name),
				Input:     toolInput,
			},
		})
	}

	if len(content) == 0 {
		content = append(content, &bedrocktypes.ContentBlockMemberText{Value: ""})
	}

	return bedrocktypes.Message{
		Role:    bedrocktypes.ConversationRoleAssistant,
		Content: content,
	}, nil
}

func buildBedrockToolResultMessage(message llm.Message) (bedrocktypes.Message, error) {
	toolUseID := strings.TrimSpace(message.ToolCallID)
	if toolUseID == "" {
		return bedrocktypes.Message{}, fmt.Errorf("tool message is missing tool_call_id")
	}

	resultContent, err := toToolResultContent(message.Content)
	if err != nil {
		return bedrocktypes.Message{}, err
	}

	return bedrocktypes.Message{
		Role: bedrocktypes.ConversationRoleUser,
		Content: []bedrocktypes.ContentBlock{
			&bedrocktypes.ContentBlockMemberToolResult{
				Value: bedrocktypes.ToolResultBlock{
					ToolUseId: bedrockPtr(toolUseID),
					Content:   resultContent,
				},
			},
		},
	}, nil
}

func contentBlocksFromMessage(message llm.Message) []bedrocktypes.ContentBlock {
	parts, err := message.ContentParts()
	if err == nil && len(parts) > 0 {
		content := make([]bedrocktypes.ContentBlock, 0, len(parts))
		for _, part := range parts {
			if part.Type != "text" || strings.TrimSpace(part.Text) == "" {
				continue
			}
			content = append(content, &bedrocktypes.ContentBlockMemberText{Value: part.Text})
		}
		return content
	}

	text := message.ContentString()
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return []bedrocktypes.ContentBlock{
		&bedrocktypes.ContentBlockMemberText{Value: text},
	}
}

func parseToolArguments(raw string) (bedrockdocument.Interface, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return bedrockdocument.NewLazyDocument(map[string]any{}), nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("parse tool arguments: %w", err)
	}
	return bedrockdocument.NewLazyDocument(parsed), nil
}

func toToolResultContent(raw json.RawMessage) ([]bedrocktypes.ToolResultContentBlock, error) {
	content := strings.TrimSpace(string(raw))
	if content == "" {
		return []bedrocktypes.ToolResultContentBlock{
			&bedrocktypes.ToolResultContentBlockMemberText{Value: ""},
		}, nil
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return []bedrocktypes.ToolResultContentBlock{
			&bedrocktypes.ToolResultContentBlockMemberText{Value: asString},
		}, nil
	}

	var asJSON any
	if err := json.Unmarshal(raw, &asJSON); err != nil {
		return nil, fmt.Errorf("parse tool result content: %w", err)
	}
	return []bedrocktypes.ToolResultContentBlock{
		&bedrocktypes.ToolResultContentBlockMemberJson{
			Value: bedrockdocument.NewLazyDocument(asJSON),
		},
	}, nil
}

func toBedrockToolConfig(tools []llm.Tool, toolChoice json.RawMessage) (*bedrocktypes.ToolConfiguration, error) {
	bedrockTools := make([]bedrocktypes.Tool, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Function.Name)
		if name == "" {
			continue
		}

		inputSchema := map[string]any{}
		if len(tool.Function.Parameters) > 0 {
			if err := json.Unmarshal(tool.Function.Parameters, &inputSchema); err != nil {
				return nil, fmt.Errorf("parse tool schema for %q: %w", name, err)
			}
		}

		spec := bedrocktypes.ToolSpecification{
			Name: bedrockPtr(name),
			InputSchema: &bedrocktypes.ToolInputSchemaMemberJson{
				Value: bedrockdocument.NewLazyDocument(inputSchema),
			},
		}
		if description := strings.TrimSpace(tool.Function.Description); description != "" {
			spec.Description = bedrockPtr(description)
		}

		bedrockTools = append(bedrockTools, &bedrocktypes.ToolMemberToolSpec{Value: spec})
	}

	if len(bedrockTools) == 0 {
		return nil, nil
	}

	cfg := &bedrocktypes.ToolConfiguration{
		Tools: bedrockTools,
	}

	choice, includeChoice, err := toBedrockToolChoice(toolChoice)
	if err != nil {
		return nil, err
	}
	if includeChoice {
		cfg.ToolChoice = choice
	}

	return cfg, nil
}

func toBedrockToolChoice(raw json.RawMessage) (choice bedrocktypes.ToolChoice, include bool, err error) {
	if len(raw) == 0 {
		return nil, false, nil
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		switch strings.ToLower(strings.TrimSpace(asString)) {
		case "auto":
			return &bedrocktypes.ToolChoiceMemberAuto{
				Value: bedrocktypes.AutoToolChoice{},
			}, true, nil
		case "required":
			return &bedrocktypes.ToolChoiceMemberAny{
				Value: bedrocktypes.AnyToolChoice{},
			}, true, nil
		case "none":
			return nil, false, nil
		default:
			return nil, false, nil
		}
	}

	var objectChoice struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(raw, &objectChoice); err != nil {
		return nil, false, fmt.Errorf("parse tool_choice: %w", err)
	}

	if strings.EqualFold(objectChoice.Type, "function") && strings.TrimSpace(objectChoice.Function.Name) != "" {
		return &bedrocktypes.ToolChoiceMemberTool{
			Value: bedrocktypes.SpecificToolChoice{
				Name: bedrockPtr(strings.TrimSpace(objectChoice.Function.Name)),
			},
		}, true, nil
	}

	return nil, false, nil
}

func parseStopSequences(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var sequences []string
	if err := json.Unmarshal(raw, &sequences); err == nil {
		return sequences, nil
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}, nil
	}

	return nil, fmt.Errorf("parse stop sequences: invalid format")
}

func (p *BedrockProvider) fromConverseOutput(out *bedrockruntime.ConverseOutput, model string) *llm.ChatCompletionResponse {
	text := ""
	toolCalls := make([]llm.ToolCall, 0)

	switch output := out.Output.(type) {
	case *bedrocktypes.ConverseOutputMemberMessage:
		text, toolCalls = extractBedrockOutputContent(output.Value.Content)
	}

	message := &llm.Message{
		Role:    "assistant",
		Content: llm.StringContent(text),
	}
	if len(toolCalls) > 0 {
		message.ToolCalls = toolCalls
	}

	response := &llm.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-bedrock-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []llm.Choice{
			{
				Index:        0,
				Message:      message,
				FinishReason: mapBedrockStopReason(out.StopReason),
			},
		},
	}

	if out.Usage != nil {
		promptTokens := int(valueOrZero(out.Usage.InputTokens))
		completionTokens := int(valueOrZero(out.Usage.OutputTokens))
		totalTokens := int(valueOrZero(out.Usage.TotalTokens))
		if totalTokens == 0 {
			totalTokens = promptTokens + completionTokens
		}

		response.Usage = &llm.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		}
	}

	return response
}

func extractBedrockOutputContent(content []bedrocktypes.ContentBlock) (text string, toolCalls []llm.ToolCall) {
	var textBuilder strings.Builder
	for _, block := range content {
		switch value := block.(type) {
		case *bedrocktypes.ContentBlockMemberText:
			textBuilder.WriteString(value.Value)
		case *bedrocktypes.ContentBlockMemberToolUse:
			toolCalls = append(toolCalls, toLLMToolCall(value.Value))
		}
	}
	return textBuilder.String(), toolCalls
}

func toLLMToolCall(block bedrocktypes.ToolUseBlock) llm.ToolCall {
	arguments := "{}"
	if block.Input != nil {
		if encoded, err := block.Input.MarshalSmithyDocument(); err == nil && len(encoded) > 0 {
			arguments = string(encoded)
		} else {
			var decoded any
			if err := block.Input.UnmarshalSmithyDocument(&decoded); err == nil {
				if encoded, err := json.Marshal(decoded); err == nil {
					arguments = string(encoded)
				}
			}
		}
	}

	return llm.ToolCall{
		ID:   valueOrEmpty(block.ToolUseId),
		Type: "function",
		Function: llm.FunctionCall{
			Name:      valueOrEmpty(block.Name),
			Arguments: arguments,
		},
	}
}

func (p *BedrockProvider) translateStream(src bedrockConverseStream, dst *io.PipeWriter, model string) {
	defer dst.Close()
	defer src.Close()

	created := time.Now().Unix()
	streamID := fmt.Sprintf("chatcmpl-bedrock-%d", time.Now().UnixNano())

	for event := range src.Events() {
		switch e := event.(type) {
		case *bedrocktypes.ConverseStreamOutputMemberContentBlockStart:
			p.emitContentBlockStart(dst, streamID, created, model, e.Value)
		case *bedrocktypes.ConverseStreamOutputMemberContentBlockDelta:
			p.emitContentBlockDelta(dst, streamID, created, model, e.Value)
		case *bedrocktypes.ConverseStreamOutputMemberMetadata:
			p.emitUsageChunk(dst, streamID, created, model, e.Value)
		case *bedrocktypes.ConverseStreamOutputMemberMessageStop:
			p.emitFinishChunk(dst, streamID, created, model, e.Value.StopReason)
			fmt.Fprint(dst, "data: [DONE]\n\n")
			return
		}
	}

	if err := src.Err(); err != nil {
		_ = dst.CloseWithError(fmt.Errorf("bedrock stream: %w", err))
		return
	}

	fmt.Fprint(dst, "data: [DONE]\n\n")
}

func (p *BedrockProvider) emitContentBlockStart(dst *io.PipeWriter, streamID string, created int64, model string, event bedrocktypes.ContentBlockStartEvent) {
	index := int(valueOrZero(event.ContentBlockIndex))
	var toolUse bedrocktypes.ToolUseBlockStart
	found := false

	switch start := event.Start.(type) {
	case *bedrocktypes.ContentBlockStartMemberToolUse:
		toolUse = start.Value
		found = true
	}

	if !found {
		return
	}

	chunk := llm.ChatCompletionChunk{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Delta: &llm.MessageDelta{
					ToolCalls: []llm.ToolCallDelta{
						{
							Index: index,
							ID:    valueOrEmpty(toolUse.ToolUseId),
							Type:  "function",
							Function: llm.FunctionCall{
								Name: valueOrEmpty(toolUse.Name),
							},
						},
					},
				},
			},
		},
	}
	writeSSEChunk(dst, chunk)
}

func (p *BedrockProvider) emitContentBlockDelta(dst *io.PipeWriter, streamID string, created int64, model string, event bedrocktypes.ContentBlockDeltaEvent) {
	index := int(valueOrZero(event.ContentBlockIndex))

	switch delta := event.Delta.(type) {
	case *bedrocktypes.ContentBlockDeltaMemberText:
		p.emitTextDeltaChunk(dst, streamID, created, model, delta.Value)
	case *bedrocktypes.ContentBlockDeltaMemberToolUse:
		p.emitToolUseDeltaChunk(dst, streamID, created, model, index, delta.Value)
	}
}

func (p *BedrockProvider) emitTextDeltaChunk(dst *io.PipeWriter, streamID string, created int64, model string, text string) {
	if text == "" {
		return
	}

	chunk := llm.ChatCompletionChunk{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Delta: &llm.MessageDelta{
					Content: llm.StringContent(text),
				},
			},
		},
	}
	writeSSEChunk(dst, chunk)
}

func (p *BedrockProvider) emitToolUseDeltaChunk(dst *io.PipeWriter, streamID string, created int64, model string, index int, delta bedrocktypes.ToolUseBlockDelta) {
	if delta.Input == nil || strings.TrimSpace(*delta.Input) == "" {
		return
	}

	chunk := llm.ChatCompletionChunk{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Delta: &llm.MessageDelta{
					ToolCalls: []llm.ToolCallDelta{
						{
							Index: index,
							Function: llm.FunctionCall{
								Arguments: *delta.Input,
							},
						},
					},
				},
			},
		},
	}
	writeSSEChunk(dst, chunk)
}

func (p *BedrockProvider) emitFinishChunk(dst *io.PipeWriter, streamID string, created int64, model string, reason bedrocktypes.StopReason) {
	chunk := llm.ChatCompletionChunk{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []llm.Choice{
			{
				Index:        0,
				Delta:        &llm.MessageDelta{},
				FinishReason: mapBedrockStopReason(reason),
			},
		},
	}
	writeSSEChunk(dst, chunk)
}

func (p *BedrockProvider) emitUsageChunk(dst *io.PipeWriter, streamID string, created int64, model string, event bedrocktypes.ConverseStreamMetadataEvent) {
	if event.Usage == nil {
		return
	}

	usage := &llm.Usage{
		PromptTokens:     int(valueOrZero(event.Usage.InputTokens)),
		CompletionTokens: int(valueOrZero(event.Usage.OutputTokens)),
		TotalTokens:      int(valueOrZero(event.Usage.TotalTokens)),
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	chunk := llm.ChatCompletionChunk{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Delta: &llm.MessageDelta{},
			},
		},
		Usage: usage,
	}
	writeSSEChunk(dst, chunk)
}

func writeSSEChunk(dst *io.PipeWriter, chunk llm.ChatCompletionChunk) {
	encoded, err := json.Marshal(chunk)
	if err != nil {
		return
	}
	fmt.Fprintf(dst, "data: %s\n\n", encoded)
}

func mapBedrockStopReason(reason bedrocktypes.StopReason) string {
	switch string(reason) {
	case "end_turn":
		return "stop"
	case "tool_use":
		return "tool_calls"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	default:
		return string(reason)
	}
}

func resolveBedrockModelID(model string) string {
	model = strings.TrimSpace(model)
	if resolved, ok := BedrockModelAliases[model]; ok {
		return resolved
	}
	return model
}

func resolveBedrockRegion(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if !strings.Contains(trimmed, "://") {
		if looksLikeBedrockHost(trimmed) {
			return regionFromBedrockHost(trimmed)
		}
		return trimmed
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		return ""
	}

	host := parsedURL.Hostname()
	if host == "" {
		return ""
	}
	if looksLikeBedrockHost(host) {
		return regionFromBedrockHost(host)
	}
	return host
}

func looksLikeBedrockHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return strings.HasPrefix(host, "bedrock-runtime.")
}

func regionFromBedrockHost(host string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(host)), ".")
	if len(parts) < 2 || parts[0] != "bedrock-runtime" {
		return ""
	}
	return parts[1]
}

func bedrockPtr[T any](value T) *T {
	return &value
}

func valueOrZero[T ~int32](value *T) T {
	if value == nil {
		return 0
	}
	return *value
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
