export interface Database {
  tables: Table[];
}

export interface Table {
  name: string;
  columns: Column[];
}

export interface Column {
  name: string;
  type?: string;
  length?: string;
}

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
