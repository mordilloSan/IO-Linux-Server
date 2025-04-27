import React, { useState, useEffect } from "react";
import { Box, Typography } from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import axios from "@/utils/axios";
import SelectCard from "@/components/cards/SelectCard";
import Loader from "@/components/Loader";
import NetworkGraph from "@/pages/dashboard/home/NetworkGraph"; // your graph

interface InterfaceStats {
  name: string;
  mtu: number;
  hardwareAddr: string;
  addresses: string[];
  rxSpeed: number;
  txSpeed: number;
  linkSpeed: number;
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
  const [history, setHistory] = useState<
    { time: number; rx: number; tx: number }[]
  >([]);

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

  const selectedInterface = filteredInterfaces.find(
    (iface) => iface.name === selected
  );

  useEffect(() => {
    if (selectedInterface) {
      setHistory((prev) => [
        ...prev.slice(-29), // keep last 30
        {
          time: Date.now(),
          rx: selectedInterface.rxSpeed / 1e3,
          tx: selectedInterface.txSpeed / 1e3,
        },
      ]);
    }
  }, [selectedInterface]);

  const options = filteredInterfaces.map((iface) => ({
    value: iface.name,
    label: iface.name,
  }));

  const content = selectedInterface ? (
    isLoading ? (
      <Loader />
    ) : (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
        <Typography variant="body2">
          <strong>IP:</strong>{" "}
          {(() => {
            const ipv4 = selectedInterface.addresses.find((addr) =>
              addr.includes(".")
            );
            return ipv4 || "None";
          })()}
        </Typography>
        <Typography variant="body2">
          <strong>MAC:</strong> {selectedInterface.hardwareAddr}
        </Typography>
        <Typography variant="body2">
          <strong>Carrier:</strong> {selectedInterface.linkSpeed} MBS/s
        </Typography>
      </Box>
    )
  ) : (
    <Typography variant="body2">No interface selected.</Typography>
  );

  const content2 = selectedInterface ? (
    isLoading ? (
      <Loader />
    ) : (
      <Box sx={{ height: "120px", width: "100%" }}>
        <NetworkGraph data={history} />
      </Box>
    )
  ) : (
    <Typography variant="body2">No graph data.</Typography>
  );

  return (
    <SelectCard
      title="Network"
      avatarIcon="mdi:ethernet"
      stats={content}
      stats2={content2}
      selectOptions={options}
      selectedOption={selected}
      selectedOptionLabel={selected}
      onSelect={(val: string) => {
        setSelected(val);
        setHistory([]);
      }}
      connectionStatus={
        selectedInterface && selectedInterface.rxSpeed > 0
          ? "online"
          : "offline"
      } // 👈 ADD THIS
    />
  );
};

export default NetworkInterfacesCard;
