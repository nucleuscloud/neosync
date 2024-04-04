import { Alert } from '@/components/ui/alert';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { ConnectionRolePrivilege } from '@neosync/sdk';
import { CheckCircledIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { IoWarning } from 'react-icons/io5';
import LearnMoreTag from '../labels/LearnMoreTag';
import { Button } from '../ui/button';
import { Skeleton } from '../ui/skeleton';
import PermissionsDataTable from './PermissionsDataTable';
import { getPermissionColumns } from './columns';

interface Props {
  data: ConnectionRolePrivilege[];
  openPermissionDialog: boolean;
  setOpenPermissionDialog: (open: boolean) => void;
  isValidating: boolean;
  validationResponse: boolean;
  connectionName: string;
}

export default function Permissions(props: Props) {
  const {
    data,
    openPermissionDialog,
    setOpenPermissionDialog,
    isValidating,
    validationResponse,
    connectionName,
  } = props;

  const columns = getPermissionColumns();

  if (isValidating) {
    return <Skeleton />;
  }

  return (
    <Dialog open={openPermissionDialog} onOpenChange={setOpenPermissionDialog}>
      <DialogContent className="max-w-5xl flex flex-col gap-4">
        <DialogHeader>
          <DialogTitle>Connection Permissions</DialogTitle>
          <div className="text-muted-foreground text-sm">
            Review the permissions that Neoynsc has to your connection.{' '}
            <LearnMoreTag href="https://docs.neosync.dev/connections/postgres#testing-your-connection" />
          </div>
        </DialogHeader>
        <TestConnectionResult
          resp={validationResponse}
          connectionName={connectionName}
          data={data}
        />
        <PermissionsDataTable data={data} columns={columns} />
        <DialogFooter className="pt-6">
          <div className="flex justify-end">
            <Button
              type="button"
              onClick={() => setOpenPermissionDialog(false)}
            >
              Close
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface TestConnectionResultProps {
  resp: boolean;
  connectionName: string;
  data: ConnectionRolePrivilege[];
}

export function TestConnectionResult(
  props: TestConnectionResultProps
): ReactElement {
  const { resp, connectionName, data } = props;

  if (resp && data.length == 0) {
    return (
      <WarningAlert
        description={`We were able to connect to: ${connectionName}, but were not able to find any schema(s) or table(s). Does your role have permissions? `}
      />
    );
  } else if (resp) {
    return (
      <SuccessAlert
        description={`Successfully connected to connection: ${connectionName}!`}
      />
    );
  }
  return <div />;
}

interface SuccessAlertProps {
  description: string;
}

function SuccessAlert(props: SuccessAlertProps): ReactElement {
  const { description } = props;
  return (
    <Alert variant="success">
      <div className="flex flex-row items-center gap-2">
        <CheckCircledIcon className="h-4 w-4 text-green-900 dark:text-green-400" />
        <div className="font-normal text-green-900 dark:text-green-400">
          {description}
        </div>
      </div>
    </Alert>
  );
}

interface WarningAlertProps {
  description: string;
}

function WarningAlert(props: WarningAlertProps): ReactElement {
  const { description } = props;
  return (
    <Alert variant="warning">
      <div className="flex flex-row items-center gap-2">
        <IoWarning className="h-4 w-4 text-orange-900 dark:text-orange-400" />
        <div className="font-normal text-orange-900 dark:text-orange-400">
          {description}
        </div>
      </div>
    </Alert>
  );
}
