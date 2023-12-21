'use client';
import { cn } from '@/lib/utils';
import { ChevronDownIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';

export interface TreeNode {
  title: string;
  nav_title: string | null;
  label: string | null;
  excerpt: string | null;
  urlPath: string;
  children: TreeNode[];
  collapsible: boolean | null;
  collapsed: boolean | null;
}

interface Props {
  tree: TreeNode[];
}

export default function DocsNavigation({ tree }: Props): ReactElement {
  const router = useRouter();

  return (
    <aside className="-ml-6 w-80">
      <div>
        <Tree tree={tree} level={0} activePath={''} />
      </div>
    </aside>
  );
}

interface NavLinkProps {
  title: string;
  label?: string;
  url: string;
  level: number;
  activePath: string;
  collapsible: boolean;
  collapsed: boolean;
  toggleCollapsed: () => void;
}

function NavLink({
  title,
  label,
  url,
  level,
  activePath,
  collapsible,
  collapsed,
  toggleCollapsed,
}: NavLinkProps): ReactElement {
  return (
    <div
      className={cn(
        'group flex h-8 items-center justify-between space-x-2 whitespace-nowrap rounded-md px-3 text-sm leading-none',
        url == activePath
          ? `${
              level == 0 ? 'font-medium' : 'font-normal'
            } bg-violet-50 text-violet-900 dark:bg-violet-500/20 dark:text-violet-50`
          : `hover:bg-gray-50 dark:hover:bg-gray-900 ${
              level == 0
                ? 'font-medium text-slate-600 hover:text-slate-700 dark:text-slate-300 dark:hover:text-slate-200'
                : 'font-normal hover:text-slate-600 dark:hover:text-slate-300'
            }`
      )}
    >
      <Link href={url}>
        <a className="flex items-center h-full space-x-2 grow">
          <span>{title}</span>
          {label && <Label text={label} />}
        </a>
      </Link>
      {collapsible && (
        <button
          aria-label="Toggle children"
          onClick={toggleCollapsed}
          className="px-2 py-1 mr-2 shrink-0"
        >
          <span
            className={`block w-2.5 ${collapsed ? '-rotate-90 transform' : ''}`}
          >
            <ChevronDownIcon />
          </span>
        </button>
      )}
    </div>
  );
}

interface NodeProps {
  node: TreeNode;
  level: number;
  activePath: string;
}

function Node({ node, level, activePath }: NodeProps): ReactElement {
  const [collapsed, setCollapsed] = useState<boolean>(node.collapsed ?? false);
  const toggleCollapsed = () => setCollapsed(!collapsed);

  useEffect(() => {
    if (
      activePath == node.urlPath ||
      node.children.map((_) => _.urlPath).includes(activePath)
    ) {
      setCollapsed(false);
    }
  }, [activePath, node.children, node.urlPath]);

  return (
    <>
      <NavLink
        title={node.nav_title || node.title}
        label={node.label || undefined}
        url={node.urlPath}
        level={level}
        activePath={activePath}
        collapsible={node.collapsible ?? false}
        collapsed={collapsed}
        toggleCollapsed={toggleCollapsed}
      />
      {node.children.length > 0 && !collapsed && (
        <Tree tree={node.children} level={level + 1} activePath={activePath} />
      )}
    </>
  );
}

interface TreeProps {
  tree: TreeNode[];
  level: number;
  activePath: string;
}

function Tree({ tree, level, activePath }: TreeProps): ReactElement {
  return (
    <div
      className={cn(
        'ml-3 space-y-2 pl-3',
        level > 0 ? 'border-l border-gray-200 dark:border-gray-800' : ''
      )}
    >
      {tree.map((treeNode, index) => (
        <Node
          key={index}
          node={treeNode}
          level={level}
          activePath={activePath}
        />
      ))}
    </div>
  );
}

interface LabelProps {
  text: string;
  theme?: 'default' | 'primary';
}

export function Label({ text, theme = 'default' }: LabelProps): ReactElement {
  return (
    <span
      className={`inline-block whitespace-nowrap rounded px-1.5 align-middle font-medium leading-4 tracking-wide [font-size:10px] ${
        theme === 'default'
          ? 'border border-slate-400/70 text-slate-500 dark:border-slate-600 dark:text-slate-400'
          : 'border border-purple-300 text-purple-400 dark:border-purple-800 dark:text-purple-600'
      }`}
    >
      {text}
    </span>
  );
}
