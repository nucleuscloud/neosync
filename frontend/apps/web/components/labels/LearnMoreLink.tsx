import { cn } from '@/libs/utils';
import { ArrowTopRightIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

interface Props {
  href: string;
  classNames?: string;
}

export default function LearnMoreLink(props: Props): ReactElement {
  const { href, classNames } = props;

  return (
    <Link
      href={href}
      target="_blank"
      className={cn('underline inline-block text-gray-600 text-sm', classNames)}
    >
      <div className="flex flex-row items-center gap-1">
        <div className="text-muted-foreground">Learn more</div>
        <ArrowTopRightIcon className="w-3 h-3" />
      </div>
    </Link>
  );
}
