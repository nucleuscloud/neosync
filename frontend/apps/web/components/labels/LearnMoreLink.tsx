import { ArrowTopRightIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

interface Props {
  href: string;
}

export default function LearnMoreLink(props: Props): ReactElement {
  const { href } = props;

  return (
    <Link
      href={href}
      target="_blank"
      className="underline inline-block text-gray-600"
    >
      <div className="flex flex-row items-center gap-1">
        <div className="text-sm text-muted-foreground">Learn more</div>
        <ArrowTopRightIcon className="w-3 h-3" />
      </div>
    </Link>
  );
}
