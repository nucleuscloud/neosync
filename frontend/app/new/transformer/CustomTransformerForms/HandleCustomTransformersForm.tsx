import { ReactElement } from 'react';
import CustomEmailTransformerForm from './CustomEmailTransformerForm';
import CustomFirstNameTransformerForm from './CustomFirstnameTransformerForm';
import CustomFullNameTransformerForm from './CustomFullnameTransformerForm';
import CustomGenderTransformerForm from './CustomGenderTransformerForm';
import CustomIntPhoneNumberTransformerForm from './CustomIntPhoneNumberTransformerForm';
import CustomLastNameTransformerForm from './CustomLastnameTransformerForm';
import CustomPhoneNumberTransformerForm from './CustomPhoneNumberTransformerForm';
import CustomRandomFloatTransformerForm from './CustomRandomFloatTransformerForm';
import CustomRandomIntTransformerForm from './CustomRandomIntTransformerForm';
import CustomRandomStringTransformerForm from './CustomRandomStringTransformerForm';
import CustomUuidTransformerForm from './CustomUuidTransformerForm';

// handles rendering custom tranformer configs
export function handleCustomTransformerForm(
  value: string | undefined
): ReactElement {
  switch (value) {
    case 'email':
      return <CustomEmailTransformerForm />;
    case 'uuid':
      return <CustomUuidTransformerForm />;
    case 'first_name':
      return <CustomFirstNameTransformerForm />;
    case 'last_name':
      return <CustomLastNameTransformerForm />;
    case 'full_name':
      return <CustomFullNameTransformerForm />;
    case 'phone_number':
      return <CustomPhoneNumberTransformerForm />;
    case 'int_phone_number':
      return <CustomIntPhoneNumberTransformerForm />;
    case 'random_string':
      return <CustomRandomStringTransformerForm />;
    case 'random_int':
      return <CustomRandomIntTransformerForm />;
    case 'random_float':
      return <CustomRandomFloatTransformerForm />;
    case 'gender':
      return <CustomGenderTransformerForm />;
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}
