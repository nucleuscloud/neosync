import { ReactElement } from 'react';

interface Props {
  icon: JSX.Element;
}

export default function IconContainer(props: Props): ReactElement {
  const { icon } = props;

  return (
    <div className="bg-[#303030] rounded-lg border border-gray-600 text-gray-100 p-2">
      {icon}
    </div>
  );
}
