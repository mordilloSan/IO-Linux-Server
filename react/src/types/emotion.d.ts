import "@emotion/react";

declare module "@emotion/react" {
  /* eslint-disable @typescript-eslint/no-empty-interface */
  export interface Theme extends CustomTheme {}
  /* eslint-enable @typescript-eslint/no-empty-interface */
}
