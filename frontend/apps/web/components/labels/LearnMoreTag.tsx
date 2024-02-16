import { ExternalLinkIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

interface Props {
  href: string;
}

export default function LearnMoreTag(props: Props): ReactElement {
  const { href } = props;

  return (
    <Link href={href} target="_blank" className="underline inline-block">
      <div className="flex flex-row items-center gap-1">
        <div className="text-sm text-muted-foreground">Learn more</div>
        <ExternalLinkIcon />
      </div>
    </Link>
  );
}
