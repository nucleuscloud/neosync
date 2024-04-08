import { ReactElement } from 'react';

interface Props {
  children: JSX.Element;
  onClick?: () => void;
}

export default function ShimmeringButton(props: Props): ReactElement {
  const { children, onClick } = props;

  return (
    <button
      className="inline-flex h-10 animate-shimmer items-center justify-center rounded-full border border-slate-400 bg-[linear-gradient(110deg,#000103,45%,#1e2631,55%,#000103)] bg-[length:200%_100%] px-6 font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2 focus:ring-offset-slate-50"
      onClick={onClick}
    >
      {children}
    </button>
  );
}
