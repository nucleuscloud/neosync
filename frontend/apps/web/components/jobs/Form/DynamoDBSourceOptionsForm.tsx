import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Transformer } from '@/shared/transformers';
import {
  DynamoDBSourceOptionsFormValues,
  DynamoDBSourceUnmappedTransformConfigFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import {
  Passthrough,
  SystemTransformer,
  TransformerSource,
} from '@neosync/sdk';
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
    <div className="flex flex-col gap-2 rounded-lg border p-3">
      <div className="p-1 flex flex-col gap-2">
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
            getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
            buttonText={getButtonText(byteTransformer)}
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
            getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
            buttonText={getButtonText(boolTransformer)}
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
            getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
            buttonText={getButtonText(numTransformer)}
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
            getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
            buttonText={getButtonText(strTransformer)}
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

function isInvalidTransformer(transformer: Transformer): boolean {
  return transformer.source === TransformerSource.UNSPECIFIED;
}

function getButtonText(transformer: Transformer): string {
  return isInvalidTransformer(transformer)
    ? 'Select Transformer'
    : transformer.name;
}

// todo: centralize this, we're copying this logic in a few different spots that we call the EditTransformerOptions component
function getTransformerFromField(
  handler: TransformerHandler,
  value: JobMappingTransformerForm
): Transformer {
  if (
    value.source === TransformerSource.USER_DEFINED &&
    value.config.case === 'userDefinedTransformerConfig'
  ) {
    return (
      handler.getUserDefinedTransformerById(value.config.value.id) ??
      new SystemTransformer({
        config: {
          config: {
            case: 'passthroughConfig',
            value: new Passthrough(),
          },
        },
      })
    );
  }
  return (
    handler.getSystemTransformerBySource(value.source) ??
    new SystemTransformer({
      config: {
        config: {
          case: 'passthroughConfig',
          value: new Passthrough(),
        },
      },
    })
  );
}
