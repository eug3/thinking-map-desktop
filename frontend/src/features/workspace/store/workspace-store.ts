/**
 * Workspace Store - Zustand state management for workspace
 * Adapted from web version for Wails desktop
 */
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { Node, Edge } from 'reactflow';
import type { CustomNodeModel } from '@/types/node';
import type { Map as ThinkingMap } from '@/types/map';

// Layout config
export interface LayoutConfig {
  direction: 'TB' | 'LR';
  nodeWidth: number;
  nodeHeight: number;
  rankSep: number;
  nodeSep: number;
}

// Animation config
export interface AnimationConfig {
  duration: number;
  easing: string;
}

// Workspace settings
export interface WorkspaceSettings {
  autoSave: boolean;
  showMinimap: boolean;
  showControls: boolean;
  layoutType: 'global' | 'local';
  layoutConfig: LayoutConfig;
  animationConfig: AnimationConfig;
}

const DEFAULT_SETTINGS: WorkspaceSettings = {
  autoSave: true,
  showMinimap: true,
  showControls: true,
  layoutType: 'global',
  layoutConfig: {
    direction: 'TB',
    nodeWidth: 200,
    nodeHeight: 100,
    rankSep: 120,
    nodeSep: 50,
  },
  animationConfig: {
    duration: 500,
    easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
  },
};

// Store state
interface WorkspaceState {
  // Map info
  mapID: string | null;
  mapInfo: ThinkingMap | null;

  // Visual state
  nodes: Node<CustomNodeModel>[];
  edges: Edge[];
  isLoading: boolean;

  // Panel state
  panelOpen: boolean;
  panelWidth: number;
  activeNodeID: string | null;

  // Selection
  selectedNodeIDs: string[];

  // Position tracking
  changedNodePositions: string[];

  // Executable nodes
  executableNodeIDs: string[];
  suggestedNodeID: string | null;

  // Settings
  settings: WorkspaceSettings;

  // Actions
  actions: WorkspaceActions;
}

interface WorkspaceActions {
  // Map
  setMapID: (id: string) => void;
  updateMap: (info: Partial<ThinkingMap>) => void;

  // Loading
  setLoading: (loading: boolean) => void;

  // Nodes
  setNodes: (nodes: Node<CustomNodeModel>[]) => void;
  addNode: (node: Node<CustomNodeModel>) => void;
  addChildNode: (parentID: string, node: Node<CustomNodeModel>) => void;
  updateNode: (nodeID: string, updates: Partial<Node<CustomNodeModel>>) => void;
  updateNodeData: (nodeID: string, updates: Partial<CustomNodeModel>) => void;
  deleteNode: (nodeID: string) => void;
  updateNodeID: (oldID: string, newID: string) => void;
  setEditing: (nodeID: string) => void;

  // Edges
  setEdges: (edges: Edge[]) => void;
  addEdge: (edge: Edge) => void;
  updateNodeDependencies: (nodeID: string, dependencies: string[]) => void;

  // Selection
  selectNode: (nodeID: string) => void;
  clearSelection: () => void;

  // Panel
  openPanel: (nodeID: string) => void;
  closePanel: () => void;
  setPanelWidth: (width: number) => void;

  // Position tracking
  addChangedNodePosition: (nodeID: string) => void;
  clearChangedNodePositions: () => void;

  // Executable nodes
  setExecutableNodes: (nodeIDs: string[], suggestedNodeID: string) => void;

  // Decomposition & conclusion
  updateNodeDecomposition: (nodeID: string, data: any) => void;
  updateNodeConclusion: (nodeID: string, data: any) => void;

  // Settings
  updateSettings: (settings: Partial<WorkspaceSettings>) => void;

  // Reset
  reset: () => void;
}

const initialState = {
  mapID: null as string | null,
  mapInfo: null as ThinkingMap | null,
  nodes: [] as Node<CustomNodeModel>[],
  edges: [] as Edge[],
  isLoading: false,
  panelOpen: false,
  panelWidth: 400,
  activeNodeID: null as string | null,
  selectedNodeIDs: [] as string[],
  changedNodePositions: [] as string[],
  executableNodeIDs: [] as string[],
  suggestedNodeID: null as string | null,
  settings: DEFAULT_SETTINGS,
};

export const useWorkspaceStore = create<WorkspaceState>()(
  devtools(
    (set, _get) => ({
      ...initialState,

      actions: {
        setMapID: (id: string) => set({ mapID: id }),

        updateMap: (info: Partial<ThinkingMap>) =>
          set((state) => ({
            mapInfo: state.mapInfo ? { ...state.mapInfo, ...info } : (info as ThinkingMap),
          })),

        setLoading: (loading: boolean) => set({ isLoading: loading }),

        setNodes: (nodes: Node<CustomNodeModel>[]) => set({ nodes }),

        addNode: (node: Node<CustomNodeModel>) =>
          set((state) => ({ nodes: [...state.nodes, node] })),

        addChildNode: (parentID: string, node: Node<CustomNodeModel>) =>
          set((state) => {
            const parentNode = state.nodes.find((n) => n.id === parentID);
            if (!parentNode) return state;

            // Position child below parent
            const childNode = {
              ...node,
              position: {
                x: parentNode.position.x,
                y: parentNode.position.y + 200,
              },
            };

            const newEdge: Edge = {
              id: `${parentID}-${node.id}`,
              source: parentID,
              target: node.id,
              type: 'default',
              style: { stroke: '#3b82f6' },
            };

            return {
              nodes: [...state.nodes, childNode],
              edges: [...state.edges, newEdge],
            };
          }),

        updateNode: (nodeID: string, updates: Partial<Node<CustomNodeModel>>) =>
          set((state) => ({
            nodes: state.nodes.map((node) => {
              if (node.id !== nodeID) return node;
              if (updates.data) {
                return {
                  ...node,
                  ...updates,
                  data: { ...node.data, ...updates.data },
                };
              }
              return { ...node, ...updates };
            }),
          })),

        updateNodeData: (nodeID: string, updates: Partial<CustomNodeModel>) =>
          set((state) => ({
            nodes: state.nodes.map((node) =>
              node.id === nodeID
                ? { ...node, data: { ...node.data, ...updates } }
                : node
            ),
          })),

        deleteNode: (nodeID: string) =>
          set((state) => ({
            nodes: state.nodes.filter((n) => n.id !== nodeID),
            edges: state.edges.filter(
              (e) => e.source !== nodeID && e.target !== nodeID
            ),
          })),

        updateNodeID: (oldID: string, newID: string) =>
          set((state) => ({
            nodes: state.nodes.map((node) =>
              node.id === oldID
                ? { ...node, id: newID, data: { ...node.data, id: newID } }
                : node
            ),
            edges: state.edges.map((edge) => ({
              ...edge,
              id: edge.id.replace(oldID, newID),
              source: edge.source === oldID ? newID : edge.source,
              target: edge.target === oldID ? newID : edge.target,
            })),
          })),

        setEditing: (nodeID: string) =>
          set((state) => ({
            nodes: state.nodes.map((node) =>
              node.id === nodeID
                ? { ...node, data: { ...node.data, isEditing: true } }
                : node
            ),
          })),

        setEdges: (edges: Edge[]) => set({ edges }),

        addEdge: (edge: Edge) =>
          set((state) => ({ edges: [...state.edges, edge] })),

        updateNodeDependencies: (nodeID: string, dependencies: string[]) =>
          set((state) => {
            // Remove old dependency edges for this node
            const filteredEdges = state.edges.filter(
              (e) => !(e.target === nodeID && e.type === 'dependency')
            );

            // Add new dependency edges
            const newEdges = dependencies.map((depID) => ({
              id: `dep-${nodeID}-${depID}`,
              source: depID,
              target: nodeID,
              type: 'dependency',
              style: { strokeDasharray: '5,5', stroke: '#8b5cf6' },
              animated: true,
              sourceHandle: 'dependency-source',
              targetHandle: 'dependency-target',
            }));

            // Update node dependencies data
            const nodes = state.nodes.map((node) =>
              node.id === nodeID
                ? { ...node, data: { ...node.data, dependencies } }
                : node
            );

            return {
              nodes,
              edges: [...filteredEdges, ...newEdges],
            };
          }),

        selectNode: (nodeID: string) =>
          set({
            selectedNodeIDs: [nodeID],
          }),

        clearSelection: () => set({ selectedNodeIDs: [] }),

        openPanel: (nodeID: string) =>
          set({
            panelOpen: true,
            activeNodeID: nodeID,
            selectedNodeIDs: [nodeID],
          }),

        closePanel: () =>
          set({
            panelOpen: false,
            activeNodeID: null,
          }),

        setPanelWidth: (width: number) => set({ panelWidth: width }),

        addChangedNodePosition: (nodeID: string) =>
          set((state) => ({
            changedNodePositions: state.changedNodePositions.includes(nodeID)
              ? state.changedNodePositions
              : [...state.changedNodePositions, nodeID],
          })),

        clearChangedNodePositions: () => set({ changedNodePositions: [] }),

        setExecutableNodes: (nodeIDs: string[], suggestedNodeID: string) =>
          set({ executableNodeIDs: nodeIDs, suggestedNodeID }),

        updateNodeDecomposition: (nodeID: string, data: any) =>
          set((state) => ({
            nodes: state.nodes.map((node) =>
              node.id === nodeID
                ? {
                    ...node,
                    data: {
                      ...node.data,
                      decomposition: { ...node.data.decomposition, ...data },
                    },
                  }
                : node
            ),
          })),

        updateNodeConclusion: (nodeID: string, data: any) =>
          set((state) => ({
            nodes: state.nodes.map((node) =>
              node.id === nodeID
                ? {
                    ...node,
                    data: {
                      ...node.data,
                      conclusion: { ...node.data.conclusion, ...data },
                    },
                  }
                : node
            ),
          })),

        updateSettings: (newSettings: Partial<WorkspaceSettings>) =>
          set((state) => ({
            settings: { ...state.settings, ...newSettings },
          })),

        reset: () => set({ ...initialState }),
      },
    }),
    { name: 'workspace-store' }
  )
);

// Selector hooks
export const useWorkspaceStoreData = () => {
  const store = useWorkspaceStore();
  return {
    mapID: store.mapID,
    mapInfo: store.mapInfo,
    nodes: store.nodes,
    edges: store.edges,
    isLoading: store.isLoading,
    settings: store.settings,
  };
};

export const usePanelState = () => {
  const store = useWorkspaceStore();
  return {
    panelOpen: store.panelOpen,
    panelWidth: store.panelWidth,
    activeNodeID: store.activeNodeID,
  };
};
