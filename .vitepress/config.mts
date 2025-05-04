import { type DefaultTheme, type UserConfig, defineConfig } from "vitepress";
import { withSidebar } from "vitepress-sidebar";
import type { VitePressSidebarOptions } from "vitepress-sidebar/types";

const config: UserConfig<DefaultTheme.Config> = {
  title: "Coding Kubernetes",
  description:
    "Implementing a minimal Kubernetes from scratch. CRI / CNI implementation included.",
  srcDir: "docs",
  themeConfig: {
    i18nRouting: true,
    nav: [
      { text: "Home", link: "/" },
      { text: "Examples", link: "/markdown-examples" },
    ],
    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/logica0419/coding-kubernetes",
      },
    ],
  },
};

const sidebarConfig: VitePressSidebarOptions = {
  documentRootPath: "docs",
  collapsed: true,
  useTitleFromFileHeading: true,
};

export default defineConfig(withSidebar(config, sidebarConfig));
