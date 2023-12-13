interface Props {
  errorMessage: string;
}

export default function FormError(props: Props) {
  const { errorMessage } = props;

  return <div className="text-red-600 text-xs">{errorMessage}</div>;
}
