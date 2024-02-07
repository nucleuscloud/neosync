import { Post, allPosts } from 'contentlayer/generated';
import { compareDesc } from 'date-fns';
import Image from 'next/image';
import Link from 'next/link';

import AuthorProfile, {
  getAuthorsByPostAuthor,
} from '@/components/AuthorProfile';
import { Badge } from '@/components/ui/badge';
import { formatDate } from '@/lib/utils';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import { Metadata } from 'next';
import { ReactElement } from 'react';

export const metadata: Metadata = {
  title: 'Neosync Blog',
};

export default async function BlogPage(): Promise<ReactElement> {
  const posts = allPosts
    .filter((post) => post.published)
    .sort((a, b) => {
      return compareDesc(new Date(a.date), new Date(b.date));
    });

  const [headerPost, ...remainingPosts] = posts;

  return (
    <div className="container lg:max-w-6xl py-6 lg:py-10">
      {headerPost && <HeaderBlog post={headerPost} />}
      <hr className="my-12" />
      {posts?.length ? (
        <div className="grid gap-10 sm:grid-cols-3">
          {remainingPosts?.map((post, index) => (
            <BlogSummary key={post._id} post={post} isPriority={index <= 1} />
          ))}
        </div>
      ) : (
        <p>No posts published.</p>
      )}
    </div>
  );
}

interface HeaderBlogProps {
  post: Post;
}

function HeaderBlog(props: HeaderBlogProps): ReactElement {
  const { post } = props;
  const authors = getAuthorsByPostAuthor(post.authors);
  return (
    <div
      key={post._id}
      className="group relative flex flex-col-reverse lg:flex-row gap-6 border border-gray-400 bg-white p-4 lg:p-10 rounded-xl shadow hover:shadow-lg"
    >
      <div className="flex flex-col gap-4">
        <div>
          <Badge variant="default">Latest blog</Badge>
        </div>
        <div className="text-3xl">{post.title}</div>
        <div>{post.description}</div>
        <div className="flex flex-col lg:flex-row gap-4 lg:items-center pt-6 ">
          {authors.length > 0 && <AuthorProfile author={authors[0]} />}
          <p className="text-sm text-muted-foreground">
            {formatDate(post.date)}
          </p>
        </div>
        <div className="text-muted-foreground flex flex-row items-center text-sm pt-6 ">
          Read this article <ArrowRightIcon className="ml-2" />
        </div>
        <Link href={post.slug} className="absolute inset-0"></Link>
      </div>
      <div>
        <Image
          src={post.image}
          alt={post.title}
          width={1792}
          height={1024}
          className="rounded-md border bg-muted transition-colors"
        />
      </div>
    </div>
  );
}

interface BlogSummaryProps {
  post: Post;
  isPriority?: boolean;
}

export function BlogSummary(props: BlogSummaryProps): ReactElement {
  const { post, isPriority } = props;
  const postAuthors = getAuthorsByPostAuthor(post.authors) ?? [];
  return (
    <article
      key={post._id}
      className="group relative flex flex-col space-y-2 border max-w-sm border-gray-400 bg-white p-2 rounded-xl shadow-sm hover:shadow-lg"
    >
      <Image
        src={post.image}
        alt={post.title}
        width={1792}
        height={1024}
        className="rounded-md border bg-muted transition-colors lg:max-w-xs"
        priority={isPriority}
      />
      <div className="flex flex-col p-2 gap-2">
        <h1 className="text-xl font-extrabold whitespace-nowrap max-w-[300px] text-ellipsis overflow-hidden">
          {post.title}
        </h1>
        <p className="text-muted-foreground lg:h-[90px]">{post.description}</p>
        <div className="flex flex-col lg:flex-row gap-4 items-center pt-6 ">
          {postAuthors.length && (
            <div className="flex gap-4">
              {postAuthors.map((author) => (
                <AuthorProfile key={author._id} author={author} />
              ))}
            </div>
          )}
          <p className="text-sm text-muted-foreground">
            {formatDate(post.date)}
          </p>
        </div>
      </div>
      <Link href={post.slug} className="absolute inset-0">
        <span className="sr-only">View Article</span>
      </Link>
    </article>
  );
}
