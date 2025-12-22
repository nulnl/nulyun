interface ISettings {
  signup: boolean;
  createUserDir: boolean;
  hideLoginButton: boolean;
  minimumPasswordLength: number;
  userHomeBasePath: string;
  defaults: SettingsDefaults;
  branding: SettingsBranding;
  tus: SettingsTus;
}

interface SettingsDefaults {
  scope: string;
  locale: string;
  viewMode: ViewModeType;
  singleClick: boolean;
  sorting: Sorting;
  perm: Permissions;
  hideDotfiles: boolean;
  dateFormat: boolean;
  aceEditorTheme: string;
}

interface SettingsBranding {
  name: string;
  disableExternal: boolean;
  disableUsedPercentage: boolean;
  files: string;
  theme: UserTheme;
  color: string;
}

interface SettingsTus {
  chunkSize: number;
  retryCount: number;
}
