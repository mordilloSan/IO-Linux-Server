import React, { useContext } from "react";
import styled from "@emotion/styled";
import { NavLink, useLocation } from "react-router-dom";
import { ListItemButton, ListItemIcon, ListItemText, ListItemButtonProps } from "@mui/material";
import { ThemeContext } from "@/contexts/ThemeContext";
import { darken, lighten } from "polished";
import { Icon } from "@iconify/react";

interface SidebarNavListItemProps extends ListItemButtonProps {
  href: string;
  title: string;
  icon?: React.ElementType | string;
}

const Item = styled(ListItemButton, {
  shouldForwardProp: (prop) => prop !== "activecolor",
})<{ activecolor: string }>`
  margin: ${(props) => props.theme.spacing(1)} ${(props) => props.theme.spacing(2)};
  padding: ${(props) => props.theme.spacing(1.5)} ${(props) => props.theme.spacing(3)};
  border-radius: 0 9999px 9999px 0;
  font-weight: ${(props) => props.theme.typography.fontWeightRegular};
  color: ${(props) => props.theme.sidebar.color};
  text-transform: none;
  width: auto;
  justify-content: flex-start;
  transition: background-color 0.3s, color 0.3s;

  svg {
    color: ${(props) => props.theme.sidebar.color};
    width: 26px;
    height: 26px;
    margin-right: ${(props) => props.theme.spacing(2)};
    transition: color 0.3s;
  }

  &.Mui-selected {
    background: linear-gradient(
      90deg,
      ${(props) => lighten(0.25, props.activecolor)} 0%,
      ${(props) => props.activecolor} 50%
    );
    color: #ffffff;

    svg {
      color: #ffffff;
    }

    span {
      color: #ffffff;
      font-weight: ${(props) => props.theme.typography.fontWeightMedium};
    }
  }

  &:hover {
  ${(props) => !props.selected && `
    background: ${props.theme.palette.mode === "light"
      ? darken(0.07, props.theme.header.background)
      : lighten(0.05, props.theme.sidebar.background)};
    box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  `}
}
`;

const Title = styled(ListItemText)`
  margin: 0;
  span {
    font-size: ${(props) => props.theme.typography.body1.fontSize}px;
    font-weight: ${(props) => props.theme.typography.fontWeightRegular};
    transition: color 0.3s;
  }
`;

const SidebarNavListItem: React.FC<SidebarNavListItemProps> = ({
  href,
  title,
  icon,
  ...rest
}) => {
  const { pathname } = useLocation();
  const { primaryColor } = useContext(ThemeContext);
  const fallbackPrimary = "#3f5efb";

  const isActive = pathname === href;

  const renderIcon = () => {
    if (!icon) return null;
    if (typeof icon === "string") {
      return <Icon icon={icon} width={24} height={24} />;
    }
    const IconComponent = icon as React.ElementType;
    return <IconComponent />;
  };

  return (
    <Item
    component={NavLink}
    activecolor={primaryColor || fallbackPrimary}
    selected={isActive}
    {...(rest as any)} // (temporary cast) or clean props manually
    // Pass the 'to' inside props properly
    to={href}
    >
      {icon && (
        <ListItemIcon
          sx={{
            minWidth: 0,
            mr: 2,
            justifyContent: "center",
            color: "inherit",
          }}
        >
          {renderIcon()}
        </ListItemIcon>
      )}
      <Title primary={title} />
    </Item>
  );
};

export default SidebarNavListItem;
