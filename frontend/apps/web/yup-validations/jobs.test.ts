/**
 * @jest-environment node
 */

import { JobMappingTransformerForm } from './jobs';

// if we wanted to be very thorough then we would test every transformerconfig ...
describe('JobMappingTransformerForm schema', () => {
  it('should pass validation with valid data', async () => {
    const validData = {
      source: 1,
      config: { case: 'generateGenderConfig', value: { abbreviate: true } },
    };

    await expect(
      JobMappingTransformerForm.isValid(validData)
    ).resolves.toBeTruthy();
  });

  it('should fail validation with invalid config', async () => {
    const invalidData = {
      source: 1,
      config: { case: 'generateGenderConfig', value: { abbreviate: 73 } },
    };

    await expect(
      JobMappingTransformerForm.isValid(invalidData)
    ).resolves.toBeFalsy();
  });

  it('should pass validation with a valid config', async () => {
    const validData = {
      source: 1,
      config: { case: 'generateGenderConfig', value: { abbreviate: true } },
    };

    await expect(JobMappingTransformerForm.isValid(validData)).resolves.toBe(
      true
    );
  });
});
