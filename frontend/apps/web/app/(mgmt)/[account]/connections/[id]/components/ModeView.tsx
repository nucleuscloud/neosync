import { ReactElement } from 'react';

interface Props {
  mode: 'view' | 'edit' | 'clone';
  view(): ReactElement;
  edit(): ReactElement;
  clone(): ReactElement;
}
export default function ModeView(props: Props): ReactElement {
  const { mode, view, edit, clone } = props;

  if (mode === 'view') {
    return view();
  }

  if (mode === 'clone') {
    return clone();
  }

  return edit();
}
