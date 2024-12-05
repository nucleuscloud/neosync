import { PlusIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { ReactElement, ReactNode } from 'react';
import ButtonText from './ButtonText';
import { Button } from './ui/button';

interface Props {
  title: string;
  description: string;
  icon: JSX.Element;

  extra?: ReactNode;
}

export default function EmptyState(props: Props): ReactElement {
  const { title, description, icon, extra } = props;
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] p-8 text-center bg-gray-50 dark:bg-gray-900/20 rounded-lg border-2 border-dashed border-gray-200 dark:border-gray-700">
      <div className="w-16 h-16 mb-4 rounded-full bg-primary/10 flex items-center justify-center">
        {icon}
      </div>
      <h2 className="text-2xl font-semibold  mb-2">{title}</h2>
      <p className="text-gray-500 mb-6 max-w-sm">{description}</p>
      {extra ? extra : null}
    </div>
  );
}

interface EmptyStateLinkButtonProps {
  href: string;
  buttonIconSide?: 'left' | 'right';
  buttonIcon?: ReactNode;
  buttonText: string;
}
export function EmptyStateLinkButton(
  props: EmptyStateLinkButtonProps
): ReactElement {
  const { buttonIconSide, href, buttonIcon, buttonText } = props;
  return (
    <NextLink href={href}>
      <Button>
        {buttonIconSide == 'right' ? (
          <ButtonText
            rightIcon={buttonIcon ? buttonIcon : <PlusIcon />}
            text={buttonText}
          />
        ) : (
          <ButtonText
            leftIcon={buttonIcon ? buttonIcon : <PlusIcon />}
            text={buttonText}
          />
        )}
      </Button>
    </NextLink>
  );
}
