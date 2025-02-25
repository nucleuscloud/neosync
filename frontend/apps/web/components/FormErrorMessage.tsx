import { cn } from '@/libs/utils';
import { ReactElement } from 'react';

interface Props {
  message?: string;

  id?: string;
  className?: string;
}

export default function FormErrorMessage(props: Props): ReactElement<any> | null {
  const { message, id, className } = props;

  if (!message) {
    return null;
  }

  return (
    <p
      id={id}
      className={cn('text-[0.8rem] font-medium text-destructive', className)}
      {...props}
    >
      {message}
    </p>
  );
}
