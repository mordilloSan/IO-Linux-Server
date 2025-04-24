import React, { useState, useEffect } from "react";
import { Box, Typography, SelectChangeEvent } from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import axios from "@/utils/axios";
import SelectCard from "@/components/cards/SelectCard";

interface InterfaceStats {
  name: string;
  mtu: number;
  hardwareAddr: string;
  addresses: string[];
  rxSpeed: number;
  txSpeed: number;
}

const NetworkInterfacesCard: React.FC = () => {
  const { data: interfaces = [], isLoading } = useQuery<InterfaceStats[]>({
    queryKey: ["networkInterfaces"],
    queryFn: async () => {
      const res = await axios.get("/system/network");
      return res.data;
    },
    refetchInterval: 1000,
  });

  const [selected, setSelected] = useState("");
  const filteredInterfaces = interfaces.filter(
    (iface) =>
      !iface.name.startsWith("veth") &&
      !iface.name.startsWith("docker") &&
      iface.name !== "lo"
  );

  useEffect(() => {
    if (filteredInterfaces.length && !selected) {
      setSelected(filteredInterfaces[0].name);
    }
  }, [filteredInterfaces, selected]);
  
  const options = filteredInterfaces.map((iface) => ({
    value: iface.name,
    label: iface.name,
  }));

  const selectedInterface = filteredInterfaces.find((iface) => iface.name === selected);

  const content = selectedInterface ? (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
      <Typography variant="body2">MAC: {selectedInterface.hardwareAddr}</Typography>
      <Typography variant="body2">MTU: {selectedInterface.mtu}</Typography>
      <Typography variant="body2">
        Sending: {(selectedInterface.txSpeed / 1e3).toFixed(2)} kB/s
      </Typography>
      <Typography variant="body2">
        Receiving: {(selectedInterface.rxSpeed / 1e3).toFixed(2)} kB/s
      </Typography>
      <Typography variant="body2">
  IP:{" "}
  {(() => {
    const ipv4 = selectedInterface.addresses.find((addr) => addr.includes("."));
    const ips = [ipv4].filter(Boolean).join(", ");
    return ips || "None";
  })()}
</Typography>

    </Box>
  ) : (
    <Typography variant="body2">No interface data.</Typography>
  );

  return (
    <SelectCard
      title="Network"
      avatarIcon="mdi:ethernet"
      stats={content}
      stats2={content}
      selectOptions={options}
      selectedOption={selected}
      selectedOptionLabel={selected}
      onSelect={(val: string) => setSelected(val)}
    />
  );
};

export default NetworkInterfacesCard;
