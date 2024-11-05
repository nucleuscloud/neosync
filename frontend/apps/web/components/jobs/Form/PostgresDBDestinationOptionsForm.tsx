import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/FormHeader';
import NumberedInput from '@/components/NumberedInput';
import SwitchCard from '@/components/switches/SwitchCard';
import { Input } from '@/components/ui/input';
import { PostgresDbDestinationOptionsFormValues } from '@/yup-validations/jobs';
import { ReactElement, ReactNode } from 'react';
import { FieldErrors } from 'react-hook-form';

interface Props {
  value: PostgresDbDestinationOptionsFormValues;
  setValue(newVal: PostgresDbDestinationOptionsFormValues): void;
  errors?: FieldErrors<PostgresDbDestinationOptionsFormValues>;
  hideInitTableSchema?: boolean;
}

export default function PostgresDBDestinationOptionsForm(
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
              truncateCascade: newVal
                ? (value?.truncateCascade ?? false)
                : false,
            });
          }}
          title="Truncate Before Insert"
          description="Truncates table before inserting data"
        />
        <FormErrorMessage message={errors?.truncateBeforeInsert?.message} />
      </FormItemContainer>
      <FormItemContainer>
        <SwitchCard
          isChecked={value.truncateCascade ?? false}
          onCheckedChange={(newVal) => {
            setValue({
              ...value,
              truncateBeforeInsert:
                newVal && !value.truncateBeforeInsert
                  ? true
                  : (value.truncateBeforeInsert ?? false),
              truncateCascade: newVal,
            });
          }}
          title="Truncate Cascade"
          description="TRUNCATE CASCADE to all tables"
        />
        <FormErrorMessage message={errors?.truncateCascade?.message} />
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
            description="Creates table(s) and their constraints. The database schema must already exist. "
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

      <FormItemContainer>
        <FormHeader
          title="Max in Flight"
          description="The max number of batched records to have in flight at a given time. Increase to improve throughput."
          isErrored={!!errors?.maxInFlight}
        />
        <div>
          <NumberedInput
            value={value.maxInFlight ?? 64}
            onChange={(val) => setValue({ ...value, maxInFlight: val })}
          />
          <FormErrorMessage message={errors?.maxInFlight?.message} />
        </div>
      </FormItemContainer>

      <FormItemContainer>
        <FormHeader
          title="Batch Count"
          description="The max allowed per batch before flushing to S3. 0 to disable count-based batching."
          isErrored={!!errors?.batch?.count}
        />
        <FormInputContainer>
          <NumberedInput
            value={value.batch?.count ?? 100}
            onChange={(val) =>
              setValue({ ...value, batch: { ...value.batch, count: val } })
            }
          />
          <FormErrorMessage message={errors?.batch?.count?.message} />
        </FormInputContainer>
      </FormItemContainer>
      <FormItemContainer>
        <FormHeader
          title="Batch Period"
          description="Time in which an incomplete batch should be flushed regardless of the count. Ex: 1s, 1m, 500ms. Empty to disable time-based batching (not recommended)."
          isErrored={!!errors?.batch?.period}
        />
        <FormInputContainer>
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
        </FormInputContainer>
      </FormItemContainer>
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
