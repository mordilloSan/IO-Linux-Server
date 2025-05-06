import { Divider, Tooltip, Menu, MenuItem, IconButton } from "@mui/material";
import { LucidePower } from "lucide-react";
import React, { useRef } from "react";
import { useNavigate } from "react-router-dom";

import useAuth from "@/hooks/useAuth";

function NavbarUserDropdown() {
  const ref = useRef(null);
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
    navigate("/sign-in");
  };

  return (
    <React.Fragment>
      <Tooltip title="Account">
        <IconButton color="inherit" ref={ref} onClick={toggleMenu} size="large">
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
