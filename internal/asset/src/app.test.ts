import { expect, it } from "vite-plus/test";

it("runs colocated tests against a DOM", () => {
  document.body.innerHTML = "<main>Hello</main>";

  expect(document.querySelector("main")?.textContent).toBe("Hello");
});
