import { cn } from '@/libs/utils';
import { create } from '@bufbuild/protobuf';
import { useQuery } from '@connectrpc/connect-query';
import {
  HasPermissionRequest_Permission,
  HasPermissionRequestSchema,
  UserAccountService,
} from '@neosync/sdk';
import { cloneElement, ReactElement, ReactNode } from 'react';
import { Skeleton } from './ui/skeleton';

interface BaseProps {
  mode: 'disable' | 'hide';
  accountId: string;
  resourceId: string;
  permission: HasPermissionRequest_Permission;

  skeletonClassName?: string;
}

interface DisableProps extends BaseProps {
  mode: 'disable';
  children:
    | ReactElement<ChildProps | any> // eslint-disable-line @typescript-eslint/no-explicit-any
    | ((props: ChildProps) => ReactNode);
}

interface HideProps extends BaseProps {
  mode: 'hide';
  children: ReactNode;
  // Optionally provide a fallback to render when the user does not have permission
  // Defaults to null, which will hide the child component
  fallback?: ReactNode;
}

type Props = DisableProps | HideProps;

interface ChildProps {
  isDisabled: boolean;
}

export default function PermissionGuard(props: Props): ReactNode {
  const {
    mode,
    accountId,
    resourceId,
    permission,
    children,
    skeletonClassName,
  } = props;

  const { data, isLoading } = useQuery(
    UserAccountService.method.hasPermission,
    create(HasPermissionRequestSchema, {
      accountId,
      resourceId,
      permission,
    }),
    {
      enabled: !!accountId && !!resourceId && permission > 0,
    }
  );

  if (isLoading) {
    return <Skeleton className={cn('h-4 w-full', skeletonClassName)} />;
  }

  if (mode === 'disable') {
    const isDisabled = !data?.hasPermission;
    return typeof children === 'function'
      ? children({
          isDisabled,
        })
      : cloneElement(children, {
          isDisabled,
        });
  }

  const fallback = props.fallback ?? null;
  return data?.hasPermission ? children : fallback;
}
