import React from "react";
import styled from "@emotion/styled";
import { Typography } from "@mui/material";
import ErrorOutlineIcon from "@mui/icons-material/ErrorOutline";

const Root = styled.div`
  justify-content: center;
  align-items: center;
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 8px;
  text-align: center;
`;

const ErrorMessage: React.FC = () => {
  return (
    <Root>
      <ErrorOutlineIcon color="error" fontSize="large" />
      <Typography color="error" variant="body1">
        Failed to load!
      </Typography>
    </Root>
  );
};

export default ErrorMessage;
