import { expect, it } from "vite-plus/test";

import { disclosure } from "./disclosure";

it("starts closed", () => {
  expect(disclosure().open).toBe(false);
});

it("toggles open and closed", () => {
  const component = disclosure();

  component.toggle();
  expect(component.open).toBe(true);

  component.toggle();
  expect(component.open).toBe(false);
});

it("closes when open", () => {
  const component = disclosure();
  component.toggle();

  component.close();

  expect(component.open).toBe(false);
});
