import { create } from '@bufbuild/protobuf';
import { TransformerConfig, TransformerConfigSchema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { FieldErrors } from 'react-hook-form';
import GenerateCardNumberForm from './GenerateCardNumber';
import GenerateCategoricalForm from './GenerateCategoricalForm';
import GenerateCountryForm from './GenerateCountryForm';
import GenerateEmailForm from './GenerateEmailForm';
import GenerateFloat64Form from './GenerateFloat64Form';
import GenerateGenderForm from './GenerateGenderForm';
import GenerateInt64Form from './GenerateInt64Form';
import GenerateInternationalPhoneNumberForm from './GenerateInternationalPhoneNumberForm';
import GenerateIpAddressForm from './GenerateIpAddressForm';
import GenerateJavascriptForm from './GenerateJavascriptForm';
import GenerateStateForm from './GenerateStateForm';
import GenerateStringForm from './GenerateStringForm';
import GenerateStringPhoneNumberNumberForm from './GenerateStringPhoneNumberForm';
import GenerateUuidForm from './GenerateUuidForm';
import TransformCharacterScrambleForm from './TransformCharacterScrambleForm';
import TransformE164NumberForm from './TransformE164PhoneNumberForm';
import TransformEmailForm from './TransformEmailForm';
import TransformFirstNameForm from './TransformFirstNameForm';
import TransformFloat64Form from './TransformFloat64Form';
import TransformFullNameForm from './TransformFullNameForm';
import TransformInt64Form from './TransformInt64Form';
import TransformIntPhoneNumberForm from './TransformInt64PhoneForm';
import TransformJavascriptForm from './TransformJavascriptForm';
import TransformLastNameForm from './TransformLastNameForm';
import TransformPhoneNumberForm from './TransformPhoneNumberForm';
import TransformStringForm from './TransformStringForm';

interface Props {
  value: TransformerConfig;
  setValue(newValue: TransformerConfig): void;
  disabled?: boolean;

  errors?: FieldErrors<TransformerConfig>;

  NoConfigComponent?: ReactElement;
}
// handles rendering custom transformer configs
export default function TransformerForm(props: Props): ReactElement {
  const { value, disabled, setValue, errors, NoConfigComponent } = props;
  const valConfig = value.config; // de-refs so that typescript is able to keep the conditional typing as it doesn't work well if you keep it on value itself

  switch (valConfig.case) {
    case 'generateCardNumberConfig':
      return (
        <GenerateCardNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateCategoricalConfig':
      return (
        <GenerateCategoricalForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateE164PhoneNumberConfig':
      return (
        <GenerateInternationalPhoneNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateFloat64Config':
      return (
        <GenerateFloat64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateGenderConfig':
      return (
        <GenerateGenderForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateInt64Config':
      return (
        <GenerateInt64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateStringConfig':
      return (
        <GenerateStringForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateStringPhoneNumberConfig':
      return (
        <GenerateStringPhoneNumberNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateStateConfig':
      return (
        <GenerateStateForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateUuidConfig':
      return (
        <GenerateUuidForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformE164PhoneNumberConfig':
      return (
        <TransformE164NumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformEmailConfig':
      return (
        <TransformEmailForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateEmailConfig':
      return (
        <GenerateEmailForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformFirstNameConfig':
      return (
        <TransformFirstNameForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformFloat64Config':
      return (
        <TransformFloat64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformFullNameConfig':
      return (
        <TransformFullNameForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformInt64Config':
      return (
        <TransformInt64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformInt64PhoneNumberConfig':
      return (
        <TransformIntPhoneNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformLastNameConfig':
      return (
        <TransformLastNameForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformPhoneNumberConfig':
      return (
        <TransformPhoneNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformStringConfig':
      return (
        <TransformStringForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformJavascriptConfig':
      return (
        <TransformJavascriptForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'transformCharacterScrambleConfig':
      return (
        <TransformCharacterScrambleForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateJavascriptConfig':
      return (
        <GenerateJavascriptForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateCountryConfig':
      return (
        <GenerateCountryForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    case 'generateIpAddressConfig':
      return (
        <GenerateIpAddressForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              create(TransformerConfigSchema, {
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    default:
      return NoConfigComponent ?? <div />;
  }
}
