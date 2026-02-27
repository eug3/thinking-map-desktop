/**
 * Panel Tabs - Tab navigation for operation panel
 */
import { useState } from 'react';
import { X } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { InfoTab } from './info-tab';
import { DecomposeTab } from './decompose-tab';
import { ConclusionTab } from './conclusion-tab';
import { Node } from 'reactflow';
import { CustomNodeModel } from '@/types/node';

interface PanelTabsProps {
  nodeID: string;
  node: Node<CustomNodeModel>;
}

export function PanelTabs({ nodeID, node }: PanelTabsProps) {
  const { actions } = useWorkspaceStore();
  const [activeTab, setActiveTab] = useState('info');

  const handleSwitchTab = (tab: 'info' | 'decompose' | 'conclusion') => {
    setActiveTab(tab);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b">
        <h3 className="text-sm font-semibold truncate flex-1">
          {node.data?.question || '节点详情'}
        </h3>
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 cursor-pointer"
          onClick={() => actions.closePanel()}
        >
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex flex-col flex-1 min-h-0">
        <TabsList className="grid w-full grid-cols-3 px-4 mt-2">
          <TabsTrigger value="info" className="cursor-pointer">
            信息
          </TabsTrigger>
          <TabsTrigger value="decompose" className="cursor-pointer">
            对话
          </TabsTrigger>
          <TabsTrigger value="conclusion" className="cursor-pointer">
            结论
          </TabsTrigger>
        </TabsList>

        <TabsContent value="info" className="flex-1 min-h-0 mt-0">
          <InfoTab nodeID={nodeID} nodeData={node.data} />
        </TabsContent>

        <TabsContent value="decompose" className="flex-1 min-h-0 mt-0">
          <DecomposeTab nodeID={nodeID} nodeData={node.data} onSwitchTab={handleSwitchTab} />
        </TabsContent>

        <TabsContent value="conclusion" className="flex-1 min-h-0 mt-0">
          <ConclusionTab nodeID={nodeID} node={node} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default PanelTabs;
