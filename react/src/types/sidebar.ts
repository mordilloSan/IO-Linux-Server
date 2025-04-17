import React from "react";

export type SidebarItemsType = {
  href: string;
  title: string;
  icon?: React.FC<any> | React.ElementType;
};