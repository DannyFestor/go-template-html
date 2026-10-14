export function disclosure() {
  return {
    open: false,
    toggle(): void {
      this.open = !this.open;
    },
    close(): void {
      this.open = false;
    },
  };
}
