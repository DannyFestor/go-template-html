// htmx swaps only <main> on boosted navigation, so the sidebar outlives the page and has to follow the URL itself.
const historyPushEvent = "htmx:after:history:push";

export function nav() {
  const listening = new AbortController();

  return {
    path: location.pathname,
    init(): void {
      document.addEventListener(
        historyPushEvent,
        () => {
          this.path = location.pathname;
        },
        { signal: listening.signal },
      );
    },
    destroy(): void {
      listening.abort();
    },
    active(prefix: string): boolean {
      return this.path === prefix || this.path.startsWith(`${prefix}/`);
    },
    current(pathname: string): "page" | false {
      return this.path === pathname ? "page" : false;
    },
  };
}
