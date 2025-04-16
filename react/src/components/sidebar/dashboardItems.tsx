import { SidebarItemsType } from "@/types/sidebar";

import { Grid, Layout, ShoppingCart, Sliders, Settings } from "lucide-react";

const pagesSection = [
  {
    href: "/",
    icon: Sliders,
    title: "Dashboard",
  },
  {
    href: "/settings",
    icon: Settings,
    title: "Settings",
  },
  {
    href: "/blank",
    icon: Layout,
    title: "Blank Page",
  },
  {
    href: "/orders",
    icon: ShoppingCart,
    title: "Orders",
  },
] as SidebarItemsType[];

const elementsSection = [
  {
    href: "/components",
    icon: Grid,
    title: "Components",
  },
] as SidebarItemsType[];

const navItems = [
  {
    title: "Pages",
    pages: pagesSection,
  },
  {
    title: "Elements",
    pages: elementsSection,
  },
];

export default navItems;
