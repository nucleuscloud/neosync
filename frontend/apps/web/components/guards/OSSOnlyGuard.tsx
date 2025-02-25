import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { ReactElement } from 'react';

interface Props {
  children: ReactElement<any>;
}

// Only renders children if the system is not Neosync Cloud
export default function OSSOnlyGuard(props: Props): ReactElement<any> | null {
  const { children } = props;
  const { data: systemAppConfig } = useGetSystemAppConfig();

  if (systemAppConfig?.isNeosyncCloud) {
    return null;
  }
  return children;
}
