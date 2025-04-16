import React from "react";
import styled from "@emotion/styled";
import { withTheme } from "@emotion/react";
import { darken } from "polished";
import { Search as SearchIcon } from "lucide-react";

import {
  Grid,
  InputBase,
  AppBar as MuiAppBar,
  IconButton as MuiIconButton,
  Toolbar,
} from "@mui/material";

import { Menu as MenuIcon } from "@mui/icons-material";

import NavbarNotificationsDropdown from "./NavbarNotificationsDropdown";
import NavbarUserDropdown from "./NavbarUserDropdown";
import Settings from "./NavbarThemeToggle";

const AppBar = styled(MuiAppBar)`
  background-color: ${(props) => props.theme.palette.background.default};
  color: ${(props) => props.theme.palette.text.primary};
`;

const IconButton = styled(MuiIconButton)`
  svg {
    width: 22px;
    height: 22px;
  }
`;

const Search = styled.div`
  border-radius: 2px;
  background-color: ${(props) => props.theme.palette.background.default};
  display: none;
  position: relative;
  width: 100%;

  &:hover {
    background-color: ${(props) =>
      darken(
        props.theme.palette.mode === "dark" ? 0.02 : 0.05,
        props.theme.palette.background.default
      )};
  }

  ${(props) => props.theme.breakpoints.up("md")} {
    display: block;
  }
`;

const SearchIconWrapper = styled.div`
  width: 50px;
  height: 100%;
  position: absolute;
  pointer-events: none;
  display: flex;
  align-items: center;
  justify-content: center;

  svg {
    width: 22px;
    height: 22px;
  }
`;

const Input = styled(InputBase)`
  color: inherit;
  width: 100%;

  > input {
    color: ${(props) => props.theme.palette.text.primary};
    padding-top: ${(props) => props.theme.spacing(2.5)};
    padding-right: ${(props) => props.theme.spacing(2.5)};
    padding-bottom: ${(props) => props.theme.spacing(2.5)};
    padding-left: ${(props) => props.theme.spacing(12)};
    width: 160px;
  }
`;

type NavbarProps = {
  onDrawerToggle: React.MouseEventHandler<HTMLElement>;
};

const Navbar: React.FC<NavbarProps> = ({ onDrawerToggle }) => {
  return (
    <React.Fragment>
      <AppBar position="sticky" elevation={1}>
        <Toolbar>
          <Grid container alignItems="center" style={{ width: "100%" }}>
            <Grid sx={{ display: { xs: "block", md: "none" } }}>
              <IconButton
                color="inherit"
                aria-label="Open drawer"
                onClick={onDrawerToggle}
                size="large"
              >
                <MenuIcon />
              </IconButton>
            </Grid>
            <Grid>
              <Search>
                <SearchIconWrapper>
                  <SearchIcon />
                </SearchIconWrapper>
                <Input placeholder={"Search"} />
              </Search>
            </Grid>
            <Grid size="grow" />
            <Grid>
              <NavbarNotificationsDropdown />
              <Settings />
              <NavbarUserDropdown />
            </Grid>
          </Grid>
        </Toolbar>
      </AppBar>
    </React.Fragment>
  );
};

export default withTheme(Navbar);
