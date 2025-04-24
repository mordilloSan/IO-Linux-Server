import React from "react";
import MuiCard from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import { styled, alpha, useTheme } from "@mui/material/styles";
import { Icon } from "@iconify/react";
import CustomAvatar from "@/components/Avatar";
import { cardHeight, cardBorderRadius } from "@/constants";
import {
  FormControl,
  Select,
  MenuItem,
  SelectChangeEvent,
} from "@mui/material";

interface SelectOption {
  value: string;
  label: string;
  id?: string;
}

interface CardWithBorderProps {
  title: string;
  stats: React.ReactNode;
  stats2?: React.ReactNode;
  avatarIcon: string;
  icon?: React.ElementType;
  iconProps?: Record<string, any>;
  icon_text?: string;

  selectOptions?: SelectOption[];
  selectedOption?: string;
  selectedOptionLabel?: string;
  onSelect?: (value: string) => void;
}

const HoverableCard = styled(MuiCard)(({ theme }) => {
  const mainColor = theme.palette.primary.main;

  return {
    transition:
      "border 0.3s ease-in-out, box-shadow 0.3s ease-in-out, margin 0.3s ease-in-out",
    borderBottomWidth: "2px",
    borderBottomStyle: "solid",
    borderBottomColor: alpha(mainColor, 0.3),

    "&:hover": {
      borderBottomWidth: "3px",
      borderBottomColor: mainColor,
      boxShadow: theme.shadows[10],
      marginBlockEnd: "-1px",
    },
  };
});

const SelectCard: React.FC<CardWithBorderProps> = ({
  title,
  stats,
  stats2,
  avatarIcon,
  icon: IconComponent,
  iconProps,
  icon_text,
  selectOptions = [],
  selectedOption = "",
  selectedOptionLabel,
  onSelect,
}) => {
  const theme = useTheme();
  const primaryColor = theme.palette.primary.main;

  const handleSelectionChange = (
    event: SelectChangeEvent<string>
  ) => {
    onSelect?.(event.target.value);
  };

  const renderSelect =
    selectOptions.length > 0 && (
      <FormControl
      size="small"
      sx={{
        "& .MuiOutlinedInput-root": {
          color: "text.secondary", // text color
        },
        "& .MuiOutlinedInput-notchedOutline": {
          border: "none",
        },
        "& .MuiSvgIcon-root": {
          color: theme.palette.text.secondary, // arrow color
        },
      }}
    >
        <Select
          id="card-select"
          value={selectedOption}
          onChange={handleSelectionChange}
          displayEmpty
          renderValue={() =>
            selectedOptionLabel ? (
              <Typography variant="body2">
                {selectedOptionLabel}
              </Typography>
            ) : (
              <Typography variant="body2" color="text.secondary">
                Select...
              </Typography>
            )
          }
        >
          {selectOptions.map((option, index) => (
            <MenuItem key={option.id ?? index} value={option.value}>
              {option.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    );

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

            {/* Conditionally show select or icon+text */}
            {renderSelect ||
              (IconComponent && icon_text && (
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
                  <Typography
                    variant="body2"
                    sx={{ color: "grey", ml: 0, lineHeight: 1 }}
                  >
                    {icon_text}
                  </Typography>
                </Box>
              ))}
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

export default SelectCard;
