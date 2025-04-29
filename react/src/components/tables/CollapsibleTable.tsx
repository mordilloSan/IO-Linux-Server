import React, { useState } from "react";
import Box from "@mui/material/Box";
import Collapse from "@mui/material/Collapse";
import IconButton from "@mui/material/IconButton";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import ComponentLoader from "@/components/ComponentLoader";
import { CollapsibleTableProps } from "@/types/collapsible";

const CollapsibleTable = <T extends Record<string, any>>({
  rows,
  columns,
  renderCollapseContent,
}: CollapsibleTableProps<T>) => {
  const CollapsibleRow = ({ row, isLast }: { row: T; isLast: boolean }) => {
    const [open, setOpen] = useState(false);

    return (
      <>
        <TableRow>
          <TableCell
            sx={{
              width: "50px",
              borderBottom:
                open || isLast ? "none" : "1px solid rgba(255,255,255,0.1)",
            }}
          >
            <IconButton
              aria-label="expand row"
              size="small"
              onClick={() => setOpen(!open)}
            >
              {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
            </IconButton>
          </TableCell>
          {columns.map((column, index) => (
            <TableCell
              key={index}
              align={column.align || "left"}
              sx={{
                borderBottom:
                  open || isLast ? "none" : "1px solid rgba(255,255,255,0.1)",
              }}
            >
              {row[column.field]}
            </TableCell>
          ))}
        </TableRow>

        {open && (
          <TableRow>
            <TableCell
              colSpan={columns.length + 1}
              sx={{
                paddingTop: 0,
                paddingBottom: 0,
                borderBottom: isLast
                  ? "none"
                  : "1px solid rgba(255,255,255,0.1)",
              }}
            >
              <Collapse in={open} timeout="auto" unmountOnExit>
                <Box sx={{ margin: 1, marginTop: 5 }}>
                  {renderCollapseContent(row)}
                </Box>
              </Collapse>
            </TableCell>
          </TableRow>
        )}
      </>
    );
  };

  return (
    <Box sx={{ padding: 2 }}>
      <TableContainer
        component={Paper}
        sx={{ paddingLeft: "16px", paddingRight: "16px" }}
      >
        <Table aria-label="collapsible table">
          <TableHead>
            <TableRow>
              <TableCell sx={{ width: "50px" }} />
              {columns.map((column, index) => (
                <TableCell key={index} align={column.align || "left"}>
                  {column.headerName}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.length > 0 ? (
              rows.map((row, index) => (
                <CollapsibleRow
                  key={index}
                  row={row}
                  isLast={index === rows.length - 1}
                />
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length + 1} align="center">
                  <ComponentLoader />
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};

export default CollapsibleTable;
