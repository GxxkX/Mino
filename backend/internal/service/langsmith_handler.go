package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	langsmith "github.com/langchain-ai/langsmith-go"
	"github.com/langchain-ai/langsmith-go/option"
	"github.com/mino/backend/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// LangSmithHandler implements callbacks.Handler and sends run data to LangSmith.
type LangSmithHandler struct {
	callbacks.SimpleHandler
	client  *langsmith.Client
	project string
	logger  *logrus.Logger

	mu        sync.Mutex
	runID     string
	traceID   string
	startTime time.Time
	inputs    map[string]interface{}
}

var _ callbacks.Handler = (*LangSmithHandler)(nil)

// NewLangSmithHandler creates a handler that traces LLM calls to LangSmith.
func NewLangSmithHandler(cfg *config.LangSmithConfig, logger *logrus.Logger) *LangSmithHandler {
	if cfg == nil || !cfg.Tracing || cfg.APIKey == "" {
		return nil
	}

	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}
	if cfg.Endpoint != "" {
		opts = append(opts, option.WithBaseURL(cfg.Endpoint))
	}

	client := langsmith.NewClient(opts...)

	return &LangSmithHandler{
		client:  client,
		project: cfg.Project,
		logger:  logger,
	}
}

func (h *LangSmithHandler) HandleLLMGenerateContentStart(_ context.Context, ms []llms.MessageContent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.runID = generateUUID()
	h.traceID = h.runID
	h.startTime = time.Now().UTC()

	var parts []map[string]string
	for _, m := range ms {
		var text strings.Builder
		for _, p := range m.Parts {
			if tc, ok := p.(llms.TextContent); ok {
				text.WriteString(tc.Text)
			}
		}
		parts = append(parts, map[string]string{
			"role":    string(m.Role),
			"content": text.String(),
		})
	}
	h.inputs = map[string]interface{}{"messages": parts}
}

func (h *LangSmithHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	h.mu.Lock()
	runID := h.runID
	traceID := h.traceID
	startTime := h.startTime
	inputs := h.inputs
	h.runID = ""
	h.mu.Unlock()

	if runID == "" {
		return
	}

	endTime := time.Now().UTC()

	outputs := map[string]interface{}{}
	if len(res.Choices) > 0 {
		outputs["content"] = res.Choices[0].Content
		if res.Choices[0].StopReason != "" {
			outputs["stop_reason"] = res.Choices[0].StopReason
		}
	}

	run := langsmith.RunParam{
		ID:           langsmith.F(runID),
		TraceID:      langsmith.F(traceID),
		DottedOrder:  langsmith.F(formatDottedOrder(startTime, runID)),
		Name:         langsmith.F("ChatModel"),
		RunType:      langsmith.F(langsmith.RunRunTypeLlm),
		StartTime:    langsmith.F(startTime.Format(time.RFC3339Nano)),
		EndTime:      langsmith.F(endTime.Format(time.RFC3339Nano)),
		Inputs:       langsmith.F(inputs),
		Outputs:      langsmith.F(outputs),
		SessionName:  langsmith.F(h.project),
	}

	go func() {
		_, err := h.client.Runs.IngestBatch(context.Background(), langsmith.RunIngestBatchParams{
			Post: langsmith.F([]langsmith.RunParam{run}),
		})
		if err != nil {
			h.logger.Warnf("langsmith: failed to send run: %v", err)
		}
	}()
}

func (h *LangSmithHandler) HandleLLMError(ctx context.Context, llmErr error) {
	h.mu.Lock()
	runID := h.runID
	traceID := h.traceID
	startTime := h.startTime
	inputs := h.inputs
	h.runID = ""
	h.mu.Unlock()

	if runID == "" {
		return
	}

	endTime := time.Now().UTC()

	run := langsmith.RunParam{
		ID:          langsmith.F(runID),
		TraceID:     langsmith.F(traceID),
		DottedOrder: langsmith.F(formatDottedOrder(startTime, runID)),
		Name:        langsmith.F("ChatModel"),
		RunType:     langsmith.F(langsmith.RunRunTypeLlm),
		StartTime:   langsmith.F(startTime.Format(time.RFC3339Nano)),
		EndTime:     langsmith.F(endTime.Format(time.RFC3339Nano)),
		Inputs:      langsmith.F(inputs),
		Error:       langsmith.F(llmErr.Error()),
		SessionName: langsmith.F(h.project),
	}

	go func() {
		_, err := h.client.Runs.IngestBatch(context.Background(), langsmith.RunIngestBatchParams{
			Post: langsmith.F([]langsmith.RunParam{run}),
		})
		if err != nil {
			h.logger.Warnf("langsmith: failed to send error run: %v", err)
		}
	}()
}

func (h *LangSmithHandler) HandleAgentAction(_ context.Context, _ schema.AgentAction) {}
func (h *LangSmithHandler) HandleAgentFinish(_ context.Context, _ schema.AgentFinish) {}

// formatDottedOrder formats a root run's dotted order: {timestamp}Z{run_id}
// Format: YYYYMMDDTHHMMSSffffffZ{uuid}
func formatDottedOrder(t time.Time, runID string) string {
	ts := t.UTC().Format("20060102T150405") + fmt.Sprintf("%06d", t.Nanosecond()/1000)
	return ts + "Z" + runID
}

// generateUUID creates a UUID v4.
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
