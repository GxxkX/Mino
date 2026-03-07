package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mino/backend/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// ──────────────────────────────────────────────────────────────────────────────
// Agent System Prompts
// ──────────────────────────────────────────────────────────────────────────────

// chatAgentSystemPrompt is used by the web/mobile/desktop frontend for interactive chat.
// It powers the RAG-based conversational Q&A experience described in INSTRUCTION.md §3.2.
const chatAgentSystemPrompt = `你是 Mino，一个隐私优先的个人 AI 助理。你的职责是帮助用户回顾、分析和检索他们的历史对话与笔记。

## 核心能力
- 根据用户的历史对话记录回答问题
- 帮助用户回忆过去的事件、决策和想法
- 从历史记忆中提炼见解和建议
- 管理和追踪用户的待办事项与行动项

## 回答规则
1. 基于提供的上下文信息回答，不要编造不存在的内容
2. 如果引用了特定的对话记录，自然地提及来源
3. 回答要简洁、准确、有帮助
4. 如果上下文中没有相关信息，诚实地告知用户
5. 使用用户的语言（中文或英文）回答
6. 保持友好、专业的语气`

// extractAgentSystemPrompt is used after STT transcription to extract structured data.
// It powers the structured information extraction described in INSTRUCTION.md §3.1 step 9.
const extractAgentSystemPrompt = `你是一个专业的语音转写内容分析助手。你的任务是从语音转写文本中提取结构化信息。

## 提取要求
请从转写文本中提取以下信息，并以 JSON 格式返回：

1. **title** (标题): 用一句简短的话概括对话主题，不超过50个字符
2. **summary** (概要): 简要总结对话的核心内容，不超过200个字符
3. **action_items** (行动项): 提取对话中提到的待办事项、需要执行的任务
4. **memories** (记忆点): 提取值得记住的关键信息，包括：
   - 见解 (insight): 有价值的想法或观点
   - 事实 (fact): 具体的事实信息
   - 偏好 (preference): 用户的偏好或习惯
   - 事件 (event): 重要的事件或约定

## 输出格式
仅返回有效的 JSON，不要包含任何其他文字：
{
  "title": "简短标题",
  "summary": "内容概要",
  "action_items": ["行动项1", "行动项2"],
  "memories": [
    {"content": "记忆内容", "category": "insight|fact|preference|event", "importance": 1-10}
  ]
}

如果没有行动项或记忆点，返回空数组。`

// summarizeTitleSystemPrompt generates a short session title from the first exchange.
const summarizeTitleSystemPrompt = `根据用户的提问和助手的回答，生成一个简短的对话标题。
规则：
- 不超过20个字符
- 概括对话的核心主题
- 不要加引号或标点
- 只返回标题文本，不要任何其他内容`

// ──────────────────────────────────────────────────────────────────────────────
// LangchainLLMService
// ──────────────────────────────────────────────────────────────────────────────

// LangchainLLMService implements LLMProvider and EmbeddingProvider using LangchainGo.
// It reads LLM_PROVIDER, LLM_API_KEY, LLM_BASE_URL, LLM_MODEL from .env via config,
// and provides two distinct agent prompts: one for chat, one for extraction.
type LangchainLLMService struct {
	llm      llms.Model
	embedder embeddings.Embedder
	cfg      *config.LLMConfig
	logger   *logrus.Logger
}

// NewLangchainLLMService creates a new LangchainGo-based LLM service.
// It initialises the LLM client based on LLM_PROVIDER from .env:
//   - "openai" / "zhipu" / "ollama" — all use the OpenAI-compatible client
//     with LLM_API_KEY, LLM_BASE_URL, LLM_MODEL.
//
// LangSmith tracing is configured via a callback handler when enabled.
func NewLangchainLLMService(cfg *config.LLMConfig, langsmithCfg *config.LangSmithConfig, logger *logrus.Logger) (*LangchainLLMService, error) {
	if logger == nil {
		logger = logrus.New()
	}

	// Build LangSmith callback handler (nil if tracing disabled)
	lsHandler := NewLangSmithHandler(langsmithCfg, logger)

	llm, err := initLLM(cfg, lsHandler)
	if err != nil {
		return nil, err
	}

	if lsHandler != nil {
		logger.Infof("LangSmith tracing enabled for project: %s", langsmithCfg.Project)
	}

	// Create embedder using the same OpenAI-compatible client.
	// The LLM client from langchaingo/llms/openai implements EmbedderClient.
	embedder, err := initEmbedder(cfg)
	if err != nil {
		// Embedding is optional — log warning but don't fail startup.
		logger.Warnf("failed to initialise embedder (vector search disabled): %v", err)
	}

	logger.Infof("LLM service initialised: provider=%s model=%s", cfg.Provider, cfg.Model)

	return &LangchainLLMService{
		llm:      llm,
		embedder: embedder,
		cfg:      cfg,
		logger:   logger,
	}, nil
}

// initLLM creates the LangchainGo LLM model based on the provider config.
// All three supported providers (openai, zhipu, ollama) expose OpenAI-compatible APIs,
// so they all use the openai client with different base URLs.
func initLLM(cfg *config.LLMConfig, lsHandler *LangSmithHandler) (llms.Model, error) {
	switch cfg.Provider {
	case "openai", "zhipu", "ollama":
		opts := []openai.Option{
			openai.WithModel(cfg.Model),
		}
		if cfg.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.APIKey))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
		}
		if lsHandler != nil {
			opts = append(opts, openai.WithCallback(lsHandler))
		}
		llm, err := openai.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialise LLM (provider=%s): %w", cfg.Provider, err)
		}
		return llm, nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (supported: openai, zhipu, ollama)", cfg.Provider)
	}
}

// initEmbedder creates a LangchainGo Embedder using the same OpenAI-compatible API.
// The embedding model defaults to the LLM model if no dedicated embedding model is configured.
func initEmbedder(cfg *config.LLMConfig) (embeddings.Embedder, error) {
	switch cfg.Provider {
	case "openai", "zhipu", "ollama":
		opts := []openai.Option{
			openai.WithModel(cfg.Model),
			openai.WithEmbeddingModel(cfg.EmbeddingModel),
		}
		if cfg.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.APIKey))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
		}
		llm, err := openai.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialise embedder (provider=%s): %w", cfg.Provider, err)
		}
		embedder, err := embeddings.NewEmbedder(llm)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedder: %w", err)
		}
		return embedder, nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.Provider)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ChatAgent — web/mobile/desktop interactive chat
// ──────────────────────────────────────────────────────────────────────────────

// ChatAgent implements LLMProvider.ChatAgent.
// It uses the chat agent system prompt and appends any RAG-retrieved context.
func (s *LangchainLLMService) ChatAgent(ctx context.Context, userMessage string, retrievedContext string) (string, error) {
	systemPrompt := chatAgentSystemPrompt
	if retrievedContext != "" {
		systemPrompt += "\n\n## 相关历史对话记录\n" + retrievedContext
	}

	return s.call(ctx, systemPrompt, userMessage)
}

// ChatAgentStream implements LLMProvider.ChatAgentStream.
// It streams token chunks via chunkFn and returns the full accumulated reply.
func (s *LangchainLLMService) ChatAgentStream(ctx context.Context, userMessage string, retrievedContext string, chunkFn func(chunk string)) (string, error) {
	systemPrompt := chatAgentSystemPrompt
	if retrievedContext != "" {
		systemPrompt += "\n\n## 相关历史对话记录\n" + retrievedContext
	}

	return s.callStream(ctx, systemPrompt, userMessage, chunkFn)
}

// ──────────────────────────────────────────────────────────────────────────────
// ExtractAgent — STT transcript → structured data
// ──────────────────────────────────────────────────────────────────────────────

// ExtractAgent implements LLMProvider.ExtractAgent.
// It uses the extraction agent system prompt to parse a raw transcript into structured data.
func (s *LangchainLLMService) ExtractAgent(ctx context.Context, transcript string) (*StructuredResult, error) {
	userPrompt := fmt.Sprintf("以下是语音转写文本，请提取结构化信息：\n\n%s", transcript)

	content, err := s.call(ctx, extractAgentSystemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Strip markdown code fences if present
	content = stripCodeFences(content)

	var result StructuredResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		s.logger.Warnf("failed to parse LLM extraction result, using fallback: %v", err)
		return &StructuredResult{
			Title:   "未命名对话",
			Summary: truncate(transcript, 200),
		}, nil
	}
	return &result, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// SummarizeTitle — generate a short session title
// ──────────────────────────────────────────────────────────────────────────────

// SummarizeTitle implements LLMProvider.SummarizeTitle.
func (s *LangchainLLMService) SummarizeTitle(ctx context.Context, userMessage, assistantReply string) (string, error) {
	userPrompt := fmt.Sprintf("用户: %s\n助手: %s", userMessage, assistantReply)
	title, err := s.call(ctx, summarizeTitleSystemPrompt, userPrompt)
	if err != nil {
		return "", err
	}
	title = strings.TrimSpace(title)
	// Enforce max length
	runes := []rune(title)
	if len(runes) > 20 {
		title = string(runes[:20])
	}
	return title, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// EmbeddingProvider — vector embedding generation
// ──────────────────────────────────────────────────────────────────────────────

// EmbedQuery implements EmbeddingProvider.EmbedQuery.
func (s *LangchainLLMService) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if s.embedder == nil {
		return nil, fmt.Errorf("embedder not initialised")
	}
	return s.embedder.EmbedQuery(ctx, text)
}

// EmbedDocuments implements EmbeddingProvider.EmbedDocuments.
func (s *LangchainLLMService) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if s.embedder == nil {
		return nil, fmt.Errorf("embedder not initialised")
	}
	return s.embedder.EmbedDocuments(ctx, texts)
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────────────────

// call sends a system+user message pair to the LLM and returns the text response.
func (s *LangchainLLMService) call(ctx context.Context, system, user string) (string, error) {
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: system},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: user},
			},
		},
	}

	response, err := s.llm.GenerateContent(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in LLM response")
	}

	return response.Choices[0].Content, nil
}

// callStream sends a system+user message pair to the LLM with streaming enabled.
// chunkFn is called for each token chunk as it arrives.
// Returns the full accumulated response text when the stream completes.
func (s *LangchainLLMService) callStream(ctx context.Context, system, user string, chunkFn func(chunk string)) (string, error) {
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: system},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: user},
			},
		},
	}

	var fullContent strings.Builder

	response, err := s.llm.GenerateContent(ctx, messages,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			text := string(chunk)
			fullContent.WriteString(text)
			if chunkFn != nil {
				chunkFn(text)
			}
			return nil
		}),
	)
	if err != nil {
		// If we got partial content before the error, return it with the error
		if fullContent.Len() > 0 {
			return fullContent.String(), fmt.Errorf("LLM streaming failed: %w", err)
		}
		return "", fmt.Errorf("LLM streaming failed: %w", err)
	}

	// Some providers return the full content in Choices even with streaming;
	// prefer the accumulated stream content if non-empty.
	if fullContent.Len() > 0 {
		return fullContent.String(), nil
	}
	if len(response.Choices) > 0 {
		return response.Choices[0].Content, nil
	}
	return "", fmt.Errorf("no content in LLM streaming response")
}

// stripCodeFences removes markdown code fences (```json ... ```) from LLM output.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) > 2 {
			s = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	return strings.TrimSpace(s)
}

// truncate returns the first n characters of s.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
