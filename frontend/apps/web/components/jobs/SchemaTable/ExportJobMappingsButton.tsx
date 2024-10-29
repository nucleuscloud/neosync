import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import SwitchCard from '@/components/switches/SwitchCard';
import { Button } from '@/components/ui/button';
import { ReactElement, useState } from 'react';

interface Props {
  onClick(shouldFormat: boolean): void | Promise<void>;
  count?: number;
}

export default function ExportJobMappingsButton(props: Props): ReactElement {
  const { onClick, count } = props;
  const [prettyPrint, setPrettyPrint] = useState<boolean>(false);
  const headerText = useHeaderText(count);
  return (
    <div>
      <ConfirmationDialog
        trigger={
          <Button type="button" variant="outline">
            <ButtonText text="Export" />
          </Button>
        }
        headerText={headerText}
        description="This will export job mappings to a JSON file and save them to disk."
        body={
          <Body prettyPrint={prettyPrint} setPrettyPrint={setPrettyPrint} />
        }
        containerClassName="max-w-xl"
        onConfirm={() => {
          // onClick can be a promise, but we don't necessarly want to wait for it
          // as it might take a long time
          onClick(prettyPrint);
          if (prettyPrint) {
            setPrettyPrint(false);
          }
        }}
      />
    </div>
  );
}

const US_NUMBER_FORMAT = new Intl.NumberFormat('en-US');

function useHeaderText(count?: number): string {
  if (!count) {
    return 'Export all Job Mappings';
  }
  return `Export ${getFormattedCount(count)} selected Job Mapping(s)`;
}

function getFormattedCount(count: number): string {
  return US_NUMBER_FORMAT.format(count);
}

interface BodyProps {
  prettyPrint: boolean;
  setPrettyPrint(value: boolean): void;
}

function Body(props: BodyProps): ReactElement {
  const { prettyPrint, setPrettyPrint } = props;

  return (
    <div className="flex flex-col gap-2">
      <SwitchCard
        isChecked={prettyPrint}
        onCheckedChange={setPrettyPrint}
        title="Format JSON"
        description="Do you want to format the JSON prior to downloading the file? This will result in a larger overall file size."
      />
    </div>
  );
}
