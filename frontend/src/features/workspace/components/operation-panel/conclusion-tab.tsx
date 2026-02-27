/**
 * Conclusion Tab - AI conclusion generation panel
 * Adapted for Wails desktop - uses Wails events instead of SSE
 */
import { useState, useEffect } from 'react';
import {
  Save,
  RotateCcw,
  Square,
  Sparkles,
  Edit,
  Eye,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { conclusion as apiConclusion, saveNodeConclusion } from '@/api';
import { toast } from 'sonner';
import { CustomNodeModel } from '@/types/node';
import { Node } from 'reactflow';
import { Events } from '@wailsio/runtime';

interface ConclusionTabProps {
  nodeID: string;
  node: Node<CustomNodeModel>;
}

export function ConclusionTab({ nodeID, node }: ConclusionTabProps) {
  const nodeData = node.data;
  const { mapID, actions } = useWorkspaceStore();
  const [content, setContent] = useState(nodeData?.conclusion?.content || '');
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [instruction, setInstruction] = useState('');

  // Sync conclusion content from node data
  useEffect(() => {
    if (nodeData?.conclusion?.content) {
      setContent(nodeData.conclusion.content);
      setHasChanges(false);
    }
  }, [nodeData?.conclusion?.content]);

  // Listen for Wails events
  useEffect(() => {
    // Handle status updates from nodeUpdated events
    const offNodeUpdated = Events.On('nodeUpdated', (event: any) => {
      const data = event.data;
      if (data?.nodeId !== nodeID) return;
      const evData = data?.data || {};

      if (evData.status) {
        actions.updateNodeData(nodeID, { status: evData.status });
      }
    });

    // Handle streaming conclusion content from messageConclusion events
    const offMessageConclusion = Events.On('messageConclusion', (event: any) => {
      const data = event.data;
      if (data?.nodeId !== nodeID) return;
      const chunk = data?.data?.message;
      if (!chunk) return;

      setContent((prev: string) => prev + chunk);
    });

    const offConclusionCompleted = Events.On(
      'conclusionCompleted',
      (event: any) => {
        const data = event.data;
        if (data?.nodeId !== nodeID) return;
        setIsGenerating(false);
        setHasChanges(true);
        actions.updateNodeData(nodeID, { status: 'pending' });
      }
    );

    return () => {
      offNodeUpdated();
      offMessageConclusion();
      offConclusionCompleted();
    };
  }, [nodeID, actions]);

  const handleSave = async () => {
    if (!hasChanges || !mapID) return;

    setIsSaving(true);
    try {
      await saveNodeConclusion(nodeID, content);
      actions.updateNodeData(nodeID, {
        conclusion: {
          lastMessageID: nodeData?.conclusion?.lastMessageID || '',
          messages: nodeData?.conclusion?.messages || [],
          content,
        },
        status: content.trim() ? 'completed' : nodeData?.status || 'pending',
      });
      toast.success('结论已保存');
      setHasChanges(false);
    } catch (error) {
      toast.error('保存失败，请重试');
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    setContent('');
    setHasChanges(false);
    actions.updateNodeData(nodeID, {
      conclusion: { lastMessageID: '', messages: [], content: '' },
    });
    toast.success('结论已重置');
  };

  const handleStartConclusion = async () => {
    if (!nodeID) {
      toast.error('请先选择一个节点');
      return;
    }

    setIsGenerating(true);
    setIsEditing(false);

    try {
      await apiConclusion(nodeID, '', instruction);
      actions.updateNodeData(nodeID, { status: 'in_conclusion' });
    } catch (error) {
      toast.error('启动结论生成失败');
      setIsGenerating(false);
    }
  };

  const handleStopConclusion = () => {
    setIsGenerating(false);
    actions.updateNodeData(nodeID, { status: 'pending' });
    toast.info('结论生成已停止');
  };

  return (
    <div className="flex flex-col h-full">
      {/* Content area */}
      <div className="flex-1 min-h-0 overflow-y-auto p-4">
        {isEditing ? (
          <Textarea
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              setHasChanges(true);
            }}
            placeholder="请输入结论..."
            className="min-h-[300px] resize-none text-sm"
            disabled={isGenerating}
          />
        ) : (
          <div className="prose prose-sm dark:prose-invert max-w-none">
            {content ? (
              <div className="whitespace-pre-wrap text-sm">{content}</div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                <p className="text-sm mb-2">暂无结论内容</p>
                <p className="text-xs">点击"开始结论"生成AI结论</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Instruction input */}
      <div className="px-4 py-2 border-t">
        <Textarea
          value={instruction}
          onChange={(e) => setInstruction(e.target.value)}
          placeholder="附加指令（可选）：对结论生成的额外要求..."
          className="min-h-[40px] max-h-[80px] resize-none text-sm"
          disabled={isGenerating}
        />
      </div>

      {/* Action buttons */}
      <div className="flex-shrink-0 border-t p-4">
        <div className="flex gap-2 flex-wrap">
          <Button
            onClick={() => setIsEditing(!isEditing)}
            variant={isEditing ? 'default' : 'outline'}
            size="sm"
            className="cursor-pointer"
          >
            {isEditing ? (
              <Eye className="w-4 h-4 mr-1" />
            ) : (
              <Edit className="w-4 h-4 mr-1" />
            )}
            {isEditing ? '预览' : '编辑'}
          </Button>

          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="cursor-pointer"
                disabled={!content}
              >
                <RotateCcw className="w-4 h-4 mr-1" />
                重置
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>确认重置</AlertDialogTitle>
                <AlertDialogDescription>
                  确定要重置当前结论内容吗？此操作不可撤销。
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>取消</AlertDialogCancel>
                <AlertDialogAction onClick={handleReset}>确认</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          {isGenerating ? (
            <Button
              onClick={handleStopConclusion}
              variant="destructive"
              size="sm"
              className="cursor-pointer"
            >
              <Square className="w-4 h-4 mr-1" />
              停止生成
            </Button>
          ) : (
            <Button
              onClick={handleStartConclusion}
              disabled={!nodeID}
              variant="default"
              size="sm"
              className="cursor-pointer"
            >
              <Sparkles className="w-4 h-4 mr-1" />
              开始结论
            </Button>
          )}

          <Button
            onClick={handleSave}
            disabled={!hasChanges || isSaving}
            size="sm"
            className="cursor-pointer ml-auto"
          >
            <Save className="w-4 h-4 mr-1" />
            {isSaving ? '保存中...' : '保存'}
          </Button>
        </div>
      </div>
    </div>
  );
}

export default ConclusionTab;
