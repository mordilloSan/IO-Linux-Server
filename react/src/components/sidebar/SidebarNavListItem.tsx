import React, { forwardRef } from "react";
import styled from "@emotion/styled";
import { NavLink, NavLinkProps } from "react-router-dom";
import { rgba, darken } from "polished";

import { ListItemProps, ListItemButton, ListItemText } from "@mui/material";

const CustomRouterLink = forwardRef<any, NavLinkProps>((props, ref) => (
  <div ref={ref}>
    <NavLink {...props} />
  </div>
));

CustomRouterLink.displayName = "CustomRouterLink";

type ItemType = {
  activeclassname?: string;
  onClick?: () => void;
  to?: string;
  component?: typeof NavLink;
};

const Item = styled(ListItemButton)<ItemType>`
  padding-top: ${(props) => props.theme.spacing(3)};
  padding-bottom: ${(props) => props.theme.spacing(3)};
  padding-left: ${(props) => props.theme.spacing(8)};
  padding-right: ${(props) => props.theme.spacing(7)};
  font-weight: ${(props) => props.theme.typography.fontWeightRegular};
  svg {
    color: ${(props) => props.theme.palette.text.secondary};
    font-size: 20px;
    width: 20px;
    height: 20px;
    opacity: 0.5;
  }

  &.${(props) => props.activeclassname} {
    background-color: ${(props) =>
      props.theme.palette.mode === "dark"
        ? darken(0.07, props.theme.palette.background.paper) // slightly darker in dark mode
        : darken(0.14, props.theme.palette.background.paper)};
    span {
      color: ${(props) => props.theme.palette.text.primary};
    }
  }
`;

const Title = styled(ListItemText)`
  margin: 0;
  span {
    color: ${(props) => rgba(props.theme.palette.text.primary, 1)};
    font-size: ${(props) => props.theme.typography.body1.fontSize}px;
    padding: 0 ${(props) => props.theme.spacing(4)};
  }
`;

type SidebarNavListItemProps = ListItemProps & {
  className?: string;
  href: string;
  icon: React.FC<any>;
  title: string;
};

const SidebarNavListItem: React.FC<SidebarNavListItemProps> = (props) => {
  const { title, href, icon: Icon } = props;

  return (
    <React.Fragment>
      <Item component={CustomRouterLink} to={href} activeclassname="active">
        {Icon && <Icon />}
        <Title>{title}</Title>
      </Item>
    </React.Fragment>
  );
};

export default SidebarNavListItem;
