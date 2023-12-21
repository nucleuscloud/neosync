import { Author, allAuthors } from '@/.contentlayer/generated';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';

interface Props {
  author: Author;
}

export default function AuthorProfile(props: Props): ReactElement {
  const { author } = props;
  if (author.twitter) {
    return (
      <div>
        <Link
          key={author._id}
          href={`https://twitter.com/${author.twitter}`}
          className="flex items-center space-x-2 text-sm"
        >
          <Image
            src={author.avatar}
            alt={author.title}
            width={42}
            height={42}
            className="rounded-full bg-white"
          />
          <div className="flex-1 text-left leading-tight">
            <p className="font-medium">{author.title}</p>
            <p className="text-[12px] text-muted-foreground">
              @{author.twitter}
            </p>
          </div>
        </Link>
      </div>
    );
  }
  return (
    <div>
      <Image
        src={author.avatar}
        alt={author.title}
        width={42}
        height={42}
        className="rounded-full bg-white"
      />
      <div className="flex-1 text-left leading-tight">
        <p className="font-medium">{author.title}</p>
      </div>
    </div>
  );
}

export function getAuthorsByPostAuthor(authors: string[]): Author[] {
  return authors
    .map((author) =>
      allAuthors.find(({ slug }) => slug === `/authors/${author}`)
    )
    .filter(filterNil);
}

function filterNil<T>(item: T | null | undefined): item is T {
  return !!item;
}
