const name: string = window.NulYun.Name || "nulyun";
const disableExternal: boolean = window.NulYun.DisableExternal;
const disableUsedPercentage: boolean = window.NulYun.DisableUsedPercentage;
const baseURL: string = window.NulYun.BaseURL;
const staticURL: string = window.NulYun.StaticURL;
const recaptcha: string = window.NulYun.ReCaptcha;
const recaptchaKey: string = window.NulYun.ReCaptchaKey;
const signup: boolean = window.NulYun.Signup;
const version: string = window.NulYun.Version;
const logoURL = `${staticURL}/img/logo.svg`;
const noAuth: boolean = window.NulYun.NoAuth;
const authMethod = window.NulYun.AuthMethod;
const logoutPage: string = window.NulYun.LogoutPage;
const loginPage: boolean = window.NulYun.LoginPage;
const theme: UserTheme = window.NulYun.Theme;
const enableThumbs: boolean = window.NulYun.EnableThumbs;
const resizePreview: boolean = window.NulYun.ResizePreview;
const tusSettings = window.NulYun.TusSettings;
const origin = window.location.origin;
const tusEndpoint = `/api/tus`;
const hideLoginButton = window.NulYun.HideLoginButton;

export {
  name,
  disableExternal,
  disableUsedPercentage,
  baseURL,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
  version,
  noAuth,
  authMethod,
  logoutPage,
  loginPage,
  theme,
  enableThumbs,
  resizePreview,
  tusSettings,
  origin,
  tusEndpoint,
  hideLoginButton,
};
