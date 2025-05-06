import { List } from "@mui/material";
import React from "react";

import SidebarNavListItem from "./SidebarNavListItem";

import { SidebarItemsType } from "@/types/sidebar";

type SidebarNavProps = {
  items: SidebarItemsType[];
  collapsed: boolean; // Make sure this is aware of the collapsed state
};

const SidebarNav: React.FC<SidebarNavProps> = ({
  items,
  collapsed = false,
}) => {
  return (
    <List disablePadding>
      {items.map((page) => (
        <SidebarNavListItem
          key={page.title}
          href={page.href}
          icon={page.icon || (() => null)}
          title={page.title}
          collapsed={collapsed}
        />
      ))}
    </List>
  );
};

export default SidebarNav;
