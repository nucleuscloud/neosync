import { ReactElement } from 'react';

interface Props {
  mode: 'view' | 'edit' | 'clone';
  view(): ReactElement<any>;
  edit(): ReactElement<any>;
  clone(): ReactElement<any>;
}
export default function ModeView(props: Props): ReactElement<any> {
  const { mode, view, edit, clone } = props;

  if (mode === 'view') {
    return view();
  }

  if (mode === 'clone') {
    return clone();
  }

  return edit();
}
