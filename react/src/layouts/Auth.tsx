import React from "react";
import { Outlet } from "react-router-dom";
import styled from "@emotion/styled";

import { CssBaseline, Paper } from "@mui/material";

import { ReactComponent as Logo } from "@/vendor/logo.svg";

import GlobalStyle from "@/components/GlobalStyle";

interface AuthType {
  children?: React.ReactNode;
}

const Root = styled.div`
  max-width: 520px;
  margin: 0 auto;
  justify-content: center;
  align-items: center;
  display: flex;
  min-height: 100%;
  flex-direction: column;
`;

const Brand = styled(Logo)`
  fill: ${(props) => props.theme.palette.primary.main};
  width: 64px;
  height: 64px;
  margin-bottom: 32px;
`;

const Wrapper = styled(Paper)`
  padding: ${(props) => props.theme.spacing(6)};

  ${(props) => props.theme.breakpoints.up("md")} {
    padding: ${(props) => props.theme.spacing(10)};
  }
`;

const Auth: React.FC<AuthType> = ({ children }) => {
  return (
    <Root>
      <CssBaseline />
      <GlobalStyle />
      <Brand />
      <Wrapper>
        {children}
        <Outlet />
      </Wrapper>
    </Root>
  );
};

export default Auth;
