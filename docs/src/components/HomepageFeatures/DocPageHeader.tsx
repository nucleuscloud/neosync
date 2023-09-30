import React from "react";

interface Props {
  title: string;
}

export const DocPageHeader = (props: Props) => {
  const { title } = props;
  return <div className="doc-page-header">{title}</div>;
};
