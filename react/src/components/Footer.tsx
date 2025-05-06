import {
  Grid,
  List,
  ListItemText,
  ListItemButton,
  Box,
  useTheme,
} from "@mui/material";
import React from "react";

function Footer() {
  const theme = useTheme();

  return (
    <Box
      sx={{
        background: theme.footer?.background || theme.palette.background.paper,
        position: "relative",
      }}
    >
      <Grid container spacing={0}>
        {/* Left side links */}
        <Grid
          size={{
            xs: 12,
            md: 6,
          }}
          sx={{ display: { xs: "none", md: "block" } }}
        >
          <List>
            {["Support", "Help Center", "Privacy", "Terms of Service"].map(
              (text) => (
                <ListItemButton
                  key={text}
                  component="a"
                  href="#"
                  sx={{
                    display: "inline-block",
                    width: "auto",
                    px: 2,
                    color: "#ff0000",
                    "&:hover, &:active": theme.footer.color,
                  }}
                >
                  <ListItemText
                    primary={text}
                    slotProps={{
                      primary: {
                        sx: { color: theme.footer?.color || "text.primary" },
                      },
                    }}
                  />
                </ListItemButton>
              ),
            )}
          </List>
        </Grid>

        {/* Right side copyright */}
        <Grid
          size={{
            xs: 12,
            md: 6,
          }}
          container
          justifyContent="flex-end"
        >
          <List>
            <ListItemButton
              sx={{
                display: "inline-block",
                width: "auto",
                px: 2,
                color: "#ff0000",
                "&:hover, &:active": { color: "#ff0000" },
              }}
            >
              <ListItemText
                primary={`© ${new Date().getFullYear()} - Mira`}
                slotProps={{
                  primary: {
                    sx: { color: theme.footer?.color || "text.primary" },
                  },
                }}
              />
            </ListItemButton>
          </List>
        </Grid>
      </Grid>
    </Box>
  );
}

export default Footer;
