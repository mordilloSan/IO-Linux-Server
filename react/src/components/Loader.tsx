import React from "react";
import styled from "@emotion/styled";
import { motion } from "framer-motion";
import { useTheme, alpha } from "@mui/material/styles";

const Root = styled.div`
  width: 100%;
  height: 100vh;
  background: transparent;
  display: flex;
  justify-content: center;
  align-items: center;
`;

const BarContainer = styled.div`
  width: 300px;
  height: 6px;
  background: #e0e0e0;
  border-radius: 3px;
  overflow: hidden;
  position: relative;
`;

const LoadingBar = styled(motion.div)<{ color: string }>`
  height: 100%;
  width: 150px; // Wider than container for trailing effect
  position: absolute;
  left: 0;
  top: 0;
  background: ${({ color }) =>
    `linear-gradient(90deg, ${color}, ${alpha(color, 0.5)})`};
  filter: blur(1px);
  border-radius: 3px;
`;

const animation = {
  animate: {
    x: ["-150px", "300px"], // Starts completely off screen to the left
    transition: {
      duration: 1.0,
      repeat: Infinity,
      ease: "easeInOut",
    },
  },
};

function Loader() {
  const theme = useTheme();
  const color = theme.palette.primary.main;

  return (
    <Root>
      <BarContainer>
        <LoadingBar color={color} variants={animation} animate="animate" />
      </BarContainer>
    </Root>
  );
}

export default Loader;
