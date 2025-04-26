import {
  Home,
  RefreshCcw,
  ServerCog,
  HardDrive,
  Users,
  Share2,
  Cpu,
  Folder,
  Network,
} from "lucide-react";

import { Icon } from "@iconify/react";

import { SidebarItemsType } from "@/types/sidebar";

const navItems: SidebarItemsType[] = [
  {
    href: "/",
    icon: Home,
    title: "Dashboard",
  },
  {
    href: "/network",
    icon: Network,
    title: "Network",
  },
  {
    href: "/updates",
    icon: RefreshCcw,
    title: "Updates",
  },
  {
    href: "/services",
    icon: ServerCog,
    title: "Services",
  },
  {
    href: "/storage",
    icon: HardDrive,
    title: "Storage",
  },
  {
    href: "/docker",
    icon: () => <Icon icon="fa-brands:docker" />,
    title: "Docker",
  },
  {
    href: "/accounts",
    icon: Users,
    title: "Accounts",
  },
  {
    href: "/shares",
    icon: Share2,
    title: "Shares",
  },
  {
    href: "/wireguard",
    icon: () => <Icon icon="cib:wireguard" width="48" height="48" />,
    title: "Wireguard",
  },
  {
    href: "/hardware",
    icon: Cpu,
    title: "Hardware",
  },
  {
    href: "/filebrowser",
    icon: Folder,
    title: "Navigator",
  },
];

export default navItems;
