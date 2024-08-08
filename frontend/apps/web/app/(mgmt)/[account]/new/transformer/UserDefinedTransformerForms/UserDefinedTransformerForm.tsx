import { PlainMessage } from '@bufbuild/protobuf';
import { TransformerConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import { FieldErrors } from 'react-hook-form';
import UserDefinedGenerateCardNumberForm from './UserDefinedGenerateCardNumber';
import UserDefinedGenerateCategoricalForm from './UserDefinedGenerateCategoricalForm';
import UserDefinedGenerateEmailForm from './UserDefinedGenerateEmailForm';
import UserDefinedGenerateFloat64Form from './UserDefinedGenerateFloat64Form';
import UserDefinedGenerateGenderForm from './UserDefinedGenerateGenderForm';
import UserDefinedGenerateInt64Form from './UserDefinedGenerateInt64Form';
import UserDefinedGenerateInternationalPhoneNumberForm from './UserDefinedGenerateInternationalPhoneNumberForm';
import UserDefinedGenerateJavascriptForm from './UserDefinedGenerateJavascriptForm';
import UserDefinedGenerateStateForm from './UserDefinedGenerateStateForm';
import UserDefinedGenerateStringForm from './UserDefinedGenerateStringForm';
import UserDefinedGenerateStringPhoneNumberNumberForm from './UserDefinedGenerateStringPhoneNumberForm';
import UserDefinedGenerateUuidForm from './UserDefinedGenerateUuidForm';
import UserDefinedTransformCharacterScrambleForm from './UserDefinedTransformCharacterScrambleForm';
import UserDefinedTransformE164NumberForm from './UserDefinedTransformE164PhoneNumberForm';
import UserDefinedTransformEmailForm from './UserDefinedTransformEmailForm';
import UserDefinedTransformFirstNameForm from './UserDefinedTransformFirstNameForm';
import UserDefinedTransformFloat64Form from './UserDefinedTransformFloat64Form';
import UserDefinedTransformFullNameForm from './UserDefinedTransformFullNameForm';
import UserDefinedTransformInt64Form from './UserDefinedTransformInt64Form';
import UserDefinedTransformIntPhoneNumberForm from './UserDefinedTransformInt64PhoneForm';
import UserDefinedTransformJavascriptForm from './UserDefinedTransformJavascriptForm';
import UserDefinedTransformLastNameForm from './UserDefinedTransformLastNameForm';
import UserDefinedTransformPhoneNumberForm from './UserDefinedTransformPhoneNumberForm';
import UserDefinedTransformStringForm from './UserDefinedTransformStringForm';

interface Props {
  value: TransformerConfig;
  setValue(newValue: TransformerConfig): void;
  disabled?: boolean;

  errors?: FieldErrors<PlainMessage<TransformerConfig>>;
}
// handles rendering custom transformer configs
export function UserDefinedTransformerForm(props: Props): ReactElement {
  const { value, disabled, setValue, errors } = props;
  const valConfig = value.config; // de-refs so that typescript is able to keep the conditional typing as it doesn't work well if you keep it on value itself

  switch (valConfig.case) {
    case 'generateCardNumberConfig':
      return (
        <UserDefinedGenerateCardNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateCategoricalForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateInternationalPhoneNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateFloat64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateGenderForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateInt64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateStringForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateStringPhoneNumberNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateStateForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateUuidForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformE164NumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformEmailForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateEmailForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformFirstNameForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformFloat64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformFullNameForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformInt64Form
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformIntPhoneNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformLastNameForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformPhoneNumberForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformStringForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformJavascriptForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedTransformCharacterScrambleForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
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
        <UserDefinedGenerateJavascriptForm
          value={valConfig.value}
          setValue={(newVal) =>
            setValue(
              new TransformerConfig({
                config: { case: valConfig.case, value: newVal },
              })
            )
          }
          isDisabled={disabled}
          errors={errors?.config?.value}
        />
      );
    default:
      <div>No transformer found</div>;
  }
  return <div></div>;
}
