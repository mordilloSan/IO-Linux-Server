import React from "react";
import styled from "@emotion/styled";

import {
  Grid,
  List,
  ListItemText as MuiListItemText,
  ListItemButtonProps as MuiListItemButtonProps,
  ListItemButton as MuiListItemButton,
  Box,
} from "@mui/material";

interface ListItemButtonProps extends MuiListItemButtonProps {
  href?: string;
}

const ListItemButton = styled(MuiListItemButton)<ListItemButtonProps>`
  display: inline-block;
  width: auto;
  padding-left: ${(props) => props.theme.spacing(2)};
  padding-right: ${(props) => props.theme.spacing(2)};

  &,
  &:hover,
  &:active {
    color: #ff0000;
  }
`;

const ListItemText = styled(MuiListItemText)`
  span {
    color: ${(props) => props.theme.palette.text.secondary};
  }
`;

function Footer() {
  return (
    <Box
      sx={{
        position: "fixed",
        bottom: 0,
        left: { xs: 0, md: "220px" },
        width: { xs: "100%", md: "calc(100% - 220px)" },
        zIndex: 1000,
        background: (theme) => theme.palette.background.default,
      }}
    >
      <Grid container spacing={0}>
        <Grid
          sx={{ display: { xs: "none", md: "block" } }}
          container
          size={{
            xs: 12,
            md: 6,
          }}
        >
          <List>
            <ListItemButton component="a" href="#">
              <ListItemText primary="Support" />
            </ListItemButton>
            <ListItemButton component="a" href="#">
              <ListItemText primary="Help Center" />
            </ListItemButton>
            <ListItemButton component="a" href="#">
              <ListItemText primary="Privacy" />
            </ListItemButton>
            <ListItemButton component="a" href="#">
              <ListItemText primary="Terms of Service" />
            </ListItemButton>
          </List>
        </Grid>
        <Grid
          container
          justifyContent="flex-end"
          size={{
            xs: 12,
            md: 6,
          }}
        >
          <List>
            <ListItemButton>
              <ListItemText
                primary={`© ${new Date().getFullYear()} - I/O Linux Server`}
              />
            </ListItemButton>
          </List>
        </Grid>
      </Grid>
    </Box>
  );
}

export default Footer;
