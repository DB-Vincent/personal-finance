import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import common from "./en/common.json";

i18n.use(initReactI18next).init({
  resources: {
    en: { common },
  },
  lng: "en",
  defaultNS: "common",
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
