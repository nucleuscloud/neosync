import Translate from '@docusaurus/Translate';
import IconEdit from '@theme/Icon/Edit';
import React from 'react';
export default function EditThisPage({ editUrl }) {
  return (
    <a
      href={editUrl}
      target="_blank"
      rel="noreferrer noopener"
      className="flex flex-row items-center text-sm"
    >
      <IconEdit />
      <Translate
        id="theme.common.editThisPage"
        description="The link label to edit the current page"
      >
        Edit this page
      </Translate>
    </a>
  );
}
