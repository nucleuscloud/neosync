import { ArrowUpIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { Button } from '../ui/button';

interface Props {
  href: string;
}

export default function UpgradeButton(props: Props): ReactElement {
  const { href } = props;

  return (
    <div>
      <Button type="button" variant="outline" size="sm">
        <div className="flex flex-row gap-2 items-center">
          <Link href={href} target="_blank">
            <div>Upgrade</div>
          </Link>
          <ArrowUpIcon />
        </div>
      </Button>
    </div>
  );
}
