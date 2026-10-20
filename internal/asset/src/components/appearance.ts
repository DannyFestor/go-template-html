const themes = ["light", "dark", "system"];

export function appearance() {
  return {
    theme: "",
    choose(theme: string): void {
      this.theme = theme;
      document.documentElement.classList.remove(...themes);
      document.documentElement.classList.add(theme);
    },
    pressed(theme: string): boolean {
      return this.theme === theme;
    },
  };
}
