'use client';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { AccountStatus, IsAccountStatusValidResponse } from '@neosync/sdk';
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
import { Progress } from '../ui/progress';
import { formatNumber } from './RecordsProgressBar';
import UpgradeButton from './UpgradeButton';

interface UpgradeProps {
  calendlyLink: string;
  isNeosyncCloud: boolean;
  count: number;
  allowedRecords: number;
  isAccountStatusValidResp: IsAccountStatusValidResponse | undefined;
  isLoading: boolean;
}

export default function Upgrade(props: UpgradeProps): ReactElement | null {
  const {
    calendlyLink,
    isNeosyncCloud,
    count,
    allowedRecords,
    isAccountStatusValidResp,
    isLoading,
  } = props;
  const { account } = useAccount();
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
      {!isAccountStatusValidResp?.isValid ? (
        <UpgradeInfoDialog
          upgradeHref={billingHref}
          accountStatus={isAccountStatusValidResp?.accountStatus}
          reason={isAccountStatusValidResp?.reason}
          count={count}
          allowedRecords={allowedRecords}
        />
      ) : (
        <UpgradeButton href={billingHref} />
      )}
    </div>
  );
}

interface UpgradeInfoDialogProps {
  upgradeHref: string;
  count: number;
  allowedRecords: number;
  accountStatus?: AccountStatus;
  reason?: string;
}

function UpgradeInfoDialog(props: UpgradeInfoDialogProps): ReactElement {
  const { upgradeHref, count, allowedRecords, accountStatus, reason } = props;
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
            <IncludedReason
              accountStatus={accountStatus}
              reason={reason}
              count={count}
              allowedRecords={allowedRecords}
            />
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
  count: number;
  allowedRecords: number;
}

function IncludedReason(props: IncludedReasonProps): ReactElement {
  const { accountStatus, reason, count, allowedRecords } = props;

  switch (accountStatus) {
    case AccountStatus.EXCEEDS_ALLOWED_LIMIT:
      return (
        <UsageLimitExceeded
          current={count}
          allowed={allowedRecords}
          title={'Usage Limit Exceeded'}
          reason={`You have used all of your allowed records for this billing cycle. `}
        />
      );
    case AccountStatus.REQUESTED_EXCEEDS_LIMIT:
      return (
        <UsageLimitExceeded
          current={count}
          allowed={allowedRecords}
          title={'Usage Limit Warning'}
          reason={`The requested record count woudl exceed your alowed records for this billing cycle. `}
        />
      );
    case AccountStatus.ACCOUNT_IN_EXPIRED_STATE:
      return (
        <Alert>
          <IoAlertCircleOutline className="h-4 w-4" />
          <AlertTitle>Account is Expired</AlertTitle>
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

interface UsageLimitExceededProps {
  current: number;
  allowed: number;
  title: string;
  reason: string;
}

function UsageLimitExceeded(props: UsageLimitExceededProps) {
  const { current, allowed, title, reason } = props;
  return (
    <div className="py-4">
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
        </CardHeader>
        <CardContent className="pb-2">
          <div className="flex items-center justify-between text-sm text-muted-foreground mb-1">
            <div>Current Usage</div>
            <div className="font-medium">
              {formatNumber(current)}/{formatNumber(allowed)}
            </div>
          </div>
          <Progress value={(current / allowed) * 100} className="h-2" />
        </CardContent>
        <CardFooter className="pt-2">
          <CardDescription className="text-xs">{reason}</CardDescription>
        </CardFooter>
      </Card>
    </div>
  );
}
