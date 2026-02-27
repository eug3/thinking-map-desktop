/**
 * Visualization Area - ReactFlow canvas for thinking map
 * Adapted from web version for Wails desktop
 */
import { useCallback, useEffect, useRef } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider,
  useReactFlow,
  NodeTypes,
  EdgeTypes,
  Connection,
  Edge,
  Node,
  NodeChange,
  EdgeChange,
  applyNodeChanges,
  applyEdgeChanges,
  BackgroundVariant,
} from 'reactflow';
import 'reactflow/dist/style.css';

import { CustomNode } from './custom-node/custom-node';
import { DependencyEdge } from './custom-edge/custom-edge';
import {
  useWorkspaceStore,
  useWorkspaceStoreData,
} from '@/features/workspace/store/workspace-store';
import {
  getMap,
  getMapNodes,
  createNode,
  updateNode,
  deleteNode as apiDeleteNode,
  executableNodes,
} from '@/api';
import { CustomNodeModel } from '@/types/node';
import { NodeResponse } from '@/types/node';
import { toast } from 'sonner';
import { Events } from '@wailsio/runtime';

// Node types registration
const nodeTypes: NodeTypes = {
  custom: CustomNode,
};

const edgeTypes: EdgeTypes = {
  dependency: DependencyEdge,
};

const defaultEdgeOptions = {
  style: { strokeWidth: 2 },
  animated: false,
};

const fitViewOptions = { padding: 0.2 };
const proOptions = { hideAttribution: true };

// Convert NodeResponse to ReactFlow Node
function toReactFlowNode(resp: NodeResponse): Node<CustomNodeModel> {
  return {
    id: resp.id,
    type: 'custom',
    position: resp.position || { x: 0, y: 0 },
    data: {
      id: resp.id,
      parentID: resp.parentID,
      nodeType: resp.nodeType,
      question: resp.question,
      target: resp.target,
      status: resp.status as CustomNodeModel['status'],
      context: resp.context,
      decomposition: resp.decomposition,
      conclusion: resp.conclusion,
      dependencies: resp.dependencies || [],
      metadata: resp.metadata,
    },
  };
}

// Build edges from node parent relationships
function buildEdgesFromNodes(nodeResponses: NodeResponse[]): Edge[] {
  const edges: Edge[] = [];

  for (const node of nodeResponses) {
    // Parent-child edge
    if (node.parentID) {
      edges.push({
        id: `${node.parentID}-${node.id}`,
        source: node.parentID,
        target: node.id,
        type: 'default',
        style: { stroke: '#3b82f6', strokeWidth: 2 },
      });
    }

    // Dependency edges
    if (node.dependencies?.length) {
      for (const depID of node.dependencies) {
        edges.push({
          id: `dep-${depID}-${node.id}`,
          source: depID,
          target: node.id,
          type: 'dependency',
        });
      }
    }
  }

  return edges;
}

interface VisualizationAreaInnerProps {
  mapID: string;
}

function VisualizationAreaInner({ mapID }: VisualizationAreaInnerProps) {
  const reactFlowInstance = useReactFlow();
  const { nodes, edges, isLoading, settings } = useWorkspaceStoreData();
  const { actions } = useWorkspaceStore();
  const isInitialLoad = useRef(true);

  // Load map data
  useEffect(() => {
    if (!mapID) return;

    const loadMapData = async () => {
      actions.setLoading(true);
      actions.setMapID(mapID);

      try {
        // Load map info and nodes in parallel
        const [mapInfo, nodeResponses] = await Promise.all([
          getMap(mapID),
          getMapNodes(mapID),
        ]);

        actions.updateMap(mapInfo);

        if (nodeResponses && nodeResponses.length > 0) {
          const rfNodes = nodeResponses.map(toReactFlowNode);
          const rfEdges = buildEdgesFromNodes(nodeResponses);

          // Inject event handlers into node data
          const nodesWithHandlers = rfNodes.map((n) => ({
            ...n,
            data: {
              ...n.data,
              onEdit: handleNodeEdit,
              onDelete: handleNodeDelete,
              onAddChild: handleAddChild,
              onSelect: handleNodeSelect,
              onDoubleClick: handleNodeDoubleClick,
            },
          }));

          actions.setNodes(nodesWithHandlers);
          actions.setEdges(rfEdges);

          // Load executable nodes
          try {
            const exec = await executableNodes(mapID);
            if (exec) {
              actions.setExecutableNodes(exec.nodeIDs || [], exec.suggestedNodeID || '');
            }
          } catch {
            // ignore
          }

          // Fit view after initial load
          if (isInitialLoad.current) {
            isInitialLoad.current = false;
            setTimeout(() => {
              reactFlowInstance.fitView({ padding: 0.2 });
            }, 100);
          }
        }
      } catch (error) {
        toast.error('加载思维导图数据失败');
        console.error('Failed to load map data:', error);
      } finally {
        actions.setLoading(false);
      }
    };

    loadMapData();
  }, [mapID]); // eslint-disable-line react-hooks/exhaustive-deps

  // Listen for node events from Go backend
  useEffect(() => {
    const offNodeCreated = Events.On('nodeCreated', (event: any) => {
      // Event data can be either:
      // 1. Direct node data (from Emit with map[string]interface{})
      // 2. EventData wrapper (from EmitNodeEvent with {type, nodeId, data})
      let nodeData = event.data;
      if (!nodeData) return;

      // If wrapped in EventData format, extract from inner data
      if (nodeData.type === 'nodeCreated' && nodeData.data) {
        nodeData = nodeData.data;
      }

      // Ensure we have an id field
      if (!nodeData.id && nodeData.nodeId) {
        nodeData.id = nodeData.nodeId;
      }

      if (!nodeData.id) {
        console.error('nodeCreated event missing id:', event.data);
        return;
      }

      // Build the ReactFlow node directly
      const node: Node<CustomNodeModel> = {
        id: nodeData.id,
        type: 'custom',
        position: nodeData.position || { x: 0, y: 0 },
        data: {
          id: nodeData.id,
          parentID: nodeData.parentID || nodeData.parentId || '',
          nodeType: nodeData.nodeType || 'sub_question',
          question: nodeData.question || '',
          target: nodeData.target || '',
          status: nodeData.status || 'pending',
          onEdit: handleNodeEdit,
          onDelete: handleNodeDelete,
          onAddChild: handleAddChild,
          onSelect: handleNodeSelect,
          onDoubleClick: handleNodeDoubleClick,
        },
      };

      actions.addNode(node);
      const parentID = nodeData.parentID || nodeData.parentId;
      if (parentID) {
        actions.addEdge({
          id: `${parentID}-${nodeData.id}`,
          source: parentID,
          target: nodeData.id,
          type: 'default',
          style: { stroke: '#3b82f6', strokeWidth: 2 },
        });
      }
    });

    const offNodeUpdated = Events.On('nodeUpdated', (event: any) => {
      const data = event.data;
      if (!data?.nodeId) return;
      if (data.data) {
        actions.updateNodeData(data.nodeId, data.data);
      }
    });

    const offNodeDeleted = Events.On('nodeDeleted', (event: any) => {
      const data = event.data;
      if (data?.nodeId) {
        actions.deleteNode(data.nodeId);
      }
    });

    return () => {
      offNodeCreated();
      offNodeUpdated();
      offNodeDeleted();
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Node change handler (position changes, selection, etc.)
  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      const updatedNodes = applyNodeChanges(changes, nodes);
      actions.setNodes(updatedNodes);

      // Track position changes for batch saving
      changes.forEach((change) => {
        if (change.type === 'position' && change.dragging === false && change.id) {
          actions.addChangedNodePosition(change.id);
        }
      });
    },
    [nodes, actions]
  );

  // Edge change handler
  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      const updatedEdges = applyEdgeChanges(changes, edges);
      actions.setEdges(updatedEdges);
    },
    [edges, actions]
  );

  // Save positions that have changed
  useEffect(() => {
    const store = useWorkspaceStore.getState();
    const changedPositions = store.changedNodePositions;
    if (changedPositions.length === 0) return;

    const saveTimeout = setTimeout(async () => {
      for (const nodeID of changedPositions) {
        const node = store.nodes.find((n) => n.id === nodeID);
        if (node) {
          try {
            await updateNode(nodeID, { position: node.position });
          } catch {
            // ignore position save errors
          }
        }
      }
      actions.clearChangedNodePositions();
    }, 500);

    return () => clearTimeout(saveTimeout);
  }, [useWorkspaceStore.getState().changedNodePositions.length]); // eslint-disable-line react-hooks/exhaustive-deps

  // Node event handlers
  const handleNodeEdit = useCallback(
    async (
      _mapID: string | null,
      nodeID: string,
      data: Partial<CustomNodeModel>
    ) => {
      try {
        const result = await updateNode(nodeID, {
          question: data.question,
          target: data.target,
          position: undefined,
        });
        actions.updateNodeData(nodeID, data);
        return result;
      } catch (error) {
        toast.error('更新节点失败');
        throw error;
      }
    },
    [actions]
  );

  const handleNodeDelete = useCallback(
    async (_mapID: string | null, nodeID: string) => {
      try {
        await apiDeleteNode(nodeID);
        actions.deleteNode(nodeID);
        toast.success('节点已删除');
      } catch (error) {
        toast.error('删除节点失败');
      }
    },
    [actions]
  );

  const handleAddChild = useCallback(
    async (parentID: string) => {
      if (!mapID) return;
      try {
        const parentNode = nodes.find((n) => n.id === parentID);
        const position = parentNode
          ? { x: parentNode.position.x, y: parentNode.position.y + 200 }
          : { x: 0, y: 0 };

        const result = await createNode({
          mapID,
          parentID,
          nodeType: 'sub_question',
          question: '新子问题',
          target: '',
          position,
        });

        const newNode = toReactFlowNode(result);
        newNode.data = {
          ...newNode.data,
          onEdit: handleNodeEdit,
          onDelete: handleNodeDelete,
          onAddChild: handleAddChild,
          onSelect: handleNodeSelect,
          onDoubleClick: handleNodeDoubleClick,
        };

        actions.addChildNode(parentID, newNode);
        toast.success('子节点已创建');
      } catch (error) {
        toast.error('创建子节点失败');
      }
    },
    [mapID, nodes, actions] // eslint-disable-line react-hooks/exhaustive-deps
  );

  const handleNodeSelect = useCallback(
    (nodeID: string) => {
      actions.selectNode(nodeID);
    },
    [actions]
  );

  const handleNodeDoubleClick = useCallback(
    (nodeID: string) => {
      actions.openPanel(nodeID);
    },
    [actions]
  );

  // ReactFlow event handlers
  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      actions.selectNode(node.id);
    },
    [actions]
  );

  const onNodeDoubleClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      actions.openPanel(node.id);
    },
    [actions]
  );

  const onPaneClick = useCallback(() => {
    actions.clearSelection();
  }, [actions]);

  const onConnect = useCallback(
    (connection: Connection) => {
      if (connection.source && connection.target) {
        const newEdge: Edge = {
          id: `dep-${connection.source}-${connection.target}`,
          source: connection.source,
          target: connection.target,
          type: 'dependency',
        };
        actions.addEdge(newEdge);
      }
    },
    [actions]
  );

  return (
    <div className="w-full h-full">
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-50">
          <div className="flex flex-col items-center gap-2">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            <span className="text-sm text-muted-foreground">加载中...</span>
          </div>
        </div>
      )}

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onNodeDoubleClick={onNodeDoubleClick}
        onPaneClick={onPaneClick}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={fitViewOptions}
        defaultEdgeOptions={defaultEdgeOptions}
        minZoom={0.1}
        maxZoom={2}
        proOptions={proOptions}
      >
        <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
        {settings.showControls && <Controls />}
        {settings.showMinimap && (
          <MiniMap
            nodeStrokeWidth={3}
            zoomable
            pannable
            className="!bg-muted/50"
          />
        )}
      </ReactFlow>
    </div>
  );
}

interface VisualizationAreaProps {
  mapID: string;
}

export function VisualizationArea({ mapID }: VisualizationAreaProps) {
  return (
    <ReactFlowProvider>
      <VisualizationAreaInner mapID={mapID} />
    </ReactFlowProvider>
  );
}

export default VisualizationArea;
