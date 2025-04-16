import React from "react";

import { Typography } from "@mui/material";

import SignInComponent from "@/components/auth/SignIn";

function SignIn() {
  return (
    <React.Fragment>
      <Typography component="h1" variant="h3" align="center" gutterBottom>
        Welcome back!
      </Typography>
      <Typography component="h2" variant="subtitle1" align="center">
        Sign in to your account to continue
      </Typography>

      <SignInComponent />
    </React.Fragment>
  );
}

export default SignIn;
