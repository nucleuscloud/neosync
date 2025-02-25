import { cn } from '@/libs/utils';
import { ReactElement, ReactNode } from 'react';

interface Props {
  leftIcon?: ReactNode;
  text: string;
  rightIcon?: ReactNode;
}

export default function ButtonText(props: Props): ReactElement<any> {
  const { leftIcon, text, rightIcon } = props;
  return (
    <div className="flex flex-row gap-1 items-center">
      {leftIcon ? leftIcon : null}
      <p
        className={cn(rightIcon ? 'hidden md:flex truncate' : 'flex truncate')}
      >
        {text}
      </p>
      {rightIcon ? rightIcon : null}
    </div>
  );
}
