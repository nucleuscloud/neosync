export type Block = HeaderBlock | DividerBlock | SectionBlock;
export type Text = PlainText | MarkdownText;

export interface HeaderBlock {
  type: 'header';
  text: Text;
}

export interface DividerBlock {
  type: 'divider';
}

export interface PlainText {
  type: 'plain_text';
  text: string;
  emoji?: boolean;
}

export interface MarkdownText {
  type: 'mrkdwn';
  text: string;
}

export type SectionBlock = SectionTextBlock | SectionFieldsBlock;

export interface SectionTextBlock {
  type: 'section';
  text: Text;
}
export interface SectionFieldsBlock {
  type: 'section';
  fields: Text[];
}
