import CardRoot from "@/components/cards/CardRoot";

const BasicCard = ({ title, icon: Icon, iconProps, icon_text, Content }) => {
  return (
    <CardRoot
      title={title}
      icon={Icon}
      iconProps={iconProps}
      icon_text={icon_text}
    >
      {null}
      {Content}
    </CardRoot>
  );
};

export default BasicCard;
