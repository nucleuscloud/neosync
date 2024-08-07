import { TransformerConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import UserDefinedGenerateCardNumberForm from './UserDefinedGenerateCardNumber';
import UserDefinedGenerateCategoricalForm from './UserDefinedGenerateCategoricalForm';

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
    // case TransformerSource.GENERATE_E164_PHONE_NUMBER:
    //   return (
    //     <UserDefinedGenerateInternationalPhoneNumberForm
    //       isDisabled={disabled}
    //     />
    //   );
    // case TransformerSource.GENERATE_FLOAT64:
    //   return <UserDefinedGenerateFloat64Form isDisabled={disabled} />;
    // case TransformerSource.GENERATE_GENDER:
    //   return <UserDefinedGenerateGenderForm isDisabled={disabled} />;
    // case TransformerSource.GENERATE_INT64:
    //   return <UserDefinedGenerateInt64Form isDisabled={disabled} />;
    // case TransformerSource.GENERATE_RANDOM_STRING:
    //   return <UserDefinedGenerateStringForm isDisabled={disabled} />;
    // case TransformerSource.GENERATE_STRING_PHONE_NUMBER:
    //   return (
    //     <UserDefinedGenerateStringPhoneNumberNumberForm isDisabled={disabled} />
    //   );
    // case TransformerSource.GENERATE_STATE:
    //   return <UserDefinedGenerateStateForm isDisabled={disabled} />;
    // case TransformerSource.GENERATE_UUID:
    //   return <UserDefinedGenerateUuidForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_E164_PHONE_NUMBER:
    //   return <UserDefinedTransformE164NumberForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_EMAIL:
    //   return <UserDefinedTransformEmailForm isDisabled={disabled} />;
    // case TransformerSource.GENERATE_EMAIL:
    //   return <UserDefinedGenerateEmailForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_FIRST_NAME:
    //   return <UserDefinedTransformFirstNameForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_FLOAT64:
    //   return <UserDefinedTransformFloat64Form isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_FULL_NAME:
    //   return <UserDefinedTransformFullNameForm isDisabled={disabled} />;
    // case TransformerSource.TRANSFORM_INT64:
    //   return <UserDefinedTransformInt64Form isDisabled={disabled} />;
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
