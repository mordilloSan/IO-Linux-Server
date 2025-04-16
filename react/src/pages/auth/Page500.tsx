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

function Page500() {
  return (
    <Wrapper>
      <Typography component="h1" variant="h1" align="center" gutterBottom>
        500
      </Typography>
      <Typography component="h2" variant="h4" align="center" gutterBottom>
        Internal server error.
      </Typography>
      <Typography
        component="h2"
        variant="subtitle1"
        align="center"
        gutterBottom
      >
        The server encountered something unexpected that didn't allow it to
        complete the request.
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

export default Page500;
