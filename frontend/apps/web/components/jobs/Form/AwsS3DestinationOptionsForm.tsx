import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input, InputProps } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getStorageClassString } from '@/util/util';
import { AwsS3DestinationOptionsFormValues } from '@/yup-validations/jobs';
import { AwsS3DestinationConnectionOptions_StorageClass } from '@neosync/sdk';
import { ReactElement, ReactNode } from 'react';
import { FieldErrors } from 'react-hook-form';

interface Props {
  value: AwsS3DestinationOptionsFormValues;
  setValue(newVal: AwsS3DestinationOptionsFormValues): void;
  errors?: FieldErrors<AwsS3DestinationOptionsFormValues>;
}

const SUPPORTED_STORAGECLASSES = [
  AwsS3DestinationConnectionOptions_StorageClass.STANDARD,
  AwsS3DestinationConnectionOptions_StorageClass.STANDARD_IA,
  AwsS3DestinationConnectionOptions_StorageClass.REDUCED_REDUNDANCY,
  AwsS3DestinationConnectionOptions_StorageClass.ONEZONE_IA,
  AwsS3DestinationConnectionOptions_StorageClass.INTELLIGENT_TIERING,
  AwsS3DestinationConnectionOptions_StorageClass.GLACIER,
  AwsS3DestinationConnectionOptions_StorageClass.DEEP_ARCHIVE,
];

export default function AwsS3DestinationOptionsForm(
  props: Props
): ReactElement {
  const { value, setValue, errors } = props;

  return (
    <div className="flex flex-col gap-6 rounded-lg border p-4">
      <div className="flex flex-col gap-2">
        <Header />
      </div>
      <div className="flex flex-col gap-2">
        <FormInputContainer>
          <FormHeader
            title="Storage Class"
            description="The storage class that will be used when records are written to in
              S3"
          />
          <div>
            <Select
              onValueChange={(newVal) => {
                setValue({ ...value, storageClass: parseInt(newVal, 10) });
              }}
              value={
                value.storageClass?.toString() ??
                AwsS3DestinationConnectionOptions_StorageClass.STANDARD.toString()
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SUPPORTED_STORAGECLASSES.map((sc) => (
                  <SelectItem
                    key={sc}
                    className="cursor-pointer"
                    value={sc.toString()}
                  >
                    {getStorageClassString(sc)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FormErrorMessage message={errors?.storageClass?.message} />
          </div>
        </FormInputContainer>
        <FormInputContainer>
          <FormHeader
            title="Max in Flight"
            description="The max number of batched records to have in flight at a given time. Increase to improve throughput."
          />
          <div>
            <NumberedInput
              value={value.maxInFlight ?? 64}
              onChange={(val) => setValue({ ...value, maxInFlight: val })}
            />
            <FormErrorMessage message={errors?.maxInFlight?.message} />
          </div>
        </FormInputContainer>
        <FormInputContainer>
          <FormHeader
            title="Upload Timeout"
            description="The maximum period to wait on an upload before abandoning and re-attempting. Ex: 5s, 1m"
          />
          <div>
            <Input
              value={value.timeout ?? '5s'}
              onChange={(e) => setValue({ ...value, timeout: e.target.value })}
            />
            <FormErrorMessage message={errors?.timeout?.message} />
          </div>
        </FormInputContainer>
        <FormInputContainer>
          <FormHeader
            title="Batch Count"
            description="The max allowed per batch before flushing to S3. 0 to disable batching."
          />
          <div>
            <NumberedInput
              value={value.batch?.count ?? 100}
              onChange={(val) =>
                setValue({ ...value, batch: { ...value.batch, count: val } })
              }
            />
            <FormErrorMessage message={errors?.batch?.count?.message} />
          </div>
        </FormInputContainer>
        <FormInputContainer>
          <FormHeader
            title="Batch Period"
            description="Time in which an incomplete batch should be flushed regardless of the count. Ex: 1s, 1m, 500ms. Empty to disable."
          />
          <div>
            <Input
              value={value.batch?.period ?? '5s'}
              onChange={(e) =>
                setValue({
                  ...value,
                  batch: { ...value.batch, period: e.target.value },
                })
              }
            />
            <FormErrorMessage message={errors?.batch?.period?.message} />
          </div>
        </FormInputContainer>
      </div>
    </div>
  );
}

function Header(): ReactElement {
  return (
    <div>
      <h2 className="text-md font-semibold tracking-tight">S3 Configuration</h2>
      <p className="text-sm tracking-tight">
        Change how Neosync handles sending records to the bucket.
      </p>
    </div>
  );
}

interface FormInputContainerProps {
  children: ReactNode;
}
function FormInputContainer(props: FormInputContainerProps): ReactElement {
  const { children } = props;
  return <div className="flex flex-col gap-2">{children}</div>;
}

interface FormHeaderProps {
  title: string;
  description: string;
  containerClassName?: string;
}
function FormHeader(props: FormHeaderProps): ReactElement {
  const { title, description, containerClassName } = props;
  return (
    <div className={containerClassName}>
      <FormLabel>{title}</FormLabel>
      <FormDescription>{description}</FormDescription>
    </div>
  );
}

interface NumberedInputProps extends Omit<InputProps, 'onChange'> {
  onChange(value: number): void;
}

function NumberedInput(props: NumberedInputProps): ReactElement {
  const { onChange, ...rest } = props;

  return (
    <Input
      {...rest}
      type="number"
      onChange={(event) => {
        const numVal = event.target.valueAsNumber;
        if (!isNaN(numVal)) onChange(numVal);
      }}
    />
  );
}
