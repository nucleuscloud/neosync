'use client';
import { useQuery } from '@connectrpc/connect-query';
import { isAccountStatusValid } from '@neosync/sdk/connectquery';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { useAccount } from '../providers/account-provider';
import { Button } from '../ui/button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog';
import UpgradeButton from './UpgradeButton';

interface UpgradeProps {
  calendlyLink: string;
  isNeosyncCloud: boolean;
}

export default function Upgrade(props: UpgradeProps): ReactElement | null {
  const { calendlyLink, isNeosyncCloud } = props;
  const { account } = useAccount();
  const accountId = account?.id;
  const { data: isAccountStatusValidResp, isLoading } = useQuery(
    isAccountStatusValid,
    { accountId },
    { enabled: !!accountId && isNeosyncCloud }
  );

  // always surface the upgrade button for non-neosynccloud users
  if (!isNeosyncCloud) {
    return <UpgradeButton href={calendlyLink} target="_blank" />;
  }
  if (isLoading || isAccountStatusValidResp?.isValid) {
    return null;
  }
  const billingHref = `/${account?.name}/settings/billing`;
  return (
    <div className="flex flex-row gap-1 items-center">
      <UpgradeInfoDialog
        upgradeHref={billingHref}
        reason={isAccountStatusValidResp?.reason}
      />
      <UpgradeButton href={billingHref} />
    </div>
  );
}

interface UpgradeInfoDialogProps {
  upgradeHref: string;
  reason?: string;
}

function UpgradeInfoDialog(props: UpgradeInfoDialogProps): ReactElement {
  const { upgradeHref, reason } = props;
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button type="button" variant="ghost">
          <ExclamationTriangleIcon className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader className="gap-3">
          <DialogTitle className="text-xl tracking-tight">
            Upgrade to continue using Neosync
          </DialogTitle>
          <DialogDescription className="text-lg tracking-tight">
            The plan you are on has expired or has used up all of its available
            records for the current cycle.
          </DialogDescription>
        </DialogHeader>
        {!!reason && <IncludedReason reason={reason} />}
        <DialogFooter className="md:justify-between">
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              Close
            </Button>
          </DialogClose>
          <UpgradeButton href={upgradeHref} />
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface IncludedReasonProps {
  reason: string;
}

function IncludedReason(props: IncludedReasonProps): ReactElement {
  const { reason } = props;
  return (
    <div className="flex flex-col gap-2">
      <span className="text-lg">Reason</span>
      <p>{reason}</p>
    </div>
  );
}
