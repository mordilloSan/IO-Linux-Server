import React from "react";
import styled from "@emotion/styled";
import { NavLink, NavLinkProps } from "react-router-dom";

import { Box, BoxProps, Drawer as MuiDrawer } from "@mui/material";

import { ReactComponent as Logo } from "@/vendor/logo.svg";
import { SidebarItemsType } from "@/types/sidebar";
import SidebarNav from "./SidebarNav";

const Drawer = styled(MuiDrawer)`
  > div {
    border-right: 0;
    scrollbar-width: none;
  }
`;
type BrandProps = BoxProps & NavLinkProps;
const Brand = styled(Box)<BrandProps>`
  display: flex;
  align-items: center;
  justify-content: center;
  padding: ${(props) => props.theme.spacing(0, 6)};
  min-height: 64px;
  font-size: ${(props) => props.theme.typography.h5.fontSize};
  font-weight: ${(props) => props.theme.typography.fontWeightMedium};
  color: ${(props) => props.theme.palette.text.primary};
  background-color: ${(props) => props.theme.palette.background.default};
  text-decoration: none;
  transition: none;
`;

const BrandIcon = styled(Logo)`
  margin-right: ${(props) => props.theme.spacing(2)};
  color: ${(props) => props.theme.palette.primary.main};
  fill: ${(props) => props.theme.palette.primary.main};
  width: 32px;
  height: 32px;
`;

export type SidebarProps = {
  PaperProps: {
    style: {
      width: number;
    };
  };
  variant?: "permanent" | "persistent" | "temporary";
  open?: boolean;
  onClose?: () => void;
  items: {
    title: string;
    pages: SidebarItemsType[];
  }[];
};

const Sidebar: React.FC<SidebarProps> = ({ items, ...rest }) => {
  return (
    <Drawer variant="permanent" {...rest}>
      <Brand component={NavLink} to="/">
        <BrandIcon />
        <Box ml={1}>Linux Server</Box>
      </Brand>
      <SidebarNav items={items} />
    </Drawer>
  );
};

export default Sidebar;
