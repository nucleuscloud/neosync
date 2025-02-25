import { cn } from '@/libs/utils';
import { ReactElement } from 'react';
import { FormDescription } from './ui/form';

interface FormHeaderProps {
  title: string;
  description: string;
  containerClassName?: string;
  isErrored?: boolean;
}

// This is intended to replace the shadcn form headers by supporting stateless error states
export default function FormHeader(props: FormHeaderProps): ReactElement<any> {
  const { title, description, containerClassName, isErrored } = props;
  return (
    <div className={containerClassName}>
      <label className={cn(isErrored ? 'text-red-500' : undefined)}>
        {title}
      </label>
      <FormDescription>{description}</FormDescription>
    </div>
  );
}
