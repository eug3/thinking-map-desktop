// API 适配层 - 将 IPC 调用适配为前端期望的接口
import { maps, nodes, thinking, config } from '@/wails-runtime';
import type {
  CreateMapRequest,
  Map,
  MapList,
  MapListQuery,
  UpdateMapRequest,
  NodeResponse,
  CreateNodeRequest,
  UpdateNodeRequest,
  ExecutableNodes,
} from '@/types';

// Map API
export async function createMap(data: CreateMapRequest): Promise<Map> {
  const result = await maps.createMap({
    title: data.title,
    problem: data.problem,
    problemType: data.problemType || '',
    target: data.target || '',
    keyPoints: data.keyPoints || [],
    constraints: data.constraints || [],
  });
  // IPC返回的MapResponse类型与Map兼容，直接返回
  return result as unknown as Map;
}

export async function fetchMapList(params: MapListQuery): Promise<MapList> {
  const result = await maps.listMaps(params);
  // 转换MapResponse[]到Map[]
  return {
    total: result.total,
    page: result.page,
    limit: result.limit,
    items: result.items as unknown as Map[],
  };
}

export async function getMap(mapId: string): Promise<Map> {
  const result = await maps.getMap(mapId);
  return result as unknown as Map;
}

export async function updateMap(mapId: string, data: UpdateMapRequest): Promise<Map> {
  const result = await maps.updateMap(mapId, data);
  return result as unknown as Map;
}

export async function deleteMap(mapId: string): Promise<void> {
  return await maps.deleteMap(mapId);
}

// Node API
export async function getMapNodes(mapId: string): Promise<NodeResponse[]> {
  const result = await nodes.listNodes(mapId);
  // IPC返回的NodeResponse与types兼容
  return result as unknown as NodeResponse[];
}

export async function createNode(data: CreateNodeRequest): Promise<NodeResponse> {
  const result = await nodes.createNode(data.mapID, {
    parentId: data.parentID,
    nodeType: data.nodeType,
    question: data.question,
    target: data.target,
    position: data.position,
  });
  return result as unknown as NodeResponse;
}

export async function updateNode(nodeId: string, data: UpdateNodeRequest): Promise<NodeResponse> {
  const result = await nodes.updateNode(nodeId, data);
  return result as unknown as NodeResponse;
}

export async function deleteNode(nodeId: string): Promise<void> {
  return await nodes.deleteNode(nodeId);
}

export async function executableNodes(mapId: string, nodeId?: string): Promise<ExecutableNodes> {
  const result = await nodes.executableNodes(mapId, nodeId || '');
  return {
    nodeIDs: result.nodeIds,
    suggestedNodeID: result.suggestedNodeId,
  };
}

export async function resetNodeContext(_nodeId: string): Promise<NodeResponse> {
  // TODO: Implement via IPC
  return {} as NodeResponse;
}

// Thinking API
export async function understanding(nodeId: string, query: string): Promise<void> {
  return await thinking.understanding({ nodeId, query });
}

export async function decomposition(nodeId: string, clarification: string, isDecomposed: boolean): Promise<void> {
  return await thinking.decomposition({ nodeId, clarification, isDecomposed });
}

export async function conclusion(nodeId: string, reference: string, instruction: string): Promise<void> {
  return await thinking.conclusion({ nodeId, reference, instruction });
}

export async function repeat(nodeId: string): Promise<void> {
  return await thinking.repeat(nodeId);
}

export async function saveNodeConclusion(nodeId: string, content: string): Promise<void> {
  // Use updateNode to persist the conclusion content
  await nodes.updateNode(nodeId, { conclusion: content } as any);
}

export async function resetNodeConclusion(_nodeId: string): Promise<void> {
  // TODO: Implement via IPC
  return;
}

// Config API
export interface LLMConfigData {
  base_url: string;
  api_key: string;
  model: string;
}

export async function getLLMConfig(): Promise<LLMConfigData> {
  return await config.getLLMConfig();
}

export async function saveLLMConfig(data: LLMConfigData): Promise<void> {
  return await config.saveLLMConfig(data);
}

export async function testLLMConfig(data: LLMConfigData): Promise<string> {
  return await config.testLLMConfig(data);
}
