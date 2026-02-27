import { useEffect, useRef } from 'react';

export function useScrollToBottom(dependencies: any[] = []) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (container) {
      container.scrollTop = container.scrollHeight;
    }
  }, dependencies);

  return containerRef;
}
