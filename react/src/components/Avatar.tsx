import { forwardRef } from "react";
import MuiAvatar, { AvatarProps as MuiAvatarProps } from "@mui/material/Avatar";
import { styled } from "@mui/material/styles";

interface CustomAvatarProps extends MuiAvatarProps {
  size?: number;
  transparent?: boolean;
}

// Styled Avatar
const StyledAvatar = styled(MuiAvatar, {
  shouldForwardProp: (prop) =>
    !["size", "transparent"].includes(prop as string),
})<Pick<CustomAvatarProps, "size" | "transparent">>(
  ({ size, transparent, theme }) => {
    const palette = theme.palette.primary;

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

// Component
const CustomAvatar = forwardRef<HTMLDivElement, CustomAvatarProps>(
  ({ size, transparent = false, ...rest }, ref) => {
    return (
      <StyledAvatar ref={ref} size={size} transparent={transparent} {...rest} />
    );
  }
);

CustomAvatar.displayName = "CustomAvatar";

export default CustomAvatar;
