import { afterEach, describe, expect, it } from "vite-plus/test";

import { nav } from "./nav";

afterEach(() => {
  history.replaceState(null, "", "/");
});

describe("active", () => {
  it("matches the current path and the paths below it", () => {
    history.replaceState(null, "", "/settings/profile");
    const component = nav();

    expect(component.active("/settings")).toBe(true);
    expect(component.active("/settings/profile")).toBe(true);
  });

  it("does not match a path that only shares a prefix", () => {
    history.replaceState(null, "", "/settings-old");
    const component = nav();

    expect(component.active("/settings")).toBe(false);
  });
});

describe("current", () => {
  it("marks only the link to the current path as the current page", () => {
    history.replaceState(null, "", "/settings/profile");
    const component = nav();

    expect(component.current("/settings/profile")).toBe("page");
    expect(component.current("/settings")).toBe(false);
  });
});

describe("after boosted navigation", () => {
  it("follows the path htmx pushed", () => {
    const component = nav();
    component.init();

    history.pushState(null, "", "/dashboard");
    document.dispatchEvent(new CustomEvent("htmx:after:history:push"));

    expect(component.current("/dashboard")).toBe("page");
    component.destroy();
  });

  it("stops following once destroyed", () => {
    const component = nav();
    component.init();
    component.destroy();

    history.pushState(null, "", "/dashboard");
    document.dispatchEvent(new CustomEvent("htmx:after:history:push"));

    expect(component.current("/dashboard")).toBe(false);
  });
});
