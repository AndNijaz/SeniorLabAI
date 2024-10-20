import React from "react";

function Subheading({ children, className }) {
  return (
    <p
      className={`text-main-800 font-semibold dark:text-main-300 mp:text-base text-lg ${className}`}
    >
      {children}
    </p>
  );
}

export default Subheading;
