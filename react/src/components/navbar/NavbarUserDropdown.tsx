import React from "react";
import styled from "@emotion/styled";
import { useNavigate } from "react-router-dom";

import {
  Divider,
  Tooltip,
  Menu,
  MenuItem,
  IconButton as MuiIconButton,
} from "@mui/material";
import { spacing } from "@mui/system";

import useAuth from "@/hooks/useAuth";
import { LucidePower } from "lucide-react";

const IconButton = styled(MuiIconButton)`
  ${spacing};

  &:hover {
    background-color: transparent;
  }
`;

function NavbarUserDropdown() {
  const [anchorMenu, setAnchorMenu] = React.useState<any>(null);
  const navigate = useNavigate();
  const { user, signOut } = useAuth();

  const toggleMenu = (event: React.SyntheticEvent) => {
    setAnchorMenu(event.currentTarget);
  };

  const closeMenu = () => {
    setAnchorMenu(null);
  };

  const handleSignOut = async () => {
    await signOut();
    navigate("/auth/sign-in");
  };

  return (
    <React.Fragment>
      <Tooltip title="Account">
        <IconButton
          aria-owns={anchorMenu ? "menu-appbar" : undefined}
          aria-haspopup="true"
          onClick={toggleMenu}
          color="inherit"
          p={0}
          mx={1}
        >
          <LucidePower />
        </IconButton>
      </Tooltip>
      <Menu
        id="menu-appbar"
        anchorEl={anchorMenu}
        open={Boolean(anchorMenu)}
        onClose={closeMenu}
      >
        {user?.name && (
          <MenuItem disabled style={{ opacity: 0.7, fontWeight: 600 }}>
            Signed in as {user.name}
          </MenuItem>
        )}
        <Divider />
        <MenuItem onClick={closeMenu}>Reboot</MenuItem>
        <MenuItem onClick={closeMenu}>Power Down</MenuItem>
        <Divider />
        <MenuItem onClick={handleSignOut}>Sign out</MenuItem>
      </Menu>
    </React.Fragment>
  );
}

export default NavbarUserDropdown;
