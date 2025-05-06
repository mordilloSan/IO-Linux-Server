import { Box, Typography, Paper } from "@mui/material";
import React from "react";

import { ReactComponent as Logo } from "@/assets/logo.svg";
import SignInComponent from "@/components/auth/SignIn";

const SignIn: React.FC = () => {
  return (
    <React.Fragment>
      {/* Logo */}
      <Box
        component={Logo}
        sx={{
          fill: (theme) => theme.palette.primary.main,
          width: 150,
          height: 64,
          mb: 4,
        }}
      />

      {/* Title and subtitle */}
      <Box textAlign="center" mb={4}>
        <Typography component="h1" variant="h3" gutterBottom>
          Welcome back!
        </Typography>
        <Typography component="h2" variant="subtitle1">
          Sign in to your account to continue
        </Typography>
      </Box>

      {/* Paper form */}
      <Paper
        sx={(theme) => ({
          p: 6,
          width: "100%",
          maxWidth: "100%",
          boxSizing: "border-box",
          [theme.breakpoints.up("md")]: {
            p: 10,
          },
        })}
      >
        <SignInComponent />
      </Paper>
    </React.Fragment>
  );
};

export default SignIn;
