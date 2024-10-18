import React from "react";

function Container({
  children,
  className,
}: {
  children: React.ReactNode;
  className: string;
}) {
  return (
    <div
      // className={`mx-auto overflow-scroll no-scrollbar scrollbar-none ${className}`}
      className={`max-w-7xl mx-auto overflow-scroll no-scrollbar scrollbar-none ${className}`}
    >
      {children}
    </div>
  );
}

export default Container;
