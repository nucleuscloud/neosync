import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Transformer } from '@/shared/transformers';
import {
  convertJobMappingTransformerToForm,
  DynamoDBSourceOptionsFormValues,
  DynamoDBSourceUnmappedTransformConfigFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import {
  GenerateBool,
  GenerateString,
  JobMappingTransformer,
  Passthrough,
  SystemTransformer,
  TransformerConfig,
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
    <div className="flex flex-col gap-2">
      <UnmappedTransformConfigForm
        value={
          value.unmappedTransformConfig ?? getDefaultUnmappedTransformConfig()
        }
        setValue={(newVal) =>
          setValue({ ...value, unmappedTransformConfig: newVal })
        }
        transformerHandler={transformerHandler}
      />
    </div>
  );
}

function getDefaultUnmappedTransformConfig(): DynamoDBSourceUnmappedTransformConfigFormValues {
  return {
    boolean: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        source: TransformerSource.GENERATE_BOOL,
        config: new TransformerConfig({
          config: {
            case: 'generateBoolConfig',
            value: new GenerateBool(),
          },
        }),
      })
    ),
    byte: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        source: TransformerSource.PASSTHROUGH,
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough(),
          },
        }),
      })
    ),
    n: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        source: TransformerSource.PASSTHROUGH,
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough(),
          },
        }),
      })
    ),
    s: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        source: TransformerSource.GENERATE_STRING,
        config: new TransformerConfig({
          config: {
            case: 'generateStringConfig',
            value: new GenerateString(),
          },
        }),
      })
    ),
  };
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
    <>
      <div>
        <FormLabel>Byte</FormLabel>
        <FormDescription>
          Set a default transformer for any unmapped Byte fields
        </FormDescription>
        <TransformerSelect
          getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
          buttonText="Select Transformer"
          value={value.byte}
          onSelect={(newVal) => setValue({ ...value, byte: newVal })}
          side="left"
          disabled={false}
        />
        <EditTransformerOptions
          transformer={byteTransformer}
          value={value.byte}
          onSubmit={(newVal) => setValue({ ...value, byte: newVal })}
          disabled={isInvalidTransformer(byteTransformer)}
        />
      </div>
      <div>
        <FormLabel>Boolean</FormLabel>
        <FormDescription>
          Set a default transformer for any unmapped Boolean fields
        </FormDescription>
        <TransformerSelect
          getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
          buttonText="Select Transformer"
          value={value.boolean}
          onSelect={(newVal) => setValue({ ...value, boolean: newVal })}
          side="left"
          disabled={false}
        />
        <EditTransformerOptions
          transformer={boolTransformer}
          value={value.boolean}
          onSubmit={(newVal) => setValue({ ...value, boolean: newVal })}
          disabled={isInvalidTransformer(boolTransformer)}
        />
      </div>
      <div>
        <FormLabel>Number</FormLabel>
        <FormDescription>
          Set a default transformer for any unmapped Number fields
        </FormDescription>
        <TransformerSelect
          getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
          buttonText="Select Transformer"
          value={value.n}
          onSelect={(newVal) => setValue({ ...value, n: newVal })}
          side="left"
          disabled={false}
        />
        <EditTransformerOptions
          transformer={numTransformer}
          value={value.n}
          onSubmit={(newVal) => setValue({ ...value, n: newVal })}
          disabled={isInvalidTransformer(numTransformer)}
        />
      </div>
      <div>
        <FormLabel>String</FormLabel>
        <FormDescription>
          Set a default transformer for any unmapped String fields
        </FormDescription>
        <TransformerSelect
          getTransformers={() => transformerHandler.getTransformers()} // todo: filter this by type
          buttonText="Select Transformer"
          value={value.s}
          onSelect={(newVal) => setValue({ ...value, s: newVal })}
          side="left"
          disabled={false}
        />
        <EditTransformerOptions
          transformer={strTransformer}
          value={value.s}
          onSubmit={(newVal) => setValue({ ...value, s: newVal })}
          disabled={isInvalidTransformer(strTransformer)}
        />
      </div>
    </>
  );
}

function isInvalidTransformer(transformer: Transformer): boolean {
  return transformer.source === TransformerSource.UNSPECIFIED;
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
      new SystemTransformer()
    );
  }
  return (
    handler.getSystemTransformerBySource(value.source) ??
    new SystemTransformer()
  );
}
