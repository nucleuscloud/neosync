import { useQuery } from '@connectrpc/connect-query';
import { UserAccountService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { IoAlertCircleOutline } from 'react-icons/io5';
import { Alert, AlertDescription, AlertTitle } from './ui/alert';

interface Props {
  title?: string;
  description?: string;
}

// Displays an alert if the user does not have a valid license
export default function SystemLicenseAlert(props: Props): ReactElement | null {
  const {
    title = 'License Required',
    description = 'This feature is only available to customers with a valid license.',
  } = props;

  const { data: systemInfo } = useQuery(
    UserAccountService.method.getSystemInformation
  );

  if (systemInfo?.license?.isValid) {
    return null;
  }

  return (
    <Alert variant="warning">
      <div className="flex flex-row items-center gap-2">
        <IoAlertCircleOutline className="h-6 w-6" />
        <AlertTitle className="font-semibold">{title}</AlertTitle>
      </div>
      <AlertDescription className="pl-8">{description}</AlertDescription>
    </Alert>
  );
}
