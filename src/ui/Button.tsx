import React from "react";

interface ButtonProps {
  onClick?: () => void;
  type?: string;
  children: React.ReactNode;
  className?: string;
  color?: string;
  size?: string;
}

const Button: React.FC<ButtonProps> = ({
  onClick,
  // type = "",
  color = "",
  size = "",
  children,
  className = "",
}) => {
  function handleClick() {
    if (onClick) onClick();
  }

  // let buttonClassName =
  //   "flex items-center justify-center gap-4 rounded-xl transition duration-200  uppercase";

  // if (type.toLowerCase() === "small")
  //   buttonClassName += ` bg-main-200 text-main-800 px-4 py-2 text-sm`;
  // else if (type.toLowerCase() === "medium")
  //   buttonClassName += `bg-main-500 text-white px-4 py-2 text-lg`;
  // else buttonClassName += ` bg-main-500 text-white px-8 py-4 text-3xl`;

  // console.log(buttonClassName);
  let buttonClassName =
    "gap-4 items-center justify-center flex rounded-xl transition duration-200 uppercase dark:bg-[#005f5f] text-white";

  // Apply size classes
  switch (size.toLowerCase()) {
    case "small":
      buttonClassName += ` px-4 py-2 text-sm !gap-1`;
      break;
    case "medium":
      buttonClassName += ` px-6 py-3 text-lg`;
      break;
    case "large":
      buttonClassName += ` px-8 py-4 text-3xl`;
      break;
    default:
      buttonClassName += ` px-6 py-3 text-lg`; // Fallback to medium size
  }

  // Apply color classes
  switch (color.toLowerCase()) {
    case "light":
      buttonClassName += ` bg-main-200 !text-main-800 dark:bg-[#005f5f]/40 dark:!text-white/90`;
      break;
    case "normal":
      buttonClassName += ` bg-main-500 text-white`;
      break;
    case "dark":
      buttonClassName += ` bg-main-800 text-white`;
      break;
    default:
      buttonClassName += ` bg-main-500 text-white`; // Fallback to normal color
  }

  return (
    <button onClick={handleClick} className={`${buttonClassName} ${className}`}>
      {children}
    </button>
  );
};

export default Button;
