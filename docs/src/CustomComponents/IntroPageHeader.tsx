import React from 'react';

interface Props {
  title: string;
}

export const IntroPageHeader = (props: Props) => {
  const { title } = props;
  return <div className="intro-page-header">{title}</div>;
};
