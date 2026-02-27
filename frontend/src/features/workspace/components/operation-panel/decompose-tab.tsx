/**
 * Decompose Tab - AI decomposition conversation panel
 * Adapted for Wails desktop - uses Wails events instead of SSE
 */
import React, { useState, useEffect, useCallback } from 'react';
import { GitBranch, Send, RefreshCcw } from 'lucide-react';
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
import { CustomNodeModel } from '@/types/node';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { decomposition } from '@/api';
import { toast } from 'sonner';
import { Events } from '@wailsio/runtime';

interface DecomposeTabProps {
  nodeID: string;
  nodeData: CustomNodeModel;
  onSwitchTab?: (tab: 'info' | 'decompose' | 'conclusion') => void;
}

interface StreamMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: number;
}

export function DecomposeTab({
  nodeID,
  nodeData: _nodeData,
  onSwitchTab,
}: DecomposeTabProps) {
  const { mapID, actions } = useWorkspaceStore();
  const [messages, setMessages] = useState<StreamMessage[]>([]);
  const [isDecomposed, setIsDecomposed] = useState(false);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const messagesEndRef = React.useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Listen for Wails events
  useEffect(() => {
    let currentAssistantId: string | null = null;

    // Handle status updates from nodeUpdated events
    const offNodeUpdated = Events.On('nodeUpdated', (event: any) => {
      const data = event.data;
      if (data?.nodeId !== nodeID) return;
      const evData = data?.data || {};

      if (evData.status) {
        actions.updateNodeData(nodeID, { status: evData.status });
      }
    });

    // Handle streaming message chunks from messageThought events
    const offMessageThought = Events.On('messageThought', (event: any) => {
      const data = event.data;
      if (data?.nodeId !== nodeID) return;
      const chunk = data?.data?.message;
      if (!chunk) return;

      setMessages((prev) => {
        if (!currentAssistantId) {
          currentAssistantId = `assistant-${Date.now()}`;
          return [
            ...prev,
            {
              id: currentAssistantId,
              role: 'assistant',
              content: chunk,
              timestamp: Date.now(),
            },
          ];
        }
        return prev.map((m) =>
          m.id === currentAssistantId
            ? { ...m, content: m.content + chunk }
            : m
        );
      });
    });

    // Also handle messageText events (used by understanding phase)
    const offMessageText = Events.On('messageText', (event: any) => {
      const data = event.data;
      if (data?.nodeId !== nodeID) return;
      const chunk = data?.data?.message;
      if (!chunk) return;

      setMessages((prev) => {
        if (!currentAssistantId) {
          currentAssistantId = `assistant-${Date.now()}`;
          return [
            ...prev,
            {
              id: currentAssistantId,
              role: 'assistant',
              content: chunk,
              timestamp: Date.now(),
            },
          ];
        }
        return prev.map((m) =>
          m.id === currentAssistantId
            ? { ...m, content: m.content + chunk }
            : m
        );
      });
    });

    const offCompleted = Events.On('decompositionCompleted', (event: any) => {
      const data = event.data;
      if (data?.nodeId !== nodeID) return;
      setLoading(false);
      currentAssistantId = null;
      actions.updateNodeData(nodeID, { status: 'pending' });
    });

    // Also handle conclusionCompleted for understanding phase completion
    const offConclusionCompleted = Events.On('conclusionCompleted', (event: any) => {
      const data = event.data;
      if (data?.nodeId !== nodeID) return;
      setLoading(false);
      currentAssistantId = null;
    });

    return () => {
      offNodeUpdated();
      offMessageThought();
      offMessageText();
      offCompleted();
      offConclusionCompleted();
    };
  }, [nodeID, actions]);

  // Submit decomposition request
  const handleSubmit = useCallback(
    async (input: string, newIsDecomposed?: boolean) => {
      if (loading || !mapID) return;

      setLoading(true);
      const currentIsDecomposed = newIsDecomposed ?? isDecomposed;

      try {
        await decomposition(nodeID, input, currentIsDecomposed);
      } catch (error) {
        toast.error('拆解请求失败');
        setLoading(false);
      }
    },
    [loading, mapID, nodeID, isDecomposed]
  );

  const handleSubmitMessage = () => {
    if (loading) return;
    if (!inputValue.trim()) {
      toast.info('请输入消息');
      return;
    }

    // Add user message
    setMessages((prev) => [
      ...prev,
      {
        id: `user-${Date.now()}`,
        role: 'user',
        content: inputValue,
        timestamp: Date.now(),
      },
    ]);
    handleSubmit(inputValue);
    setInputValue('');
  };

  const handleStartAnalysis = () => {
    setMessages((prev) => [
      ...prev,
      {
        id: `user-${Date.now()}`,
        role: 'user',
        content: '开始拆解分析',
        timestamp: Date.now(),
      },
    ]);
    handleSubmit('');
  };

  const handleStartDecompose = () => {
    setIsDecomposed(true);
    setMessages((prev) => [
      ...prev,
      {
        id: `user-${Date.now()}`,
        role: 'user',
        content: '开始拆解',
        timestamp: Date.now(),
      },
    ]);
    handleSubmit('开始拆解', true);
  };

  const handleReset = () => {
    setMessages([]);
    setIsDecomposed(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmitMessage();
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Messages area */}
      <div className="flex-1 min-h-0 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && !loading && (
          <div className="text-center text-muted-foreground py-8">
            <p className="text-sm mb-4">选择一个操作开始对话</p>
            <div className="flex gap-2 justify-center">
              <Button
                variant="outline"
                size="sm"
                onClick={handleStartAnalysis}
              >
                <GitBranch className="w-3 h-3 mr-1" />
                拆解分析
              </Button>
            </div>
          </div>
        )}

        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[85%] rounded-lg px-3 py-2 text-sm ${
                msg.role === 'user'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted'
              }`}
            >
              <p className="whitespace-pre-wrap">{msg.content}</p>
            </div>
          </div>
        ))}

        {loading && messages.length > 0 && (
          <div className="flex justify-start">
            <div className="bg-muted rounded-lg px-3 py-2 text-sm">
              <span className="animate-pulse">思考中...</span>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input area */}
      <div className="flex-shrink-0 border-t p-4 space-y-2">
        <div className="flex gap-2">
          <Textarea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入消息..."
            className="min-h-[40px] max-h-[120px] resize-none text-sm"
            disabled={loading}
          />
          <Button
            size="icon"
            onClick={handleSubmitMessage}
            disabled={loading || !inputValue.trim()}
            className="shrink-0"
          >
            <Send className="w-4 h-4" />
          </Button>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleStartAnalysis}
            disabled={loading}
            className="cursor-pointer"
          >
            <GitBranch className="w-3 h-3 mr-1" />
            拆解分析
          </Button>
          {!isDecomposed && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleStartDecompose}
              disabled={loading}
              className="cursor-pointer"
            >
              开始拆解
            </Button>
          )}
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                disabled={loading || messages.length === 0}
                className="cursor-pointer"
              >
                <RefreshCcw className="w-3 h-3 mr-1" />
                重置
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>确认重置</AlertDialogTitle>
                <AlertDialogDescription>
                  确定要重置当前拆解吗？此操作将清空消息。
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>取消</AlertDialogCancel>
                <AlertDialogAction onClick={handleReset}>确认</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onSwitchTab?.('conclusion')}
            className="cursor-pointer ml-auto"
          >
            跳转结论
          </Button>
        </div>
      </div>
    </div>
  );
}

export default DecomposeTab;
