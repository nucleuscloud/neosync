/**
 * @jest-environment jsdom
 */

import '@testing-library/jest-dom';
import { fireEvent, screen } from '@testing-library/react';

import { composeStories } from '@storybook/nextjs';

import * as stories from './PasswordComponent.stories'; // 👈 Our stories imported here.

const story = composeStories(stories);
test('Checks if the form is valid', async () => {
  // Renders the composed story
  await story.Default.run();

  const passwordInput = screen.getByLabelText('password', {
    selector: 'input',
  });

  fireEvent.change(passwordInput, {
    target: { value: 'topsecret' },
  });

  fireEvent.click(screen.getByRole('button'));
  expect(screen.getByRole('textbox')).toBeInTheDocument();
  fireEvent.click(screen.getByRole('button'));
  expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
});
