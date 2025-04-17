import { forwardRef } from "react";
import MuiAvatar, { AvatarProps as MuiAvatarProps } from "@mui/material/Avatar";
import { styled } from "@mui/material/styles";
import { OverridableStringUnion } from "@mui/types";
import { Theme } from "@mui/material/styles";

interface CustomAvatarProps extends MuiAvatarProps {
  color?: OverridableStringUnion<"primary" | "secondary" | "error", object>;
  size?: number;
  transparent?: boolean;
}

const Avatar = styled(MuiAvatar, {
  shouldForwardProp: (prop) =>
    !["color", "size", "transparent"].includes(prop as string),
})<CustomAvatarProps>(
  ({
    color = "primary",
    size,
    transparent,
    theme,
  }: CustomAvatarProps & { theme: Theme }) => {
    const palette = theme.palette[color];

    return {
      backgroundColor: transparent ? "transparent" : palette.main,
      color: palette.contrastText,
      ...(size && {
        width: size,
        height: size,
        fontSize: size * 0.45,
      }),
    };
  }
);

const CustomAvatar = forwardRef<HTMLDivElement, CustomAvatarProps>(
  ({ color = "primary", size, transparent = false, ...rest }, ref) => {
    return (
      <Avatar
        ref={ref}
        color={color}
        size={size}
        transparent={transparent}
        {...rest}
      />
    );
  }
);

CustomAvatar.displayName = "CustomAvatar";

export default CustomAvatar;
