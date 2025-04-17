import { Select, MenuItem, FormControl, Typography } from "@mui/material";
import CardRoot from "@/components/cards/CardRoot";

const SelectCard = ({
  title,
  icon: Icon,
  iconProps,
  icon_text,
  Content,
  options,
  onSelect,
  selectedOption,
  selectedOptionLabel,
}) => {
  const handleSelectionChange = (event) => {
    if (onSelect) {
      onSelect(event);
    }
  };

  const renderSelect = options?.length > 0 && (
    <FormControl
      size="small"
      sx={{
        minWidth: 120,
        "& .MuiOutlinedInput-notchedOutline": { border: "none" },
      }}
    >
      <Select
        id="drive-select"
        name="drive"
        value={selectedOption}
        onChange={handleSelectionChange}
        displayEmpty
        IconComponent={null}
        renderValue={() =>
          selectedOption ? (
            <Typography variant="body2">{selectedOptionLabel}</Typography>
          ) : (
            <Typography variant="body2">Select...</Typography>
          )
        }
      >
        {options.map((option, index) => (
          <MenuItem key={option.id || index} value={option.value}>
            {option.label}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );

  return (
    <CardRoot
      title={title}
      icon={Icon}
      iconProps={iconProps}
      icon_text={icon_text}
    >
      {renderSelect}
      {Content}
    </CardRoot>
  );
};

export default SelectCard;
