'use client';
import { CheckConnectionConfigResponse } from '@neosync/sdk';
import { ArrowTopRightIcon, CheckCircledIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { MdErrorOutline } from 'react-icons/md';
import { TiWarningOutline } from 'react-icons/ti';

interface TestConnectionBadgeProps {
  validationResponse: CheckConnectionConfigResponse | undefined;
  id: string | undefined;
  accountName: string;
}

export default function TestConnectionBadge(props: TestConnectionBadgeProps) {
  const { validationResponse, id, accountName } = props;

  return (
    <ValidationResponseBadge
      validationResponse={validationResponse}
      accountName={accountName}
      id={id ?? ''}
    />
  );
}

interface ValidationResponseBadgeProps {
  validationResponse: CheckConnectionConfigResponse | undefined;
  accountName: string;
  id: string;
}

function ValidationResponseBadge(props: ValidationResponseBadgeProps) {
  const { validationResponse, accountName, id } = props;
  const url = `/${accountName}/connections/${id}/permissions`;

  if (
    validationResponse?.isConnected &&
    validationResponse.privileges.length > 0
  ) {
    return (
      <div className="flex flex-row items-center gap-2 rounded-xl px-2 py-1 h-auto text-green-900 dark:text-green-100 border border-green-700 bg-green-100 dark:bg-green-900 transition-colors">
        <CheckCircledIcon />
        <div className="text-nowrap text-xs font-medium">
          Successfully connected
        </div>
      </div>
    );
  } else if (
    validationResponse?.isConnected &&
    validationResponse.privileges.length === 0
  ) {
    return (
      <Link href={url} passHref target="_blank">
        <div className="flex flex-row items-center gap-2 rounded-xl px-2 py-1 h-auto text-orange-900 dark:text-orange-100 border border-orange-700 bg-orange-100 dark:bg-orange-900 hover:bg-orange-200 hover:dark:bg-orange-950/90 transition-colors">
          <TiWarningOutline />
          <div className="text-nowrap text-xs font-medium">
            Connection Warning - No tables found.{' '}
            <a
              href={url}
              className="underline"
              target="_blank"
              rel="noopener noreferrer"
            >
              More info
            </a>
          </div>
          <ArrowTopRightIcon />
        </div>
      </Link>
    );
  } else if (validationResponse && !validationResponse.isConnected) {
    return (
      <Link href={url} passHref target="_blank">
        <div className="flex flex-row items-center gap-2 rounded-xl px-2 py-1 h-auto text-red-900 dark:text-red-100 border border-red-700 bg-red-100 dark:bg-red-950 hover:dark:bg-red-950/90 hover:bg-red-200 transition-colors">
          <MdErrorOutline />
          <div className="text-nowrap text-xs pl-2 font-medium">
            Connection Error - Unable to connect.{' '}
            <a
              href={url}
              className="underline"
              target="_blank"
              rel="noopener noreferrer"
            >
              More info
            </a>
          </div>
          <ArrowTopRightIcon />
        </div>
      </Link>
    );
  } else {
    return null;
  }
}
