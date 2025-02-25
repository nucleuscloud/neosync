import { Label } from '@/components/ui/label';
import { cn } from '@/libs/utils';
import { ReactElement, ReactNode } from 'react';

interface FormHeaderProps {
  title: string;
  description: string;
  containerClassName?: string;
  isErrored?: boolean;
  htmlFor?: string;
  isRequired?: boolean;
}

// This is intended to replace the shadcn form entirely by being completely stateless to any specific form(s)
export default function FormHeader(props: FormHeaderProps): ReactElement<any> {
  const {
    title,
    description,
    containerClassName,
    isErrored,
    htmlFor,
    isRequired,
  } = props;
  return (
    <div className={containerClassName}>
      <Label
        htmlFor={htmlFor}
        className={cn(isErrored ? 'text-destructive' : undefined)}
      >
        {isRequired && <span className="text-destructive mr-1">*</span>}
        {title}
      </Label>
      <Description>{description}</Description>
    </div>
  );
}

interface DescriptionProps {
  children: ReactNode;
}
function Description(props: DescriptionProps): ReactElement<any> {
  const { children } = props;
  return <p className="text-[0.8rem] text-muted-foreground">{children}</p>;
}
