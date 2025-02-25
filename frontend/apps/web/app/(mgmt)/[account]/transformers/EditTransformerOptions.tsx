import LearnMoreLink from '@/components/labels/LearnMoreLink';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import {
  isSystemTransformer,
  isUserDefinedTransformer,
  Transformer,
} from '@/shared/transformers';
import {
  getTransformerDataTypesString,
  getTransformerSourceString,
} from '@/util/util';
import {
  convertTransformerConfigSchemaToTransformerConfig,
  convertTransformerConfigToForm,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  TransformerConfig,
  TransformerConfigSchema,
  TransformerSource,
  TransformersService,
} from '@neosync/sdk';
import {
  EyeOpenIcon,
  MixerHorizontalIcon,
  Pencil1Icon,
} from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import {
  EditJobMappingTransformerConfigFormContext,
  EditJobMappingTransformerConfigFormValues,
} from '../../../../yup-validations/transformer-validations';
import TransformerForm from '../new/transformer/TransformerForms/TransformerForm';

interface Props {
  transformer: Transformer;
  disabled: boolean;
  value: JobMappingTransformerForm;
  onSubmit(newValue: JobMappingTransformerForm): void;
}

// Note: this has issues with re-rendering due to being embedded within the tanstack table.
// This will cause the sheet to close when the user clicks back onto the page.
// This is partially solved by memoizing the tanstack columns, but any time the columns need to re-render, this sheet will close if it's open.
export default function EditTransformerOptions(props: Props): ReactElement<any> {
  const { transformer, disabled, value, onSubmit } = props;
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleClose = () => {
    setIsDialogOpen(false);
  };

  return (
    <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          disabled={disabled}
          className="h-[36px] lg:flex"
          type="button"
          onClick={() => setIsDialogOpen(true)}
        >
          {isUserDefinedTransformer(transformer) ? (
            <EyeOpenIcon />
          ) : (
            <Pencil1Icon />
          )}
        </Button>
      </DialogTrigger>
      <DialogContent
        className="max-w-3xl"
        onPointerDownOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <div className="flex flex-row w-full">
            <div className="flex flex-col space-y-2 w-full">
              <div className="flex flex-row justify-between items-center">
                <div className="flex flex-row gap-4">
                  <DialogTitle className="text-xl">
                    {transformer.name}
                  </DialogTitle>
                  <Badge variant="outline">
                    {getTransformerDataTypesString(transformer.dataTypes)}
                  </Badge>
                </div>
              </div>
              <div className="flex flex-row items-center gap-2">
                <DialogDescription>
                  {transformer?.description}{' '}
                  <LearnMoreLink href={constructDocsLink(transformer.source)} />
                </DialogDescription>
              </div>
            </div>
          </div>
          <Separator />
        </DialogHeader>
        <div className="pt-4">
          {transformer && (
            <EditTransformerConfig
              value={
                isSystemTransformer(transformer)
                  ? convertTransformerConfigSchemaToTransformerConfig(
                      value.config
                    )
                  : (transformer.config ?? create(TransformerConfigSchema))
              }
              onSubmit={(newval) => {
                onSubmit({
                  ...value,
                  config: convertTransformerConfigToForm(newval),
                });
                handleClose();
              }}
              isDisabled={disabled || isUserDefinedTransformer(transformer)}
              onClose={handleClose}
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

interface EditTransformerConfigProps {
  value: TransformerConfig;
  onSubmit(newValue: TransformerConfig): void;
  isDisabled: boolean;
  onClose: () => void;
}

function EditTransformerConfig(
  props: EditTransformerConfigProps
): ReactElement<any> {
  const { value, onSubmit, isDisabled, onClose } = props;

  const { account } = useAccount();
  const { mutateAsync: isJavascriptCodeValid } = useMutation(
    TransformersService.method.validateUserJavascriptCode
  );

  const form = useForm<
    EditJobMappingTransformerConfigFormValues,
    EditJobMappingTransformerConfigFormContext
  >({
    mode: 'onChange',
    resolver: yupResolver(EditJobMappingTransformerConfigFormValues),
    defaultValues: { config: convertTransformerConfigToForm(value) },
    context: {
      accountId: account?.id ?? '',
      isUserJavascriptCodeValid: isJavascriptCodeValid,
    },
  });
  return (
    <Form {...form}>
      <div className="flex flex-col gap-8">
        <FormField
          control={form.control}
          name="config"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <TransformerForm
                  value={convertTransformerConfigSchemaToTransformerConfig(
                    field.value
                  )}
                  setValue={(newValue) => {
                    field.onChange(convertTransformerConfigToForm(newValue));
                  }}
                  disabled={isDisabled}
                  errors={form.formState.errors}
                  NoConfigComponent={<NoAdditionalTransformerConfigurations />}
                />
              </FormControl>
            </FormItem>
          )}
        />
        <div className="flex justify-between w-full">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>

          <Button
            type="button"
            disabled={isDisabled || !form.formState.isValid}
            onClick={(e) => {
              form.handleSubmit((formValues) => {
                onSubmit(
                  convertTransformerConfigSchemaToTransformerConfig(
                    formValues.config
                  )
                );
              })(e);
            }}
          >
            Save
          </Button>
        </div>
      </div>
    </Form>
  );
}

function NoAdditionalTransformerConfigurations(): ReactElement<any> {
  return (
    <Alert className="border-gray-200 dark:border-gray-700 shadow-xs">
      <div className="flex flex-row items-center gap-4">
        <MixerHorizontalIcon className="h-4 w-4" />
        <AlertDescription className="text-gray-500">
          There are no additional configurations for this Transformer
        </AlertDescription>
      </div>
    </Alert>
  );
}

export function constructDocsLink(source: TransformerSource): string {
  const name = getTransformerSourceString(source).replaceAll('_', '-');

  if (
    source == TransformerSource.GENERATE_JAVASCRIPT ||
    source == TransformerSource.TRANSFORM_JAVASCRIPT
  ) {
    return `https://docs.neosync.dev/guides/custom-code-transformers`;
  } else {
    return `https://docs.neosync.dev/transformers/system#${name}`;
  }
}
