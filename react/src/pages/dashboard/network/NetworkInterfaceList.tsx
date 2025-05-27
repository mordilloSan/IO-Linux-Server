import { Icon } from "@iconify/react";
import {
  Box,
  Typography,
  Grid,
  TextField,
  Button,
  Collapse,
  Tooltip,
  Fade,
} from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import { useState, Suspense } from "react";

import FrostedCard from "@/components/cards/RootCard";
import ComponentLoader from "@/components/loaders/ComponentLoader";
import axios from "@/utils/axios";

export interface NetworkInterface {
  name: string;
  type: string; // ethernet, wifi, loopback, etc.
  mac: string;
  mtu: number;
  speed: string;
  duplex: string;
  state: number;
  ipv4: string[];
  ipv6: string[];
  rx_speed: number;
  tx_speed: number;
}

const getStatusColor = (state: number) => {
  if (state === 100) return "success.main"; // activated
  if (state === 30 || state === 20) return "warning.main"; // connecting/disconnected
  return "error.main";
};

const getStatusTooltip = (state: number) => {
  if (state === 100) return "Connected";
  if (state === 30) return "Connecting";
  if (state === 20) return "Disconnected";
  return "Unknown";
};

const getInterfaceIcon = (type?: string) => {
  if (type === "wifi") return "mdi:wifi";
  if (type === "ethernet") return "mdi:ethernet";
  if (type === "loopback") return "mdi:lan-connect";
  return "mdi:network";
};

const formatBps = (bps?: number) =>
  typeof bps === "number" ? `${(bps / 1024).toFixed(1)} kB/s` : "N/A";

const NetworkInterfaceList = () => {
  const [expanded, setExpanded] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<Record<string, any>>({});

  const handleToggle = (iface: NetworkInterface) => {
    if (expanded === iface.name) {
      setExpanded(null);
    } else {
      setEditForm({
        ipv4: iface.ipv4.join(", "),
        ipv6: iface.ipv6.join(", "),
        dns: "",
        gateway: "",
        mtu: iface.mtu.toString(),
      });
      setExpanded(iface.name);
    }
  };

  const handleChange = (field: string, value: string) => {
    setEditForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSave = (iface: NetworkInterface) => {
    console.log("Save", iface.name, editForm);
    setExpanded(null);
  };

  const { data: interfaces = [] } = useQuery<NetworkInterface[]>({
    queryKey: ["networkInterfaces"],
    queryFn: async () => {
      const res = await axios.get("/system/network");
      return res.data.map((iface: any) => ({
        ...iface,
        type: iface.name.startsWith("wl")
          ? "wifi"
          : iface.name.startsWith("lo")
            ? "loopback"
            : "ethernet",
        // optional: convert state from number to string
        state: iface.state === 100 ? "up" : "down",
      }));
    },

    refetchInterval: 1000,
  });

  return (
    <Suspense fallback={<ComponentLoader />}>
      <Box>
        <Typography variant="h4" sx={{ mb: 2 }}>
          Network Interfaces
        </Typography>
        <Grid container spacing={2}>
          {interfaces.map((iface) => (
            <Grid key={iface.name} size={{ xs: 6, sm: 4, md: 4, lg: 3, xl: 2 }}>
              <FrostedCard
                sx={{
                  p: 2,
                  position: "relative",
                  transition: "transform 0.2s, box-shadow 0.2s",
                  cursor: "pointer",
                  "&:hover": {
                    transform: "translateY(-4px)",
                    boxShadow: "0 8px 24px rgba(0,0,0,0.35)",
                  },
                }}>
                <Tooltip
                  title={getStatusTooltip(iface.state)}
                  placement="top"
                  arrow
                  slots={{ transition: Fade }}
                  slotProps={{ transition: { timeout: 300 } }}>
                  <Box
                    sx={{
                      position: "absolute",
                      top: 16,
                      right: 8,
                      width: 10,
                      height: 10,
                      borderRadius: "50%",
                      backgroundColor: getStatusColor(iface.state),
                      cursor: "default",
                    }}
                  />
                </Tooltip>

                <Box
                  display="flex"
                  alignItems="flex-start"
                  onClick={() => handleToggle(iface)}>
                  <Box
                    sx={{
                      width: 44,
                      height: 44,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      mr: 1.5,
                    }}>
                    <Icon
                      icon={getInterfaceIcon(iface.type)}
                      width={36}
                      height={36}
                    />
                  </Box>
                  <Box flexGrow={1}>
                    <Typography variant="subtitle1" fontWeight={600} noWrap>
                      {iface.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap>
                      IPv4: {iface.ipv4.join(", ")}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap>
                      MAC: {iface.mac}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap>
                      Link Speed: {iface.speed} ({iface.duplex})
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap>
                      RX/s: {formatBps(iface.rx_speed)} | TX/s:{" "}
                      {formatBps(iface.tx_speed)}
                    </Typography>
                  </Box>
                </Box>

                <Collapse
                  in={expanded === iface.name}
                  timeout="auto"
                  unmountOnExit>
                  <Box mt={2}>
                    <TextField
                      fullWidth
                      label="IPv4"
                      value={editForm.ipv4 || ""}
                      onChange={(e) => handleChange("ipv4", e.target.value)}
                      sx={{ mb: 1 }}
                    />
                    <TextField
                      fullWidth
                      label="IPv6"
                      value={editForm.ipv6 || ""}
                      onChange={(e) => handleChange("ipv6", e.target.value)}
                      sx={{ mb: 1 }}
                    />
                    <TextField
                      fullWidth
                      label="DNS"
                      value={editForm.dns || ""}
                      onChange={(e) => handleChange("dns", e.target.value)}
                      sx={{ mb: 1 }}
                    />
                    <TextField
                      fullWidth
                      label="Gateway"
                      value={editForm.gateway || ""}
                      onChange={(e) => handleChange("gateway", e.target.value)}
                      sx={{ mb: 1 }}
                    />
                    <TextField
                      fullWidth
                      type="number"
                      label="MTU"
                      value={editForm.mtu || ""}
                      onChange={(e) => handleChange("mtu", e.target.value)}
                      sx={{ mb: 2 }}
                    />
                    <Box display="flex" justifyContent="flex-end" gap={1}>
                      <Button onClick={() => setExpanded(null)}>Cancel</Button>
                      <Button
                        variant="contained"
                        onClick={() => handleSave(iface)}>
                        Save
                      </Button>
                    </Box>
                  </Box>
                </Collapse>
              </FrostedCard>
            </Grid>
          ))}
        </Grid>
      </Box>
    </Suspense>
  );
};

export default NetworkInterfaceList;
