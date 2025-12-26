export {};

declare global {
  interface Window {
    NulYun: any;
    grecaptcha: any;
  }

  interface IUserForm {
    storageQuota?: string; // Optional storage quota in human-readable format like "10M", "5G"
  }

  interface HTMLElement {
    // TODO: no idea what the exact type is
    __vue__: any;
  }
}
