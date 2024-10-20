import React from "react";

function Subheading({ children }) {
  return (
    <p className="text-main-800 font-semibold dark:text-main-300 mp:text-base">
      {children}
    </p>
  );
}

export default Subheading;
