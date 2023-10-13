import { ReactElement, ReactNode } from 'react';

interface Props {
  leftIcon?: ReactNode;
  text: string;
  rightIcon?: ReactNode;
}

export default function ButtonText(props: Props): ReactElement {
  const { leftIcon, text, rightIcon } = props;
  return (
    <div className="flex flex-row gap-1 items-center">
      {leftIcon ? leftIcon : null}
      <p>{text}</p>
      {rightIcon ? rightIcon : null}
    </div>
  );
}
