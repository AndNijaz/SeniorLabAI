import React from "react";

function ResaultHeading({ children }) {
  return (
    <h3 className="text-main-800 font-semibold text-4xl mb-4 dark:text-main-300 mp:text-3xl">
      {children}
    </h3>
  );
}

export default ResaultHeading;
