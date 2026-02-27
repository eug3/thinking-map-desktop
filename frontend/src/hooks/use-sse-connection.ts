// SSE连接钩子 - 桌面版使用Wails Events替代HTTP SSE
import { useEffect, useState } from 'react';
import { events, EventTypes } from '@/wails-runtime';

export interface SSEConnectionState {
  connected: boolean;
  reconnecting: boolean;
  error: string | null;
}

export function useSSEConnection(mapId: string | null) {
  const [state, _setState] = useState<SSEConnectionState>({
    connected: true,
    reconnecting: false,
    error: null,
  });

  // 在桌面版中，我们使用Wails Events而不是HTTP SSE
  // 这里只是提供一个兼容的接口
  useEffect(() => {
    // TODO: 初始化Wails Events监听
    const unsubscribes = [
      events.on(EventTypes.NodeCreated, (data) => {
        console.log('Node created:', data);
      }),
      events.on(EventTypes.NodeUpdated, (data) => {
        console.log('Node updated:', data);
      }),
      events.on(EventTypes.NodeDeleted, (data) => {
        console.log('Node deleted:', data);
      }),
      events.on(EventTypes.Error, (data) => {
        console.error('Error:', data);
      }),
    ];

    return () => {
      unsubscribes.forEach(unsub => unsub());
    };
  }, [mapId]);

  return state;
}

// 导出类型用于外部使用
