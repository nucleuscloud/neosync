'use client';
import { getErrorMessage } from '@/util/util';
import {
  CheckConnectionConfigByIdResponse,
  CheckConnectionConfigResponse,
} from '@neosync/sdk';
import { ArrowTopRightIcon, CheckCircledIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { MdErrorOutline } from 'react-icons/md';
import { TiWarningOutline } from 'react-icons/ti';

interface TestConnectionBadgeProps {
  validationResponse:
    | CheckConnectionConfigResponse
    | CheckConnectionConfigByIdResponse
    | undefined;
  connectionId: string | undefined;
  accountName: string;
}

export default function TestConnectionBadge(
  props: TestConnectionBadgeProps
): ReactElement {
  const { validationResponse, connectionId, accountName } = props;

  return (
    <ValidationResponseBadge
      validationResponse={validationResponse}
      accountName={accountName}
      connectionId={connectionId ?? ''}
    />
  );
}

interface ValidationResponseBadgeProps {
  validationResponse:
    | CheckConnectionConfigResponse
    | CheckConnectionConfigByIdResponse
    | undefined;
  accountName: string;
  connectionId: string;
}

function ValidationResponseBadge(
  props: ValidationResponseBadgeProps
): ReactElement | null {
  const { validationResponse, accountName, connectionId } = props;
  const url = `/${accountName}/connections/${connectionId}/permissions`;

  if (!validationResponse) {
    return null;
  }
  if (validationResponse.isConnected) {
    if (validationResponse.privileges.length === 0) {
      return (
        <div className="inline-flex">
          <Link
            href={url}
            passHref
            target="_blank"
            className="flex flex-row items-center gap-2 rounded-xl px-2 py-1 h-auto text-orange-900 dark:text-orange-100 border border-orange-700 bg-orange-100 dark:bg-orange-900 hover:bg-orange-200 dark:hover:bg-orange-950/90 transition-colors"
          >
            <TiWarningOutline />
            <div className="text-nowrap text-xs pl-2 font-medium">
              Successfully connected - found no tables{' '}
              <span className="underline">More info</span>
            </div>
            <ArrowTopRightIcon />
          </Link>
        </div>
      );
    } else {
      return (
        <div className="flex flex-row items-center gap-2 rounded-xl px-2 py-1 h-auto text-green-900 dark:text-green-100 border border-green-700 bg-green-100 dark:bg-green-900 transition-colors">
          <CheckCircledIcon />
          <div className="text-nowrap text-xs font-medium">
            Successfully connected
          </div>
        </div>
      );
    }
  }
  return (
    <div className="inline-flex">
      <Link
        href={url}
        passHref
        target="_blank"
        className="flex flex-row items-center gap-2 rounded-xl px-2 py-1 h-auto text-red-900 dark:text-red-100 border border-red-700 bg-red-100 dark:bg-red-950 dark:hover:bg-red-950/90 hover:bg-red-200 transition-colors"
      >
        <MdErrorOutline />
        <div className="text-nowrap text-xs pl-2 font-medium">
          Connection Error - Unable to connect.{' '}
          <span className="underline">More info</span>
          <p>{getErrorMessage(validationResponse.connectionError)}</p>
        </div>
        <ArrowTopRightIcon />
      </Link>
    </div>
  );
}
