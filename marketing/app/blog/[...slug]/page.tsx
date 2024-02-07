import { allPosts } from 'contentlayer/generated';
import { notFound } from 'next/navigation';

import '@/styles/mdx.css';
import { Metadata } from 'next';
import Link from 'next/link';

import AuthorProfile, {
  getAuthorsByPostAuthor,
} from '@/components/AuthorProfile';
import { Mdx } from '@/components/mdx-components';
import { getTableOfContents } from '@/components/nav/toc';
import { buttonVariants } from '@/components/ui/button';
import { env } from '@/env';
import { absoluteUrl, cn, formatDate } from '@/lib/utils';
import { ChevronLeftIcon } from '@radix-ui/react-icons';
import { compareDesc } from 'date-fns';
import Image from 'next/image';
import { ReactElement } from 'react';
import { BlogSummary } from '../page';
import { BlogTableOfContents } from '../tableofcontents';

interface PostPageProps {
  params: {
    slug: string[];
  };
}

async function getPostFromParams(params: PostPageProps['params']) {
  const slug = params?.slug?.join('/');
  const post = allPosts.find((post) => post.slugAsParams === slug);

  if (!post) {
    null;
  }

  return post;
}

async function getFooterPosts(params: PostPageProps['params']) {
  const slug = params?.slug?.join('/');
  const footerPosts = allPosts.filter((item) => item.slugAsParams !== slug);
  return footerPosts;
}

export async function generateMetadata({
  params,
}: PostPageProps): Promise<Metadata> {
  const post = await getPostFromParams(params);

  if (!post) {
    return {};
  }

  const url = env.NEXT_PUBLIC_APP_URL;

  const ogUrl = new URL(`${url}/api/og/blog`);
  ogUrl.searchParams.set('heading', post.title);
  ogUrl.searchParams.set('type', 'Blog Post');
  ogUrl.searchParams.set('mode', 'dark');

  return {
    title: post.title,
    description: post.description,
    authors: post.authors.map((author) => ({
      name: author,
    })),
    openGraph: {
      title: post.title,
      description: post.description,
      type: 'article',
      url: absoluteUrl(post.slug),
      images: [
        {
          url: ogUrl.toString(),
          width: 1200,
          height: 630,
          alt: post.title,
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      title: post.title,
      description: post.description,
      images: [ogUrl.toString()],
    },
    metadataBase: new URL(url),
  };
}

export async function generateStaticParams(): Promise<
  PostPageProps['params'][]
> {
  return allPosts.map((post) => ({
    slug: post.slugAsParams.split('/'),
  }));
}

export default async function PostPage({
  params,
}: PostPageProps): Promise<ReactElement> {
  const post = await getPostFromParams(params);
  const footerPosts = await getFooterPosts(params);

  if (!post) {
    notFound();
  }

  const authors = getAuthorsByPostAuthor(post.authors);
  const toc = await getTableOfContents(post.body.raw);

  return (
    <div className="container max-w-4xl relative py-6 lg:grid lg:grid-cols-[1fr_300px] lg:py-10 gap-20 ">
      <div className="py-6 lg:py-10">
        <Link
          href="/blog"
          className={cn(
            buttonVariants({ variant: 'ghost' }),
            'absolute left-[-200px] top-14 hidden xl:inline-flex'
          )}
        >
          <ChevronLeftIcon className="mr-2 h-4 w-4" />
          See all posts
        </Link>
        <h1 className="mt-2 inline-block font-heading text-4xl leading-tight lg:text-4xl">
          {post.title}
        </h1>
        {post.image && (
          <Image
            src={post.image}
            alt={post.title}
            width={720}
            height={405}
            className="my-8 rounded-md border bg-muted transition-colors"
            priority
          />
        )}
        {authors?.length ? (
          <div className="flex gap-4 flex-col lg:flex-row items-center mt-6 mb-20">
            {authors.map((author) => (
              <AuthorProfile key={author._id} author={author} />
            ))}
            <time
              dateTime={post.date}
              className="block text-sm text-muted-foreground"
            >
              {formatDate(post.date)}
            </time>
          </div>
        ) : null}
        <div className="max-w-2xl">
          <Mdx code={post.body.code} />
        </div>
        <hr className="mt-12" />
        <div className="flex flex-col lg:flex-row items-center gap-8 pt-10">
          {footerPosts
            .filter((item) => item.date < post.date)
            .sort((a, b) => {
              return compareDesc(new Date(a.date), new Date(b.date));
            })
            .slice(0, 2)
            ?.map((post, index) => (
              <BlogSummary key={post._id} post={post} isPriority={index <= 1} />
            ))}
        </div>
        <div className="flex justify-center py-6 lg:py-10">
          <Link
            href="/blog"
            className={cn(buttonVariants({ variant: 'ghost' }))}
          >
            <ChevronLeftIcon className="mr-2 h-4 w-4" />
            See all posts
          </Link>
        </div>
      </div>
      <div className="hidden lg:block sticky top-10 max-h-[calc(100vh-4rem)] overflow-y-auto">
        <BlogTableOfContents toc={toc} />
      </div>
    </div>
  );
}
