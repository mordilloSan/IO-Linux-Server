import { Card, CardContent, Typography, Box } from "@mui/material";
import { cardHeight, cardBorderRadius } from "@/constants";

const CardRoot = ({ title, icon: Icon, iconProps, icon_text, children }) => {
  return (
    <Card
      elevation={2}
      sx={{
        height: cardHeight,
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
            mb: 0,
            mt: -1,
          }}
        >
          <Box sx={{ display: "flex", alignItems: "center", gap: 0 }}>
            <Typography variant="h4" sx={{ fontWeight: "bold" }}>
              {title}
            </Typography>
            {Icon ? <Icon {...iconProps} /> : null}
            <Typography variant="body2" sx={{ color: "grey" }}>
              {icon_text}
            </Typography>
          </Box>
          {children && children[0] && <Box>{children[0]}</Box>}
        </Box>
        <Box mt={1} sx={{ flex: "1 0 auto" }}>
          {children && children[1]}
        </Box>
      </CardContent>
    </Card>
  );
};

export default CardRoot;
