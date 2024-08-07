import { TransformerConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import UserDefinedGenerateCardNumberForm from './UserDefinedGenerateCardNumber';
import UserDefinedGenerateCategoricalForm from './UserDefinedGenerateCategoricalForm';
import UserDefinedGenerateEmailForm from './UserDefinedGenerateEmailForm';
import UserDefinedGenerateFloat64Form from './UserDefinedGenerateFloat64Form';
import UserDefinedGenerateGenderForm from './UserDefinedGenerateGenderForm';
import UserDefinedGenerateInt64Form from './UserDefinedGenerateInt64Form';
import UserDefinedGenerateInternationalPhoneNumberForm from './UserDefinedGenerateInternationalPhoneNumberForm';
import UserDefinedGenerateStateForm from './UserDefinedGenerateStateForm';
import UserDefinedGenerateStringForm from './UserDefinedGenerateStringForm';
import UserDefinedGenerateStringPhoneNumberNumberForm from './UserDefinedGenerateStringPhoneNumberForm';
import UserDefinedGenerateUuidForm from './UserDefinedGenerateUuidForm';
import UserDefinedTransformE164NumberForm from './UserDefinedTransformE164PhoneNumberForm';
import UserDefinedTransformEmailForm from './UserDefinedTransformEmailForm';
import UserDefinedTransformFirstNameForm from './UserDefinedTransformFirstNameForm';
import UserDefinedTransformFloat64Form from './UserDefinedTransformFloat64Form';
import UserDefinedTransformFullNameForm from './UserDefinedTransformFullNameForm';
import UserDefinedTransformInt64Form from './UserDefinedTransformInt64Form';

interface Props {
  value: TransformerConfig;
  setValue(newValue: TransformerConfig): void;
  disabled: boolean;
}
// handles rendering custom transformer configs
export function UserDefinedTransformerForm(props: Props): ReactElement {
  const { value, disabled, setValue } = props;
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
        />
      );
    // case TransformerSource.TRANSFORM_INT64_PHONE_NUMBER:
    //   return <UserDefinedTransformIntPhoneNumberForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_LAST_NAME:
    //   return <UserDefinedTransformLastNameForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_PHONE_NUMBER:
    //   return <UserDefinedTransformPhoneNumberForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_STRING:
    //   return <UserDefinedTransformStringForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_JAVASCRIPT:
    //   return <UserDefinedTransformJavascriptForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_CHARACTER_SCRAMBLE:
    //   return (
    //     <UserDefinedTransformCharacterScrambleForm isDisabled={disabled} />
    //   );
    // case TransformerSource.GENERATE_JAVASCRIPT:
    //   return <UserDefinedGenerateJavascriptForm isDisabled={disabled} />;
    default:
      <div>No transformer found</div>;
  }
  return <div></div>;
}
