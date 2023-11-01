import { cn } from '@/libs/utils';
import { ReactElement, ReactNode } from 'react';

interface Props {
  containerClassName?: string;
  Header: ReactNode;
  children: ReactNode;
  childrenStackClassnames?: string;
}

/**
 * The purpose of this component is to offer a standardized way of defining pages with a header
 */
export default function OverviewContainer(props: Props): ReactElement {
  const { containerClassName, Header, children, childrenStackClassnames } =
    props;

  return (
    <div className={containerClassName}>
      <div className="header-container my-8">{Header}</div>
      <div className={cn('flex', 'flex-col', 'gap-5', childrenStackClassnames)}>
        {children}
      </div>
    </div>
  );
}
