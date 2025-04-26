import React from "react";
import MuiCard from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import { styled, alpha, useTheme } from "@mui/material/styles";
import CustomAvatar from "@/components/Avatar";
import { cardHeight, cardBorderRadius } from "@/constants";
import { Icon } from "@iconify/react";

interface CardWithBorderProps {
  title: string;
  stats: React.ReactNode;
  stats2?: React.ReactNode;
  avatarIcon: string;
  icon?: React.ElementType;
  iconProps?: Record<string, any>;
  icon_text?: string;
}

// Styled card with hover border animation
const HoverableCard = styled(MuiCard)(({ theme }) => {
  const mainColor = theme.palette.primary.main;

  return {
    transition: "border-color 0.3s ease-in-out, box-shadow 0.3s ease-in-out",
    borderBottomWidth: "2px",
    borderBottomStyle: "solid",
    borderBottomColor: alpha(mainColor, 0.3),
    marginBlockEnd: 0,

    "&:hover": {
      borderBottomColor: mainColor,
      boxShadow: theme.shadows[10],
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
}) => {
  const theme = useTheme();
  const primaryColor = theme.palette.primary.main;

  return (
    <HoverableCard
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
                transform: "translateY(-1px)",
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
                    mr: "-4px",
                  }}
                >
                  <IconComponent
                    {...iconProps}
                    sx={{
                      verticalAlign: "middle",
                      color: primaryColor,
                      ...iconProps?.sx,
                    }}
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

          <CustomAvatar transparent size={40}>
            <Icon
              icon={avatarIcon}
              width="32px"
              height="32px"
              color={primaryColor}
            />
          </CustomAvatar>
        </Box>

        {stats2 ? (
          <Box
            sx={{
              mt: 3,
              display: "flex",
              flexDirection: { xs: "column", sm: "row", xl: "row" },
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
                flex: { sm: 1, xl: 1 },
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
