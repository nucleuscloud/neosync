import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import {
  getFilterdTransformersByType,
  getTransformerFromField,
  getTransformerSelectButtonText,
  isInvalidTransformer,
} from '@/util/util';
import {
  DynamoDBSourceOptionsFormValues,
  DynamoDBSourceUnmappedTransformConfigFormValues,
} from '@/yup-validations/jobs';
import { TransformerDataType } from '@neosync/sdk';
import { ExternalLinkIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import TransformerSelect from '../SchemaTable/TransformerSelect';
import { TransformerHandler } from '../SchemaTable/transformer-handler';

interface Props {
  value: DynamoDBSourceOptionsFormValues;
  setValue(newVal: DynamoDBSourceOptionsFormValues): void;
  transformerHandler: TransformerHandler;
}
export default function DynamoDBSourceOptionsForm(props: Props): ReactElement {
  const { value, setValue, transformerHandler } = props;
  return (
    <div className="flex flex-col gap-6 rounded-lg border p-4">
      <div className="flex flex-col gap-2">
        <div>
          <h2 className="text-md font-semibold tracking-tight">
            Unmapped Key Transformer Defaults
          </h2>
          <p className="text-sm tracking-tight">
            Set default transformers for any unmapped keys to automatically
            anonymize generic fields of data by data type.
          </p>
        </div>
        <UnmappedTransformConfigForm
          value={value.unmappedTransformConfig}
          setValue={(newVal) =>
            setValue({ ...value, unmappedTransformConfig: newVal })
          }
          transformerHandler={transformerHandler}
        />
      </div>
      <div className="flex flex-col gap-2">
        <div>
          <h2 className="text-md font-semibold tracking-tight">
            Read Consistency
          </h2>
          <div className="inline-flex gap-1 flex-row">
            <p className="text-sm tracking-tight">
              Configures the read consistency, with the default being eventually
              consistent reads. <br />
              Strongly consistent ensures the most up to date data, reflecting
              updates from all prior successful write operations.
              <br />
              Read more on the{' '}
              <NextLink
                className="hover:underline inline-flex gap-1 flex-row items-center"
                href="https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.ReadConsistency.html"
                target="_blank"
              >
                AWS Docs
                <ExternalLinkIcon className="text-gray-800 w-4 h-4" />
              </NextLink>
            </p>
          </div>
        </div>
        <ConsistentReadForm
          value={value.enableConsistentRead}
          setValue={(checked) =>
            setValue({ ...value, enableConsistentRead: checked })
          }
        />
      </div>
    </div>
  );
}

interface UnmappedTransformConfigFormProps {
  value: DynamoDBSourceUnmappedTransformConfigFormValues;
  setValue(newVal: DynamoDBSourceUnmappedTransformConfigFormValues): void;
  transformerHandler: TransformerHandler;
}
function UnmappedTransformConfigForm(
  props: UnmappedTransformConfigFormProps
): ReactElement {
  const { value, setValue, transformerHandler } = props;
  const byteTransformer = getTransformerFromField(
    transformerHandler,
    value.byte
  );
  const boolTransformer = getTransformerFromField(
    transformerHandler,
    value.boolean
  );
  const numTransformer = getTransformerFromField(transformerHandler, value.n);
  const strTransformer = getTransformerFromField(transformerHandler, value.s);
  return (
    <div className="flex flex-row gap-2 flex-wrap justify-between">
      <div className="flex flex-col gap-2">
        <div>
          <FormLabel>Byte</FormLabel>
          <FormDescription className="text-ellipsis">
            Set a default for unmapped Byte fields
          </FormDescription>
        </div>
        <div className="flex flex-row gap-2">
          <TransformerSelect
            getTransformers={() =>
              getFilterdTransformersByType(
                transformerHandler,
                TransformerDataType.ANY
              )
            }
            buttonText={getTransformerSelectButtonText(byteTransformer)}
            value={value.byte}
            onSelect={(newVal) => setValue({ ...value, byte: newVal })}
            disabled={false}
          />
          <EditTransformerOptions
            transformer={byteTransformer}
            value={value.byte}
            onSubmit={(newVal) => setValue({ ...value, byte: newVal })}
            disabled={isInvalidTransformer(byteTransformer)}
          />
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <div>
          <FormLabel>Boolean</FormLabel>
          <FormDescription>
            Set a default for unmapped Boolean fields
          </FormDescription>
        </div>
        <div className="flex flex-row gap-2">
          <TransformerSelect
            getTransformers={() =>
              getFilterdTransformersByType(
                transformerHandler,
                TransformerDataType.BOOLEAN
              )
            }
            buttonText={getTransformerSelectButtonText(boolTransformer)}
            value={value.boolean}
            onSelect={(newVal) => setValue({ ...value, boolean: newVal })}
            disabled={false}
          />
          <EditTransformerOptions
            transformer={boolTransformer}
            value={value.boolean}
            onSubmit={(newVal) => setValue({ ...value, boolean: newVal })}
            disabled={isInvalidTransformer(boolTransformer)}
          />
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <div>
          <FormLabel>Number</FormLabel>
          <FormDescription>
            Set a default for unmapped Number fields
          </FormDescription>
        </div>
        <div className="flex flex-row gap-2">
          <TransformerSelect
            getTransformers={() =>
              getFilterdTransformersByType(
                transformerHandler,
                TransformerDataType.INT64
              )
            }
            buttonText={getTransformerSelectButtonText(numTransformer)}
            value={value.n}
            onSelect={(newVal) => setValue({ ...value, n: newVal })}
            disabled={false}
          />
          <EditTransformerOptions
            transformer={numTransformer}
            value={value.n}
            onSubmit={(newVal) => setValue({ ...value, n: newVal })}
            disabled={isInvalidTransformer(numTransformer)}
          />
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <div>
          <FormLabel>String</FormLabel>
          <FormDescription>
            Set a default for unmapped String fields
          </FormDescription>
        </div>
        <div className="flex flex-row gap-2">
          <TransformerSelect
            getTransformers={() =>
              getFilterdTransformersByType(
                transformerHandler,
                TransformerDataType.STRING
              )
            } // todo: filter this by type
            buttonText={getTransformerSelectButtonText(strTransformer)}
            value={value.s}
            onSelect={(newVal) => setValue({ ...value, s: newVal })}
            disabled={false}
          />
          <EditTransformerOptions
            transformer={strTransformer}
            value={value.s}
            onSubmit={(newVal) => setValue({ ...value, s: newVal })}
            disabled={isInvalidTransformer(strTransformer)}
          />
        </div>
      </div>
    </div>
  );
}

interface ConsistentReadFormProps {
  value: boolean;
  setValue(newValue: boolean): void;
}

function ConsistentReadForm(props: ConsistentReadFormProps): ReactElement {
  const { value, setValue } = props;

  return (
    <ToggleGroup
      type="single"
      className="flex justify-start items-start flex-wrap"
      onValueChange={(newValue) => {
        // on value change is triggered with an empty string if the currently selected toggle is clicked
        if (newValue) {
          setValue(newValue === 'strong');
        }
      }}
      value={value ? 'strong' : 'eventual'}
    >
      <ToggleGroupItem className="border" value="eventual">
        Eventual
      </ToggleGroupItem>
      <ToggleGroupItem className="border" value="strong">
        Strong
      </ToggleGroupItem>
    </ToggleGroup>
  );
}
