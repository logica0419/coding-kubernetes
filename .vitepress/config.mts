import { type DefaultTheme, type UserConfig, defineConfig } from "vitepress";
import { withSidebar } from "vitepress-sidebar";
import type { VitePressSidebarOptions } from "vitepress-sidebar/types";

const locales = {
  root: {
    label: "English",
    lang: "en",
  },
  ja: {
    label: "日本語",
    lang: "ja",
  },
};

const config: UserConfig<DefaultTheme.Config> = {
  title: "Coding Kubernetes",
  description:
    "Implementing a minimal Kubernetes from scratch. CRI / CNI implementation included.",
  srcDir: "docs",
  locales: locales,
  rewrites: {
    "en/:content*": ":content*",
  },
  themeConfig: {
    i18nRouting: true,
    nav: [{ text: "Home", link: "/" }],
    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/logica0419/coding-kubernetes",
      },
    ],
  },
};

const sidebarConfigs: VitePressSidebarOptions[] = Object.entries(locales).map(
  (locale) => {
    return {
      ...(locale[0] === "root" ? {} : { basePath: `/${locale[1].lang}/` }),
      documentRootPath: `/docs/${locale[1].lang}`,
      resolvePath: locale[0] === "root" ? "/" : `/${locale[1].lang}/`,
      collapsed: true,
      useTitleFromFileHeading: true,
      useFolderTitleFromIndexFile: true,
      useFolderLinkFromIndexFile: true,
    };
  },
);

export default defineConfig(withSidebar(config, sidebarConfigs));
