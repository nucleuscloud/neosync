import React from "react";

interface Props {
  title: string;
  subtitle: string;
}

export const DocPageHeader = (props: Props) => {
  const { title, subtitle } = props;
  return (
    <div className="flex flex-col text-left gap-[8px]">
      <div className="font-bold text-3xl">{title}</div>
      <div className="font-light text-xl text-[#666e7a]">{subtitle}</div>
    </div>
  );
};
