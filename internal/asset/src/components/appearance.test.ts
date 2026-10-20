import { afterEach, describe, expect, it } from "vite-plus/test";

import { appearance } from "./appearance";

afterEach(() => {
  document.documentElement.className = "";
});

describe("choose", () => {
  it("puts the chosen theme on the document in place of the previous one", () => {
    document.documentElement.className = "h-full system";
    const component = appearance();

    component.choose("dark");

    expect([...document.documentElement.classList]).toEqual(["h-full", "dark"]);
  });

  it("marks only the chosen theme as pressed", () => {
    const component = appearance();

    component.choose("light");

    expect(component.pressed("light")).toBe(true);
    expect(component.pressed("dark")).toBe(false);
    expect(component.pressed("system")).toBe(false);
  });
});
