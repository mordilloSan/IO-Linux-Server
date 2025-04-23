export type Update = {
  name: string;
  version: string;
  severity: string;
  changelog?: string;
  packages?: string[];
};
