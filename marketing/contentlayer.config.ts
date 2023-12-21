import { DocumentGen } from 'contentlayer/core';
import {
  ComputedFields,
  defineDocumentType,
  makeSource,
} from 'contentlayer/source-files';
import { mdxToMarkdown } from 'mdast-util-mdx';
import { toMarkdown } from 'mdast-util-to-markdown';
import { bundleMDX } from 'mdx-bundler';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';
import rehypePrettyCode from 'rehype-pretty-code';
import rehypeSlug from 'rehype-slug';
import remarkGfm from 'remark-gfm';

import type * as unified from 'unified';

export type DocHeading = { level: 1 | 2 | 3; title: string };

const computedFields: ComputedFields = {
  slug: {
    type: 'string',
    resolve: (doc) => `/${doc._raw.flattenedPath}`,
  },
  slugAsParams: {
    type: 'string',
    resolve: (doc) => doc._raw.flattenedPath.split('/').slice(1).join('/'),
  },
};

export const Post = defineDocumentType(() => ({
  name: 'Post',
  filePathPattern: `blog/**/*.mdx`,
  contentType: 'mdx',
  fields: {
    title: {
      type: 'string',
      required: true,
    },
    description: {
      type: 'string',
    },
    date: {
      type: 'date',
      required: true,
    },
    published: {
      type: 'boolean',
      default: true,
    },
    image: {
      type: 'string',
      required: true,
    },
    authors: {
      // Reference types are not embedded.
      // Until this is fixed, we can use a simple list.
      // type: "reference",
      // of: Author,
      type: 'list',
      of: { type: 'string' },
      required: true,
    },
  },
  computedFields,
}));

export const Author = defineDocumentType(() => ({
  name: 'Author',
  filePathPattern: `authors/**/*.mdx`,
  contentType: 'mdx',
  fields: {
    title: {
      type: 'string',
      required: true,
    },
    description: {
      type: 'string',
    },
    avatar: {
      type: 'string',
      required: true,
    },
    twitter: {
      type: 'string',
    },
  },
  computedFields,
}));

export const Doc = defineDocumentType(() => ({
  name: 'Doc',
  filePathPattern: `docs/**/*.mdx`,
  contentType: 'mdx',
  fields: {
    title: {
      type: 'string',
      required: true,
    },
    nav_title: {
      type: 'string',
      description: 'Override the title for display in nav',
      required: false,
    },
    published: {
      type: 'boolean',
      required: false,
    },
    excerpt: {
      type: 'string',
      required: false,
    },
    collapsible: {
      type: 'boolean',
      required: false,
      default: false,
    },
    collapsed: {
      type: 'boolean',
      required: false,
      default: false,
    },
  },
  computedFields: {
    url: {
      type: 'string',
      resolve(doc) {
        return `/${doc._raw.flattenedPath}`;
      },
    },
    pathSegments: {
      type: 'json',
      resolve: (doc) =>
        urlFromFilePath(doc)
          .split('/')
          // skip `/docs` prefix
          .slice(2)
          .map((dirName) => {
            const re = /^((\d+)-)?(.*)$/;
            const [, , orderStr, pathName] = dirName.match(re) ?? [];
            const order = orderStr ? parseInt(orderStr) : 0;
            return { order, pathName };
          }),
    },
    headings: {
      type: 'json',
      resolve: async (doc) => {
        const headings: DocHeading[] = [];

        await bundleMDX({
          source: doc.body.raw,
          mdxOptions: (opts) => {
            opts.remarkPlugins = [
              ...((opts.remarkPlugins ?? []) as any),
              tocPlugin(headings),
            ];
            return opts;
          },
        });

        return [{ level: 1, title: doc.title }, ...headings];
      },
    },
  },
}));

export default makeSource({
  contentDirPath: './content',
  documentTypes: [Post, Author, Doc],
  mdx: {
    remarkPlugins: [remarkGfm],
    rehypePlugins: [
      rehypeSlug,
      [
        rehypePrettyCode,
        {
          theme: 'github-dark',
          onVisitLine(node: any) {
            // Prevent lines from collapsing in `display: grid` mode, and allow empty
            // lines to be copy/pasted
            if (node.children.length === 0) {
              node.children = [{ type: 'text', value: ' ' }];
            }
          },
          onVisitHighlightedLine(node: any) {
            node.properties.className.push('line--highlighted');
          },
          onVisitHighlightedWord(node: any) {
            node.properties.className = ['word--highlighted'];
          },
        },
      ],
      [
        rehypeAutolinkHeadings,
        {
          properties: {
            className: ['subheading-anchor'],
            ariaLabel: 'Link to section',
          },
        },
      ],
    ],
  },
});

function urlFromFilePath(doc: DocumentGen): string {
  let urlPath = doc._raw.flattenedPath.replace(/^pages\/?/, '/');
  if (!urlPath.startsWith('/')) urlPath = `/${urlPath}`;
  // Remove preceding indexes from path segments
  urlPath = urlPath
    .split('/')
    .map((segment) => segment.replace(/^\d\d\d\-/, ''))
    .join('/');
  return urlPath;
}

const tocPlugin =
  (headings: DocHeading[]): unified.Plugin =>
  () => {
    return (node: any) => {
      for (const element of node.children.filter(
        (_: any) => _.type === 'heading' || _.name === 'OptionsTable'
      )) {
        if (element.type === 'heading') {
          const title = toMarkdown(
            { type: 'paragraph', children: element.children },
            { extensions: [mdxToMarkdown()] }
          )
            .trim()
            .replace(/<.*$/g, '')
            .replace(/\\/g, '')
            .trim();
          headings.push({ level: element.depth, title });
        } else if (element.name === 'OptionsTable') {
          element.children
            .filter((_: any) => _.name === 'OptionTitle')
            .forEach((optionTitle: any) => {
              optionTitle.children
                .filter((_: any) => _.type === 'heading')
                .forEach((heading: any) => {
                  const title = toMarkdown(
                    { type: 'paragraph', children: heading.children },
                    { extensions: [mdxToMarkdown()] }
                  )
                    .trim()
                    .replace(/<.*$/g, '')
                    .replace(/\\/g, '')
                    .trim();
                  headings.push({ level: heading.depth, title });
                });
            });
        }
      }
    };
  };
