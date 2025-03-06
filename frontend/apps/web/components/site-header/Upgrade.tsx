'use client';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { useQuery } from '@connectrpc/connect-query';
import {
  AccountStatus,
  IsAccountStatusValidResponse,
  UserAccountService,
} from '@neosync/sdk';
import { ArrowUpIcon, ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { IoAlertCircleOutline } from 'react-icons/io5';
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
  isAccountStatusValidResp: IsAccountStatusValidResponse | undefined;
  isLoading: boolean;
}

export default function Upgrade(props: UpgradeProps): ReactElement | null {
  const { calendlyLink, isAccountStatusValidResp, isLoading } = props;
  const { account } = useAccount();
  const { data: systemInfo } = useQuery(
    UserAccountService.method.getSystemInformation
  );
  // always surface the upgrade button for non-neosynccloud users
  if (!systemInfo?.license?.isValid && !systemInfo?.license?.isNeosyncCloud) {
    return <UpgradeButton href={calendlyLink} target="_blank" />;
  }

  if (isLoading || isAccountStatusValidResp?.isValid) {
    return null;
  }

  const billingHref = `/${account?.name}/settings/billing`;

  return (
    <div className="flex flex-row gap-1 items-center">
      {!isAccountStatusValidResp?.isValid ? (
        <UpgradeInfoDialog
          upgradeHref={billingHref}
          accountStatus={isAccountStatusValidResp?.accountStatus}
          reason={isAccountStatusValidResp?.reason}
        />
      ) : (
        <UpgradeButton href={billingHref} />
      )}
    </div>
  );
}

interface UpgradeInfoDialogProps {
  upgradeHref: string;
  accountStatus?: AccountStatus;
  reason?: string;
}

function UpgradeInfoDialog(props: UpgradeInfoDialogProps): ReactElement {
  const { upgradeHref, accountStatus, reason } = props;
  const [open, onOpenChange] = useState(false);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button type="button" variant="outline" size="sm">
          <ExclamationTriangleIcon className="w-4 h-4 text-yellow-600 dark:text-yellow-400 mr-2" />
          Upgrade <ArrowUpIcon className="ml-2" />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader className="gap-3">
          <DialogTitle className="text-xl tracking-tight">
            Upgrade Required
          </DialogTitle>
          <DialogDescription className="tracking-tight">
            Upgrade to a Team or Enterprise plan to continue using Neosync.
          </DialogDescription>
        </DialogHeader>
        {!!accountStatus && !!reason && (
          <div className="py-6">
            <IncludedReason accountStatus={accountStatus} reason={reason} />
          </div>
        )}
        <DialogFooter className="md:justify-between">
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              Close
            </Button>
          </DialogClose>
          <UpgradeButton
            href={upgradeHref}
            onClick={() => onOpenChange(false)}
          />
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface IncludedReasonProps {
  accountStatus: AccountStatus;
  reason: string;
}

function IncludedReason(props: IncludedReasonProps): ReactElement {
  const { accountStatus, reason } = props;

  switch (accountStatus) {
    case AccountStatus.ACCOUNT_IN_EXPIRED_STATE:
      return (
        <Alert>
          <IoAlertCircleOutline className="h-4 w-4" />
          <AlertTitle>Account is Expired</AlertTitle>
          <AlertDescription>{reason}</AlertDescription>
        </Alert>
      );
    case AccountStatus.ACCOUNT_TRIAL_EXPIRED:
      return (
        <Alert>
          <IoAlertCircleOutline className="h-4 w-4" />
          <AlertTitle>Account Trial is Expired</AlertTitle>
          <AlertDescription>{reason}</AlertDescription>
        </Alert>
      );
    default:
      return (
        <Alert>
          <IoAlertCircleOutline className="h-4 w-4" />
          <AlertTitle>Warning</AlertTitle>
          <AlertDescription>
            Your account is expired or you have no more records remaining for
            this billing cycle.
          </AlertDescription>
        </Alert>
      );
  }
}
