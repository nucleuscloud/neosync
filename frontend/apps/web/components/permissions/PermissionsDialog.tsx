import { Alert } from '@/components/ui/alert';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { CheckConnectionConfigResponse } from '@neosync/sdk';
import {
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { ReactElement, useMemo } from 'react';
import { IoWarning } from 'react-icons/io5';
import Spinner from '../Spinner';
import LearnMoreLink from '../labels/LearnMoreLink';
import { Button } from '../ui/button';
import PermissionsDataTable from './PermissionsDataTable';
import { PermissionConnectionType, getPermissionColumns } from './columns';

interface Props {
  checkResponse: CheckConnectionConfigResponse;
  openPermissionDialog: boolean;
  setOpenPermissionDialog(open: boolean): void;
  isValidating: boolean;
  connectionName: string;
  connectionType: PermissionConnectionType;
}

export default function PermissionsDialog(props: Props): ReactElement {
  const {
    openPermissionDialog,
    setOpenPermissionDialog,
    connectionName,
    checkResponse,
    isValidating,
    connectionType,
  } = props;

  const columns = useMemo(
    () => getPermissionColumns(connectionType),
    [connectionType]
  );

  return (
    <Dialog open={openPermissionDialog} onOpenChange={setOpenPermissionDialog}>
      <DialogContent className="max-w-5xl flex flex-col gap-4">
        <DialogHeader>
          <div className="flex flex-col md:flex-row gap-2 items-center">
            <DialogTitle>Connection Permissions</DialogTitle>
            {isValidating ? <Spinner /> : null}
          </div>
          <DialogDescription className="text-muted-foreground text-sm">
            Review the permissions that Neosync needs for your connection.{' '}
            <LearnMoreLink href="https://docs.neosync.dev/connections/postgres#permissions" />{' '}
          </DialogDescription>
        </DialogHeader>
        <PermissionsDataTable
          ConnectionAlert={
            <TestConnectionResult
              isConnected={checkResponse.isConnected}
              connectionName={connectionName}
              hasPrivileges={checkResponse.privileges.length > 0}
            />
          }
          data={checkResponse.privileges}
          columns={columns}
        />
        <DialogFooter className="pt-28">
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
  isConnected: boolean;
  connectionName: string;
  hasPrivileges: boolean;
}

export function TestConnectionResult(
  props: TestConnectionResultProps
): ReactElement {
  const { isConnected, connectionName, hasPrivileges } = props;

  if (isConnected && !hasPrivileges) {
    return (
      <WarningAlert
        description={`We were able to connect to: ${connectionName}, but were not able to find any schema(s) or table(s). Does your role have permissions? `}
      />
    );
  } else if (isConnected) {
    return (
      <SuccessAlert
        description={`Successfully connected to: ${connectionName}!`}
      />
    );
  }
  return <ErrorAlert description="Not currently connected." />;
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

interface ErrorAlertProps {
  description: string;
}

function ErrorAlert(props: ErrorAlertProps): ReactElement {
  const { description } = props;
  return (
    <Alert variant="destructive">
      <div className="flex flex-row items-center gap-2">
        <ExclamationTriangleIcon className="h-4 w-4" />
        <div className="font-normal">{description}</div>
      </div>
    </Alert>
  );
}
