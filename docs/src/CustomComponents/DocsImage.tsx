import React, { ReactElement } from 'react';

interface Props {
  href: string;
}

export function DocsImage(props: Props): ReactElement {
  const { href } = props;
  return (
    <div className="docsImage">
      <img src={href} />
    </div>
  );
}
