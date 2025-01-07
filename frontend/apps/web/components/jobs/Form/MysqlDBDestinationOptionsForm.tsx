import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/FormHeader';
import NumberedInput from '@/components/NumberedInput';
import SwitchCard from '@/components/switches/SwitchCard';
import { Input } from '@/components/ui/input';
import { MysqlDbDestinationOptionsFormValues } from '@/yup-validations/jobs';
import { ReactElement, ReactNode } from 'react';
import { FieldErrors } from 'react-hook-form';

interface Props {
  value: MysqlDbDestinationOptionsFormValues;
  setValue(newVal: MysqlDbDestinationOptionsFormValues): void;
  errors?: FieldErrors<MysqlDbDestinationOptionsFormValues>;
  hideInitTableSchema?: boolean;
}

export default function MysqlDBDestinationOptionsForm(
  props: Props
): ReactElement {
  const { value, setValue, errors, hideInitTableSchema } = props;
  return (
    <div className="flex flex-col gap-2">
      <FormItemContainer>
        <SwitchCard
          isChecked={value.truncateBeforeInsert ?? false}
          onCheckedChange={(newVal) => {
            setValue({
              ...value,
              truncateBeforeInsert: newVal,
            });
          }}
          title="Truncate Before Insert"
          description="Truncates table before inserting data"
        />
        <FormErrorMessage message={errors?.truncateBeforeInsert?.message} />
      </FormItemContainer>
      {!hideInitTableSchema && (
        <FormItemContainer>
          <SwitchCard
            isChecked={value.initTableSchema ?? false}
            onCheckedChange={(newVal) => {
              setValue({
                ...value,
                initTableSchema: newVal ?? false,
              });
            }}
            title="Init Table Schema"
            description="Creates schemas, tables, and their constraints."
          />
          <FormErrorMessage message={errors?.initTableSchema?.message} />
        </FormItemContainer>
      )}
      <FormItemContainer>
        <SwitchCard
          isChecked={value.onConflictDoNothing ?? false}
          onCheckedChange={(newVal) => {
            setValue({
              ...value,
              onConflictDoNothing: newVal,
            });
          }}
          title="On Conflict Do Nothing"
          description="If there is a conflict when inserting data do not insert"
        />
        <FormErrorMessage message={errors?.onConflictDoNothing?.message} />
      </FormItemContainer>
      <FormItemContainer>
        <SwitchCard
          isChecked={value.skipForeignKeyViolations ?? false}
          onCheckedChange={(newVal) => {
            setValue({
              ...value,
              skipForeignKeyViolations: newVal,
            });
          }}
          title="Skip Foreign Key Violations"
          description="Insert all valid records, bypassing any that violate foreign key constraints."
        />
        <FormErrorMessage message={errors?.skipForeignKeyViolations?.message} />
      </FormItemContainer>

      <FormItemBorder>
        <FormItemContainer>
          <FormHeader
            title="Max in Flight"
            description="The max number of batched records to have in flight at a given time. Increase to improve throughput."
            isErrored={!!errors?.maxInFlight}
          />
          <div>
            <NumberedInput
              value={value.maxInFlight ?? -1}
              onChange={(val) => setValue({ ...value, maxInFlight: val })}
            />
            <FormErrorMessage message={errors?.maxInFlight?.message} />
          </div>
        </FormItemContainer>
      </FormItemBorder>

      <FormItemBorder>
        <FormItemContainer>
          <FormHeader
            title="Batch Count"
            description="The max allowed per batch before flushing to S3. 0 to disable count-based batching."
            isErrored={!!errors?.batch?.count}
          />
          <FormInputContainer>
            <NumberedInput
              value={value.batch?.count ?? -1}
              onChange={(val) =>
                setValue({ ...value, batch: { ...value.batch, count: val } })
              }
            />
            <FormErrorMessage message={errors?.batch?.count?.message} />
          </FormInputContainer>
        </FormItemContainer>
      </FormItemBorder>
      <FormItemBorder>
        <FormItemContainer>
          <FormHeader
            title="Batch Period"
            description="Time in which an incomplete batch should be flushed regardless of the count. Ex: 1s, 1m, 500ms. Empty to disable time-based batching (not recommended)."
            isErrored={!!errors?.batch?.period}
          />
          <FormInputContainer>
            <Input
              value={value.batch?.period ?? ''}
              onChange={(e) =>
                setValue({
                  ...value,
                  batch: { ...value.batch, period: e.target.value },
                })
              }
            />
            <FormErrorMessage message={errors?.batch?.period?.message} />
          </FormInputContainer>
        </FormItemContainer>
      </FormItemBorder>
    </div>
  );
}

interface FormItemContainerProps {
  children: ReactNode;
}
function FormItemContainer(props: FormItemContainerProps): ReactElement {
  const { children } = props;
  return <div className="flex flex-col gap-2">{children}</div>;
}

interface FormInputContainerProps {
  children: ReactNode;
}
function FormInputContainer(props: FormInputContainerProps): ReactElement {
  const { children } = props;
  return <div className="flex flex-col gap-1">{children}</div>;
}

interface FormItemBorderProps {
  children: ReactNode;
}
function FormItemBorder(props: FormItemBorderProps): ReactElement {
  const { children } = props;
  return (
    <div className="rounded-lg border p-3 shadow-sm dark:border-gray-700">
      {children}
    </div>
  );
}
