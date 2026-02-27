/**
 * Operation Panel - Right-side panel for node operations
 * Resizable panel that shows info/decompose/conclusion tabs
 */
import React, { useCallback, useRef, useState, useEffect } from 'react';
import { useWorkspaceStore, usePanelState, useWorkspaceStoreData } from '@/features/workspace/store/workspace-store';
import { PanelTabs } from './panel-tabs';
import { Node } from 'reactflow';
import { CustomNodeModel } from '@/types/node';

const MIN_PANEL_WIDTH = 320;
const MAX_PANEL_WIDTH = 600;

export function OperationPanel() {
  const panelState = usePanelState();
  const { panelOpen, activeNodeID, panelWidth } = panelState;
  const { nodes } = useWorkspaceStoreData();
  const { actions } = useWorkspaceStore();
  const resizeRef = useRef<HTMLDivElement>(null);
  const [isResizing, setIsResizing] = useState(false);

  // Find active node
  const activeNode = nodes.find((n) => n.id === activeNodeID) as
    | Node<CustomNodeModel>
    | undefined;

  // Resize handlers
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsResizing(true);
  }, []);

  useEffect(() => {
    if (!isResizing) return;

    const handleMouseMove = (e: MouseEvent) => {
      const newWidth = window.innerWidth - e.clientX;
      const clampedWidth = Math.max(MIN_PANEL_WIDTH, Math.min(MAX_PANEL_WIDTH, newWidth));
      actions.setPanelWidth(clampedWidth);
    };

    const handleMouseUp = () => {
      setIsResizing(false);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isResizing, actions]);

  if (!panelOpen || !activeNodeID || !activeNode) {
    return null;
  }

  return (
    <div
      className="relative h-full bg-background border-l flex-shrink-0"
      style={{ width: panelWidth }}
    >
      {/* Resize handle */}
      <div
        ref={resizeRef}
        className="absolute left-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-primary/30 active:bg-primary/50 z-10"
        onMouseDown={handleMouseDown}
      />

      {/* Panel content */}
      <div className="h-full overflow-hidden">
        <PanelTabs nodeID={activeNodeID} node={activeNode} />
      </div>
    </div>
  );
}

export default OperationPanel;
