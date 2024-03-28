export type Block = HeaderBlock | DividerBlock | SectionBlock;
type Text = PlainText | MarkdownText;

interface HeaderBlock {
  type: 'header';
  text: Text;
}

interface DividerBlock {
  type: 'divider';
}

interface PlainText {
  type: 'plain_text';
  text: string;
  emoji?: boolean;
}

interface MarkdownText {
  type: 'mrkdwn';
  text: string;
}

type SectionBlock = SectionTextBlock | SectionFieldsBlock;

interface SectionTextBlock {
  type: 'section';
  text: Text;
}
interface SectionFieldsBlock {
  type: 'section';
  fields: Text[];
}
