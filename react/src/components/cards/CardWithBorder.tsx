import React from "react";
import MuiCard from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import { styled, alpha } from "@mui/material/styles";
import CustomAvatar from "@/components/Avatar";
import { cardHeight, cardBorderRadius } from "@/constants";
import { Icon } from "@iconify/react";

type AllowedColor = "primary" | "secondary";

interface CardWithBorderProps {
  title: string;
  stats: React.ReactNode;
  stats2?: React.ReactNode;
  avatarIcon: string;
  icon?: React.ElementType;
  iconProps?: Record<string, any>;
  icon_text?: string;
  color?: AllowedColor;
}

// Styled card with hover border animation
const HoverableCard = styled(MuiCard, {
  shouldForwardProp: (prop) => prop !== "color",
})<{ color?: AllowedColor }>(({ theme, color = "primary" }) => {
  const palette = theme.palette[color];

  return {
    transition:
      "border 0.3s ease-in-out, box-shadow 0.3s ease-in-out, margin 0.3s ease-in-out",
    borderBottomWidth: "2px",
    borderBottomStyle: "solid",
    borderBottomColor: alpha(palette.main, 0.3),

    "&:hover": {
      borderBottomWidth: "3px",
      borderBottomColor: palette.main,
      boxShadow: theme.shadows[10],
      marginBlockEnd: "-1px",
    },
  };
});

const CardWithBorder: React.FC<CardWithBorderProps> = ({
  title,
  stats,
  stats2,
  avatarIcon,
  icon: IconComponent,
  iconProps,
  icon_text,
  color = "primary",
}) => {
  return (
    <HoverableCard
      color={color}
      elevation={2}
      sx={{
        minHeight: cardHeight,
        m: 1,
        display: "flex",
        flexDirection: "column",
        borderRadius: cardBorderRadius,
      }}
    >
      <CardContent>
        <Box
          sx={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            mb: 1,
          }}
        >
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
            <Typography
              variant="h4"
              sx={{
                fontWeight: "bold",
                transform: "translateY(-1px)", // subtle upward shift
              }}
            >
              {title}
            </Typography>
            {IconComponent && icon_text && (
              <Box
                sx={{
                  display: "inline-flex",
                  alignItems: "center",
                  gap: 0,
                  lineHeight: 1,
                }}
              >
                <Box
                  sx={{
                    display: "inline-flex",
                    alignItems: "center",
                    mr: "-4px", // trims space between icon and text
                  }}
                >
                  <IconComponent
                    {...iconProps}
                    sx={{ verticalAlign: "middle", ...iconProps?.sx }}
                  />
                </Box>
                <Typography
                  variant="body2"
                  sx={{ color: "grey", ml: 0, lineHeight: 1 }}
                >
                  {icon_text}
                </Typography>
              </Box>
            )}
          </Box>

          <CustomAvatar color={color} transparent size={40}>
            <Icon icon={avatarIcon} width="32px" height="32px" />
          </CustomAvatar>
        </Box>

        {stats2 ? (
          <Box
            sx={{
              mt: 3,
              display: "flex",
              flexDirection: { xs: "column", sm: "column", xl: "row" },
            }}
          >
            <Box
              sx={{
                flex: 1,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              {stats}
            </Box>
            <Box
              sx={{
                flex: { xl: 1 },
                display: "flex",
                height: 120,
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              {stats2}
            </Box>
          </Box>
        ) : (
          <Box sx={{ mt: 7 }}>{stats}</Box>
        )}
      </CardContent>
    </HoverableCard>
  );
};

export default CardWithBorder;
