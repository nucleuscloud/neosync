import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import SwitchCard from '@/components/switches/SwitchCard';
import { Button } from '@/components/ui/button';
import { ReactElement, useState } from 'react';

interface Props {
  onClick(shouldFormat: boolean): void | Promise<void>;
}

export default function ExportJobMappingsButton(props: Props): ReactElement {
  const { onClick } = props;
  const [prettyPrint, setPrettyPrint] = useState<boolean>(false);
  return (
    <div>
      <ConfirmationDialog
        trigger={
          <Button type="button" variant="outline">
            <ButtonText text="Export" />
          </Button>
        }
        headerText="Export Job Mappings"
        description="This will export job mappings to a JSON file and save them to disk."
        body={
          <ConfirmBody
            prettyPrint={prettyPrint}
            setPrettyPrint={setPrettyPrint}
          />
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

interface ConfirmBodyProps {
  prettyPrint: boolean;
  setPrettyPrint(value: boolean): void;
}

function ConfirmBody(props: ConfirmBodyProps): ReactElement {
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
