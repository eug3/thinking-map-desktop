/**
 * Wails v3 IPC Client Wrapper
 * Uses @wailsio/runtime Call.ByName to invoke Go bindings
 */
import { Call } from '@wailsio/runtime';

// Package path for the Go bindings
const PKG = 'github.com/PGshen/thinking-map-desktop/bindings';

// Type definitions for the Go bindings
export interface CreateMapRequest {
  title: string;
  problem: string;
  problemType: string;
  target: string;
  keyPoints: string[];
  constraints: string[];
}

export interface UpdateMapRequest {
  status?: number;
  problem?: string;
  problemType?: string;
  target?: string;
  keyPoints?: string[];
  constraints?: string[];
  conclusion?: string;
}

export interface MapResponse {
  id: string;
  title: string;
  problem: string;
  problemType: string;
  target: string;
  keyPoints: string[];
  constraints: string[];
  conclusion: string;
  status: string;
  createdAt: number;
  updatedAt: number;
}

export interface MapListQuery {
  page: number;
  limit: number;
  status?: number;
  problemType?: string;
  search?: string;
  dateRange?: string;
}

export interface MapListResponse {
  total: number;
  page: number;
  limit: number;
  items: MapResponse[];
}

export interface CreateNodeRequest {
  parentId: string;
  nodeType: string;
  question: string;
  target: string;
  position: { x: number; y: number };
}

export interface UpdateNodeRequest {
  question?: string;
  target?: string;
  position?: { x: number; y: number };
  status?: string;
}

export interface NodeResponse {
  id: string;
  mapId: string;
  parentId: string;
  nodeType: string;
  question: string;
  target: string;
  context: any;
  decomposition: any;
  conclusion: any;
  status: string;
  position: { x: number; y: number };
  dependencies: string[];
  createdAt: number;
  updatedAt: number;
}

export interface ExecutableNodesResponse {
  nodeIds: string[];
  suggestedNodeId: string;
}

// Map binding - calls Go MapBinding methods via Wails v3 IPC
export class MapBinding {
  async createMap(data: CreateMapRequest): Promise<MapResponse> {
    return Call.ByName(`${PKG}.MapBinding.CreateMap`, data);
  }

  async listMaps(params: MapListQuery): Promise<MapListResponse> {
    return Call.ByName(`${PKG}.MapBinding.ListMaps`, params);
  }

  async getMap(mapId: string): Promise<MapResponse> {
    return Call.ByName(`${PKG}.MapBinding.GetMap`, mapId);
  }

  async updateMap(mapId: string, data: UpdateMapRequest): Promise<MapResponse> {
    return Call.ByName(`${PKG}.MapBinding.UpdateMap`, mapId, data);
  }

  async deleteMap(mapId: string): Promise<void> {
    return Call.ByName(`${PKG}.MapBinding.DeleteMap`, mapId);
  }
}

// Node binding - calls Go NodeBinding methods via Wails v3 IPC
export class NodeBinding {
  async listNodes(mapId: string): Promise<NodeResponse[]> {
    return Call.ByName(`${PKG}.NodeBinding.ListNodes`, mapId);
  }

  async createNode(mapId: string, data: CreateNodeRequest): Promise<NodeResponse> {
    return Call.ByName(`${PKG}.NodeBinding.CreateNode`, mapId, data);
  }

  async updateNode(nodeId: string, data: UpdateNodeRequest): Promise<NodeResponse> {
    return Call.ByName(`${PKG}.NodeBinding.UpdateNode`, nodeId, data);
  }

  async deleteNode(nodeId: string): Promise<void> {
    return Call.ByName(`${PKG}.NodeBinding.DeleteNode`, nodeId);
  }

  async executableNodes(mapId: string, nodeId?: string): Promise<ExecutableNodesResponse> {
    return Call.ByName(`${PKG}.NodeBinding.ExecutableNodes`, mapId, nodeId || '');
  }
}

// Thinking binding - calls Go ThinkingBinding methods via Wails v3 IPC
export class ThinkingBinding {
  async understanding(data: { nodeId: string; query: string }): Promise<void> {
    return Call.ByName(`${PKG}.ThinkingBinding.Understanding`, data);
  }

  async decomposition(data: { nodeId: string; clarification: string; isDecomposed: boolean }): Promise<void> {
    return Call.ByName(`${PKG}.ThinkingBinding.Decomposition`, data);
  }

  async conclusion(data: { nodeId: string; reference: string; instruction: string }): Promise<void> {
    return Call.ByName(`${PKG}.ThinkingBinding.Conclusion`, data);
  }

  async repeat(nodeId: string): Promise<void> {
    return Call.ByName(`${PKG}.ThinkingBinding.Repeat`, nodeId);
  }
}

// Config binding - calls Go ConfigBinding methods via Wails v3 IPC
export class ConfigBinding {
  async getLLMConfig(): Promise<{ base_url: string; api_key: string; model: string }> {
    return Call.ByName(`${PKG}.ConfigBinding.GetLLMConfig`);
  }

  async saveLLMConfig(data: { base_url: string; api_key: string; model: string }): Promise<void> {
    return Call.ByName(`${PKG}.ConfigBinding.SaveLLMConfig`, data);
  }

  async testLLMConfig(data: { base_url: string; api_key: string; model: string }): Promise<string> {
    return Call.ByName(`${PKG}.ConfigBinding.TestLLMConfig`, data);
  }
}

// Export singleton instances
export const maps = new MapBinding();
export const nodes = new NodeBinding();
export const thinking = new ThinkingBinding();
export const config = new ConfigBinding();

// Export everything as a default object for convenience
export default {
  maps,
  nodes,
  thinking,
  config,
};
