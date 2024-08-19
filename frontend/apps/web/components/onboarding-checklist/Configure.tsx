import { ArrowRightIcon, CaretSortIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';

interface Props {
  currentStep: number;
  setCurrentStep: (val: number) => void;
}

export default function Config(props: Props): ReactElement {
  const { currentStep, setCurrentStep } = props;

  return (
    <div className="flex flex-col gap-12 justify-center items-center text-center">
      <h1 className="font-semibold text-2xl">Configure</h1>
      <p className="text-sm px-10">
        Configure your schema, transformers and subsetting rules in order to
        anonymize and generate data to sync to lower level environments.
      </p>
      <div className=" flex flex-col gap-4 p-4 border border-gray-300 dark:dark:border-[#0D47F0] rounded-lg text-xs shadow-lg">
        <div className="flex flex-row gap-4 justify-between" id="header">
          <div className="w-[100px] font-semibold">Column</div>
          <div className="w-[100px] font-semibold">Constraints</div>
          <div className="w-[100px] font-semibold">Data Type</div>
          <div className="w-[100px] font-semibold">Transformer</div>
        </div>
        <div className="flex flex-row items-center gap-4 justify-between">
          <div className="w-[100px]">id</div>
          <Badge variant="outline" className="w-[100px]">
            Primary Key
          </Badge>
          <Badge variant="outline" className="w-[100px]">
            UUID
          </Badge>
          <TransformerSelector text="UUID ..." />
        </div>
        <div className="flex flex-row items-center gap-4">
          <div className="w-[100px]">first_name</div>
          <div className="w-[100px]" />
          <Badge variant="outline" className="w-[100px]">
            varchar(255)
          </Badge>
          <TransformerSelector text="First Name" />
        </div>
        <div className="flex flex-row items-center gap-4">
          <div className="w-[100px]">last_name</div>
          <div className="w-[100px]" />
          <Badge variant="outline" className="w-[100px]">
            varchar(255)
          </Badge>
          <TransformerSelector text="Last Name ..." />
        </div>
        <div className="flex flex-row items-center gap-4">
          <div className="w-[100px]">email</div>

          <Badge variant="outline" className="w-[100px]">
            Foreign Key
          </Badge>
          <Badge variant="outline" className="w-[100px]">
            varchar(255)
          </Badge>
          <TransformerSelector text="Email ..." />
        </div>
      </div>
      <div className="flex flex-row justify-between w-full py-6">
        <Button
          variant="outline"
          type="reset"
          onClick={() => setCurrentStep(currentStep - 1)}
        >
          Back
        </Button>
        <Button onClick={() => setCurrentStep(currentStep + 1)}>
          <div className="flex flex-row items-center gap-2">
            <div>Next</div> <ArrowRightIcon />
          </div>
        </Button>
      </div>
    </div>
  );
}

interface TransformerSelectorProps {
  text: string;
}

function TransformerSelector(props: TransformerSelectorProps): ReactElement {
  const { text } = props;
  return (
    <div className="flex justify-between border border-gray-300 dark:border-gray-600 p-2 rounded w-[100px]">
      <div className="whitespace-nowrap truncate lg:w-[200px] text-left">
        {text}{' '}
      </div>
      <div>
        <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </div>
    </div>
  );
}
