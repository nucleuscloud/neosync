// Offset is important here as without it, things get pretty strange I believe due to the container

import { useMemo } from 'react';
import { useResizeDetector } from 'react-resize-detector';
import { OnRefChangeType } from 'react-resize-detector/build/types/types';

// Lower offsets cause the resize to happen at a glacial pace, and without one, not at all.
const WIDTH_OFFSET = 10;

interface UseMonacoResizerReturn {
  ref: OnRefChangeType<HTMLDivElement>;
  width: string;
}

export default function useMonacoResizer(): UseMonacoResizerReturn {
  const { ref, width } = useResizeDetector<HTMLDivElement>({
    handleHeight: false,
    handleWidth: true,
    refreshMode: 'debounce',
    refreshRate: 10,
    skipOnMount: false,
  });

  const editorWidth = useMemo(
    () =>
      width != null && width > WIDTH_OFFSET
        ? `${width - WIDTH_OFFSET}px`
        : '100%',
    [width]
  );

  return {
    ref,
    width: editorWidth,
  };
}
