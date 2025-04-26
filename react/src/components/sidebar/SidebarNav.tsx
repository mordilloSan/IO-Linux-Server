import React from "react";
import styled from "@emotion/styled";
import { List } from "@mui/material";

import { SidebarItemsType } from "@/types/sidebar";
import SidebarNavListItem from "./SidebarNavListItem";

const Wrapper = styled.div`
  background-color: ${(props) => props.theme.sidebar.background};
  flex-grow: 1;
`;

const Items = styled.div`
  padding-top: ${(props) => props.theme.spacing(2.5)};
  padding-bottom: ${(props) => props.theme.spacing(2.5)};
`;

type SidebarNavProps = {
  items: SidebarItemsType[]; // ✅ flat array
};

const SidebarNav: React.FC<SidebarNavProps> = ({ items }) => {
  return (
    <Wrapper>
      <List disablePadding>
        <Items>
          {items.map((page) => (
            <SidebarNavListItem
              key={page.title}
              href={page.href}
              icon={page.icon || (() => null)}
              title={page.title}
            />
          ))}
        </Items>
      </List>
    </Wrapper>
  );
};

export default SidebarNav;
