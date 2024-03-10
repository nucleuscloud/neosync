import { ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

export function DotBackground(props: Props) {
  const { children } = props;
  return (
    <div className="h-auto w-full bg-white bg-dot-gray-800/[0.2] relative flex items-center justify-center">
      <div className="absolute pointer-events-none inset-0 flex py-10 items-center justify-center  bg-white [mask-image:radial-gradient(ellipse_at_center_700px,transparent_20%,black)]"></div>
      {children}
    </div>
  );
}
