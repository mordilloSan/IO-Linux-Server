import React from "react";
import { List } from "@mui/material";
import { SidebarItemsType } from "@/types/sidebar";
import SidebarNavListItem from "./SidebarNavListItem";

type SidebarNavProps = {
  items: SidebarItemsType[];
  collapsed?: boolean;
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
