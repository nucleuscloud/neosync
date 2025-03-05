'use client';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { useNeosyncUser } from '@/libs/hooks/useNeosyncUser';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { UserAccountService } from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { signIn, useSession } from 'next-auth/react';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';

export default function InvitePage(): ReactElement {
  const { status } = useSession();
  const { setAccount, mutateUserAccount } = useAccount();
  const { isLoading: isUserLoading, error: userError } = useNeosyncUser();
  const searchParams = useSearchParams();
  const token = searchParams.get('token');
  const router = useRouter();
  const [error, setError] = useState<string>();
  const isFirstRender = useRef(true);
  const { data: systemData, isLoading: isSystemDataLoading } =
    useGetSystemAppConfig();
  const { mutateAsync: acceptTeamInvite } = useMutation(
    UserAccountService.method.acceptTeamAccountInvite
  );

  useEffect(() => {
    if (isSystemDataLoading) {
      return;
    }
    if (status === 'unauthenticated') {
      // signin must be called on the client for this page so the redirectUrl is properly set
      signIn(systemData?.signInProviderId).catch((err) => {
        toast.error('Unable to redirect to signin page', {
          description: err,
        });
      });
      return;
    }
    // Don't accept invite until we know the user is authenticated and the user has been loaded
    // We wait for the user to load because sometimes the acceptTeamInvite runs before the user has been created
    // and results in a race that fails to invite unless the user refreshes the page.
    if (
      status === 'authenticated' &&
      !isUserLoading &&
      !userError &&
      token &&
      isFirstRender.current
    ) {
      isFirstRender.current = false;
      acceptTeamInvite({ token })
        .then((res) => {
          if (res.account) {
            toast.success('Invite accepted');
            setAccount(res.account);
            mutateUserAccount();
            router.replace(`/${res.account.name}`);
          }
        })
        .catch((err) => setError(getErrorMessage(err)));
    }
  }, [status, token, isUserLoading, userError, isSystemDataLoading]);

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
