'use client';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { GearIcon } from '@radix-ui/react-icons';
import { signOut, useSession } from 'next-auth/react';
import { usePostHog } from 'posthog-js/react';
import { ReactElement } from 'react';
import { toast } from 'sonner';

export function UserNav(): ReactElement | null {
  const session = useSession();
  const posthog = usePostHog();

  const avatarImageSrc = session.data?.user?.image ?? '';
  const avatarImageAlt = session.data?.user?.name ?? 'unknown';
  const avatarFallback = getAvatarFallback(session.data?.user?.name);

  if (session.status === 'unauthenticated') {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="relative h-8 w-8 rounded-full">
          <Avatar className="h-8 w-8">
            <AvatarImage src={avatarImageSrc} alt={avatarImageAlt} />
            <AvatarFallback>{avatarFallback}</AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" align="end" forceMount>
        {(session.data?.user?.email || session.data?.user?.name) && (
          <div>
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col space-y-1">
                {session.data?.user?.name && (
                  <p className="text-sm font-medium leading-none">
                    {session.data.user.name}
                  </p>
                )}
                {session.data?.user?.email && (
                  <p className="text-xs leading-none text-muted-foreground">
                    {session.data.user.email}
                  </p>
                )}
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
          </div>
        )}
        <DropdownMenuGroup>
          <DropdownMenuItem disabled>Profile</DropdownMenuItem>
          <DropdownMenuItem disabled>Settings</DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        {session.status === 'authenticated' && (
          <DropdownMenuItem
            className="cursor-pointer"
            onClick={async () => {
              posthog.reset();
              try {
                await signOut({
                  callbackUrl: `/api/auth/provider-sign-out?idToken=${session.data.idToken}`,
                });
              } catch (err) {
                toast.error('Unable to sign out of provider session');
              }
            }}
          >
            Log out
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function getAvatarFallback(name?: string | null): string | ReactElement {
  return !!name ? name[0].toUpperCase() : <GearIcon />;
}
