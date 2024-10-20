import React from "react";

interface ButtonProps {
  onClick?: () => void;
  type?: string;
  children: React.ReactNode;
  className?: string;
  color?: string;
  size?: string;
  disabled?: boolean;
}

const Button: React.FC<ButtonProps> = ({
  disabled = false,
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

  let buttonClassName =
    "gap-4 items-center justify-center flex rounded-xl transition duration-200 uppercase dark:bg-[#005f5f] text-white dark:disabled:bg-[#003030] disabled:bg-main-700 disabled:text-gray-400 dark:disabled:text-gray-500 hover:bg-main-600 hover:shadow-xl hover:scale-110 dark:hover:bg-[#337f7f] disabled:hover:scale-100 disabled:hover:shadow-none";

  // Apply size classes
  switch (size.toLowerCase()) {
    case "small":
      buttonClassName += ` px-4 py-2 text-sm !gap-1`;
      break;
    case "medium":
      buttonClassName += ` px-6 py-3 text-lg`;
      break;
    case "large":
      buttonClassName += ` px-8 py-4 text-3xl st:px-4 st:text-2xl st:py-2`;
      break;
    default:
      buttonClassName += ` px-6 py-3 text-lg`; // Fallback to medium size
  }

  // Apply color classes
  switch (color.toLowerCase()) {
    case "light":
      buttonClassName += ` bg-main-200 !text-main-800 dark:bg-[#005f5f]/40 dark:!text-white/90 hover:bg-main-400 hover:dark:bg-[#337f7f]/40`;
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
    <button
      onClick={handleClick}
      className={`${buttonClassName} ${className}`}
      disabled={disabled}
    >
      {children}
    </button>
  );
};

export default Button;
