// Match Redoc's standalone hydration entry point, but bundle its browser library
// with our locked dependencies instead of shipping its prebuilt sanitizer.
import { createElement } from 'react';
import { hydrateRoot } from 'react-dom/client';
import { configure } from 'mobx';
import { AppStore, Redoc } from 'redoc';

configure({ useProxies: 'ifavailable' });

export function hydrate(state, element, onRecoverableError) {
  const store = AppStore.fromJS(state);
  setTimeout(() => {
    hydrateRoot(element, createElement(Redoc, { store }), { onRecoverableError });
  }, 0);
}
