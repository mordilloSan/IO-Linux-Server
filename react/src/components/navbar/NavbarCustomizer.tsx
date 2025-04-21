import { useEffect, useState } from "react";
import {
  IconButton,
  Tooltip,
  Popover,
  Typography,
  useTheme as useMuiTheme,
} from "@mui/material";
import { Paintbrush } from "lucide-react"; // You can switch this to Palette, Droplet, etc.
import { ColorPicker, useColor, type IColor } from "react-color-palette";
import "react-color-palette/css";

import useTheme from "@/hooks/useTheme";

function NavbarColorCustomizer() {
  const { primaryColor, setPrimaryColor } = useTheme();
  const muiTheme = useMuiTheme();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [color, setColor] = useColor(primaryColor || "#4782da");

  // Keep picker synced with context
  useEffect(() => {
    if (primaryColor && primaryColor !== color.hex) {
      setColor({ ...color, hex: primaryColor });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [primaryColor, color.hex]);

  const handleChangeComplete = (newColor: IColor) => {
    setPrimaryColor(newColor.hex);
  };

  const handleOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const open = Boolean(anchorEl);

  return (
    <>
      <Tooltip title="Customize primary color">
        <IconButton color="inherit" onClick={handleOpen} size="large">
          <Paintbrush />
        </IconButton>
      </Tooltip>

      <Popover
        open={open}
        anchorEl={anchorEl}
        onClose={handleClose}
        anchorOrigin={{
          vertical: "bottom",
          horizontal: "right",
        }}
        transformOrigin={{
          vertical: "top",
          horizontal: "right",
        }}
        slotProps={{
          paper: {
            elevation: 6,
            sx: {
              p: 2,
              bgcolor: muiTheme.palette.background.paper,
              borderRadius: 2,
              width: 250,
            },
          },
        }}
      >
        <Typography variant="h6" gutterBottom>
          Primary Color
        </Typography>

        <ColorPicker
          height={150}
          color={color}
          onChange={setColor}
          onChangeComplete={handleChangeComplete}
          hideInput={["rgb", "hsv"]}
        />
      </Popover>
    </>
  );
}

export default NavbarColorCustomizer;
