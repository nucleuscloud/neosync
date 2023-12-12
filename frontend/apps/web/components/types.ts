import { ReactNode } from 'react';

/**
 * Props that are available on top-level next.js pages
 * ```ts
 * export default function Page(props: PageProps): ReactElement {}
 * ```
 */
export interface PageProps {
  params?: Record<string, string>;
  /**
   * You can use this...but it's suggested to instead use the `useSearchParams` hook instead
   */
  searchParams?: Record<string, string | string[]>;
}

/**
 * Props that are available on top-level next.js layout.js components
 * ```ts
 * export default function Layout(props: LayoutProps): ReactElement {}
 * ```
 */
export interface LayoutProps {
  params?: Record<string, string>;
  children: ReactNode;
}
