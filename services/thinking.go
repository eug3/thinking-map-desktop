/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/services/thinking.go
 */
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/PGshen/thinking-map-desktop/events"
	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/PGshen/thinking-map-desktop/repositories"
)

// ThinkingService handles AI thinking operations
type ThinkingService struct {
	nodeRepo    repositories.ThinkingNodeRepository
	mapRepo     repositories.ThinkingMapRepository
	messageRepo repositories.MessageRepository
	events      *events.EventManager
	llm         *LLMClient
}

// NewThinkingService creates a new thinking service
func NewThinkingService(
	nodeRepo repositories.ThinkingNodeRepository,
	mapRepo repositories.ThinkingMapRepository,
	messageRepo repositories.MessageRepository,
	events *events.EventManager,
	llm *LLMClient,
) *ThinkingService {
	return &ThinkingService{
		nodeRepo:    nodeRepo,
		mapRepo:     mapRepo,
		messageRepo: messageRepo,
		events:      events,
		llm:         llm,
	}
}

// StreamCallback is called for each chunk of streaming response
type StreamCallback func(chunk string, eventType string)

// Understanding initiates the understanding process
func (s *ThinkingService) Understanding(ctx context.Context, nodeID, query string, callback StreamCallback) error {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return err
	}

	// Update node status
	node.Status = "running"
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return err
	}

	// Emit event
	s.events.EmitNodeEvent(events.EventNodeUpdated, nodeID, node.MapID, map[string]interface{}{
		"status": "running",
	})

	// Stream AI response
	go func() {
		defer func() {
			node.Status = "completed"
			s.nodeRepo.Update(context.Background(), node)
			s.events.EmitNodeEvent(events.EventNodeUpdated, nodeID, node.MapID, map[string]interface{}{
				"status": "completed",
			})
		}()

		if !s.llm.IsConfigured() {
			callback("⚠️ AI 未配置。请点击右上角设置按钮，填入 API 配置。", events.EventMessageText)
			callback("", events.EventConclusionCompleted)
			return
		}

		// Get map info for context
		mapModel, _ := s.mapRepo.FindByID(ctx, node.MapID)
		systemPrompt := buildUnderstandingPrompt(mapModel, node)

		messages := []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: query},
		}

		var fullResponse strings.Builder
		err := s.llm.StreamChat(messages, func(chunk string) error {
			fullResponse.WriteString(chunk)
			callback(chunk, events.EventMessageText)
			return nil
		})

		if err != nil {
			slog.Error("LLM stream error", "error", err)
			callback("\n\n❌ AI 调用失败: "+err.Error(), events.EventMessageText)
		}

		callback("", events.EventConclusionCompleted)
	}()

	return nil
}

// Decomposition initiates the decomposition process
func (s *ThinkingService) Decomposition(ctx context.Context, nodeID, clarification string, isDecomposed bool, callback StreamCallback) error {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return err
	}

	// Update node status
	if isDecomposed {
		node.Status = "in_decomposition"
	} else {
		node.Status = "running"
	}
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return err
	}

	s.events.EmitNodeEvent(events.EventNodeUpdated, nodeID, node.MapID, map[string]interface{}{
		"status": node.Status,
	})

	// Stream AI decomposition
	go func() {
		defer func() {
			node.Status = "in_decomposition"
			node.Decomposition.IsDecomposed = isDecomposed
			s.nodeRepo.Update(context.Background(), node)
			s.events.EmitNodeEvent(events.EventDecompositionCompleted, nodeID, node.MapID, map[string]interface{}{
				"status": "completed",
			})
		}()

		if !s.llm.IsConfigured() {
			callback("⚠️ AI 未配置。请点击右上角设置按钮，填入 API 配置。", events.EventMessageThought)
			return
		}

		mapModel, _ := s.mapRepo.FindByID(ctx, node.MapID)
		systemPrompt := buildDecompositionPrompt(mapModel, node)

		userMsg := "请分析并拆解这个问题。"
		if clarification != "" {
			userMsg = clarification
		}

		messages := []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMsg},
		}

		err := s.llm.StreamChat(messages, func(chunk string) error {
			callback(chunk, events.EventMessageThought)
			return nil
		})

		if err != nil {
			slog.Error("LLM decomposition error", "error", err)
			callback("\n\n❌ AI 调用失败: "+err.Error(), events.EventMessageThought)
		}
	}()

	return nil
}

// Conclusion initiates the conclusion process
func (s *ThinkingService) Conclusion(ctx context.Context, nodeID, reference, instruction string, callback StreamCallback) error {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return err
	}

	// Update node status
	node.Status = "in_conclusion"
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return err
	}

	s.events.EmitNodeEvent(events.EventNodeUpdated, nodeID, node.MapID, map[string]interface{}{
		"status": "in_conclusion",
	})

	// Stream AI conclusion
	go func() {
		defer func() {
			node.Status = "completed"
			s.nodeRepo.Update(context.Background(), node)
			s.events.EmitNodeEvent(events.EventConclusionCompleted, nodeID, node.MapID, map[string]interface{}{
				"status": "completed",
			})
		}()

		if !s.llm.IsConfigured() {
			callback("⚠️ AI 未配置。请点击右上角设置按钮，填入 API 配置。", events.EventMessageConclusion)
			return
		}

		mapModel, _ := s.mapRepo.FindByID(ctx, node.MapID)
		systemPrompt := buildConclusionPrompt(mapModel, node)

		userMsg := "请基于分析生成结论。"
		if reference != "" {
			userMsg = "参考信息: " + reference
		}
		if instruction != "" {
			userMsg += "\n要求: " + instruction
		}

		messages := []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMsg},
		}

		var fullResponse strings.Builder
		err := s.llm.StreamChat(messages, func(chunk string) error {
			fullResponse.WriteString(chunk)
			callback(chunk, events.EventMessageConclusion)
			return nil
		})

		if err != nil {
			slog.Error("LLM conclusion error", "error", err)
			callback("\n\n❌ AI 调用失败: "+err.Error(), events.EventMessageConclusion)
		} else {
			node.Conclusion.Content = fullResponse.String()
		}
	}()

	return nil
}

// Repeat re-runs the thinking process
func (s *ThinkingService) Repeat(ctx context.Context, nodeID string, callback StreamCallback) error {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return err
	}

	// Reset node status
	node.Status = "pending"
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return err
	}

	s.events.EmitNodeEvent(events.EventNodeUpdated, nodeID, node.MapID, map[string]interface{}{
		"status": "pending",
	})

	return nil
}

// buildUnderstandingPrompt creates the system prompt for understanding phase
func buildUnderstandingPrompt(mapModel *models.ThinkingMap, node *models.ThinkingNode) string {
	problem := ""
	if mapModel != nil {
		problem = mapModel.Problem
	}

	return fmt.Sprintf(`你是一个思维分析助手。你的任务是帮助用户深入理解问题。

问题背景: %s

请对用户的问题进行深入分析：
1. 理解问题的核心要素
2. 识别关键概念和术语
3. 分析问题的范围和边界
4. 提出可能的分析方向
5. 指出需要注意的要点

请用清晰、结构化的方式回答。`, problem)
}

// buildDecompositionPrompt creates the system prompt for decomposition phase
func buildDecompositionPrompt(mapModel *models.ThinkingMap, node *models.ThinkingNode) string {
	problem := ""
	if mapModel != nil {
		problem = mapModel.Problem
	}

	return fmt.Sprintf(`你是一个问题拆解专家。你的任务是将复杂问题拆解为子问题。

原始问题: %s

请分析并拆解这个问题：
1. 将问题分解为可独立解决的子问题
2. 说明每个子问题的关系（顺序依赖/并行/层次）
3. 为每个子问题提供简要分析方向
4. 建议解决顺序

请用结构化的方式呈现拆解结果。`, problem)
}

// buildConclusionPrompt creates the system prompt for conclusion phase
func buildConclusionPrompt(mapModel *models.ThinkingMap, node *models.ThinkingNode) string {
	problem := ""
	if mapModel != nil {
		problem = mapModel.Problem
	}

	return fmt.Sprintf(`你是一个总结分析专家。你的任务是基于已有分析生成结论。

原始问题: %s

请生成综合结论：
1. 总结关键发现
2. 提出解决方案或建议
3. 指出潜在风险和注意事项
4. 给出下一步行动建议

请用清晰、可执行的方式表达结论。`, problem)
}
