import { useTheme } from 'next-themes';
import { useMemo } from 'react';

const DARK_THEME = 'vs-dark';
const LIGHT_THEME = 'light';

export default function useMonacoTheme(): string {
  const { resolvedTheme } = useTheme();
  return useMemo(
    () => (resolvedTheme === 'dark' ? DARK_THEME : LIGHT_THEME),
    [resolvedTheme]
  );
}
