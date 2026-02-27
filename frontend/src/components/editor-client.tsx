'use client';

import React, { forwardRef, useEffect, useRef } from 'react';

interface EditorClientProps {
  initContent?: string;
  placeholder?: string;
  onChange?: (value: string) => void;
  editable?: boolean;
  className?: string;
  hideToolbar?: boolean;
  isEditing?: boolean;
}

const EditorClient = forwardRef<HTMLTextAreaElement, EditorClientProps>(
  function EditorClient({ initContent = '', placeholder, onChange, editable = true, className, hideToolbar: _hideToolbar, isEditing: _isEditing }, ref) {
    const internalRef = useRef<HTMLTextAreaElement>(null);
    const textareaRef = (ref as React.RefObject<HTMLTextAreaElement>) || internalRef;

    useEffect(() => {
      if (textareaRef.current && initContent !== textareaRef.current.value) {
        textareaRef.current.value = initContent;
      }
    }, [initContent, textareaRef]);

    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      onChange?.(e.target.value);
    };

    return (
      <textarea
        ref={textareaRef}
        placeholder={placeholder}
        onChange={handleChange}
        disabled={!editable}
        className={`w-full min-h-[200px] p-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${className || ''}`}
        defaultValue={initContent}
      />
    );
  }
);

export default EditorClient;
