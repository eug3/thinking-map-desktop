import { useEffect } from 'react';

export function useTextareaResize(
  textareaRef: React.RefObject<HTMLTextAreaElement>,
  value: string,
  minHeight: number = 44,
  maxHeight: number = 200
) {
  useEffect(() => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    textarea.style.height = `${minHeight}px`;
    const newHeight = Math.min(textarea.scrollHeight, maxHeight);
    textarea.style.height = `${newHeight}px`;
  }, [value, textareaRef, minHeight, maxHeight]);
}
