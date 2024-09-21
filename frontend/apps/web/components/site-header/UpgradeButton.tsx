import { ArrowUpIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { Button } from '../ui/button';

interface Props {
  href: string;
  target?: string;
  // Mostly used for side effects since the user click will result in the link being routed to
  onClick?(): void;
}

export default function UpgradeButton(props: Props): ReactElement {
  const { href, target, onClick } = props;

  return (
    <div>
      <Button type="button" variant="outline" size="sm" onClick={onClick}>
        <div className="flex flex-row gap-2 items-center">
          <Link href={href} target={target}>
            <div>Upgrade</div>
          </Link>
          <ArrowUpIcon />
        </div>
      </Button>
    </div>
  );
}
