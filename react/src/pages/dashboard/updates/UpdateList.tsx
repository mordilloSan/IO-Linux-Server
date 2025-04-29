// components/system/UpdateList.tsx
import React from "react";
import { Card, Box, Typography } from "@mui/material";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import CollapsibleTable from "@/components/tables/CollapsibleTable";
import ComponentLoader from "@/components/ComponentLoader";
import { Update } from "@/types/update";
import { CollapsibleColumn } from "@/types/collapsible";

interface Props {
  updates: Update[];
  loading: boolean;
  renderCollapseContent: (row: Update) => React.ReactNode;
}

const columns: CollapsibleColumn[] = [
  { field: "name", headerName: "Name" },
  { field: "version", headerName: "Version" },
  { field: "severity", headerName: "Severity" },
];

const UpdateList: React.FC<Props> = ({
  updates,
  loading,
  renderCollapseContent,
}) => {
  if (loading) {
    return (
      <Box sx={{ padding: 2 }}>
        <Card>
          <Box sx={{ py: 2.8 }}>
            <ComponentLoader />
          </Box>
        </Card>
      </Box>
    );
  }

  if (updates.length === 0) {
    return (
      <Box sx={{ padding: 2 }}>
        <Card>
          <Box sx={{ display: "flex", alignItems: "center", py: 5 }}>
            <CheckCircleIcon
              color="success"
              sx={{ ml: 9, mr: 8, fontSize: 22 }}
            />
            <Typography variant="body1" fontSize={15}>
              System is up to date
            </Typography>
          </Box>
        </Card>
      </Box>
    );
  }

  return (
    <CollapsibleTable
      rows={updates}
      columns={columns}
      renderCollapseContent={renderCollapseContent}
    />
  );
};

export default UpdateList;
