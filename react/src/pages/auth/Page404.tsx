import React from "react";
import styled from "@emotion/styled";
import { Link } from "react-router-dom";

import { Button as MuiButton, Typography } from "@mui/material";
import { spacing, SpacingProps } from "@mui/system";

interface ButtonProps extends SpacingProps {
  component?: React.ElementType;
  to?: string;
  target?: string;
}

const Button = styled(MuiButton)<ButtonProps>(spacing);

const Wrapper = styled.div`
  text-align: center;
`;

function Page404() {
  return (
    <Wrapper>
      <Typography component="h1" variant="h1" align="center" gutterBottom>
        404
      </Typography>
      <Typography component="h2" variant="h4" align="center" gutterBottom>
        Page not found.
      </Typography>
      <Typography
        component="h2"
        variant="subtitle1"
        align="center"
        gutterBottom
      >
        The page you are looking for might have been removed.
      </Typography>

      <Button
        component={Link}
        to="/"
        variant="contained"
        color="secondary"
        mt={2}
      >
        Return to website
      </Button>
    </Wrapper>
  );
}

export default Page404;
