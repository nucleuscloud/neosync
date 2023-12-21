interface Props {
  children: string;
}

export default function InlineBlogCode(props: Props) {
  const { children } = props;
  return (
    <code className="bg-gray-100 text-gray-800 px-1 rounded-md border border-gray-300 text-sm">
      {children}
    </code>
  );
}
