import { PropertyDict } from 'mixpanel';

interface Mixpanel {
  name: string;
  props: PropertyDict;
}

export function FireMixpanel(name: string, props: PropertyDict): void {
  const data: Mixpanel = {
    name: name,
    props: props,
  };

  fetch(`api/mixpanel`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
}
