'use client';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AcceptTeamAccountInviteResponse } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { getErrorMessage } from '@/util/util';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useEffect, useRef, useState } from 'react';

export default function InvitePage(): ReactElement {
  const { setAccount } = useAccount();
  const searchParams = useSearchParams();
  const token = searchParams.get('token');
  const router = useRouter();
  const [error, setError] = useState<string>();
  const isFirstRender = useRef(true);

  useEffect(() => {
    if (token) {
      if (isFirstRender.current) {
        isFirstRender.current = false;
        acceptTeamInvite(token)
          .then((res) => {
            if (res.account) {
              setAccount(res.account);
              router.replace(`/`);
            }
          })
          .catch((err) => setError(getErrorMessage(err)));
      }
    }
  }, [token]);

  return (
    <div className="flex justify-center mt-24">
      <Card className="w-1/2 h-/12 py-24">
        <CardHeader className="space-y-1 items-center">
          <CardTitle className="text-2xl">Welcome to the team!</CardTitle>
        </CardHeader>
        <CardContent className="flex gap-4 justify-center">
          {error ? (
            <Alert variant="destructive">
              <ExclamationTriangleIcon className="h-4 w-4" />
              <AlertTitle>Unable to accept invite</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : (
            <div className="flex flex-row items-center gap-4">
              <Spinner />
              Please wait. You will be redirected in a moment
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

async function acceptTeamInvite(
  token: string
): Promise<AcceptTeamAccountInviteResponse> {
  const res = await fetch(`/api/users/accept-invite?token=${token}`, {
    method: 'POST',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return AcceptTeamAccountInviteResponse.fromJson(await res.json());
}
