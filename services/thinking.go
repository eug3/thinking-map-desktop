/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/services/thinking.go
 */
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/PGshen/thinking-map-desktop/events"
	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/PGshen/thinking-map-desktop/repositories"
	"github.com/google/uuid"
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

// SubQuestion represents a sub-question extracted from AI decomposition
type SubQuestion struct {
	Question string `json:"question"`
	Target   string `json:"target"`
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

		var fullResponse strings.Builder
		err := s.llm.StreamChat(messages, func(chunk string) error {
			fullResponse.WriteString(chunk)
			callback(chunk, events.EventMessageThought)
			return nil
		})

		if err != nil {
			slog.Error("LLM decomposition error", "error", err)
			callback("\n\n❌ AI 调用失败: "+err.Error(), events.EventMessageThought)
			return
		}

		// When isDecomposed=true, extract sub-questions and create child nodes
		if isDecomposed {
			s.createChildNodesFromDecomposition(ctx, node, mapModel, fullResponse.String(), callback)
		}
	}()

	return nil
}

// createChildNodesFromDecomposition extracts sub-questions from analysis and creates child nodes
func (s *ThinkingService) createChildNodesFromDecomposition(ctx context.Context, parentNode *models.ThinkingNode, mapModel *models.ThinkingMap, analysisText string, callback StreamCallback) {
	callback("\n\n---\n\n🔄 正在提取子问题并创建节点...\n", events.EventMessageThought)

	// Build extraction prompt
	extractPrompt := buildSubQuestionExtractionPrompt(parentNode, analysisText)

	messages := []ChatMessage{
		{Role: "system", Content: extractPrompt},
		{Role: "user", Content: "请从上述分析中提取子问题，以JSON格式输出。"},
	}

	response, err := s.llm.Chat(messages)
	if err != nil {
		slog.Error("Failed to extract sub-questions", "error", err)
		callback("❌ 提取子问题失败: "+err.Error(), events.EventMessageThought)
		return
	}

	// Parse JSON from response
	subQuestions, err := parseSubQuestions(response)
	if err != nil {
		slog.Error("Failed to parse sub-questions", "error", err, "response", response)
		callback("❌ 解析子问题失败: "+err.Error(), events.EventMessageThought)
		return
	}

	if len(subQuestions) == 0 {
		callback("⚠️ 未提取到子问题。", events.EventMessageThought)
		return
	}

	// Create child nodes
	for i, sq := range subQuestions {
		childNode := &models.ThinkingNode{
			ID:       uuid.NewString(),
			MapID:    parentNode.MapID,
			ParentID: parentNode.ID,
			NodeType: "sub_question",
			Question: sq.Question,
			Target:   sq.Target,
			Status:   "pending",
			Position: models.Position{
				X: parentNode.Position.X + float64(i)*250 - float64(len(subQuestions)-1)*125,
				Y: parentNode.Position.Y + 200,
			},
			Decomposition: models.Decomposition{
				IsDecomposed: false,
			},
			Conclusion: models.Conclusion{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.nodeRepo.Create(context.Background(), childNode); err != nil {
			slog.Error("Failed to create child node", "error", err)
			callback(fmt.Sprintf("❌ 创建子节点失败: %s\n", err.Error()), events.EventMessageThought)
			continue
		}

		// Emit nodeCreated event with full node data
		s.events.Emit(events.EventNodeCreated, map[string]interface{}{
			"id":       childNode.ID,
			"mapID":    childNode.MapID,
			"parentID": childNode.ParentID,
			"nodeType": childNode.NodeType,
			"question": childNode.Question,
			"target":   childNode.Target,
			"status":   childNode.Status,
			"position": map[string]interface{}{
				"x": childNode.Position.X,
				"y": childNode.Position.Y,
			},
		})

		callback(fmt.Sprintf("✅ 子节点已创建: %s\n", sq.Question), events.EventMessageThought)
	}

	callback(fmt.Sprintf("\n🎉 共创建了 %d 个子节点。", len(subQuestions)), events.EventMessageThought)
}

// buildSubQuestionExtractionPrompt creates a prompt to extract structured sub-questions
func buildSubQuestionExtractionPrompt(node *models.ThinkingNode, analysisText string) string {
	return fmt.Sprintf(`你是一个问题拆解助手。根据以下分析内容，提取出所有子问题。

原始问题: %s

分析内容:
%s

请严格以以下JSON格式输出子问题列表，不要输出其他任何内容：
{"subQuestions": [{"question": "子问题1的描述", "target": "子问题1的目标"}, {"question": "子问题2的描述", "target": "子问题2的目标"}]}

要求：
1. 每个子问题应该是可独立分析和解决的
2. question 字段描述子问题本身
3. target 字段描述该子问题的分析目标或预期产出
4. 只输出纯JSON，不要添加markdown代码块标记或其他文字`, node.Question, analysisText)
}

// parseSubQuestions parses JSON response to extract sub-questions
func parseSubQuestions(response string) ([]SubQuestion, error) {
	// Try to find JSON in the response (may be wrapped in markdown code block)
	jsonStr := response

	// Remove markdown code block if present
	if idx := strings.Index(jsonStr, "```json"); idx != -1 {
		jsonStr = jsonStr[idx+7:]
		if endIdx := strings.Index(jsonStr, "```"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx]
		}
	} else if idx := strings.Index(jsonStr, "```"); idx != -1 {
		jsonStr = jsonStr[idx+3:]
		if endIdx := strings.Index(jsonStr, "```"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)

	// Try to find JSON object boundaries
	startIdx := strings.Index(jsonStr, "{")
	if startIdx == -1 {
		// Try parsing as array directly
		startIdx = strings.Index(jsonStr, "[")
		if startIdx != -1 {
			endIdx := strings.LastIndex(jsonStr, "]")
			if endIdx != -1 {
				jsonStr = jsonStr[startIdx : endIdx+1]
				var questions []SubQuestion
				if err := json.Unmarshal([]byte(jsonStr), &questions); err != nil {
					return nil, fmt.Errorf("failed to parse JSON array: %w", err)
				}
				return questions, nil
			}
		}
		return nil, fmt.Errorf("no JSON found in response")
	}

	endIdx := strings.LastIndex(jsonStr, "}")
	if endIdx == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}
	jsonStr = jsonStr[startIdx : endIdx+1]

	// Parse as object with subQuestions field
	var result struct {
		SubQuestions []SubQuestion `json:"subQuestions"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Try parsing as array
		var questions []SubQuestion
		if err2 := json.Unmarshal([]byte(jsonStr), &questions); err2 != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w (original: %w)", err2, err)
		}
		return questions, nil
	}

	return result.SubQuestions, nil
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
