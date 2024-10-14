import { PlusIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import ButtonText from './ButtonText';
import { Button } from './ui/button';

interface Props {
  title: string;
  description: string;
  buttonText: string;
  icon: JSX.Element;
  href: string;
  buttonIcon?: JSX.Element;
  buttonIconSide?: string;
}

export default function EmptyState(props: Props) {
  const {
    title,
    description,
    buttonText,
    icon,
    href,
    buttonIcon,
    buttonIconSide,
  } = props;
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] p-8 text-center bg-gray-50 dark:bg-gray-900/20 rounded-lg border-2 border-dashed border-gray-200 dark:border-gray-700">
      <div className="w-16 h-16 mb-4 rounded-full bg-primary/10 flex items-center justify-center">
        {icon}
      </div>
      <h2 className="text-2xl font-semibold  mb-2">{title}</h2>
      <p className="text-gray-500 mb-6 max-w-sm">{description}</p>
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
    </div>
  );
}
