import { Mdx } from '@/components/mdx-components';
import { Doc, allDocs } from 'contentlayer/generated';
import { notFound } from 'next/navigation';
import type { ReactElement } from 'react';
import DocsFooter from './components/DocsFooter';
import DocsHeader, { Breadcrumb } from './components/DocsHeader';
import DocsNavigation, { TreeNode } from './components/DocsNavigation';
import PageNavigation from './components/PageNavigation';

type PathSegment = { order: number; pathName: string };

interface PageParams {
  slug?: string[];
}

interface Props {
  params: PageParams;
}

export default async function DocsPage(props: Props): Promise<ReactElement> {
  const {
    params: { slug },
  } = props;

  const doc = slug
    ? allDocs.find((d) => d.url === `/docs/${slug.join('/')}`)
    : allDocs.find((d) => d.url === '/docs');

  const slugs = slug ? ['docs', ...slug] : [];

  if (!doc) {
    notFound();
  }

  const tree = buildDocsTree(allDocs);
  const breadcrumbs = buildBreadcrumbs(slugs);
  // const childrenTree = buildDocsTree(
  //   allDocs,
  //   doc?.pathSegments.map((ps: PathSegment) => ps.pathName)
  // );
  return (
    <div className="relative w-full mx-auto max-w-screen-2xl lg:flex lg:items-start">
      <div
        style={{ height: 'calc(100vh - 64px)' }}
        className="sticky hidden border-r border-gray-200 top-16 shrink-0 dark:border-gray-800 lg:block"
      >
        <div className="h-full p-8 pl-16 -ml-3 overflow-y-scroll">
          <DocsNavigation tree={tree} />
        </div>
        <div className="absolute inset-x-0 top-0 h-8 bg-gradient-to-t from-white/0 to-white/100 dark:from-gray-950/0 dark:to-gray-950/100" />
        <div className="absolute inset-x-0 bottom-0 h-8 bg-gradient-to-b from-white/0 to-white/100 dark:from-gray-950/0 dark:to-gray-950/100" />
      </div>

      <div className="relative w-full grow">
        <DocsHeader tree={tree} breadcrumbs={breadcrumbs} title={doc.title} />
        <div className="w-full max-w-3xl p-4 pb-8 mx-auto mb-4 prose docs prose-slate prose-violet shrink prose-headings:font-semibold prose-a:font-normal prose-code:font-normal prose-code:before:content-none prose-code:after:content-none prose-hr:border-gray-200 dark:prose-invert dark:prose-a:text-violet-400 dark:prose-hr:border-gray-800 md:mb-8 md:px-8 lg:mx-0 lg:max-w-full lg:px-16">
          {<Mdx code={doc?.body.code} />}
          {/* {doc.show_child_cards && (
            <>
              <hr />
              <div className="grid grid-cols-1 gap-6 mt-12 md:grid-cols-2">
                {childrenTree.map((card: any, index: number) => (
                  <div
                    key={index}
                    onClick={() => router.push(card.urlPath)}
                    className="cursor-pointer"
                  >
                    <ChildCard className="h-full p-6 py-4 hover:border-violet-100 hover:bg-violet-50 dark:hover:border-violet-900/50 dark:hover:bg-violet-900/20">
                      <h3 className="mt-0 no-underline">{card.title}</h3>
                      {card.label && <Label text={card.label} />}
                      <div className="text-sm text-slate-500 dark:text-slate-400">
                        <p>{card.excerpt}</p>
                      </div>
                    </ChildCard>
                  </div>
                ))}
              </div>
            </>
          )} */}
          <DocsFooter doc={doc} />
        </div>
      </div>
      <div
        style={{ maxHeight: 'calc(100vh - 128px)' }}
        className="sticky top-32 hidden w-80 shrink-0 overflow-y-scroll p-8 pr-16 1.5xl:block"
      >
        <PageNavigation headings={doc.headings} />
        <div className="absolute inset-x-0 top-0 h-8 bg-gradient-to-t from-white/0 to-white/100 dark:from-gray-950/0 dark:to-gray-950/100" />
        <div className="absolute inset-x-0 bottom-0 h-8 bg-gradient-to-b from-white/0 to-white/100 dark:from-gray-950/0 dark:to-gray-950/100" />
      </div>
    </div>
  );
}

function buildBreadcrumbs(slugs: string[]): Breadcrumb[] {
  const breadcrumbs: Breadcrumb[] = [];
  let path = '';
  for (const slug of slugs) {
    path += `/${slug}`;
    const breadcrumbDoc = allDocs.find((doc) => {
      return doc.url === path;
    });
    if (!breadcrumbDoc) continue;
    breadcrumbs.push({
      path: breadcrumbDoc.url,
      title: breadcrumbDoc?.nav_title || breadcrumbDoc?.title,
    });
  }
  return breadcrumbs;
}

function buildDocsTree(
  docs: Doc[],
  parentPathNames: string[] = []
): TreeNode[] {
  const level = parentPathNames.length;

  // Remove ID from parent path
  // parentPathNames = parentPathNames
  //   .join('/')
  //   .split('-')
  //   .slice(0, -1)
  //   .join('-')
  //   .split('/');

  return docs
    .filter(
      (doc) =>
        doc.pathSegments.length === level + 1 &&
        doc.pathSegments
          .map((ps: PathSegment) => ps.pathName)
          .join('/')
          .startsWith(parentPathNames.join('/'))
    )
    .sort((a, b) => a.pathSegments[level].order - b.pathSegments[level].order)
    .map<TreeNode>((doc) => ({
      nav_title: doc.nav_title ?? null,
      title: doc.title,
      label: '', //doc.label ?? null,
      excerpt: doc.excerpt ?? null,
      urlPath: doc.url, //doc.url_path,
      collapsible: doc.collapsible ?? null,
      collapsed: doc.collapsed ?? null,
      children: buildDocsTree(
        docs,
        doc.pathSegments.map((ps: PathSegment) => ps.pathName)
      ),
    }));
}
