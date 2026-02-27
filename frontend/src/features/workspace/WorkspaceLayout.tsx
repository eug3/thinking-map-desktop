/*
 * Desktop Workspace Layout
 * Full workspace with ReactFlow visualization and operation panel
 */
'use client';

import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { VisualizationArea } from './components/visualization-area';
import { OperationPanel } from './components/operation-panel/operation-panel';
import {
  useWorkspaceStore,
  useWorkspaceStoreData,
  usePanelState,
} from './store/workspace-store';

export function WorkspaceLayout() {
  const { id: mapID } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { mapInfo } = useWorkspaceStoreData();
  const { panelOpen } = usePanelState();
  const { actions } = useWorkspaceStore();

  // Reset store on unmount
  useEffect(() => {
    return () => {
      actions.reset();
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  if (!mapID) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <h2 className="text-xl font-bold text-red-500">无效的地图 ID</h2>
          <Button onClick={() => navigate('/')} className="mt-4 cursor-pointer">
            返回首页
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col bg-background">
      {/* Header */}
      <header className="border-b px-4 py-2 flex items-center gap-4 flex-shrink-0">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate('/')}
          className="cursor-pointer"
        >
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div className="flex-1 min-w-0">
          <h1 className="font-semibold truncate">
            {mapInfo?.title || '加载中...'}
          </h1>
          {mapInfo?.problem && (
            <p className="text-xs text-muted-foreground truncate">
              {mapInfo.problem}
            </p>
          )}
        </div>
      </header>

      {/* Main content */}
      <main className="flex-1 flex min-h-0">
        {/* Visualization area */}
        <div className="flex-1 min-w-0 relative">
          <VisualizationArea mapID={mapID} />
        </div>

        {/* Operation panel */}
        {panelOpen && <OperationPanel />}
      </main>
    </div>
  );
}

export default WorkspaceLayout;
