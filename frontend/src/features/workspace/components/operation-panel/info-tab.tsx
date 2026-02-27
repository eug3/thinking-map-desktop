/**
 * Info Tab - Node information editing panel
 */
import { useState, useEffect } from 'react';
import { Save, RotateCcw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { CustomNodeModel } from '@/types/node';
import { updateNode } from '@/api';
import { toast } from 'sonner';

interface InfoTabProps {
  nodeID: string;
  nodeData: CustomNodeModel;
}

export function InfoTab({ nodeID, nodeData }: InfoTabProps) {
  const { mapID } = useWorkspaceStore();
  const { actions } = useWorkspaceStore();

  const [formData, setFormData] = useState({
    question: nodeData?.question || '',
    target: nodeData?.target || '',
  });
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  // Watch node data changes
  useEffect(() => {
    setFormData({
      question: nodeData?.question || '',
      target: nodeData?.target || '',
    });
    setHasChanges(false);
  }, [nodeData]);

  // Check for unsaved changes
  useEffect(() => {
    const changed =
      formData.question !== (nodeData?.question || '') ||
      formData.target !== (nodeData?.target || '');
    setHasChanges(changed);
  }, [formData, nodeData]);

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSave = async () => {
    if (!hasChanges || !mapID) return;

    setIsSaving(true);
    try {
      await updateNode(nodeID, formData);
      actions.updateNodeData(nodeID, { ...formData });
      toast.success('节点信息已保存');
      setHasChanges(false);
    } catch (error) {
      toast.error('保存失败，请重试');
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    setFormData({
      question: nodeData?.question || '',
      target: nodeData?.target || '',
    });
    setHasChanges(false);
  };

  return (
    <div className="h-full flex flex-col space-y-6 py-4">
      {/* Node basic info */}
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="question">当前问题</Label>
          <Textarea
            id="question"
            value={formData.question}
            onChange={(e) => handleInputChange('question', e.target.value)}
            placeholder="描述当前需要解决的问题..."
            className="min-h-[80px] resize-none"
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="target">目标描述</Label>
          <Textarea
            id="target"
            value={formData.target}
            onChange={(e) => handleInputChange('target', e.target.value)}
            placeholder="描述期望达到的目标..."
            className="min-h-[80px] resize-none"
          />
        </div>

        {/* Show conclusion if completed */}
        {nodeData?.status === 'completed' && nodeData?.conclusion && (
          <div className="space-y-2">
            <Label>结论内容</Label>
            <div className="p-3 bg-green-50 border border-green-200 rounded-md">
              <p className="text-sm text-green-800 whitespace-pre-wrap">
                {nodeData.conclusion.content}
              </p>
            </div>
          </div>
        )}
      </div>

      <Separator />

      {/* Node status */}
      <div className="space-y-2">
        <Label>节点状态</Label>
        <div className="text-sm text-muted-foreground">
          <span className="font-medium">类型: </span>
          {nodeData?.nodeType || '未知'}
          <span className="ml-4 font-medium">状态: </span>
          {nodeData?.status || '未知'}
        </div>
        {nodeData?.parentID && (
          <div className="text-sm text-muted-foreground">
            <span className="font-medium">父节点: </span>
            {nodeData.parentID}
          </div>
        )}
      </div>

      <Separator />

      {/* Action buttons */}
      <div className="flex gap-2">
        <Button
          onClick={handleSave}
          disabled={!hasChanges || isSaving}
          className="flex-1 cursor-pointer"
          variant="default"
        >
          <Save className="w-4 h-4 mr-2" />
          {isSaving ? '保存中...' : '保存修改'}
        </Button>

        <Button
          onClick={handleReset}
          variant="outline"
          className="flex-1 cursor-pointer"
          disabled={!hasChanges}
        >
          <RotateCcw className="w-4 h-4 mr-2" />
          重置
        </Button>
      </div>
    </div>
  );
}

export default InfoTab;
